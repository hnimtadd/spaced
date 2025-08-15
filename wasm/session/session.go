//go:build js && wasm

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"slices"
	"sort"
	"syscall/js"
	"time"

	handler "github.com/hnimtadd/spaced/api/sound"
	internalfsrs "github.com/hnimtadd/spaced/src/core/fsrs"
	"github.com/hnimtadd/spaced/src/core/model"
	"github.com/hnimtadd/spaced/src/core/session"
	"github.com/hnimtadd/spaced/src/core/utils"
	"github.com/hnimtadd/spaced/src/crafter"
	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// Global JavaScript AudioContext instance
var (
	audioContext js.Value
	soundSource  any
)

func init() {
	// This function runs once when the WASM module is initialized.
	// Create the AudioContext. It will likely be in 'suspended' state until user interaction.
	audioContext = js.Global().Get("AudioContext").New()
	fmt.Println("Go WASM: AudioContext created (state:", audioContext.Get("state").String(), ")")
}

type SpacedManager struct {
	cards  internalfsrs.Cards
	lookup map[int]*model.Card
	fsrs   *fsrs.FSRS

	targetNum   int
	currSession *session.Session

	records []*session.Record
}

func NewSpacedManger() (*SpacedManager, error) {
	fsrss := fsrs.NewFSRS(fsrs.DefaultParam())
	m := &SpacedManager{
		fsrs:      fsrss,
		targetNum: 10,
		records:   []*session.Record{},
		lookup:    map[int]*model.Card{},
	}
	return m, nil
}

func (m *SpacedManager) JSInit(_ js.Value, args []js.Value) any {
	fmt.Println("args", args)
	if err := m.parsedFromLocalState(); err != nil {
		req := utils.Request{
			URL:       "/assets/cards.json",
			Method:    http.MethodGet,
			Header:    nil,
			BodyBytes: nil,
		}
		resolveFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
			resp := args[0]
			jsonPromiseFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
				jsonString := js.Global().Get("JSON").Call("stringify", args[0]).String()
				fmt.Println(jsonString)
				if err := json.Unmarshal([]byte(jsonString), &m.cards); err != nil {
					fmt.Println("failed to unmarshal cards:", err)
				} else {
					// hack, indexing cards on init
					for i := range m.cards {
						m.cards[i].ID = i
					}
					m.handleSaveState()
					fmt.Println("cards loaded successfully")
					// this is a hack, JS land and Go land should have a
					// way to do some pub sub or callback management.
					// so after create this, we could do something, like:
					// this.get("craft-onload").Call()
					// so the logic will be controlled by user in JS land.
					js.Global().Get("location").Call("reload")
				}
				return js.ValueOf(nil)
			})
			resp.Call("json").Call("then", jsonPromiseFunc)

			return nil
		})

		return utils.HTTPRequest(req, resolveFunc, utils.NopFunc)
	}

	for card := range slices.Values(m.cards) {
		m.lookup[card.ID] = card
	}
	return js.ValueOf(nil)
}

// startSession based on the list of most urgent due date cards
// prepare the list of cards to be review in this session.
func (m *SpacedManager) startSession() any {
	revieweds := internalfsrs.Cards{}
	news := internalfsrs.Cards{}
	for _, card := range m.cards {
		if card.Due.IsZero() {
			news = append(news, card)
		} else {
			revieweds = append(revieweds, card)
		}
	}
	sort.Sort(revieweds)
	sort.Sort(news)
	numCards := min(m.targetNum, len(m.cards))
	cards := make(internalfsrs.Cards, numCards)

	numReviewed := int(0.2 * float64(numCards))
	numNews := numCards - numReviewed

	i := 0
	for idx := 0; idx < numReviewed; idx, i = idx+1, i+1 {
		cards[i] = revieweds[idx]
	}

	for idx := 0; idx < numNews; idx, i = idx+1, i+1 {
		cards[i] = news[idx]
	}

	rand.Shuffle(numCards, func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})

	m.currSession = session.NewSession(cards)
	return model.PayloadResponse("success")
}

func (m *SpacedManager) addRecord(record session.Record) error {
	id := len(m.records)

	ptr := &record
	ptr.ID = id
	m.records = append(m.records, ptr)
	return nil
}

