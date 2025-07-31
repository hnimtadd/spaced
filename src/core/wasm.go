//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"slices"
	"sort"
	"strings"
	"syscall/js"
	"time"

	internalfsrs "github.com/hnimtadd/spaced/src/core/fsrs"
	"github.com/hnimtadd/spaced/src/core/model"
	"github.com/hnimtadd/spaced/src/core/session"
	"github.com/hnimtadd/spaced/src/core/utils"
	"github.com/open-spaced-repetition/go-fsrs/v3"
)

type SpacedManager struct {
	cards        internalfsrs.Cards
	lookup       map[int]*model.Card
	localStorage js.Value
	fsrs         *fsrs.FSRS

	targetNum   int
	currSession *session.Session

	records []session.Record
}

func NewSpacedManger() (*SpacedManager, error) {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.Truthy() {
		return nil, errors.New("localStorage from JS is not truthy")
	}
	fsrss := fsrs.NewFSRS(fsrs.DefaultParam())
	m := &SpacedManager{
		localStorage: localStorage,
		fsrs:         fsrss,
		targetNum:    10,
		records:      []session.Record{},
		lookup:       map[int]*model.Card{},
	}
	return m, nil
}

func (m *SpacedManager) JSInit(js.Value, []js.Value) any {
	if err := m.handlePullState(); err != nil {

		// Use a Go routine to fetch the cards data without blocking the main thread.
		go func() {
			respPromise := js.Global().Call("fetch", "/assets/cards.json")

			// Handle the promise returned by fetch.
			respPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				// On success, parse the response body as JSON.
				return args[0].Call("json")
			})).Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				// On successful parsing, store this into the local storage
				jsonString := js.Global().Get("JSON").Call("stringify", args[0]).String()

				if err := json.Unmarshal([]byte(jsonString), &m.cards); err != nil {
					fmt.Println("failed to unmarshal cards:", err)
				} else {
					// hack, indexing cards on init
					for i := range m.cards {
						m.cards[i].ID = i
					}
					m.handlePushState()
					for card := range slices.Values(m.cards) {
						m.lookup[card.ID] = card
					}
					fmt.Println("cards loaded successfully")
					// this is a hack, JS land and Go land should have a
					// way to do some pub sub or callback management.
					// so after create this, we could do something, like:
					// this.get("craft-onload").Call()
					// so the logic will be controlled by user in JS land.
					js.Global().Get("location").Call("reload")
				}
				return js.ValueOf(nil)
			}))
		}()
		return js.ValueOf(nil)
	}

	for card := range slices.Values(m.cards) {
		m.lookup[card.ID] = card
	}
	return js.ValueOf(nil)
}

// startSession based on the list of most urgent due date cards
// prepare the list of cards to be review in this session.
func (m *SpacedManager) startSession() any {
	sort.Sort(internalfsrs.Cards(m.cards))
	numCards := min(m.targetNum, len(m.cards))
	cards := make(internalfsrs.Cards, numCards)

	for i := range numCards {
		cards[i] = m.cards[i]
	}

	m.currSession = session.NewSession(cards)
	return model.PayloadResponse("success")
}

func (m *SpacedManager) completeSession() {
	record := session.NewRecordFromSession(m.currSession)
	m.currSession = nil
	m.records = append(m.records, record)
	m.push("records", m.records)
	m.push("flashcards", m.cards)
}

func (m *SpacedManager) pull(key string, to any) error {
	dataRaw := m.localStorage.Call("getItem", key)
	if !dataRaw.Truthy() {
		return errors.New("flashcards in localStorage is in invalid form")
	}
	data := dataRaw.String()
	if data == "" {
		return errors.New("empty flashcards item, skipping")
	}
	err := utils.DeserializeTo([]byte(data), to)
	if err != nil {
		return errors.New("could not deserialize the data, got: " + err.Error())
	}
	return nil
}

func (m *SpacedManager) push(key string, data any) error {
	dataBytes, err := utils.Serialize(data)
	if err != nil {
		return fmt.Errorf("failed to push state to localStorage: %w", err)
	}

	m.localStorage.Call("setItem", key, string(dataBytes))
	return nil
}

// handlePullState pull the state passed from web browser.
func (m *SpacedManager) handlePullState() error {
	if err := m.pull("flashcards", &m.cards); err != nil {
		return fmt.Errorf("failed to pull flashcards, err: %v", err)
	}
	if err := m.pull("records", &m.records); err != nil {
		return fmt.Errorf("failed to pull sessions, err: %v", err)
	}
	return nil
}

