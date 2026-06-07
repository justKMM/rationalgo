package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"rationalgo/internal/models"
	"rationalgo/internal/repository"
	"rationalgo/internal/services/decision"
	"rationalgo/internal/services/outcome"
	"rationalgo/internal/services/reasoning"
	"rationalgo/internal/services/research"
	"rationalgo/pkg/provenance"
)

// EventType identifies a step in the hero demo stream.
type EventType string

const (
	EventAgentThinking     EventType = "agent.thinking"
	EventDecisionPending   EventType = "decision.pending"
	EventDecisionCommitted EventType = "decision.committed"
	EventPaymentSent       EventType = "payment.sent"
	EventDecisionOutcome   EventType = "decision.outcome"
	EventDecisionBlocked   EventType = "decision.blocked"
	EventAlertFired        EventType = "alert.fired"
	EventResearchSummary   EventType = "research.summary"
)

// Event is emitted on the scenario channel for SSE consumers.
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// ScenarioType selects normal or anomaly demo flow.
type ScenarioType string

const (
	ScenarioNormal  ScenarioType = "normal"
	ScenarioAnomaly ScenarioType = "anomaly"
)

// AlgorandCommitter commits provenance envelopes on-chain.
type AlgorandCommitter interface {
	CommitProvenance(env *provenance.Envelope) (string, error)
	CommitOutcome(env *provenance.OutcomeEnvelope) (string, error)
}

// X402Payer pays for and fetches a protected resource.
type X402Payer interface {
	PayAndFetch(ctx context.Context, url string, amountEURQ float64) ([]byte, error)
}

// Deps holds injected services for the hero orchestrator.
type Deps struct {
	Reasoning  *reasoning.Service
	Outcome    *outcome.Service
	Algorand   AlgorandCommitter
	X402       X402Payer
	Store      *repository.Store
	Vendors    func() []models.VendorOption
	Policy     func(models.VendorOption, float64, float64, float64, []string, map[string][]float64) models.PolicyResult
	Allowed    func() []string
	PriceHist  func() map[string][]float64
	Inject     func(map[string][]float64, string, float64) map[string][]float64
	AgentID    string
	DailySpent float64
	DailyLimit float64
}

const stepDelay = 600 * time.Millisecond

// Run executes the hero demo and streams events until completion or ctx cancel.
func Run(ctx context.Context, scenario ScenarioType, deps Deps) (<-chan Event, error) {
	if deps.Reasoning == nil || deps.Outcome == nil || deps.Store == nil {
		return nil, fmt.Errorf("scenario: missing required dependencies")
	}

	ch := make(chan Event, 16)
	go func() {
		defer close(ch)
		runScenario(ctx, scenario, deps, ch)
	}()
	return ch, nil
}

// planStep is one knapsack-selected endpoint awaiting purchase, in attempt order.
type planStep struct {
	vendor models.VendorOption
	score  float64
}

// researchEnvelope extracts the confidence field from a /company/* ApiResponse body.
type researchEnvelope struct {
	Confidence float64 `json:"confidence"`
}

