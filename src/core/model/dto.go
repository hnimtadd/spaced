// go: build js && wasm
package model

import (
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
)

type Card struct {
	Word       string `json:"word"`
	IPA        string `json:"ipa"`
	Definition string `json:"definition"`
	Example    string `json:"example"`

	// fsrs fields
	Due        time.Time   `json:"due"`
	Stability  float64     `json:"stability"`
	Difficulty float64     `json:"difficulty"`
	Elapsed    int64       `json:"elapsed"`
	Scheduled  int64       `json:"scheduled"`
	Reps       int         `json:"reps"`
	Lapses     int         `json:"lapses"`
	State      fsrs.State  `json:"state"`
	LastReview time.Time   `json:"last_review"`
}

func (c *Card) ToFsrsCard() fsrs.Card {
	return fsrs.Card{
		Due:        c.Due,
		Stability:  c.Stability,
		Difficulty: c.Difficulty,
		Elapsed:    c.Elapsed,
		Scheduled:  c.Scheduled,
		Reps:       c.Reps,
		Lapses:     c.Lapses,
		State:      c.State,
		LastReview: c.LastReview,
	}
}
