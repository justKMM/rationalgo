package x402

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rationalgo/internal/config"
	"rationalgo/internal/models"
	"rationalgo/internal/services/algorand"
)

// Service probes and pays x402-protected HTTP resources via GoPlausible.
type Service struct {
	cfg              config.Config
	httpClient       *http.Client
	algClient        *algorand.Client
	lastSettlementTx string
}

// NewService creates an x402 client service. Pass algClient when the server already
// holds a wallet-connected client (hero demo); otherwise PayAndFetch builds one on demand.
func NewService(cfg config.Config, algClient ...*algorand.Client) *Service {
	s := &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.X402HTTPTimeout(),
		},
	}
	if len(algClient) > 0 {
		s.algClient = algClient[0]
	}
	return s
}

// LastSettlementTx returns the on-chain tx id from the most recent PayAndFetch.
func (s *Service) LastSettlementTx() string {
	return s.lastSettlementTx
}

// RunProbe issues an unpaid GET and reports whether the server returns HTTP 402.
func (s *Service) RunProbe() (models.X402ProbeResult, error) {
	return s.probeURL(s.cfg.X402ProbeURL)
}

func (s *Service) probeURL(url string) (models.X402ProbeResult, error) {
	if url == "" {
		return models.X402ProbeResult{}, fmt.Errorf("x402 url is empty")
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

	paymentHeader := paymentRequiredHeader(resp.Header)

	return models.X402ProbeResult{
		URL:             url,
		StatusCode:      resp.StatusCode,
		PaymentRequired: resp.StatusCode == http.StatusPaymentRequired,
		PaymentHeader:   paymentHeader,
		BodySnippet:     strings.TrimSpace(string(body)),
	}, nil
}

// PayAndFetch completes the x402 v2 flow: 402 → sign ASA transfer → retry with PAYMENT-SIGNATURE.
// Requires a funded wallet with the requested ASA (testnet USDC/EURQ ASA 10458941 by default).
func (s *Service) PayAndFetch(ctx context.Context, url string, amountEURQ float64) ([]byte, error) {
	_ = amountEURQ // price comes from PAYMENT-REQUIRED.accepts

	if url == "" {
		url = s.cfg.X402ProbeURL
	}
	if url == "" {
		return nil, fmt.Errorf("x402: url is empty")
	}

	algClient, err := s.algorandClient()
	if err != nil {
		return nil, fmt.Errorf("x402: wallet required for payment: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("x402: build request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("x402: initial GET: %w", err)
	}
	paymentHeader := paymentRequiredHeader(resp.Header)
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body.Close()

	if resp.StatusCode != http.StatusPaymentRequired {
		if resp.StatusCode == http.StatusOK {
			return body, nil
		}
		return nil, fmt.Errorf("x402: expected 402, got %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	required, err := decodePaymentRequired(paymentHeader)
	if err != nil {
		return nil, fmt.Errorf("x402: %w", err)
	}

	network, err := algClient.NetworkIdentifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("x402: network: %w", err)
	}

	accept, err := selectAlgorandAccept(required.Accepts, network)
	if err != nil {
		return nil, fmt.Errorf("x402: %w", err)
	}

	paymentSig, err := buildAlgorandPaymentSignature(ctx, algClient, required, accept)
	if err != nil {
		return nil, fmt.Errorf("x402: build payment: %w", err)
	}

	paidReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("x402: build paid request: %w", err)
	}
	paidReq.Header.Set(headerPaymentSignature, paymentSig)
	paidReq.Header.Set(headerLegacyPayment, paymentSig)

	paidResp, err := s.httpClient.Do(paidReq)
	if err != nil {
		return nil, fmt.Errorf("x402: paid GET: %w", err)
	}
	defer paidResp.Body.Close()

	paidBody, err := io.ReadAll(paidResp.Body)
	if err != nil {
		return nil, fmt.Errorf("x402: read paid body: %w", err)
	}

	if paidResp.StatusCode != http.StatusOK {
		reason := paymentRejectionReason(paidResp.Header, paidBody)
		return nil, fmt.Errorf("x402: payment failed with %d: %s", paidResp.StatusCode, reason)
	}

	if hdr := paidResp.Header.Get(headerPaymentResponse); hdr != "" {
		if settlement, err := decodePaymentResponse(hdr); err == nil && settlement.Transaction != "" {
			s.lastSettlementTx = settlement.Transaction
		}
	}

	return paidBody, nil
}

func paymentRequiredHeader(h http.Header) string {
	if v := h.Get(headerPaymentRequired); v != "" {
		return v
	}
	return h.Get("X-PAYMENT-REQUIRED")
}

func paymentRejectionReason(h http.Header, body []byte) string {
	if hdr := paymentRequiredHeader(h); hdr != "" {
		if req, err := decodePaymentRequired(hdr); err == nil && req.Error != "" {
			return humanizeSettlementError(req.Error)
		}
	}
	if trimmed := strings.TrimSpace(string(body)); trimmed != "" {
		return humanizeSettlementError(trimmed)
	}
	return "payment rejected"
}

func (s *Service) algorandClient() (*algorand.Client, error) {
	if s.algClient != nil {
		return s.algClient, nil
	}
	client, err := algorand.NewClient(s.cfg)
	if err != nil {
		return nil, err
	}
	s.algClient = client
	return client, nil
}
