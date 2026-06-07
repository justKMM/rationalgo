package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"rationalgo/internal/util"
)

// defaultSettlementAssetID is the testnet USDC ASA used for x402 settlement
// (6 decimals; see README — mainnet USDC is 31566704).
const defaultSettlementAssetID = 10458941

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

	// SettlementAssetID is the ASA the x402 seller charges in (testnet USDC by default).
	SettlementAssetID uint64
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

		SettlementAssetID: envOrUint("RATIONALGO_SETTLEMENT_ASSET_ID", defaultSettlementAssetID),
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
		return fmt.Errorf("set RATIONALGO_MNEMONIC in backend/.env (24-word passphrase from Pera)")
	}
	if err := util.ValidateWalletMnemonic(c.Mnemonic, c.WalletAddress); err != nil {
		return fmt.Errorf("RATIONALGO_MNEMONIC: %w", err)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envOrUint(key string, fallback uint64) uint64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

// PublicBaseURL returns the local base URL the API server is reachable at — used to
// build self-referential x402 seller endpoint URLs for the catalog (e.g. "http://localhost:8080").
func (c Config) PublicBaseURL() string {
	addr := strings.TrimSpace(c.HTTPAddr)
	if addr == "" {
		addr = ":8080"
	}
	if strings.Contains(addr, "://") {
		return addr
	}
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	return "http://" + addr
}
