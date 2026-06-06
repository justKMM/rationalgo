package x402

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"rationalgo/internal/config"
	"rationalgo/internal/models"
)

// Service probes x402-protected HTTP resources.
type Service struct {
	cfg        config.Config
	httpClient *http.Client
}

// NewService creates an x402 probe service.
func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

// RunProbe issues an unpaid GET and reports whether the server returns HTTP 402.
func (s *Service) RunProbe() (models.X402ProbeResult, error) {
	url := s.cfg.X402ProbeURL
	if url == "" {
		return models.X402ProbeResult{}, fmt.Errorf("RATIONALGO_X402_PROBE_URL is empty")
	}

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return models.X402ProbeResult{}, fmt.Errorf("probe GET: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil {
		return models.X402ProbeResult{}, fmt.Errorf("read body: %w", err)
	}

	paymentHeader := resp.Header.Get("PAYMENT-REQUIRED")
	if paymentHeader == "" {
		paymentHeader = resp.Header.Get("X-PAYMENT-REQUIRED")
	}

	return models.X402ProbeResult{
		URL:             url,
		StatusCode:      resp.StatusCode,
		PaymentRequired: resp.StatusCode == http.StatusPaymentRequired,
		PaymentHeader:   paymentHeader,
		BodySnippet:     strings.TrimSpace(string(body)),
	}, nil
}

// PayAndFetch probes 402 then returns demo payload (real EURQ payment in Phase 2).
func (s *Service) PayAndFetch(ctx context.Context, url string, amountEURQ float64) ([]byte, error) {
	_ = ctx
	if _, err := s.RunProbe(); err != nil {
		return nil, fmt.Errorf("x402: pay and fetch: probe: %w", err)
	}
	_ = url
	_ = amountEURQ
	return []byte(`{"precipitation_pct":14.0,"source":"demo-stub-until-phase-2"}`), nil
}
