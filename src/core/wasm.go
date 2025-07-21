//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"syscall/js"

	"github.com/hnimtadd/spaced/src/core/model"
)

type SpacedManager struct {
	Cards []model.Card
}

func (m *SpacedManager) store(this js.Value, args []js.Value) any {
	fmt.Println("called store with args", args)
	data := args[0].String()
	if data == "" {
		return js.ValueOf(map[string]any{"error": "empty arg"})
	}

	cards := []model.Card{}
	if err := json.Unmarshal([]byte(data), &cards); err != nil {
		return js.ValueOf(map[string]any{"error": "could not unmarshal the data, got: " + err.Error()})
	}
	m.Cards = cards
	return js.ValueOf(map[string]any{"error": nil})
}

func (m *SpacedManager) suggest(_ js.Value, _ []js.Value) any {
	idx := rand.Intn(len(m.Cards))
	jsonBytes, err := json.Marshal(m.Cards[idx])
	if err != nil {
		return js.ValueOf(map[string]any{"error": "could not marshal the card, got: " + err.Error()})
	}
	return js.ValueOf(string(jsonBytes))
}

func main() {
	c := make(chan struct{})

	m := SpacedManager{}
	// Map of exposed Go methods for easy lookup in JS land.
	goFuncs := map[string]js.Func{
		"store":   js.FuncOf(m.store),
		"suggest": js.FuncOf(m.suggest),
	}

	wasmBridge := js.Global().Get("Object").New()
	for name, fn := range goFuncs {
		wasmBridge.Set(name, fn)
	}

	js.Global().Set("wasmBridge", wasmBridge)

	// Keep the Go program running so the WASM module doesn't exit.
	<-c
}
