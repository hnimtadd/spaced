//go:build js && wasm

package model

import "syscall/js"

func ErrorResponse(err string) js.Value {
	return js.ValueOf(map[string]any{
		"success": false,
		"error":   err,
	})
}

func PayloadResponse(payload any) js.Value {
	return js.ValueOf(map[string]any{
		"success": true,
		"payload": payload,
	})
}

func StopResponse() js.Value {
	return js.ValueOf(map[string]any{
		"success": true,
		"stop":    true,
	})
}
