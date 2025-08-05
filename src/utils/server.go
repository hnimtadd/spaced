package utils

import (
	"encoding/json"
	"io"
)

func SMarshal(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	return enc.Encode(data)
}
