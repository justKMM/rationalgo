package outcome

import (
	"context"
	"fmt"
	"math"
)

// Record is the verified outcome (alias for models.OutcomeRecord at service boundary).
type Record struct {
	Predicted   string
	Actual      string
	Score       float64
	GroundTruth string
	Verdict     string
	TrustDelta  float64
}

// Service verifies purchased research-endpoint confidence against the agent's pre-purchase estimate.
type Service struct{}

// NewService returns an outcome verification service.
func NewService() *Service {
	return &Service{}
}

// Verify compares the confidence the agent expected to get for its money against the
// confidence the endpoint's response actually reported, scoring how well the purchase paid off.
func (s *Service) Verify(ctx context.Context, expectedConfidence, actualConfidence float64) (*Record, error) {
	_ = ctx
	delta := math.Abs(actualConfidence - expectedConfidence)
	score := math.Max(0, 1-delta/0.5)

	verdict := "Good purchase"
	trustDelta := 0.08
	if score < 0.7 {
		verdict = "Within tolerance"
		trustDelta = 0.02
	}

	return &Record{
		Predicted:   fmt.Sprintf("confidence ≈ %.2f", expectedConfidence),
		Actual:      fmt.Sprintf("confidence %.2f (mock)", actualConfidence),
		Score:       score,
		GroundTruth: "endpoint-reported confidence",
		Verdict:     verdict,
		TrustDelta:  trustDelta,
	}, nil
}
