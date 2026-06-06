package x402

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	headerPaymentRequired  = "PAYMENT-REQUIRED"
	headerPaymentSignature = "PAYMENT-SIGNATURE"
	headerPaymentResponse  = "PAYMENT-RESPONSE"
	headerLegacyPayment    = "X-PAYMENT"
)

// paymentRequired is the v2 PAYMENT-REQUIRED body (base64 JSON).
type paymentRequired struct {
	X402Version int              `json:"x402Version"`
	Error       string           `json:"error"`
	Resource    *resourceInfo    `json:"resource"`
	Accepts     []paymentAccept  `json:"accepts"`
	Extensions  map[string]any   `json:"extensions"`
}

type resourceInfo struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// paymentAccept is one payable option from PAYMENT-REQUIRED.accepts.
type paymentAccept struct {
	Scheme            string         `json:"scheme"`
	Network           string         `json:"network"`
	Amount            string         `json:"amount"`
	Asset             string         `json:"asset"`
	PayTo             string         `json:"payTo"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds"`
	Extra             map[string]any `json:"extra"`
}

// paymentSignature is the v2 PAYMENT-SIGNATURE / X-PAYMENT header payload.
type paymentSignature struct {
	X402Version int            `json:"x402Version"`
	Scheme      string         `json:"scheme"`
	Network     string         `json:"network"`
	Resource    *resourceInfo  `json:"resource,omitempty"`
	Accepted    paymentAccept  `json:"accepted"`
	Extensions  map[string]any `json:"extensions,omitempty"`
	Payload     avmPayload     `json:"payload"`
}

type avmPayload struct {
	PaymentGroup []string `json:"paymentGroup"`
	PaymentIndex int      `json:"paymentIndex"`
}

type paymentResponse struct {
	Success     bool   `json:"success"`
	ErrorReason string `json:"errorReason"`
	Payer       string `json:"payer"`
	Transaction string `json:"transaction"`
	Network     string `json:"network"`
}

func decodePaymentRequired(header string) (*paymentRequired, error) {
	if header == "" {
		return nil, fmt.Errorf("empty PAYMENT-REQUIRED header")
	}
	raw, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf("decode PAYMENT-REQUIRED: %w", err)
	}
	var req paymentRequired
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("parse PAYMENT-REQUIRED json: %w", err)
	}
	if len(req.Accepts) == 0 {
		return nil, fmt.Errorf("PAYMENT-REQUIRED has no accepts")
	}
	return &req, nil
}

func decodePaymentResponse(header string) (*paymentResponse, error) {
	if header == "" {
		return nil, fmt.Errorf("empty PAYMENT-RESPONSE header")
	}
	raw, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf("decode PAYMENT-RESPONSE: %w", err)
	}
	var resp paymentResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse PAYMENT-RESPONSE json: %w", err)
	}
	return &resp, nil
}

func selectAlgorandAccept(accepts []paymentAccept, preferredNetwork string) (paymentAccept, error) {
	var algorand []paymentAccept
	for _, a := range accepts {
		if a.Scheme != "exact" || !strings.HasPrefix(a.Network, "algorand:") {
			continue
		}
		algorand = append(algorand, a)
	}
	if len(algorand) == 0 {
		return paymentAccept{}, fmt.Errorf("no algorand exact payment option in PAYMENT-REQUIRED")
	}
	if preferredNetwork != "" {
		for _, a := range algorand {
			if a.Network == preferredNetwork {
				return a, nil
			}
		}
	}
	return algorand[0], nil
}

func feePayerFromExtra(extra map[string]any) string {
	if extra == nil {
		return ""
	}
	v, ok := extra["feePayer"].(string)
	if !ok {
		return ""
	}
	return v
}
