package html

import (
	"slices"
	"strings"

	coreHtml "golang.org/x/net/html"
)

func GetAttr(attrs []coreHtml.Attribute, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func HasAttr(attrs []coreHtml.Attribute, key string, values ...string) bool {
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
