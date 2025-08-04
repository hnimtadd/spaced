package main

import (
	"fmt"
	"strings"
)

var ipaToCMU = map[string]string{
	"ɑ":  "AA",
	"æ":  "AE",
	"ə":  "AH0",
	"ʌ":  "AH",
	"ɔ":  "AO",
	"aʊ": "AW",
	"aɪ": "AY",
	"ɛ":  "EH",
	"ɝ":  "ER",
	"ɚ":  "ER",
	"eɪ": "EY",
	"ɪ":  "IH",
	"i":  "IY",
	"oʊ": "OW",
	"ɔɪ": "OY",
	"ʊ":  "UH",
	"u":  "UW",
	"b":  "B",
	"tʃ": "CH",
	"d":  "D",
	"ð":  "DH",
	"f":  "F",
	"ɡ":  "G",
	"h":  "HH",
	"dʒ": "JH",
	"k":  "K",
	"l":  "L",
	"m":  "M",
	"n":  "N",
	"ŋ":  "NG",
	"p":  "P",
	"r":  "R",
	"s":  "S",
	"ʃ":  "SH",
	"t":  "T",
	"θ":  "TH",
	"v":  "V",
	"w":  "W",
	"j":  "Y",
	"z":  "Z",
	"ʒ":  "ZH",
}

func IPAToCMU(ipa string) string {
	ipaChars := strings.Split(ipa, "")
	cmdChars := make([]string, len(ipaChars))
	for idx, ipaChar := range ipaChars {
		cmdChars[idx] = ipaToCMU[ipaChar]
	}
	return strings.Join(cmdChars, " ")
}

func TestIPAToCMUConvertion() {
	tcs := []struct {
		name string
		ipa  string
		cmu  string
	}{
		{name: "agree", ipa: "əˈɡri", cmu: "AH0 G R IY1"},
		{name: "nginx", ipa: "ˈɛndʒɪnˈɛks", cmu: "EH1 N JH IH0 N EH1 K S"},
		{name: "tomato", ipa: "təˈmeɪtoʊ", cmu: "T AH0 M EY1 T OW0"},
		{name: "apple", ipa: "ˈæpl̩", cmu: "AE1 P AH0 L"},
		{name: "actually", ipa: "ˈæktʃuəli", cmu: "AE1 K CH UH0 AH0 L IY0"},
		{name: "Madison", ipa: "ˈmædɪsən", cmu: "M AE1 D IH0 S AH0 N"},
		{name: "pronunciation", ipa: "prəˌnʌnsiˈeɪʃən", cmu: "P R AH0 N AH0 N S IY EY1 SH AH0 N"},
		{name: "agree", ipa: "əˈɡriː", cmu: "AH0 G R IY1"},
		{name: "water", ipa: "ˈwɔːtər", cmu: "W AO1 T ER0"},
		{name: "computer", ipa: "kəmˈpjuːtər", cmu: "K AH0 M P Y UW1 T ER0"},
		{name: "beautiful", ipa: "ˈbjuːtɪfəl", cmu: "B Y UW1 T AH0 F AH0 L"},
	}

	for _, tc := range tcs {
		cmu := IPAToCMU(tc.ipa)
		if cmu != tc.cmu {
			fmt.Println("expeteced", tc.cmu, "got", cmu)
			panic(tc)
		}
	}
}

func main() {
	TestIPAToCMUConvertion()
}
