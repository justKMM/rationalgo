package scenario

import (
	"context"
	"fmt"
	"time"

	"rationalgo/internal/models"
	"rationalgo/internal/repository"
	"rationalgo/internal/services/decision"
	"rationalgo/internal/services/outcome"
	"rationalgo/internal/services/reasoning"
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
	Reasoning *reasoning.Service
	Outcome   *outcome.Service
	Algorand  AlgorandCommitter
	X402      X402Payer
	Store     *repository.Store
	Vendors   func() []models.VendorOption
	Policy    func(models.VendorOption, float64, float64, float64, []string, map[string][]float64) models.PolicyResult
	Allowed   func() []string
	PriceHist func() map[string][]float64
	Inject    func(map[string][]float64, string, float64) map[string][]float64
	AgentID   string
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

func runScenario(ctx context.Context, scenario ScenarioType, deps Deps, ch chan<- Event) {
	emit := func(t EventType, payload interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		case ch <- Event{Type: t, Payload: payload}:
			return sleep(ctx, stepDelay)
		}
	}

	intent := reasoning.DemoIntent()
	emit(EventAgentThinking, map[string]string{"intent": intent})

	vendors := deps.Vendors()
	if len(vendors) == 0 {
		return
	}

	record, err := deps.Reasoning.GenerateDecision(ctx, intent, vendors)
	if err != nil {
		return
	}
	emit(EventDecisionPending, record)

	priceHist := deps.PriceHist()
	if scenario == ScenarioAnomaly && deps.Inject != nil {
		priceHist = deps.Inject(priceHist, record.VendorID, 10)
	}

	chosen := findVendor(vendors, record.VendorID)
	policyResult := deps.Policy(
		chosen,
		record.AmountEURQ,
		deps.DailySpent,
		deps.DailyLimit,
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
		return
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
			Vendor:       record.VendorID,
			AmountEURQ:   record.AmountEURQ,
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

	if deps.X402 != nil {
		body, err := deps.X402.PayAndFetch(ctx, chosen.URL, record.AmountEURQ)
		if err == nil {
			emit(EventPaymentSent, map[string]interface{}{
				"vendor":  record.VendorName,
				"amount":  record.AmountEURQ,
				"bytes":   len(body),
				"stub":    true,
			})
		}
	}

	precip := outcome.DemoPaidPrecip()
	out, err := deps.Outcome.Verify(ctx, precip)
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
