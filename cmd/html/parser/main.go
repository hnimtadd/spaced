package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	coreHtml "github.com/hnimtadd/spaced/src/html"
	"golang.org/x/net/html"
)

func main() {
	file, _ := os.Open("./sample.html")
	doc, _ := html.Parse(file)
	region := "us"
	ipa := "əˈɡriː"

	soundURLs := []string{}
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

	err := coreHtml.Walk(doc, func(node *html.Node) error {
		if node.Type == html.ElementNode && node.Data == "div" && coreHtml.HasAttr(node.Attr, "class", "pos-header", "dpos-h") {
			coreHtml.Walk(node, processWordNode)
			return coreHtml.ErrWalkSkip
		}
		return nil
	})
	fmt.Println(err)

	fmt.Println(soundURLs)

	if len(soundURLs) == 0 {
		panic("empty sound url")
	}
	url := "https://dictionary.cambridge.org/%s"
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf(url, soundURLs[0]), http.NoBody)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, _ := http.DefaultClient.Do(req)

	file, _ = os.OpenFile("./sample.mp3", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o777)

	io.Copy(file, resp.Body)
}
