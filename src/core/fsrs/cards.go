//go:build js && wasm

package fsrs

import (
	"strings"

	"github.com/hnimtadd/spaced/src/core/model"
)

type Cards []*model.Card

func (a Cards) Len() int      { return len(a) }
func (a Cards) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Cards) Less(i, j int) bool {
	compare := a[i].Due.Compare(a[j].Due)
	if compare < 0 {
		return true
	} else if compare > 0 {
		return false
	}

	// if we have cards with same urgencity, then sort based on state.
	return a[i].State < a[j].State
}

func (a Cards) String() string {
	words := make([]string, a.Len())
	for i, card := range a {
		words[i] = card.Word
	}
	return strings.Join(words, ",")
}
