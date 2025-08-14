//go:build js && wasm

package crafter

import (
	"errors"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/hnimtadd/spaced/src/core/utils"
)

func StorageSetItem(key string, data any) error {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.Truthy() {
		return errors.New("localStorage from JS is not truthy")
	}

	dataBytes, err := utils.Serialize(data)
	if err != nil {
		return fmt.Errorf("failed to push state to localStorage: %w", err)
	}

	localStorage.Call("setItem", key, string(dataBytes))
	return nil
}

func StorageRemoveItem(key string) error {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.Truthy() {
		return errors.New("localStorage from JS is not truthy")
	}

	localStorage.Call("removeItem", key)
	return nil
}

func StorageGetItem[T any](key string, to T) error {
	localStorage := js.Global().Get("localStorage")
	if !localStorage.Truthy() {
		return errors.New("localStorage from JS is not truthy")
	}
	dataRaw := localStorage.Call("getItem", key)
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

func NavigateTo(addr string) {
	js.Global().Get("location").Call("assign", addr)
}

func Reload() {
	js.Global().Get("location").Call("reload")
}

func Call(method string, args ...any) js.Value {
	return js.Global().Call(method, args...)
}

func Get(objectKey string) js.Value {
	object := js.Global()
	for segment := range strings.SplitSeq(objectKey, ".") {
		object = object.Get(segment)
	}
	return object
}
