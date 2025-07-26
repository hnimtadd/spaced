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

	// unique index of the cards.
	ID int

	// Due is The date the card is schduled for next review.
	Due time.Time `json:"due"`
	// The card's current memory stabitlity value.
	Stability float64 `json:"stability"`
	// The card's current difficulty value.
	Difficulty float64 `json:"difficulty"`

	// The number of days since the last review.
	ElapsedDays uint64 `json:"elapsed"`
	// The length of the last of interval.
	ScheduledDays uint64 `json:"scheduled"`
	// The total number of times the card has been reviews.
	Reps uint64 `json:"reps"`
	// The total number of times the card has been forgotten (rated "Again").
	Lapses uint64 `json:"lapses"`
	// The current state of the card (New, Learning, Review, ReLearning).
	State fsrs.State `json:"state"`
	// The timestamp of the last review.
	LastReview time.Time `json:"last_review"`
}

func (c *Card) ToFsrsCard() fsrs.Card {
	return fsrs.Card{
		Due:           c.Due,
		Stability:     c.Stability,
		Difficulty:    c.Difficulty,
		ElapsedDays:   c.ElapsedDays,
		ScheduledDays: c.ScheduledDays,
		Reps:          c.Reps,
		Lapses:        c.Lapses,
		State:         c.State,
		LastReview:    c.LastReview,
	}
}

func (c *Card) SyncFromFSRSCard(from fsrs.Card) {
	c.Difficulty = from.Difficulty
	c.Due = from.Due
	c.Stability = from.Stability
	c.Difficulty = from.Difficulty
	c.ElapsedDays = from.ElapsedDays
	c.ScheduledDays = from.ScheduledDays
	c.Reps = from.Reps
	c.Lapses = from.Lapses
	c.State = from.State
	c.LastReview = from.LastReview
}

type Review struct {
	CardID int         `json:"cardID"`
	Rate   fsrs.Rating `json:"rate"`
}