// handlePushState push the state from wasm land to js land
func (m *SpacedManager) handlePushState() error {
	if err := m.push("flashcards", &m.cards); err != nil {
		return fmt.Errorf("failed to push flashcards, err: %v", err)
	}
	if err := m.push("records", &m.records); err != nil {
		return fmt.Errorf("failed to push sessions, err: %v", err)
	}
	return nil
}

func (m *SpacedManager) next() any {
	if len(m.cards) == 0 {
		return model.ErrorResponse("no cards found")
	}
	if m.currSession == nil {
		return model.ErrorResponse("not start session yet")
	}

	if m.currSession.ShouldStop() {
		m.completeSession()
		return model.StopResponse()
	}

	sort.Sort(m.currSession.Cards)

	card := m.currSession.Cards[0]
	jsonBytes, err := utils.Serialize(card)
	if err != nil {
		return model.ErrorResponse("could not marshal the card, got: " + err.Error())
	}
	return model.PayloadResponse(string(jsonBytes))
}

func (m *SpacedManager) JSNext(_ js.Value, _ []js.Value) any {
	return m.next()
}

func (m *SpacedManager) JSSubmit(_ js.Value, args []js.Value) any {
	if len(args) < 2 {
		return model.ErrorResponse("must provide card ID and rating")
	}

	cardID, err := utils.Deserialize[int]([]byte(args[0].String()))
	if err != nil {
		return model.ErrorResponse("invalid card ID: " + err.Error())
	}

	rating, err := utils.Deserialize[fsrs.Rating]([]byte(args[1].String()))
	if err != nil {
		return model.ErrorResponse("invalid rating: " + err.Error())
	}
	m.currSession.Looked[*cardID] = true
	if *rating != 0 {
		if *rating == fsrs.Again {
			m.currSession.AgainsID[*cardID] = true
		} else {
			delete(m.currSession.AgainsID, *cardID)
		}
		// assume that we have a very little latency, from when the user provide
		// feedback to when this path is reached.
		// so using current timestamp
		card, exists := m.lookup[*cardID]
		if !exists {
			return model.ErrorResponse("submit for not exists card")
		}
		fmt.Println("handle submit for", "id", *cardID, card)
		state := m.fsrs.Repeat(card.ToFsrsCard(), time.Now())
		m.lookup[*cardID].SyncFromFSRSCard(state[*rating].Card)
		return model.PayloadResponse("updated")
	}

	return model.PayloadResponse("not updated")
}

func (m *SpacedManager) JSStart(js.Value, []js.Value) any {
	m.startSession()
	return model.PayloadResponse("ready")
}

func (m *SpacedManager) JSStats(js.Value, []js.Value) any {
	tpl := `<div class="max-w-5xl sm:w-[30rem] md:w-[40rem] lg:w-[50rem] mx-auto h-screen p-4 space-y-4">{{range .Sessions}}{{.}}{{end}}</div>`
	tmpl, err := template.New("stats").Parse(tpl)
	if err != nil {
		return model.ErrorResponse("failed to init tmpl: " + err.Error())
	}
	eles := make([]template.HTML, len(m.records))

	for i, session := range m.records {
		eles[i] = session.ToHTML()
	}
	slices.Reverse(eles)
	data := struct {
		Sessions []template.HTML
	}{
		Sessions: eles,
	}
	buf := &strings.Builder{}
	if err := tmpl.Execute(buf, data); err != nil {
		return model.ErrorResponse("failed to execute" + err.Error())
	}
	return js.ValueOf(buf.String())
}

func main() {
	c := make(chan struct{})

	m, err := NewSpacedManger()
	if err != nil {
		fmt.Println("failed to init: " + err.Error())
	}

	// Map of exposed Go methods for easy lookup in JS land.
	goFuncs := map[string]js.Func{
		"init":   js.FuncOf(m.JSInit),
		"start":  js.FuncOf(m.JSStart),
		"next":   js.FuncOf(m.JSNext),
		"submit": js.FuncOf(m.JSSubmit),
		"stats":  js.FuncOf(m.JSStats),
	}

	wasmBridge := js.Global().Get("Object").New()
	for name, fn := range goFuncs {
		wasmBridge.Set(name, fn)
	}

	js.Global().Set("wasmBridge", wasmBridge)

	// Keep the Go program running so the WASM module doesn't exit.
	<-c
}
