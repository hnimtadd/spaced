package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/hnimtadd/spaced/src/core/model"
)

type data [4]string

func main() {
	fileBytes, err := os.ReadFile("./src/assets/data.txt")
	if err != nil {
		panic("failed to read notes file: " + err.Error())
	}

	lines := strings.Split(string(fileBytes), "\n")
	fmt.Println(len(lines))

	datas := make([]data, len(lines))

	for cardIdx, line := range lines {
		parts := strings.Split(line, "\t")
		item := data{}
		idx := 0
		for part := range slices.Values(parts) {
			if part != "" {
				item[idx] = part
				idx++
			}
		}
		datas[cardIdx] = item
	}

	cards := make([]model.Card, len(datas))
	for idx, sample := range datas {
		cards[idx] = model.Card{
			Word:       sample[0],
			Definition: sample[1],
			Example:    sample[2],
			IPA:        sample[3],
		}
	}
	jsonBytes, err := json.Marshal(cards)
	if err != nil {
		panic("failed to marshal cards data: " + err.Error())
	}
	if err := os.WriteFile("./src/assets/cards.json", jsonBytes, os.ModePerm); err != nil {
		panic("failed to write marshalled cards: " + err.Error())
	}

	jsonBytes, err = json.Marshal(cards[0:10])
	if err != nil {
		panic("failed to marshal sample cards data: " + err.Error())
	}
	if err := os.WriteFile("./src/assets/cards_sample.json", jsonBytes, os.ModePerm); err != nil {
		panic("failed to write marshalled sample cards: " + err.Error())
	}
	fmt.Println("🚀 complete")
}
