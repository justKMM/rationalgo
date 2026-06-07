package models

// DecisionStatus is the policy outcome for a spend request.
type DecisionStatus string

const (
	StatusApproved DecisionStatus = "APPROVED"
	StatusBlocked  DecisionStatus = "BLOCKED"
	StatusPending  DecisionStatus = "PENDING"
)

// LegacyAlternative describes a vendor the agent considered but did not choose (API/frontend).
type LegacyAlternative struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// VendorOption is a catalog vendor candidate for agent discovery and policy evaluation.
type VendorOption struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category,omitempty"`
	URL          string  `json:"url"`
	PriceEURQ    float64 `json:"price_eurq"`
	TrustScore   float64 `json:"trust_score"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs int     `json:"avg_latency_ms,omitempty"`
	Description  string  `json:"description,omitempty"`
}

// PolicyResult is the output of policy evaluation for a spend request.
type PolicyResult struct {
	Approved      bool   `json:"approved"`
	BudgetOK      bool   `json:"budget_ok"`
	VendorAllowed bool   `json:"vendor_allowed"`
	PriceAnomaly  bool   `json:"price_anomaly"`
	BlockReason   string `json:"block_reason,omitempty"`
}

// Alternative describes a vendor the agent considered but rejected (service layer).
type Alternative struct {
	Vendor         VendorOption `json:"vendor"`
	ReasonRejected string       `json:"reason_rejected"`
}

// OutcomeRecord records post-purchase results vs expectations.
type OutcomeRecord struct {
	Predicted   string  `json:"predicted"`
	Actual      string  `json:"actual"`
	Score       float64 `json:"score"`
	GroundTruth string  `json:"ground_truth"`
	Verdict     string  `json:"verdict,omitempty"`
	TrustDelta  float64 `json:"trust_delta,omitempty"`
}

// DecisionRecord is the full audit record for an agent spend decision.
type DecisionRecord struct {
	ID            string         `json:"id"`
	AgentID       string         `json:"agent_id,omitempty"`
	SessionID     string         `json:"session_id,omitempty"`
	TaskIntent    string         `json:"task_intent"`
	VendorChosen  VendorOption   `json:"vendor_chosen"`
	Alternatives  []Alternative  `json:"alternatives"`
	ExpectedValue string         `json:"expected_value"`
	Confidence    float64        `json:"confidence"`
	Policy        PolicyResult   `json:"policy"`
	Status        DecisionStatus `json:"status"`
	ReasoningHash string         `json:"reasoning_hash"`
	CommittedTx   string         `json:"committed_tx,omitempty"`
	SettlementTx  string         `json:"settlement_tx,omitempty"`
	OutcomeTx     string         `json:"outcome_tx,omitempty"`
	Outcome       *OutcomeRecord `json:"outcome,omitempty"`
	Timestamp     int64          `json:"timestamp"`
}

// ToDecision converts a DecisionRecord to the dashboard Decision shape.
func (r DecisionRecord) ToDecision() Decision {
	d := Decision{
		ID:            r.ID,
		Vendor:        r.VendorChosen.Name,
		Status:        r.Status,
		AmountEURQ:    r.VendorChosen.PriceEURQ,
		Intent:        r.TaskIntent,
		Alternatives:  alternativesToLegacy(r.Alternatives),
		ExpectedValue: r.ExpectedValue,
		Confidence:    r.Confidence,
		Policy: PolicyChecks{
			BudgetOk:      r.Policy.BudgetOK,
			Reputation:    r.VendorChosen.TrustScore,
			Anomaly:       anomalyLabel(r.Policy.PriceAnomaly),
			VendorAllowed: r.Policy.VendorAllowed,
		},
		ReasoningHash: r.ReasoningHash,
		Timestamp:     r.Timestamp,
		CommittedTx:   r.CommittedTx,
		SettlementTx:  r.SettlementTx,
		OutcomeTx:     r.OutcomeTx,
	}
	if !r.Policy.Approved {
		d.BlockedReason = r.Policy.BlockReason
	}
	if r.Outcome != nil {
		d.Outcome = &Outcome{
			Predicted:  r.Outcome.Predicted,
			Actual:     r.Outcome.Actual,
			Verdict:    r.Outcome.Verdict,
			TrustDelta: r.Outcome.TrustDelta,
		}
	}
	return d
}

func alternativesToLegacy(alts []Alternative) []LegacyAlternative {
	out := make([]LegacyAlternative, len(alts))
	for i, a := range alts {
		out[i] = LegacyAlternative{
			Name:   a.Vendor.Name,
			Reason: a.ReasonRejected,
		}
	}
	return out
}

func anomalyLabel(flagged bool) string {
	if flagged {
		return "flagged"
	}
	return "none"
}

// PolicyChecks captures policy engine results for a decision.
type PolicyChecks struct {
	BudgetOk      bool    `json:"budgetOk"`
	Reputation    float64 `json:"reputation"`
	Anomaly       string  `json:"anomaly"`
	VendorAllowed bool    `json:"vendorAllowed"`
}

// Outcome records post-purchase results vs expectations.
type Outcome struct {
	Predicted  string  `json:"predicted"`
	Actual     string  `json:"actual"`
	Verdict    string  `json:"verdict"`
	TrustDelta float64 `json:"trustDelta"`
}

// Decision is the audit record for an agent spend request (API/frontend model).
type Decision struct {
	ID            string              `json:"id"`
	Vendor        string              `json:"vendor"`
	Status        DecisionStatus      `json:"status"`
	AmountEURQ    float64             `json:"amountEURQ"`
	Intent        string              `json:"intent"`
	Alternatives  []LegacyAlternative `json:"alternatives"`
	ExpectedValue string              `json:"expectedValue"`
	Confidence    float64             `json:"confidence"`
	Policy        PolicyChecks        `json:"policy"`
	ReasoningHash string              `json:"reasoningHash"`
	Round         int64               `json:"round"`
	Timestamp     int64               `json:"timestamp"`
	Outcome       *Outcome            `json:"outcome,omitempty"`
	BlockedReason string              `json:"blockedReason,omitempty"`
	CommittedTx   string              `json:"committedTx,omitempty"`
	SettlementTx  string              `json:"settlementTx,omitempty"`
	OutcomeTx     string              `json:"outcomeTx,omitempty"`
	ExplorerURL   string              `json:"explorerUrl,omitempty"`
}

// Vendor tracks trust score for a vendor.
type Vendor struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// Alert is a policy or anomaly notification.
type Alert struct {
	ID      string `json:"id"`
	Level   string `json:"level"`
	Message string `json:"message"`
	At      int64  `json:"at"`
}

// AppState is the dashboard snapshot served to the frontend.
type AppState struct {
	Agent          string     `json:"agent"`
	Balance        float64    `json:"balance"`
	Spent          float64    `json:"spent"`
	DailyLimit     float64    `json:"dailyLimit"`
	Decisions []Decision `json:"decisions"`
	Vendors   []Vendor   `json:"vendors"`
	Alerts    []Alert    `json:"alerts"`
	SelectedID     *string    `json:"selectedId"`
}
