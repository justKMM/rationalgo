package algorand

import (
	"fmt"
	"time"

	"rationalgo/internal/config"
	"rationalgo/internal/models"
	"rationalgo/internal/services/decision"
	"rationalgo/internal/util"
	"rationalgo/pkg/provenance"
)

// Service orchestrates Algorand commitment operations.
type Service struct {
	cfg    config.Config
	client *Client
}

// NewService creates an Algorand service backed by a live testnet client.
func NewService(cfg config.Config) (*Service, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{cfg: cfg, client: client}, nil
}

// RunSpike commits a sample decision hash to Algorand Testnet (Phase 0).
func (s *Service) RunSpike() (models.AlgorandSpikeResult, error) {
	info, err := s.client.AccountInfo()
	if err != nil {
		return models.AlgorandSpikeResult{}, fmt.Errorf("account info: %w", err)
	}

	sample := map[string]any{
		"project":   "RationAlgo",
		"phase":     0,
		"intent":    "spike: weather data purchase reasoning",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	hash, err := decision.HashCanonicalJSON(sample)
	if err != nil {
		return models.AlgorandSpikeResult{}, err
	}

	txID, err := s.client.CommitHash(hash)
	if err != nil {
		return models.AlgorandSpikeResult{}, err
	}

	return models.AlgorandSpikeResult{
		Address:       s.client.Address(),
		MicroAlgos:    info.Amount,
		ReasoningHash: hash,
		TxID:          txID,
		ExplorerURL:   util.TxURL(txID),
	}, nil
}

// RunProvenanceSpike commits a sample RAv1 envelope to Algorand Testnet.
func (s *Service) RunProvenanceSpike() (models.AlgorandSpikeResult, error) {
	info, err := s.client.AccountInfo()
	if err != nil {
		return models.AlgorandSpikeResult{}, fmt.Errorf("account info: %w", err)
	}

	taskHash, err := decision.HashCanonicalJSON(reasoningDemoIntent())
	if err != nil {
		return models.AlgorandSpikeResult{}, err
	}

	record := map[string]any{
		"intent": reasoningDemoIntent(),
		"vendor": "company-basic-info",
		"phase":  "provenance-spike",
	}
	decisionHash, err := decision.HashCanonicalJSON(record)
	if err != nil {
		return models.AlgorandSpikeResult{}, err
	}

	env := &provenance.Envelope{
		Version:      1,
		AgentID:      "research-agent-01",
		SessionID:    fmt.Sprintf("spike-%d", time.Now().Unix()),
		TaskHash:     taskHash,
		DecisionHash: decisionHash,
		Vendor:       "company-basic-info",
		AmountEURQ:   0.01,
		Intent:       reasoningDemoIntent(),
		Expected:     "92% confidence basic company facts",
		Confidence:   0.87,
		CommittedAt:  time.Now().Unix(),
	}

	txID, err := s.client.CommitProvenance(env)
	if err != nil {
		return models.AlgorandSpikeResult{}, err
	}

	return models.AlgorandSpikeResult{
		Address:       s.client.Address(),
		MicroAlgos:    info.Amount,
		ReasoningHash: decisionHash,
		TxID:          txID,
		ExplorerURL:   util.TxURL(txID),
	}, nil
}

func reasoningDemoIntent() string {
	return "Research Atlas Robotics GmbH within a €1.00 data budget — which paid sources are worth buying?"
}

