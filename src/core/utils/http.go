//go:build js && wasm

package utils

import (
	"bytes"
	"io"
	"net/http"
	"syscall/js"
	"time"
)

type Request struct {
	URL       string
	Method    string
	Header    http.Header
	BodyBytes []byte
}

type Response struct {
	Header    http.Header
	BodyBytes []byte
	Status    int
}

type Options struct {
	Timeout time.Duration
}

var DefaultOptions = Options{
	Timeout: time.Second * 2,
}

var NopFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
	return js.ValueOf(nil)
})

func HTTPRequest(req Request, resolveFn js.Func, rejectFn js.Func) js.Value {
	// Handler for the promise
	// We need to return a promise because HTTP requests are blocking in Go.
	handler := js.FuncOf(func(this js.Value, args []js.Value) any {
		resolveHandler := args[0]
		rejectHandler := args[1]
		// Run the logic asynchrnously.
		go func() {
			client := &http.Client{}
			r, _ := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.BodyBytes))
			r.Header = req.Header
			res, err := client.Do(r)
			if err != nil {
				// Handle errors: reject the Promise if we have an error
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				rejectHandler.Invoke(rejectFn.Invoke(errorObject))
				return
			}
			defer res.Body.Close()

			// Read the response body.
			dataBody, err := io.ReadAll(res.Body)
			if err != nil {
				// Handle errors here too
				errorConstructor := js.Global().Get("Error")
				errorObject := errorConstructor.New(err.Error())
				rejectHandler.Invoke(rejectFn.Invoke(errorObject))
				return
			}
			// "data" is a byte slice, so we need to convert it to a JS Uint8Array object
			arrayConstructor := js.Global().Get("Uint8Array")
			dataJS := arrayConstructor.New(len(dataBody))
			js.CopyBytesToJS(dataJS, dataBody)

			// Create a Response object and pass the data
			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(dataJS)

			// Resolve the Promise
			resolveHandler.Invoke(resolveFn.Invoke(response))
		}()

		// The handler of a Promise doesn't return any value.
		return nil
	})

	// Create and return the Promise object.
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}
