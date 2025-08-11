//go:build js && wasm

package session

import (
	"html/template"
	"slices"
	"strings"
	"time"

	"github.com/hnimtadd/spaced/src/core/fsrs"
)

type Session struct {
	Cards fsrs.Cards

	AgainsID map[int]bool

	Looked map[int]bool

	StartedAt time.Time
}

func NewSession(cards fsrs.Cards) *Session {
	return &Session{
		Cards:     cards,
		AgainsID:  map[int]bool{},
		Looked:    map[int]bool{},
		StartedAt: time.Now(),
	}
}

func (s Session) ShouldStop() bool {
	if len(s.AgainsID) > 0 {
		return false
	}
	// gurantee every cards need to be take a looked at least 1 time.
	if len(s.Looked) != len(s.Cards) {
		return false
	}
	for card := range slices.Values(s.Cards) {
		if card.Due.Before(time.Now()) {
			return false
		}
		if card.LastReview.IsZero() {
			return false
		}
	}
	return true
}

type Record struct {
	ID          int       `json:"id"`
	Cards       []int     `json:"cardIDs"`
	StartedAt   time.Time `json:"staredAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func NewRecordFromSession(session *Session) Record {
	ids := make([]int, len(session.Cards))
	for i, card := range session.Cards {
		ids[i] = card.ID
	}

	return Record{
		Cards:       ids,
		StartedAt:   session.StartedAt,
		CompletedAt: time.Now(),
	}
}

var sessionTpml = `
	<div class="bg-white p-4 rounded-lg shadow-md border border-gray-200 space-y-2 cursor-pointer" craft-name="replay" id="stat-{{.ID}}" data="{{.ID}}" craft-input="#stat-{{.ID}}:[data]" craft-trigger="click">
    <div class="flex items-center justify-between">
        <span class="font-bold text-gray-700">ID:</span>
        <span class="font-bold text-gray-700">{{.ID}}</span>
    </div>
    <div class="flex items-center justify-between">
        <span class="font-bold text-gray-700">Status:</span>
        <span class="px-2 py-1 text-xs font-semibold text-white bg-green-500 rounded-full">Completed</span>
    </div>
    <div class="flex items-center justify-between">
        <span class="font-bold text-gray-700">Card reviewed:</span>
        <span class="font-normal text-gray-600">{{.CardReviewed}}</span>
    </div>
    <div class="flex items-center justify-between">
        <span class="font-bold text-gray-700">Duration:</span>
        <span class="font-normal text-gray-600">{{.Duration}}</span>
    </div>
    <div class="flex items-center justify-between">
        <span class="font-bold text-gray-700">Date:</span>
        <span class="font-normal text-gray-600">{{.Date}}</span>
    </div>
</div>
`

func (s Record) ToHTML() template.HTML {
	tmpl, err := template.New("session").Parse(sessionTpml)
	if err != nil {
		return template.HTML("failed to parse to HTML" + err.Error())
	}
	buf := &strings.Builder{}
	data := struct {
		Date         string
		Duration     string
		CardReviewed int
		ID           int
	}{
		ID:           s.ID,
		Date:         s.StartedAt.Format(time.DateTime),
		Duration:     s.CompletedAt.Sub(s.StartedAt).Round(time.Second).String(),
		CardReviewed: len(s.Cards),
	}
	if err := tmpl.Execute(buf, data); err != nil {
		return template.HTML("failed to execute template" + err.Error())
	}
	return template.HTML(buf.String())
}