func (m *SpacedManager) completeSession() {
	record := session.NewRecordFromSession(m.currSession)
	m.addRecord(record)
	m.currSession = nil
	crafter.StorageSetItem("records", m.records)
	crafter.StorageSetItem("flashcards", m.cards)
	crafter.StorageRemoveItem("currentSession")
}

// parsedFromLocalState pull the state passed from web browser.
func (m *SpacedManager) parsedFromLocalState() error {
	if err := crafter.StorageGetItem("flashcards", &m.cards); err != nil {
		return fmt.Errorf("failed to pull flashcards, err: %v", err)
	}
	if err := crafter.StorageGetItem("records", &m.records); err != nil {
		return fmt.Errorf("failed to pull sessions, err: %v", err)
	}

	if err := crafter.StorageGetItem("currentSession", &m.currSession); err != nil {
		return fmt.Errorf("failed to save current session, err: %v", err)
	}
	return nil
}

// handleSaveState push the state from wasm land to js land
func (m *SpacedManager) handleSaveState() error {
	if err := crafter.StorageSetItem("flashcards", &m.cards); err != nil {
		return fmt.Errorf("failed to push flashcards, err: %v", err)
	}
	if err := crafter.StorageSetItem("records", &m.records); err != nil {
		return fmt.Errorf("failed to push sessions, err: %v", err)
	}

	if err := crafter.StorageSetItem("currentSession", &m.currSession); err != nil {
		return fmt.Errorf("failed to save current session, err: %v", err)
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
		crafter.NavigateTo("/stats")
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
	if m.currSession != nil {
		askPromiseFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
			confirmJS := crafter.Call("confirm", "the current session is not finished, do you want to continue it first.")
			confirm := confirmJS.Bool()
			crafter.Get("console.log").Invoke(confirmJS)
			if !confirm {
				crafter.Get("console.log").Invoke(js.ValueOf("here"))
				m.currSession = nil
				m.handleSaveState()
				crafter.Reload()
			}
			return js.ValueOf(nil)
		})
		askPromiseFunc.Invoke()
		return model.PayloadResponse("ready")
	}
	m.startSession()
	return model.PayloadResponse("ready")
}

func JSPlay(_ js.Value, args []js.Value) any {
	if len(args) != 3 {
		return model.PayloadResponse(map[string]any{"error": "number of args pass to this method should = 3!"})
	}
	var sound64 string
	// if the request include non-empty base64 encoded sound payload, then
	// try our-best to play it first.
	if args[2].Truthy() && args[2].Type() == js.TypeString {
		sound64 = args[2].String()
	}

	// if we reach this part, it's mean the sound encoded payload is not
	// ready, feetch ones from proxy server and return to the js land.
	if sound64 == "" {
		word := args[0].String()
		ipa := args[1].String()
		encodedIPA := base64.StdEncoding.EncodeToString([]byte(ipa))

		headers := http.Header{}
		headers.Set(handler.CraftIPAHeader, encodedIPA)
		headers.Set(handler.CraftWordHeader, word)
		req := utils.Request{
			URL:    "/api/sound/index",
			Method: http.MethodGet,
			Header: headers,
		}
		resolveFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
			jsonPromiseFunc := js.FuncOf(func(this js.Value, args []js.Value) (result any) {
				jsonString := js.Global().Get("JSON").Call("stringify", args[0]).String()

				respPayload := make(map[string]string)
				err := json.Unmarshal([]byte(jsonString), &respPayload)
				if err != nil {
					fmt.Printf("failed to unmarshal cards: %v", err)
					result = js.ValueOf(nil)
					return result
				}
				if errMsg, exists := respPayload["error"]; exists {
					fmt.Println(errMsg)
					result = js.ValueOf(nil)
					return result
				}
				if payload, exists := respPayload["payload"]; exists {
					result = js.ValueOf(payload)
					if err := playSound64(payload); err != nil {
						fmt.Println("failed to playsound", err)
					}
					return result
				}
				fmt.Println("not reached case")
				result = js.ValueOf(nil)
				return result
			})
			// defer jsonPromiseFunc.Release()

			// Don't release here as this workload will be later execute by JS engine.
			// Release here, mean in the future, when the JS engine execute this, the method will
			// not ready.
			//
			// NOTE, JS call here is a async call, we don't expect the callBackHandlerFunc or any func
			// complete exe3ctue in this scope context.

			return args[0].Call("json").Call("then", jsonPromiseFunc)
		})

		return utils.HTTPRequest(req, resolveFunc, utils.NopFunc)
		// js.Global().Get("console").Call("log", "response from http request", resp)
	}

	if sound64 != "" {
		if err := playSound64(sound64); err != nil {
			fmt.Println("failed to playsound64", err)
		}
		return js.ValueOf(sound64)
	}
	return js.ValueOf(nil)
}

