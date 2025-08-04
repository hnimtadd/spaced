package main

import (
	"fmt"
	"strings"
)

var ipaToCMU = map[string]string{
	"f":  "F",
	"m":  "M",
	"n":  "N",
	"k":  "K",
	"l":  "L",
	"d":  "D",
	"b":  "B",
	"h":  "HH",
	"p":  "P",
	"t":  "T",
	"s":  "S",
	"r":  "R",
	"æ":  "AE",
	"w":  "W",
	"z":  "Z",
	"v":  "V",
	"ɡ":  "G",
	"ŋ":  "NG",
	"ð":  "DH",
	"ə":  "AX",
	"ɑː": "AA",
	"ʌ":  "AH",
	"ɔr": "AO",
	"aʊ": "AW",
	"ɜr": "AXR",
	"aɪ": "AY",
	"tʃ": "CH",
	"ɛ":  "EH",
	"ʌr": "ER",
	"eɪ": "EY",
	"ɪ":  "IH",
	"i":  "IX",
	"iː": "IY",
	"dʒ": "JH",
	"oʊ": "OW",
	"ɔɪ": "OY",
	"ʃ":  "SH",
	"θ":  "TH",
	"ʊ":  "UH",
	"uː": "UW",
	"j":  "Y",
	"ɹ":  "R",
	"ɚ":  "R",
	"ɐ":  "AH",
	"ɒ":  "AA",
	" ":  "SIL",
	"ɫ":  "L",
	"ɬ":  "L",
	"ɾ":  "R",
	"ᵻ":  "IH",
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
