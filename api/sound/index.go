package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func jsonResponse(w io.Writer, data any) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataBytes)
	if err != nil {
		return err
	}
	return nil
}

type requestBody struct {
	IPA string `json:"ipa"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		jsonResponse(w, map[string]any{"error": "only POST allowed"})
		return
	}

	reqBody := new(requestBody)
	if err := json.NewDecoder(r.Body).Decode(reqBody); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		jsonResponse(w, map[string]any{"error": "could not parse request body " + err.Error()})
		return
	}

	token := os.Getenv("ELEVENLABS_API_KEY")
	if token == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		jsonResponse(w, map[string]any{"error": "empty token found"})
		return
	}

	body := map[string]any{
		"text":     fmt.Sprintf(`<phoneme alphabet="ipa" ph="%s"></phoneme>`, reqBody.IPA),
		"model_id": "eleven_monolingual_v1",
		"voice_settings": map[string]any{
			"speed": 0.75,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		jsonResponse(w, map[string]any{"error": "could not marshal upstream request body " + err.Error()})
		return
	}

	payload := bytes.NewReader(bodyBytes)

	url := "https://api.elevenlabs.io/v1/text-to-speech/21m00Tcm4TlvDq8ikWAM?output_format=mp3_44100_128"
	req, _ := http.NewRequest(http.MethodPost, url, payload)
	req.Header.Add("xi-api-key", token)
	req.Header.Add("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	defer func() { _ = resp.Body.Close() }()
	respBytes, _ := io.ReadAll(resp.Body)
	soundPayload := base64.RawStdEncoding.EncodeToString(respBytes)
	jsonResponse(w, map[string]any{"payload": soundPayload})
}
