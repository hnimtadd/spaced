package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"golang.org/x/net/html"
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

const url = "https://dictionary.cambridge.org/us/dictionary/english/%s"

var defaultHeaders = http.Header{
	http.CanonicalHeaderKey("User-Agent"): []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"},
}

func GetAttr(attrs []html.Attribute, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func HasAttr(attrs []html.Attribute, key string, values ...string) bool {
	for _, attr := range attrs {
		if attr.Key != key {
			continue
		}

		attrValues := strings.Split(attr.Val, " ")
		for _, val := range values {
			if !slices.Contains(attrValues, val) {
				return false
			}
		}
		return true
	}
	return false
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		jsonResponse(w, map[string]any{"error": "only GET allowed"})
		return
	}
	word := r.Header.Get("Craft-word")
	region := r.Header.Get("Craft-region")
	if region == "" {
		region = "us"
	}
	ipa := r.Header.Get("Craft-word")

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
		jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}

	processWordNode := func(node *html.Node) (string, error) {
	sibLoop:
		for sib := range node.ChildNodes() {
			if sib.Type != html.ElementNode ||
				sib.Data != "span" ||
				!HasAttr(sib.Attr, "class", region, "dpron-i") {
				continue
			}

			var foundIPANode bool
			var soundURL string

			for child := range sib.ChildNodes() {
				if child.Type == html.ElementNode &&
					child.Data == "span" &&
					HasAttr(child.Attr, "class", region, "prop", "dpron") {
					for grandChild := range child.ChildNodes() {
						if grandChild.Type == html.ElementNode &&
							grandChild.Data == "span" &&
							HasAttr(grandChild.Attr, "class", "ipa", "dipa") {
							if grandChild.FirstChild.Data != ipa {
								continue sibLoop
							}
							foundIPANode = true
						}
					}
				}

				if sib.Type == html.ElementNode &&
					sib.Data == "span" &&
					HasAttr(sib.Attr, "class", region, "daud") {
					for grandChild := range child.ChildNodes() {
						if grandChild.Type == html.ElementNode &&
							grandChild.Data == "audio" {

							for grandGrandChild := range grandChild.ChildNodes() {
								if grandGrandChild.Type == html.ElementNode &&
									grandGrandChild.Data == "source" &&
									HasAttr(grandGrandChild.Attr, "type", "audio/mpeg") {
									soundURL = GetAttr(grandGrandChild.Attr, "src")

									if foundIPANode {
										goto returnSound
									}
								}
							}
						}
					}
				}
			}

		returnSound:
			if foundIPANode {
				return soundURL, nil
			}
		}
		return "", fmt.Errorf("not found")
	}

	processDocNode := func(node *html.Node) (string, error) {
		if node.Type == html.ElementNode && node.Data == "div" {
			for _, attr := range node.Attr {
				if attr.Key == "class" &&
					strings.Contains(attr.Val, "pos-header") &&
					strings.Contains(attr.Val, "dpos-h") {
					soundURL, err := processWordNode(node)
					if err != nil {
						return soundURL, nil
					}
				}
			}
		}
		return "", fmt.Errorf("not found")
	}

	soundURL, err := processDocNode(doc)
	if err != nil {
		fmt.Println(3)
		w.WriteHeader(http.StatusServiceUnavailable)
		jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}
	fmt.Println(soundURL)
	jsonResponse(w, map[string]any{"url": soundURL})
}
