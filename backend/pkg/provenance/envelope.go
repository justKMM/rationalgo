package provenance

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const prefix = "RAv1:"
const outcomePrefix = "RAv1out:"

// Envelope is committed to the Algorand note field BEFORE payment fires.
// Note field value: "RAv1:<base64url(canonical_json)>"
type Envelope struct {
	Version      int     `json:"v"`
	AgentID      string  `json:"agent_id"`
	SessionID    string  `json:"session_id"`
	TaskHash     string  `json:"task_hash"`
	DecisionHash string  `json:"decision_hash"`
	Vendor       string  `json:"vendor"`
	AmountEURQ   float64 `json:"amount_eurq"`
	Intent       string  `json:"intent"`
	Expected     string  `json:"expected"`
	Confidence   float64 `json:"confidence"`
	CommittedAt  int64   `json:"committed_at"`
}

// OutcomeEnvelope is committed after the agent evaluates the response.
// Note field value: "RAv1out:<base64url(canonical_json)>"
type OutcomeEnvelope struct {
	Version     int     `json:"v"`
	OriginalTx  string  `json:"original_tx"`
	Actual      string  `json:"actual"`
	Score       float64 `json:"score"`
	GroundTruth string  `json:"ground_truth"`
	ComputedAt  int64   `json:"computed_at"`
}

// Encode serializes an Envelope to the RAv1 note string.
func Encode(e *Envelope) (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("provenance: encode: %w", err)
	}
	return prefix + base64.URLEncoding.EncodeToString(b), nil
}

// Decode parses an RAv1 note string into an Envelope.
func Decode(note string) (*Envelope, error) {
	if !strings.HasPrefix(note, prefix) {
		return nil, fmt.Errorf("provenance: not a RAv1 note")
	}
	b, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(note, prefix))
	if err != nil {
		return nil, fmt.Errorf("provenance: decode base64: %w", err)
	}
	var e Envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, fmt.Errorf("provenance: decode json: %w", err)
	}
	return &e, nil
}

// IsProvenance reports whether a note uses the RAv1 prefix.
func IsProvenance(note string) bool {
	return strings.HasPrefix(note, prefix)
}

// EncodeOutcome serializes an OutcomeEnvelope to the RAv1out note string.
func EncodeOutcome(e *OutcomeEnvelope) (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("provenance: encode outcome: %w", err)
	}
	return outcomePrefix + base64.URLEncoding.EncodeToString(b), nil
}

// DecodeOutcome parses an RAv1out note string into an OutcomeEnvelope.
func DecodeOutcome(note string) (*OutcomeEnvelope, error) {
	if !strings.HasPrefix(note, outcomePrefix) {
		return nil, fmt.Errorf("provenance: not a RAv1out note")
	}
	b, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(note, outcomePrefix))
	if err != nil {
		return nil, fmt.Errorf("provenance: decode outcome base64: %w", err)
	}
	var e OutcomeEnvelope
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, fmt.Errorf("provenance: decode outcome json: %w", err)
	}
	return &e, nil
}

// IsOutcome reports whether a note uses the RAv1out prefix.
func IsOutcome(note string) bool {
	return strings.HasPrefix(note, outcomePrefix)
}
