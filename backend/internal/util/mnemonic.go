package util

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/tyler-smith/go-bip39"
)

//go:embed algorand_wordlist.txt
var algorandWordlistRaw string

var algorandWordlist = func() []string {
	lines := strings.Split(strings.TrimSpace(algorandWordlistRaw), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		w := strings.TrimSpace(line)
		if w != "" {
			out = append(out, w)
		}
	}
	return out
}()

// NormalizeMnemonic accepts a 24- or 25-word Algorand passphrase.
// Pera displays 24 words; the 25th checksum word is derived automatically.
// If you already have 25 words, pass them as-is.
func NormalizeMnemonic(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'`)
	words := strings.Fields(raw)
	for i := range words {
		words[i] = strings.ToLower(strings.Trim(words[i], `"'`))
	}

	switch len(words) {
	case 25:
		phrase := strings.Join(words, " ")
		if _, err := mnemonic.ToKey(phrase); err != nil {
			return "", fmt.Errorf("mnemonic: invalid 25-word passphrase: %w", err)
		}
		return phrase, nil
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

	return "", invalid24WordMnemonicError(strings.Join(words, " "))
}

func invalid24WordMnemonicError(phrase string) error {
	if bip39.IsMnemonicValid(phrase) {
		return fmt.Errorf(
			"valid BIP39 seed phrase, but it did not derive your RATIONALGO_WALLET_ADDRESS — " +
				"confirm the wallet address and passphrase are from the same Pera account",
		)
	}
	return fmt.Errorf(
		"24-word passphrase is not a valid Algorand recovery phrase — " +
			"the 25th checksum word is derived automatically from your 24 words, but these did not validate. " +
			"Re-copy from Pera → Settings → Security and check for typos or wrong word order",
	)
}

// AlgorandWordlist is the 2048-word Algorand passphrase dictionary.
func AlgorandWordlist() []string {
	return algorandWordlist
}

func wordIndex(word string) int {
	for i, w := range algorandWordlist {
		if w == word {
			return i
		}
	}
	return -1
}
