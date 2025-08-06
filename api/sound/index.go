package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	coreHtml "github.com/hnimtadd/spaced/src/html"
	"github.com/hnimtadd/spaced/src/utils"
	"golang.org/x/net/html"
)

const (
	baseURL           = "https://dictionary.cambridge.org"
	CraftWordHeader   = "Craft-word"
	CraftRegionHeader = "Craft-region"
	CraftIPAHeader    = "Craft-ipa"
)

var defaultHeaders = http.Header{
	http.CanonicalHeaderKey("User-Agent"): []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		utils.SMarshal(w, map[string]any{"error": "only GET allowed"})
		return
	}

	word := strings.TrimSpace(r.Header.Get(CraftWordHeader))
	region := strings.TrimSpace(r.Header.Get(CraftRegionHeader))
	if region == "" {
		region = "us"
	}

	encodedIPA := r.Header.Get(CraftIPAHeader)
	ipaBytes, err := base64.StdEncoding.DecodeString(encodedIPA)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.SMarshal(w, map[string]any{"error": "invalid IPA header, expect base64 encoded value"})
		return
	}
	ipa := strings.TrimSpace(string(ipaBytes))
	if word == "" || ipa == "" {
		w.WriteHeader(http.StatusBadRequest)
		utils.SMarshal(w, map[string]any{"error": "bad craft headers"})
		return
	}

	req, _ := http.NewRequest(http.MethodGet, baseURL+"/us/dictionary/english/"+word, http.NoBody)

	req.Header = defaultHeaders

	client := http.Client{
		Timeout: time.Second * 10,
	}

	dictionaryResponse, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.SMarshal(w, map[string]any{"error": "failed to request dict: " + err.Error()})
		return
	}

	if dictionaryResponse.StatusCode != http.StatusOK {
		w.WriteHeader(dictionaryResponse.StatusCode)
		io.Copy(w, dictionaryResponse.Body)
		return
	}

	doc, err := html.Parse(dictionaryResponse.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		utils.SMarshal(w, map[string]any{"error": err.Error()})
		return
	}

	soundURLs := []string{}
	candidateSoundURL := []string{} // candidate means we don't have exact ipa,
	// but we has same word, some word might has multiple IPA flavors.

	var processWordNode coreHtml.WalkFunc = func(node *html.Node) error {
		if node.Type != html.ElementNode ||
			node.Data != "source" ||
			!coreHtml.HasAttr(node.Attr, "type", "audio/mpeg") {
			return nil
		}

		greatGrandParent := node.Parent.Parent.Parent

		if greatGrandParent.Type != html.ElementNode ||
			greatGrandParent.Data != "span" ||
			!coreHtml.HasAttr(greatGrandParent.Attr, "class", region, "dpron-i") {
			return nil
		}

		for child := range greatGrandParent.ChildNodes() {
			if child.Type == html.ElementNode &&
				child.Data == "span" &&
				coreHtml.HasAttr(child.Attr, "class", "pron", "dpron") {
				for grandChild := range child.ChildNodes() {
					if grandChild.Type == html.ElementNode &&
						grandChild.Data == "span" &&
						coreHtml.HasAttr(grandChild.Attr, "class", "ipa", "dipa") {
						if strings.TrimSpace(grandChild.FirstChild.Data) == ipa {
							soundURLs = append(soundURLs, coreHtml.GetAttr(node.Attr, "src"))
						} else {
							candidateSoundURL = append(candidateSoundURL, coreHtml.GetAttr(node.Attr, "src"))
						}

						return coreHtml.ErrWalkSkip
					}
				}
			}
		}
		return coreHtml.ErrWalkSkip
	}

	err = coreHtml.Walk(doc, func(node *html.Node) error {
		if node.Type == html.ElementNode && node.Data == "div" && coreHtml.HasAttr(node.Attr, "class", "pos-header", "dpos-h") {
			coreHtml.Walk(node, processWordNode)
			return coreHtml.ErrWalkSkip
		}
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		utils.SMarshal(w, map[string]any{"error": err.Error()})
		return
	}

	fmt.Println(soundURLs)

	if len(soundURLs) == 0 && len(candidateSoundURL) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		utils.SMarshal(w, map[string]any{"error": "could not find any sound url."})
		return
	}

	var soundURL string
	if len(soundURLs) != 0 {
		soundURL = soundURLs[0]
	} else {
		soundURL = candidateSoundURL[0]
	}
	req, _ = http.NewRequest(http.MethodGet, baseURL+soundURL, http.NoBody)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, _ := http.DefaultClient.Do(req)
	defer func() { _ = resp.Body.Close() }()
	respBytes, _ := io.ReadAll(resp.Body)
	soundPayload := base64.RawStdEncoding.EncodeToString(respBytes)
	utils.SMarshal(w, map[string]any{"payload": soundPayload})
}
