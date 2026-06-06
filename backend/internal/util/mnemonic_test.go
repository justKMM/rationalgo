package util

import (
	"strings"
	"testing"

	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
)

func TestNormalizeMnemonic25Words(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	full, err := mnemonic.FromKey(seed)
	if err != nil {
		t.Fatal(err)
	}

	got, err := NormalizeMnemonic(full)
	if err != nil {
		t.Fatal(err)
	}
	if got != full {
		t.Fatal("expected unchanged 25-word mnemonic")
	}
}

func TestNormalizeMnemonic24WordsAddsChecksum(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	full, err := mnemonic.FromKey(seed)
	if err != nil {
		t.Fatal(err)
	}
	words := strings.Fields(full)
	twentyFour := strings.Join(words[:24], " ")

	got, err := NormalizeMnemonic(twentyFour)
	if err != nil {
		t.Fatal(err)
	}
	if got != full {
		t.Fatalf("24-word input should expand to full mnemonic\ngot:  %s\nwant: %s", got, full)
	}
}

func TestNormalizeMnemonicInvalidCount(t *testing.T) {
	if _, err := NormalizeMnemonic("one two three"); err == nil {
		t.Fatal("expected error for short mnemonic")
	}
}
