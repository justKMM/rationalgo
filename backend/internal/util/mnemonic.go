package util

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
)

//go:embed algorand_wordlist.txt
var algorandWordlistRaw string

var algorandWordlist = strings.Split(strings.TrimSpace(algorandWordlistRaw), "\n")

// NormalizeMnemonic accepts a 24- or 25-word Algorand passphrase.
// Pera often displays 24 words (key material only); the 25th checksum word is derived automatically.
func NormalizeMnemonic(raw string) (string, error) {
	words := strings.Fields(strings.TrimSpace(raw))
	switch len(words) {
	case 25:
		if _, err := mnemonic.ToKey(strings.Join(words, " ")); err != nil {
			return "", fmt.Errorf("mnemonic: invalid 25-word passphrase: %w", err)
		}
		return strings.Join(words, " "), nil
	case 24:
		return complete24WordMnemonic(words)
	default:
		return "", fmt.Errorf("mnemonic must be 24 or 25 Algorand words (got %d)", len(words))
	}
}

// complete24WordMnemonic finds the checksum word Algorand appends as word 25.
func complete24WordMnemonic(words []string) (string, error) {
	for _, w := range words {
		if wordIndex(w) == -1 {
			return "", fmt.Errorf("%q is not in the Algorand word list", w)
		}
	}

	base := strings.Join(words, " ")
	for _, checksumWord := range algorandWordlist {
		candidate := base + " " + checksumWord
		if _, err := mnemonic.ToKey(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("24-word passphrase is not a valid Algorand recovery phrase")
}

func wordIndex(word string) int {
	for i, w := range algorandWordlist {
		if w == word {
			return i
		}
	}
	return -1
}
