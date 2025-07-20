//go:build js && wasm

package main

import (
	"syscall/js"
)

func hello(_ js.Value, args []js.Value) any {
	return js.ValueOf("hello from Go land")
}

func main() {
	c := make(chan struct{})

	// Map of exposed Go methods for easy lookup in JS land.
	goFuncs := map[string]js.Func{
		"hello": js.FuncOf(hello),
	}

	wasmBridge := js.Global().Get("Object").New()
	for name, fn := range goFuncs {
		wasmBridge.Set(name, fn)
	}

	js.Global().Set("wasmBridge", wasmBridge)

	// Keep the Go program running so the WASM module doesn't exit.
	<-c
}