func runScenario(ctx context.Context, scenario ScenarioType, deps Deps, ch chan<- Event) {
	emit := func(t EventType, payload interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		case ch <- Event{Type: t, Payload: payload}:
			return sleep(ctx, stepDelay)
		}
	}

	budget := deps.DailyLimit
	intent := reasoning.ResearchIntent(reasoning.DemoCompany, budget)
	emit(EventAgentThinking, map[string]interface{}{
		"intent":      intent,
		"company":     reasoning.DemoCompany,
		"budget_eurq": budget,
	})

	vendors := deps.Vendors()
	if len(vendors) == 0 {
		return
	}

	plan, excluded := buildResearchPlan(vendors, budget)
	if len(plan) == 0 {
		return
	}

	priceHist := deps.PriceHist()
	if scenario == ScenarioAnomaly && deps.Inject != nil {
		priceHist = deps.Inject(priceHist, plan[0].vendor.ID, 10)
	}

	spent := deps.DailySpent
	var purchased, blocked []string

	for i, step := range plan {
		chosen := step.vendor
		alternatives := researchAlternatives(plan, excluded, i, budget)
		expectedValue := fmt.Sprintf("%s — advertised confidence around %.0f%%", chosen.Description, chosen.SuccessRate*100)

		record, err := deps.Reasoning.GenerateResearchDecision(intent, chosen, alternatives, expectedValue)
		if err != nil {
			return
		}
		emit(EventDecisionPending, record)

		policyResult := deps.Policy(
			chosen,
			chosen.PriceEURQ,
			spent,
			budget,
			deps.Allowed(),
			priceHist,
		)
		record.Policy = policyResult

		hash, err := decision.HashCanonicalJSON(record)
		if err != nil {
			return
		}
		record.ReasoningHash = hash

		if !policyResult.Approved {
			record.Status = models.StatusBlocked
			emit(EventDecisionBlocked, record)
			alert := models.Alert{
				ID:      fmt.Sprintf("alert-%d", models.NowMillis()),
				Level:   "amber",
				Message: policyResult.BlockReason,
				At:      models.NowMillis(),
			}
			emit(EventAlertFired, alert)
			deps.Store.AddDecision(record.ToDecision())
			blocked = append(blocked, chosen.Name)
			continue
		}

		record.Status = models.StatusApproved

		if deps.Algorand != nil {
			taskHash, _ := decision.HashCanonicalJSON(intent)
			env := &provenance.Envelope{
				Version:      1,
				AgentID:      deps.AgentID,
				SessionID:    fmt.Sprintf("sess-%d", models.NowMillis()),
				TaskHash:     taskHash,
				DecisionHash: record.ReasoningHash,
				Vendor:       record.VendorChosen.ID,
				AmountEURQ:   record.VendorChosen.PriceEURQ,
				Intent:       record.TaskIntent,
				Expected:     record.ExpectedValue,
				Confidence:   record.Confidence,
				CommittedAt:  time.Now().Unix(),
			}
			txID, err := deps.Algorand.CommitProvenance(env)
			if err == nil {
				record.CommittedTx = txID
				emit(EventDecisionCommitted, record)
			}
		}

		// actualConfidence falls back to the catalog's advertised SuccessRate if payment
		// or response parsing fails, so every approved purchase still gets a full outcome.
		actualConfidence := chosen.SuccessRate
		if deps.X402 != nil {
			body, err := deps.X402.PayAndFetch(ctx, chosen.URL, chosen.PriceEURQ)
			if err == nil {
				emit(EventPaymentSent, map[string]interface{}{
					"vendor": record.VendorChosen.Name,
					"amount": record.VendorChosen.PriceEURQ,
					"bytes":  len(body),
					"paid":   true,
				})
				var env researchEnvelope
				if jsonErr := json.Unmarshal(body, &env); jsonErr == nil && env.Confidence > 0 {
					actualConfidence = env.Confidence
				}
			}
		}

		expectedConfidence := math.Round(chosen.SuccessRate*20) / 20
		out, err := deps.Outcome.Verify(ctx, expectedConfidence, actualConfidence)
		if err != nil {
			return
		}
		record.Outcome = &models.OutcomeRecord{
			Predicted:   out.Predicted,
			Actual:      out.Actual,
			Score:       out.Score,
			GroundTruth: out.GroundTruth,
			Verdict:     out.Verdict,
			TrustDelta:  out.TrustDelta,
		}

		if deps.Algorand != nil && record.CommittedTx != "" {
			outEnv := &provenance.OutcomeEnvelope{
				Version:     1,
				OriginalTx:  record.CommittedTx,
				Actual:      out.Actual,
				Score:       out.Score,
				GroundTruth: out.GroundTruth,
				ComputedAt:  time.Now().Unix(),
			}
			if txID, err := deps.Algorand.CommitOutcome(outEnv); err == nil {
				record.OutcomeTx = txID
			}
		}

		emit(EventDecisionOutcome, record)
		deps.Store.AddDecision(record.ToDecision())
		spent += chosen.PriceEURQ
		purchased = append(purchased, chosen.Name)
	}

	emit(EventResearchSummary, map[string]interface{}{
		"company":     reasoning.DemoCompany,
		"budget_eurq": budget,
		"spent_eurq":  spent,
		"purchased":   purchased,
		"blocked":     blocked,
	})
}

// buildResearchPlan runs budgeted 0/1 knapsack selection over the catalog's research
// endpoints, returning the purchase order (best value-per-price first) plus the
// endpoints that didn't fit the budget at all.
func buildResearchPlan(vendors []models.VendorOption, budgetEURQ float64) ([]planStep, []models.VendorOption) {
	selections := research.Select(research.Pricing, research.ConfidenceMap(), budgetEURQ)

	plan := make([]planStep, 0, len(selections))
	chosen := make(map[string]bool, len(selections))
	for _, sel := range selections {
		id := "company-" + sel.Endpoint.Name
		v := findVendor(vendors, id)
		plan = append(plan, planStep{vendor: v, score: sel.Score})
		chosen[id] = true
	}

	var excluded []models.VendorOption
	for _, v := range vendors {
		if !chosen[v.ID] {
			excluded = append(excluded, v)
		}
	}
	return plan, excluded
}

// researchAlternatives explains, for the purchase at index i, why every endpoint not yet
// bought was passed over — either it scored lower (queued for later) or it didn't fit
// the budget at all (excluded from the plan).
func researchAlternatives(plan []planStep, excluded []models.VendorOption, i int, budgetEURQ float64) []models.Alternative {
	var alts []models.Alternative
	chosenScore := plan[i].score
	for _, step := range plan[i+1:] {
		alts = append(alts, models.Alternative{
			Vendor: step.vendor,
			ReasonRejected: fmt.Sprintf(
				"Lower value/price score (%.2f vs %.2f) — queued later in the budgeted plan",
				step.score, chosenScore,
			),
		})
	}
	for _, v := range excluded {
		alts = append(alts, models.Alternative{
			Vendor: v,
			ReasonRejected: fmt.Sprintf(
				"Excluded from the €%.2f research budget — value/price score too low to fit",
				budgetEURQ,
			),
		})
	}
	return alts
}

func findVendor(vendors []models.VendorOption, id string) models.VendorOption {
	for _, v := range vendors {
		if v.ID == id {
			return v
		}
	}
	return vendors[0]
}

func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
