// go: build js && wasm
package model

type Card struct {
	Word       string `json:"word"`
	IPA        string `json:"ipa"`
	Definition string `json:"definition"`
	Example    string `json:"example"`
}
