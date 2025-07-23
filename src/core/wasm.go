//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"syscall/js"

	"github.com/hnimtadd/spaced/src/core/model"
)

type SpacedManager struct {
	Cards        []model.Card
	localStorage js.Value
}

func NewSpacedManger() (*SpacedManager, error) {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.Truthy() {
		return nil, errors.New("localStorage from JS is not truthy")
	}
	m := &SpacedManager{
		localStorage: localStorage,
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

				if err := json.Unmarshal([]byte(jsonString), &m.Cards); err != nil {
					fmt.Println("failed to unmarshal cards:", err)
				} else {
					m.handlePushState()
					fmt.Println("cards loaded successfully")
				}
				return nil
			}))
		}()
	}
	fmt.Println("pull state from localStorage completed")
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

	cards := []model.Card{}
	if err := json.Unmarshal([]byte(data), &cards); err != nil {
		return errors.New("could not unmarshal the data, got: " + err.Error())
	}
	m.Cards = cards

	return nil
}

// handlePushState push the state from wasm land to js land
func (m *SpacedManager) handlePushState() error {
	fmt.Println("handle push state to localStorage")
	dataBytes, err := json.Marshal(m.Cards)
	if err != nil {
		return fmt.Errorf("failed to push state to localStorage: %w", err)
	}

	m.localStorage.Call("setItem", "flashcards", string(dataBytes))
	fmt.Println("handle push state to localStorage, complete")
	return nil
}

func (m *SpacedManager) next(_ js.Value, _ []js.Value) any {
	idx := rand.Intn(len(m.Cards))
	jsonBytes, err := json.Marshal(m.Cards[idx])
	if err != nil {
		return js.ValueOf(map[string]any{"error": "could not marshal the card, got: " + err.Error()})
	}
	return js.ValueOf(string(jsonBytes))
}

func main() {
	c := make(chan struct{})

	m, err := NewSpacedManger()
	if err != nil {
		fmt.Println("failed to init: " + err.Error())
	}

	// Map of exposed Go methods for easy lookup in JS land.
	goFuncs := map[string]js.Func{
		"next": js.FuncOf(m.next),
	}

	wasmBridge := js.Global().Get("Object").New()
	for name, fn := range goFuncs {
		wasmBridge.Set(name, fn)
	}

	js.Global().Set("wasmBridge", wasmBridge)

	// Keep the Go program running so the WASM module doesn't exit.
	<-c
}
