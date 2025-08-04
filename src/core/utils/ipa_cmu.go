//go:build js && wasm

package utils

import (
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
