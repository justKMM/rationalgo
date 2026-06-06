package reasoning

import (
	"context"
	"fmt"

	"rationalgo/internal/models"
	"rationalgo/internal/services/decision"
)

const demoIntent = "Should drone deliveries operate in Frankfurt in the next 2 hours?"

// Service generates decision records from vendor options (stub — no LLM yet).
type Service struct{}

// NewService returns a reasoning service.
func NewService() *Service {
	return &Service{}
}

// GenerateDecision picks the best vendor for the task and builds a DecisionRecord.
func (s *Service) GenerateDecision(
	ctx context.Context,
	intent string,
	vendors []models.VendorOption,
) (*models.DecisionRecord, error) {
	_ = ctx
	if len(vendors) == 0 {
		return nil, fmt.Errorf("reasoning: no vendors provided")
	}

	chosen := pickBest(vendors)
	alts := buildAlternatives(vendors, chosen.ID)

	record := &models.DecisionRecord{
		ID:            fmt.Sprintf("dec-%d", models.NowMillis()),
		TaskIntent:    intent,
		VendorID:      chosen.ID,
		VendorName:    chosen.Name,
		AmountEURQ:    chosen.PriceEURQ,
		Alternatives:  alts,
		ExpectedValue: fmt.Sprintf("%.0f%% forecast accuracy", chosen.TrustScore),
		Confidence:    chosen.SuccessRate,
		VendorTrust:   chosen.TrustScore / 20,
		Status:        models.StatusPending,
		Timestamp:     models.NowMillis(),
	}

	hash, err := decision.HashCanonicalJSON(record)
	if err != nil {
		return nil, fmt.Errorf("reasoning: hash decision: %w", err)
	}
	record.ReasoningHash = hash
	return record, nil
}

// DemoIntent returns the Frankfurt drone task string.
func DemoIntent() string {
	return demoIntent
}

func pickBest(vendors []models.VendorOption) models.VendorOption {
	best := vendors[0]
	for _, v := range vendors[1:] {
		if v.TrustScore > best.TrustScore {
			best = v
		}
	}
	return best
}

func buildAlternatives(vendors []models.VendorOption, chosenID string) []models.Alternative {
	var alts []models.Alternative
	for _, v := range vendors {
		if v.ID == chosenID {
			continue
		}
		reason := fmt.Sprintf("trust %.0f vs chosen; price %.4f EURQ", v.TrustScore, v.PriceEURQ)
		if v.PriceEURQ == 0 {
			reason = fmt.Sprintf("free but %.0f%% accuracy vs paid option", v.TrustScore)
		}
		alts = append(alts, models.Alternative{Name: v.Name, Reason: reason})
	}
	return alts
}
