package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

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

func main() {
	file, _ := os.Open("./sample.html")
	doc, _ := html.Parse(file)
	region := "us"
	ipa := "əˈɡriː"

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
					HasAttr(child.Attr, "class", "pron", "dpron") {
					for grandChild := range child.ChildNodes() {
						if grandChild.Type == html.ElementNode &&
							grandChild.Data == "span" &&
							HasAttr(grandChild.Attr, "class", "ipa", "dipa") {
							fmt.Println("IPA", grandChild.FirstChild.Data)
							if grandChild.FirstChild.Data != ipa {
								fmt.Println("found different IPA, skip this node")
								continue sibLoop
							}
							foundIPANode = true
						}
					}
				}

				if child.Type == html.ElementNode &&
					child.Data == "span" &&
					HasAttr(child.Attr, "class", "daud") {
					for grandChild := range child.ChildNodes() {
						if grandChild.Type == html.ElementNode &&
							grandChild.Data == "audio" {

							for grandGrandChild := range grandChild.ChildNodes() {
								if grandGrandChild.Type == html.ElementNode &&
									grandGrandChild.Data == "source" &&
									HasAttr(grandGrandChild.Attr, "type", "audio/mpeg") {
									soundURL = GetAttr(grandGrandChild.Attr, "src")
									fmt.Println("found SourceURL", soundURL, foundIPANode)
								}
							}
						}
					}
				}
				if foundIPANode && soundURL != "" {
					fmt.Println("found both ipa and sound, early return")
					break
				}
			}
			fmt.Println("return", foundIPANode, soundURL)
			if foundIPANode {
				return soundURL, nil
			}
		}
		return "", fmt.Errorf("not found")
	}

	var processDocNode func(node *html.Node) (string, error)
	processDocNode = func(node *html.Node) (string, error) {
		if node.Type == html.ElementNode && node.Data == "div" {
			for _, attr := range node.Attr {
				if attr.Key == "class" &&
					strings.Contains(attr.Val, "pos-header") &&
					strings.Contains(attr.Val, "dpos-h") {
					soundURL, err := processWordNode(node)
					fmt.Println("receive", soundURL, err)
					if err == nil {
						return soundURL, nil
					}
				}
			}
		}
		for child := range node.ChildNodes() {
			if soundURL, err := processDocNode(child); err == nil {
				return soundURL, nil
			}
		}
		return "", fmt.Errorf("not found")
	}

	soundURL, err := processDocNode(doc)
	if err != nil {
		panic(err)
	}
	url := "https://dictionary.cambridge.org/%s"
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(url, soundURL), http.NoBody)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, _ := http.DefaultClient.Do(req)

	file, _ = os.OpenFile("./sample.mp3", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o777)

	io.Copy(file, resp.Body)
}