func playSound64(sound64 string) error {
	payload, err := base64.StdEncoding.DecodeString(sound64)
	if err != nil {
		return err
	}
	go func() {
		audioDataLength := len(payload)
		fmt.Printf("Go WASM: Payload audio data length: %d bytes\n", audioDataLength)
		state := audioContext.Get("state").String()
		switch state {
		case "suspended":
			// Ensure AudioContext is in a runnable state.
			fmt.Println("Go WASM: AudioContext is suspended, attempting to resume.")
			resumePromise := audioContext.Call("resume")
			resumePromise.Call("then", js.FuncOf(func(this js.Value, pArgs []js.Value) any {
				fmt.Println("Go WASM: AudioContext resumed successfully.")
				return nil
			}), js.FuncOf(func(this js.Value, pArgs []js.Value) any {
				err := pArgs[0]
				errMsg := fmt.Sprintf("failed to resume AudioContext: %v", err.String())
				fmt.Printf("Go WASM Error: %s\n", errMsg)
				js.Global().Get("alert").Call("alert", errMsg)
				return nil
			}))
			// Give the promise some time to resolve before proceeding.
			// In a real application, you might use a channel or a more robust promise handling pattern.
			// For this example, a small sleep might work, but it's not ideal for production.
			// A better approach would be to chain the decode/play calls after the resume promise resolves.
			time.Sleep(100 * time.Millisecond) // This is a simple, but not robust, way to wait.
		case "running":
			if soundSource != nil {
				// any non-null soundsource mean there's an playing audio.
				// stop it first
				fmt.Println("stop")
				soundSource.(js.Value).Call("stop")
				soundSource.(js.Value).Call("disconnect")
			}
		}

		// Create a JavaScript ArrayBuffer to hold the audio data
		jsAudioBuffer := js.Global().Get("ArrayBuffer").New(audioDataLength)

		// Create a Uint8Array view over the ArrayBuffer
		jsUint8Array := js.Global().Get("Uint8Array").New(jsAudioBuffer)

		// Copy bytes from Go []byte to the JS Uint8Array
		js.CopyBytesToJS(jsUint8Array, payload)
		// 5. Decode the ArrayBuffer into an AudioBuffer (asynchronously)
		decodePromise := audioContext.Call("decodeAudioData", jsAudioBuffer)

		// Handle the decode Promise resolution and rejection from Go
		decodePromise.Call("then", js.FuncOf(func(this js.Value, pArgs []js.Value) any {
			audioBuffer := pArgs[0] // The decoded AudioBuffer
			fmt.Println("Go WASM: Audio decoded successfully.")

			source := js.Global().Get("AudioBufferSourceNode").New(audioContext)
			soundSource = source
			source.Set("buffer", audioBuffer) // Set the decoded audio data
			source.Call("connect", audioContext.Get("destination"))
			source.Call("start", 0) // Play from the beginning
			source.Set("onended", js.FuncOf(func(this js.Value, _ []js.Value) any {
				fmt.Println("Go WASM: Audio playback finished.")
				source.Call("disconnect")
				soundSource = nil
				return js.ValueOf(nil)
			}))

			return nil
		}), js.FuncOf(func(this js.Value, pArgs []js.Value) any {
			err := pArgs[0] // The error object
			errMsg := fmt.Sprintf("Error decoding audio data: %v", err.String())
			fmt.Printf("Go WASM Error: %s\n", errMsg)
			js.Global().Get("alert").Call("alert", errMsg)
			return nil
		}))
	}()
	return nil
}

func main() {
	m, err := NewSpacedManger()
	if err != nil {
		fmt.Println("failed to init: " + err.Error())
	}

	wasm := crafter.NewWasm()
	wasm.HandleFunc("init", m.JSInit)
	wasm.HandleFunc("start", m.JSStart)
	wasm.HandleFunc("next", m.JSNext)
	wasm.HandleFunc("submit", m.JSSubmit)
	wasm.HandleFunc("play", JSPlay)

	fmt.Println(wasm.ListenAndServe())
}
