// go: build js && wasm
package utils

import "encoding/json"

func Deserialize[To any](from []byte) (*To, error) {
	to := new(To)
	if err := json.Unmarshal(from, to); err != nil {
		return to, err
	}
	return to, nil
}

func Serialize[To any](from To) ([]byte, error) {
	result, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}
	return result, nil
}
