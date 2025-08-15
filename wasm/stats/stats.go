//go:build js && wasm

package main

import (
	"fmt"
	"html/template"
	"slices"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/hnimtadd/spaced/src/core/model"
	"github.com/hnimtadd/spaced/src/core/session"
	"github.com/hnimtadd/spaced/src/crafter"
)

func JSStats(this js.Value, args []js.Value) any {
	records := []session.Record{}
	if err := crafter.StorageGetItem("records", &records); err != nil {
		return model.ErrorResponse("failed to read from records: " + err.Error())
	}
	tpl := `<div class="max-w-5xl sm:w-[30rem] md:w-[40rem] lg:w-[50rem] mx-auto h-screen p-4 space-y-4">{{range .Sessions}}{{.}}{{end}}</div>`
	tmpl, err := template.New("stats").Parse(tpl)
	if err != nil {
		fmt.Println("failed to init tmpl: " + err.Error())
		return model.ErrorResponse("failed to init tmpl: " + err.Error())
	}
	eles := make([]template.HTML, len(records))

	for i, session := range records {
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

func JSReplaySession(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		fmt.Println("number of args is not valid")
		return model.ErrorResponse("number of args passed to this method should = 1")
	}

	records := []session.Record{}
	if err := crafter.StorageGetItem("records", &records); err != nil {
		return model.ErrorResponse("failed to read from records: " + err.Error())
	}

	recordIDString := args[0].String()
	recordID, err := strconv.Atoi(recordIDString)
	if err != nil {
		fmt.Println("could not convert to int from", recordIDString)
		return model.ErrorResponse(err.Error())
	}
	if recordID < 0 || recordID >= len(records) {
		fmt.Println("invalid record")
		return model.ErrorResponse("invalid recordID")
	}
	crafter.NavigateTo("/session?id=" + recordIDString)
	return nil
}

func main() {
	wasm := crafter.NewWasm()
	wasm.HandleFunc("stats", JSStats)
	wasm.HandleFunc("replay", JSReplaySession)

	fmt.Println(wasm.ListenAndServe())
}
