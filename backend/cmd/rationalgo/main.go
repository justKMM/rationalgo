package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"rationalgo/internal/api"
	"rationalgo/internal/config"
	algosvc "rationalgo/internal/services/algorand"
	"rationalgo/internal/services/reasoning"
	x402svc "rationalgo/internal/services/x402"
	"rationalgo/internal/util"
)

func main() {
	config.LoadEnv()

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}

	if len(os.Args) < 2 {
		printStatus(cfg)
		return
	}

	switch os.Args[1] {
	case "status":
		printStatus(cfg)
	case "serve":
		runServe(cfg)
	case "spike":
		runSpike(cfg, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func runServe(cfg config.Config) {
	fmt.Println("RationAlgo — Phase 2 API")
	fmt.Printf("listening:   %s\n", cfg.HTTPAddr)
	fmt.Println("endpoints:   GET /health  GET /api/state  GET /api/decisions")
	fmt.Println("             POST /api/state/reset  POST /api/scenario/run  POST /api/decide")
	fmt.Println("             GET /pricing  GET /company/* (x402-protected research marketplace)")
	fmt.Println()
	reasoningSvc := reasoning.New(cfg.AnthropicKey)
	if err := api.NewServer(cfg, reasoningSvc).ListenAndServe(); err != nil {
		fail(err)
	}
}

func runSpike(cfg config.Config, args []string) {
	target := "all"
	if len(args) > 0 {
		target = args[0]
	}

	switch target {
	case "algorand":
		spikeAlgorand(cfg)
	case "provenance":
		spikeProvenance(cfg)
	case "x402":
		spikeX402(cfg, args[1:])
	case "all":
		spikeAlgorand(cfg)
		fmt.Println()
		spikeProvenance(cfg)
		fmt.Println()
		spikeX402(cfg, nil)
	default:
		fail(fmt.Errorf("unknown spike target %q (use algorand, provenance, x402, or all)", target))
	}
}

func spikeProvenance(cfg config.Config) {
	fmt.Println("=== Algorand spike (RAv1 provenance) ===")
	svc, err := algosvc.NewService(cfg)
	if err != nil {
		fail(err)
	}
	result, err := svc.RunProvenanceSpike()
	if err != nil {
		fail(err)
	}
	fmt.Printf("wallet:         %s\n", result.Address)
	fmt.Printf("balance:        %d microAlgos\n", result.MicroAlgos)
	fmt.Printf("decision_hash:  %s\n", result.ReasoningHash)
	fmt.Printf("tx_id:          %s\n", result.TxID)
	fmt.Printf("explorer:       %s\n", result.ExplorerURL)
	fmt.Println("ok: RAv1 provenance commitment confirmed")
}

func spikeAlgorand(cfg config.Config) {
	fmt.Println("=== Algorand spike (hash commitment) ===")
	svc, err := algosvc.NewService(cfg)
	if err != nil {
		fail(err)
	}
	result, err := svc.RunSpike()
	if err != nil {
		fail(err)
	}
	fmt.Printf("wallet:         %s\n", result.Address)
	fmt.Printf("balance:        %d microAlgos\n", result.MicroAlgos)
	fmt.Printf("reasoning_hash: %s\n", result.ReasoningHash)
	fmt.Printf("tx_id:          %s\n", result.TxID)
	fmt.Printf("explorer:       %s\n", result.ExplorerURL)
	fmt.Println("ok: testnet commitment confirmed")
}

func spikeX402(cfg config.Config, args []string) {
	if len(args) > 0 && args[0] == "pay" {
		spikeX402Pay(cfg, cfg.X402ProbeURL, "GoPlausible")
		return
	}
	if len(args) > 0 && args[0] == "pay-local" {
		url := cfg.PublicBaseURL() + "/company/basic-info?company=" + url.QueryEscape(reasoning.DemoCompany)
		spikeX402Pay(cfg, url, "local /company/basic-info")
		return
	}
	fmt.Println("=== x402 spike (402 probe) ===")
	result, err := x402svc.NewService(cfg).RunProbe()
	if err != nil {
		fail(err)
	}
	fmt.Printf("url:         %s\n", result.URL)
	fmt.Printf("status:      %d\n", result.StatusCode)
	if result.PaymentRequired {
		fmt.Println("payment:     HTTP 402 Payment Required (expected)")
	} else {
		fmt.Println("payment:     no 402 — endpoint may have changed; check URL")
	}
	if result.PaymentHeader != "" {
		fmt.Printf("header:      PAYMENT-REQUIRED present (%d chars)\n", len(result.PaymentHeader))
	}
	if result.BodySnippet != "" {
		fmt.Printf("body:        %s\n", result.BodySnippet)
	}
	fmt.Println("ok: x402 probe complete (use spike x402 pay for real payment)")
}

func spikeX402Pay(cfg config.Config, targetURL, label string) {
	fmt.Printf("=== x402 spike (real payment — %s) ===\n", label)
	svc := x402svc.NewService(cfg)
	body, err := svc.PayAndFetch(context.Background(), targetURL, 0.001)
	if err != nil {
		fail(err)
	}
	fmt.Printf("url:           %s\n", targetURL)
	fmt.Printf("body:          %s\n", string(body))
	if tx := svc.LastSettlementTx(); tx != "" {
		fmt.Printf("settlement_tx: %s\n", tx)
	}
	fmt.Println("ok: x402 payment settled — resource returned")
}

func printStatus(cfg config.Config) {
	fmt.Println("RationAlgo — Phase 2")
	fmt.Println()
	fmt.Printf("wallet:      %s\n", displayWallet(cfg))
	fmt.Printf("algod:       %s\n", cfg.AlgodURL)
	fmt.Printf("algod token: %s\n", displayToken(cfg.AlgodToken))
	fmt.Printf("x402 probe:  %s\n", cfg.X402ProbeURL)
	fmt.Printf("http addr:   %s (go run ./cmd/rationalgo serve)\n", cfg.HTTPAddr)
	fmt.Println()
	if err := cfg.ValidateForSpike(); err != nil {
		fmt.Printf("spike ready: no — %v\n", err)
		fmt.Println()
		printUsage()
		return
	}
	fmt.Printf("account:     %s\n", util.AccountURL(cfg.WalletAddress))
	if err := cfg.ValidateForSeller(); err != nil {
		fmt.Printf("x402 seller: no — %v\n", err)
	} else {
		fmt.Printf("x402 seller: %s\n", cfg.SellerWalletAddressEffective())
	}
	fmt.Println("spike ready: yes")
	fmt.Println()
	printUsage()
}

func displayWallet(cfg config.Config) string {
	if !cfg.WalletConfigured() {
		return "(not set — paste RATIONALGO_WALLET_ADDRESS in backend/.env)"
	}
	return cfg.WalletAddress
}

func displayToken(token string) string {
	if token == "" {
		return "(empty — OK for public AlgoNode)"
	}
	return "(set)"
}

func printUsage() {
	fmt.Println(`Usage:
  go run ./cmd/rationalgo                 # config status
  go run ./cmd/rationalgo serve           # HTTP API for dashboard (Phase 1)
  go run ./cmd/rationalgo spike all           # hash + RAv1 + x402 probe
  go run ./cmd/rationalgo spike algorand      # legacy hash commitment
  go run ./cmd/rationalgo spike provenance    # RAv1 envelope commitment
  go run ./cmd/rationalgo spike x402          # unpaid 402 probe (GoPlausible)
  go run ./cmd/rationalgo spike x402 pay      # real payment + fetch (GoPlausible)
  go run ./cmd/rationalgo spike x402 pay-local # real payment against local /company/* (serve must be running)

Setup:
  cp .env.example .env
  # edit backend/.env — wallet address, algod token, mnemonic`)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
