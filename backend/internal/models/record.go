package models

// VendorOption is a vendor candidate for agent discovery and policy evaluation.
type VendorOption struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	URL          string  `json:"url"`
	PriceEURQ    float64 `json:"price_eurq"`
	TrustScore   float64 `json:"trust_score"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs int     `json:"avg_latency_ms"`
	Allowed      bool    `json:"allowed"`
	Description  string  `json:"description"`
}

// PolicyResult is the output of policy evaluation for a spend request.
type PolicyResult struct {
	Approved      bool   `json:"approved"`
	BudgetOK      bool   `json:"budget_ok"`
	VendorAllowed bool   `json:"vendor_allowed"`
	PriceAnomaly  bool   `json:"price_anomaly"`
	BlockReason   string `json:"block_reason,omitempty"`
}

// OutcomeRecord captures post-purchase verification against ground truth.
type OutcomeRecord struct {
	Predicted   string  `json:"predicted"`
	Actual      string  `json:"actual"`
	Score       float64 `json:"score"`
	GroundTruth string  `json:"ground_truth"`
	Verdict     string  `json:"verdict"`
	TrustDelta  float64 `json:"trust_delta"`
}

// DecisionRecord is the full audit record for an agent spend decision.
type DecisionRecord struct {
	ID            string         `json:"id"`
	TaskIntent    string         `json:"task_intent"`
	VendorID      string         `json:"vendor_id"`
	VendorName    string         `json:"vendor_name"`
	AmountEURQ    float64        `json:"amount_eurq"`
	Alternatives  []Alternative  `json:"alternatives"`
	ExpectedValue string         `json:"expected_value"`
	Confidence    float64        `json:"confidence"`
	VendorTrust   float64        `json:"vendor_trust"`
	Policy        PolicyResult   `json:"policy"`
	Status        DecisionStatus `json:"status"`
	ReasoningHash string         `json:"reasoning_hash"`
	CommittedTx   string         `json:"committed_tx,omitempty"`
	OutcomeTx     string         `json:"outcome_tx,omitempty"`
	Outcome       *OutcomeRecord `json:"outcome,omitempty"`
	Timestamp     int64          `json:"timestamp"`
}

// ToDecision converts a DecisionRecord to the dashboard Decision shape.
func (r DecisionRecord) ToDecision() Decision {
	d := Decision{
		ID:            r.ID,
		Vendor:        r.VendorName,
		Status:        r.Status,
		AmountEURQ:    r.AmountEURQ,
		Intent:        r.TaskIntent,
		Alternatives:  r.Alternatives,
		ExpectedValue: r.ExpectedValue,
		Confidence:    r.Confidence,
		Policy: PolicyChecks{
			BudgetOk:      r.Policy.BudgetOK,
			Reputation:    r.VendorTrust,
			Anomaly:       anomalyLabel(r.Policy.PriceAnomaly),
			VendorAllowed: r.Policy.VendorAllowed,
		},
		ReasoningHash: r.ReasoningHash,
		Timestamp:     r.Timestamp,
		CommittedTx:   r.CommittedTx,
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

func anomalyLabel(flagged bool) string {
	if flagged {
		return "flagged"
	}
	return "none"
}
