package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	coreHtml "github.com/hnimtadd/spaced/src/html"
	"github.com/hnimtadd/spaced/src/utils"
	"golang.org/x/net/html"
)

const url = "https://dictionary.cambridge.org/us/dictionary/english/%s"

var defaultHeaders = http.Header{
	http.CanonicalHeaderKey("User-Agent"): []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		utils.SMarshal(w, map[string]any{"error": "only GET allowed"})
		return
	}

	word := r.Header.Get("Craft-word")
	region := r.Header.Get("Craft-region")
	if region == "" {
		region = "us"
	}

	ipa := r.Header.Get("Craft-ipa")
	if word == "" || ipa == "" {
		w.WriteHeader(http.StatusBadRequest)
		utils.SMarshal(w, map[string]any{"error": "bad craft headers"})
		return
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(url, word), http.NoBody)

	req.Header = defaultHeaders

	client := http.Client{
		Timeout: time.Second * 10,
	}

	dictionaryResponse, _ := client.Do(req)

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

	var processWordNode coreHtml.WalkFunc = func(node *html.Node) error {
		if node.Type != html.ElementNode ||
			node.Data != "source" ||
			!coreHtml.HasAttr(node.Attr, "type", "audio/mpeg") {
			return nil
		}

		greatGrandParent := node.Parent.Parent.Parent

		for child := range greatGrandParent.ChildNodes() {
			if child.Type == html.ElementNode &&
				child.Data == "span" &&
				coreHtml.HasAttr(child.Attr, "class", "pron", "dpron") {
				for grandChild := range child.ChildNodes() {
					if grandChild.Type == html.ElementNode &&
						grandChild.Data == "span" &&
						coreHtml.HasAttr(grandChild.Attr, "class", "ipa", "dipa") {
						fmt.Println("IPA", grandChild.FirstChild.Data)
						if grandChild.FirstChild.Data == ipa {
							soundURLs = append(soundURLs, coreHtml.GetAttr(node.Attr, "src"))
						}

						return coreHtml.ErrWalkSkip
					}
				}
			}
		}
		return coreHtml.ErrWalkSkip
	}

	coreHtml.Walk(doc, func(node *html.Node) error {
		if node.Type == html.ElementNode && node.Data == "div" && coreHtml.HasAttr(node.Attr, "class", "pos-header", "dpos-h") {
			coreHtml.Walk(node, processWordNode)
			return coreHtml.ErrWalkSkip
		}
		return nil
	})

	fmt.Println(soundURLs)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		utils.SMarshal(w, map[string]any{"error": err.Error()})
		return
	}
	// url := "https://dictionary.cambridge.org/%s"
	// req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf(url, soundURL), http.NoBody)
	//
	// req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	//
	// resp, _ := http.DefaultClient.Do(req)
	// io.Copy(w, resp.Body)
}
