package x402

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"

	"rationalgo/internal/services/algorand"
)

// PriceInfo describes what a protected resource costs, independent of any pricing registry.
type PriceInfo struct {
	ResourcePath    string // e.g. "/company/basic-info" — used in PAYMENT-REQUIRED.resource.url
	Description     string // used in PAYMENT-REQUIRED.resource.description
	AmountBaseUnits uint64 // required ASA transfer amount in base units (e.g. testnet USDC has 6 decimals: $0.01 == 10000 base units)
}

// Seller protects HTTP handlers behind an on-chain Algorand ASA payment requirement.
type Seller struct {
	client  *algorand.Client
	assetID uint64
	network string
	mu      sync.Mutex
	seen    map[string]bool
}

// NewSeller builds a Seller, opting the wallet into the ASA so it can receive payments.
func NewSeller(ctx context.Context, client *algorand.Client, assetID uint64) (*Seller, error) {
	network, err := client.NetworkIdentifier(ctx)
	if err != nil {
		return nil, fmt.Errorf("x402: seller: network identifier: %w", err)
	}
	if err := client.EnsureAssetOptIn(ctx, assetID); err != nil {
		return nil, fmt.Errorf("x402: seller: asa opt-in: %w", err)
	}
	return &Seller{
		client:  client,
		assetID: assetID,
		network: network,
		seen:    make(map[string]bool),
	}, nil
}

// Protect wraps next with an x402 paywall that requires a verified on-chain ASA payment.
func (s *Seller) Protect(meta PriceInfo, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(headerPaymentSignature)
		if header == "" {
			header = r.Header.Get(headerLegacyPayment)
		}
		if header == "" {
			s.respondPaymentRequired(w, meta, "")
			return
		}

		sig, err := s.decodePaymentSignature(header)
		if err != nil {
			s.respondPaymentRequired(w, meta, fmt.Sprintf("invalid payment: %v", err))
			return
		}

		raw, err := s.extractSignedTxn(sig)
		if err != nil {
			s.respondPaymentRequired(w, meta, fmt.Sprintf("invalid payment: %v", err))
			return
		}

		var stxn types.SignedTxn
		if err := msgpack.Decode(raw, &stxn); err != nil {
			s.respondPaymentRequired(w, meta, fmt.Sprintf("decode signed txn: %v", err))
			return
		}

		txID := crypto.GetTxID(stxn.Txn)
		if err := s.verify(stxn, meta, txID); err != nil {
			s.respondPaymentRequired(w, meta, err.Error())
			return
		}

		if _, err := s.client.SubmitSignedTxn(r.Context(), raw); err != nil {
			s.respondPaymentRequired(w, meta, fmt.Sprintf("settlement failed: %v", err))
			return
		}

		s.markSeen(txID)

		resp := paymentResponse{
			Success:     true,
			Payer:       stxn.Txn.Sender.String(),
			Transaction: txID,
			Network:     s.network,
		}
		if encoded, err := encodePaymentResponse(resp); err == nil {
			w.Header().Set(headerPaymentResponse, encoded)
		}

		next(w, r)
	}
}

// decodePaymentSignature base64-decodes and JSON-unmarshals a PAYMENT-SIGNATURE header value.
func (s *Seller) decodePaymentSignature(header string) (*paymentSignature, error) {
	raw, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	var sig paymentSignature
	if err := json.Unmarshal(raw, &sig); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	return &sig, nil
}

// extractSignedTxn returns the msgpack-encoded signed txn bytes for the simple single-txn,
// no-fee-payer case this seller supports.
func (s *Seller) extractSignedTxn(sig *paymentSignature) ([]byte, error) {
	group := sig.Payload.PaymentGroup
	if len(group) != 1 || sig.Payload.PaymentIndex != 0 {
		return nil, fmt.Errorf("unsupported payment group shape (len=%d, index=%d)", len(group), sig.Payload.PaymentIndex)
	}
	raw, err := base64.StdEncoding.DecodeString(group[0])
	if err != nil {
		return nil, fmt.Errorf("base64 decode payment group entry: %w", err)
	}
	return raw, nil
}

// verify checks that a decoded signed txn is a correctly addressed, signed, unseen ASA payment.
func (s *Seller) verify(stxn types.SignedTxn, meta PriceInfo, txID string) error {
	txn := stxn.Txn
	if txn.Type != types.AssetTransferTx {
		return fmt.Errorf("not an asset transfer transaction")
	}
	if txn.AssetTransferTxnFields.XferAsset != types.AssetIndex(s.assetID) {
		return fmt.Errorf("wrong asset id")
	}
	if txn.AssetTransferTxnFields.AssetReceiver.String() != s.client.Address() {
		return fmt.Errorf("wrong asset receiver")
	}
	if txn.AssetTransferTxnFields.AssetAmount != meta.AmountBaseUnits {
		return fmt.Errorf("wrong asset amount")
	}
	if len(stxn.Sig) == 0 && stxn.Msig.Blank() && len(stxn.Lsig.Logic) == 0 {
		return fmt.Errorf("missing signature")
	}
	if s.isSeen(txID) {
		return fmt.Errorf("payment already settled")
	}
	return nil
}

func (s *Seller) isSeen(txID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen[txID]
}

func (s *Seller) markSeen(txID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[txID] = true
}

// respondPaymentRequired writes a 402 response carrying a fresh PAYMENT-REQUIRED header.
func (s *Seller) respondPaymentRequired(w http.ResponseWriter, meta PriceInfo, reason string) {
	required := paymentRequired{
		X402Version: 2,
		Error:       reason,
		Resource: &resourceInfo{
			URL:         meta.ResourcePath,
			Description: meta.Description,
			MimeType:    "application/json",
		},
		Accepts: []paymentAccept{
			{
				Scheme:            "exact",
				Network:           s.network,
				Amount:            strconv.FormatUint(meta.AmountBaseUnits, 10),
				Asset:             strconv.FormatUint(s.assetID, 10),
				PayTo:             s.client.Address(),
				MaxTimeoutSeconds: 60,
				Extra:             map[string]any{},
			},
		},
	}

	raw, err := json.Marshal(required)
	if err == nil {
		w.Header().Set(headerPaymentRequired, base64.StdEncoding.EncodeToString(raw))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":    "payment required",
		"resource": meta.ResourcePath,
	})
}

// encodePaymentResponse JSON-marshals and base64-encodes a PAYMENT-RESPONSE payload.
func encodePaymentResponse(resp paymentResponse) (string, error) {
	raw, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("marshal payment response: %w", err)
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}
