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

func Serialize[From any](from From) ([]byte, error) {
	result, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DeserializeTo[To any](from []byte, to To) error {
	if err := json.Unmarshal(from, to); err != nil {
		return err
	}
	return nil
}
