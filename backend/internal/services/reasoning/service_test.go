//go:build integration

package reasoning_test

import (
	"context"
	"os"
	"testing"

	"rationalgo/internal/services/policy"
	"rationalgo/internal/services/reasoning"
	"rationalgo/internal/services/vendor"
)

func TestGenerateDecision_Integration(t *testing.T) {
	key := os.Getenv("RATIONALGO_ANTHROPIC_KEY")
	if key == "" {
		t.Skip("RATIONALGO_ANTHROPIC_KEY not set")
	}

	svc := reasoning.New(key)
	vendorSvc := vendor.NewService("http://localhost:8080")
	vendors := vendorSvc.GetResearchEndpoints()
	pol := policy.Evaluate(
		vendors[0],
		vendors[0].PriceEURQ,
		0.0,
		10.0,
		[]string{vendors[0].Name, vendors[1].Name},
		vendorSvc.GetPriceHistory(),
	)

	record, err := svc.GenerateDecision(
		context.Background(),
		"agent-01", "sess-001",
		reasoning.ResearchIntent(reasoning.DemoCompany, 1.0),
		vendors,
		pol,
	)
	if err != nil {
		t.Fatalf("GenerateDecision: %v", err)
	}

	if record.VendorChosen.Name == "" {
		t.Error("VendorChosen.Name is empty")
	}
	if len(record.Alternatives) == 0 {
		t.Error("Alternatives is empty — model should list rejected vendors")
	}
	if record.ReasoningHash == "" {
		t.Error("ReasoningHash not set")
	}

	t.Logf("chosen: %s | confidence: %.2f | hash: %s",
		record.VendorChosen.Name, record.Confidence, record.ReasoningHash[:8])
}
