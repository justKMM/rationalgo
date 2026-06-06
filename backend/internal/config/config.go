package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// WalletAddressPlaceholder is the default sentinel in .env before you paste a real address.
const WalletAddressPlaceholder = "PASTE_YOUR_PERA_WALLET_ADDRESS_HERE"

// Config holds runtime settings loaded from the environment.
type Config struct {
	WalletAddress string
	AlgodToken    string
	Mnemonic      string
	AlgodURL      string
	IndexerURL    string
	IndexerToken  string
	X402ProbeURL  string
	HTTPAddr      string
	AnthropicKey  string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		WalletAddress: strings.TrimSpace(os.Getenv("RATIONALGO_WALLET_ADDRESS")),
		AlgodToken:    strings.TrimSpace(os.Getenv("RATIONALGO_ALGOD_TOKEN")),
		Mnemonic:      strings.TrimSpace(os.Getenv("RATIONALGO_MNEMONIC")),
		AlgodURL:      envOr("RATIONALGO_ALGOD_URL", "https://testnet-api.algonode.cloud"),
		IndexerURL:    envOr("RATIONALGO_INDEXER_URL", "https://testnet-idx.algonode.cloud"),
		IndexerToken:  strings.TrimSpace(os.Getenv("RATIONALGO_INDEXER_TOKEN")),
		X402ProbeURL:  envOr("RATIONALGO_X402_PROBE_URL", "https://example.x402.goplausible.xyz/avm/weather"),
		HTTPAddr:      envOr("RATIONALGO_HTTP_ADDR", ":8080"),
		AnthropicKey:  strings.TrimSpace(os.Getenv("RATIONALGO_ANTHROPIC_KEY")),
	}
	// Do NOT fatal here — spike commands don't need it.
	// The reasoning service will fail loudly at call time if key is empty.
	if cfg.AnthropicKey == "" {
		log.Println("warning: RATIONALGO_ANTHROPIC_KEY not set; reasoning unavailable")
	}
	return cfg, nil
}

// WalletConfigured reports whether a real wallet address is set.
func (c Config) WalletConfigured() bool {
	return c.WalletAddress != "" && c.WalletAddress != WalletAddressPlaceholder
}

// ValidateForSpike checks fields required for Phase 0 integration spikes.
func (c Config) ValidateForSpike() error {
	if !c.WalletConfigured() {
		return fmt.Errorf("set RATIONALGO_WALLET_ADDRESS in backend/.env (your Pera Testnet address)")
	}
	if c.Mnemonic == "" {
		return fmt.Errorf("set RATIONALGO_MNEMONIC in backend/.env (24- or 25-word passphrase from Pera)")
	}
	words := len(strings.Fields(c.Mnemonic))
	if words != 24 && words != 25 {
		return fmt.Errorf("RATIONALGO_MNEMONIC must be 24 or 25 Algorand words (got %d) — copy the passphrase from Pera → Settings → Security", words)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
