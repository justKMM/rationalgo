package store

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"rationalgo/internal/models"
)

// Seed returns the initial dashboard state matching the frontend mock.
func Seed() models.AppState {
	now := time.Now().UnixMilli()

	return models.AppState{
		Agent:      "fleet-router-01",
		Balance:    9.41,
		Spent:      0.59,
		DailyLimit: 10.0,
		Decisions: []models.Decision{
			{
				ID: "fp1", Vendor: "FuelPriceAPI", Status: models.StatusApproved,
				AmountEURQ: 0.15,
				Intent:     "Fuel index lookup BE/NL/DE for nightly fleet rebalancing",
				Alternatives: []models.LegacyAlternative{
					{Name: "EIA-public", Reason: "EU coverage missing; US-only feed"},
					{Name: "cached @ 06:00", Reason: "8h stale — diesel moved 2.1% intraday"},
				},
				ExpectedValue: "+11% rebalancing margin", Confidence: 0.74,
				Policy: models.PolicyChecks{BudgetOk: true, Reputation: 4.0, Anomaly: "none", VendorAllowed: true},
				ReasoningHash: mockHash(), Round: mockRound(), Timestamp: now - 14*60*1000,
				Outcome: &models.Outcome{Predicted: "+11%", Actual: "+9%", Verdict: "Within band", TrustDelta: -0.02},
			},
			{
				ID: "tg1", Vendor: "TollGuru", Status: models.StatusApproved,
				AmountEURQ: 0.04,
				Intent:     "Toll cost lookup BE→NL corridor, truck class N3",
				Alternatives: []models.LegacyAlternative{
					{Name: "ViaMichelin", Reason: "no N3 truck class endpoint"},
					{Name: "internal-table", Reason: "last updated 2024-11; rates revised Q1"},
				},
				ExpectedValue: "+4% quote accuracy", Confidence: 0.88,
				Policy: models.PolicyChecks{BudgetOk: true, Reputation: 4.5, Anomaly: "none", VendorAllowed: true},
				ReasoningHash: mockHash(), Round: mockRound(), Timestamp: now - 9*60*1000,
				Outcome: &models.Outcome{Predicted: "+4%", Actual: "+5%", Verdict: "Good purchase", TrustDelta: 0.05},
			},
			{
				ID: "os1", Vendor: "OSRM-Pro", Status: models.StatusApproved,
				AmountEURQ: 0.08,
				Intent:     "Recompute route after traffic spike on A4 Antwerp→Breda",
				Alternatives: []models.LegacyAlternative{
					{Name: "OSRM-public", Reason: "no live-traffic layer"},
					{Name: "Google Directions", Reason: "1.2 EURQ per 1k req — 15× quote"},
				},
				ExpectedValue: "+18% ETA accuracy", Confidence: 0.83,
				Policy: models.PolicyChecks{BudgetOk: true, Reputation: 4.7, Anomaly: "none", VendorAllowed: true},
				ReasoningHash: mockHash(), Round: mockRound(), Timestamp: now - 4*60*1000,
				Outcome: &models.Outcome{Predicted: "+18%", Actual: "+21%", Verdict: "Good purchase", TrustDelta: 0.08},
			},
			{
				ID: "ss1", Vendor: "ScrapeShack", Status: models.StatusBlocked,
				AmountEURQ: 0.9,
				Intent:     "Competitor pricing scrape, rotating residential proxy",
				Alternatives: []models.LegacyAlternative{
					{Name: "RateIntel", Reason: "same dataset @ lower quote when price normalizes"},
				},
				ExpectedValue: "—", Confidence: 0.41,
				Policy: models.PolicyChecks{BudgetOk: true, Reputation: 1.8, Anomaly: "flagged", VendorAllowed: true},
				ReasoningHash: mockHash(), Round: mockRound(), Timestamp: now - 2*60*1000,
				BlockedReason: "Price anomaly: current quote exceeds 5× 7-day median",
			},
		},
		Vendors: []models.Vendor{
			{Name: "OSRM-Pro", Score: 4.7},
			{Name: "TollGuru", Score: 4.5},
			{Name: "WeatherAPI", Score: 4.2},
			{Name: "FuelPriceAPI", Score: 4.0},
			{Name: "MetricsHub.xyz", Score: 1.1},
		},
		Alerts: []models.Alert{
			{
				ID: "a1", Level: "amber",
				Message: "WeatherAPI price +12% vs 7d avg — within tolerance, monitoring",
				At:      now - 22*60*1000,
			},
		},
		SelectedID: nil,
	}
}

func mockHash() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "0x" + hex.EncodeToString(b)
}

func mockRound() int64 {
	return 41_238_900 + int64(time.Now().UnixNano()%9999)
}
