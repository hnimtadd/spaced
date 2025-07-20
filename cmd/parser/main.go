package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type data [4]string

func main() {
	fileBytes, err := os.ReadFile("./assets/notes.txt")
	if err != nil {
		panic("failed to read notes file: " + err.Error())
	}

	lines := strings.Split(string(fileBytes), "\n")
	fmt.Println(len(lines))

	cards := make([]data, len(lines))

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
		cards[cardIdx] = item
	}

	for sample := range slices.Values(cards[0:10]) {
		fmt.Println("word: ", sample[0])
		fmt.Println("definition: ", sample[1])
		fmt.Println("example: ", sample[2])
		fmt.Println("ipa: ", sample[3])
		fmt.Println("========================")
	}
}
