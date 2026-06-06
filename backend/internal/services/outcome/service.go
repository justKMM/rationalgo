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

// Service verifies paid forecast responses against ground truth (stub).
type Service struct{}

// NewService returns an outcome verification service.
func NewService() *Service {
	return &Service{}
}

// Verify compares paid precipitation forecast to a simulated OpenMeteo ground truth.
func (s *Service) Verify(ctx context.Context, paidPrecipPct float64) (*Record, error) {
	_ = ctx
	groundTruth := 12.0 + math.Mod(paidPrecipPct, 5)
	delta := math.Abs(paidPrecipPct - groundTruth)
	score := math.Max(0, 1-delta/20)

	verdict := "Good purchase"
	trustDelta := 0.08
	if score < 0.7 {
		verdict = "Within tolerance"
		trustDelta = 0.02
	}

	return &Record{
		Predicted:   fmt.Sprintf("%.0f%% precip", paidPrecipPct),
		Actual:      fmt.Sprintf("%.0f%% precip", groundTruth),
		Score:       score,
		GroundTruth: "OpenMeteo historical",
		Verdict:     verdict,
		TrustDelta:  trustDelta,
	}, nil
}

// DemoPaidPrecip returns a simulated paid API precipitation value for the demo.
func DemoPaidPrecip() float64 {
	return 14.0
}
