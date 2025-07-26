//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"syscall/js"
	"time"

	internalfsrs "github.com/hnimtadd/spaced/src/core/fsrs"
	"github.com/hnimtadd/spaced/src/core/model"
	"github.com/hnimtadd/spaced/src/core/utils"
	"github.com/open-spaced-repetition/go-fsrs/v3"
)

type SpacedManager struct {
	cards        internalfsrs.Cards
	localStorage js.Value
	fsrs         *fsrs.FSRS

	sessionCards internalfsrs.Cards

	schedulingCards fsrs.RecordLog
	againsID        map[int]bool
	targetNum       int
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
		againsID:     map[int]bool{},
	}
	m.init()
	return m, nil
}

func (m *SpacedManager) init() {
	if err := m.handlePullState(); err != nil {
		fmt.Println("could not pull the state from localStorage try to fetch")

		// Use a Go routine to fetch the cards data without blocking the main thread.
		go func() {
			respPromise := js.Global().Call("fetch", "/assets/cards_sample.json")

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
					m.handlePushState()
					fmt.Println("cards loaded successfully")
				}
				return nil
			}))
		}()
	}

	// hack, indexing cards on init
	for i := range m.cards {
		m.cards[i].ID = i
	}
	fmt.Println("pull state from localStorage completed")
}

func (m *SpacedManager) shouldStop() bool {
	if len(m.againsID) > 0 {
		return false
	}
	for card := range slices.Values(m.sessionCards) {
		if card.Due.Before(time.Now()) {
			return false
		}
		if card.LastReview.IsZero() {
			return false
		}
	}
	return true
}

// startSession based on the list of most urgent due date cards
// prepare the list of cards to be review in this session.
func (m *SpacedManager) startSession(js.Value, []js.Value) any {
	sort.Sort(internalfsrs.Cards(m.cards))
	len := min(m.targetNum, len(m.cards))
	m.sessionCards = make(internalfsrs.Cards, len)
	for i := range len {
		m.sessionCards[i] = m.cards[i]
	}
	return model.PayloadResponse("success")
}

// handlePullState pull the state passed from web browser.
func (m *SpacedManager) handlePullState() error {
	fmt.Println("handle pull state from localStorage")
	dataRaw := m.localStorage.Call("getItem", "flashcards")
	if !dataRaw.Truthy() {
		return errors.New("flashcards in localStorage is in invalid form")
	}
	data := dataRaw.String()
	if data == "" {
		return errors.New("empty flashcards item, skipping")
	}

	if err := json.Unmarshal([]byte(data), &m.cards); err != nil {
		return errors.New("could not unmarshal the data, got: " + err.Error())
	}

	return nil
}

// handlePushState push the state from wasm land to js land
func (m *SpacedManager) handlePushState() error {
	fmt.Println("handle push state to localStorage")
	dataBytes, err := json.Marshal(m.cards)
	if err != nil {
		return fmt.Errorf("failed to push state to localStorage: %w", err)
	}

	m.localStorage.Call("setItem", "flashcards", string(dataBytes))
	fmt.Println("handle push state to localStorage, complete")
	return nil
}

func (m *SpacedManager) next(_ js.Value, args []js.Value) any {
	if len(m.cards) == 0 {
		return model.ErrorResponse("no cards found")
	}
	if m.shouldStop() {
		return model.StopResponse()
	}

	sort.Sort(m.sessionCards)
	return cardToJsValue(m.sessionCards[0])
}

func (m *SpacedManager) submit(this js.Value, args []js.Value) any {
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
	if *rating != 0 {
		if *rating == fsrs.Again {
			m.againsID[*cardID] = true
		} else {
			delete(m.againsID, *cardID)
		}
		// assume that we have a very little latency, from when the user provide
		// feedback to when this path is reached.
		// so using current timestamp
		fmt.Println("handle submit for", "id", *cardID, m.cards[*cardID])
		state := m.fsrs.Repeat(m.cards[*cardID].ToFsrsCard(), time.Now())
		m.cards[*cardID].SyncFromFSRSCard(state[*rating].Card)
		fmt.Println("after", m.cards[*cardID])
		return model.PayloadResponse("updated")
	}

	return model.PayloadResponse("not updated")
}

func cardToJsValue(card *model.Card) any {
	jsonBytes, err := utils.Serialize(card)
	if err != nil {
		return model.ErrorResponse("could not marshal the card, got: " + err.Error())
	}
	return model.PayloadResponse(string(jsonBytes))
}

func (m *SpacedManager) start(js.Value, []js.Value) any {
	m.startSession(js.Value{}, []js.Value{})
	return m.next(js.Value{}, []js.Value{})
}

func helloworld(js.Value, []js.Value) any {
	return js.ValueOf("Helloworld")
}

func main() {
	c := make(chan struct{})

	m, err := NewSpacedManger()
	if err != nil {
		fmt.Println("failed to init: " + err.Error())
	}

	// Map of exposed Go methods for easy lookup in JS land.
	goFuncs := map[string]js.Func{
		"start":  js.FuncOf(m.start),
		"next":   js.FuncOf(m.next),
		"submit": js.FuncOf(m.submit),
		"stats":  js.FuncOf(helloworld),
	}

	wasmBridge := js.Global().Get("Object").New()
	for name, fn := range goFuncs {
		wasmBridge.Set(name, fn)
	}

	js.Global().Set("wasmBridge", wasmBridge)

	// Keep the Go program running so the WASM module doesn't exit.
	<-c
}
