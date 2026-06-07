package policy

import (
	"math"

	"rationalgo/internal/models"
)

const anomalyMultiplier = 5.0

// Evaluate checks budget and price anomaly for a vendor spend.
func Evaluate(
	chosen models.VendorOption,
	amountEURQ, dailySpent, dailyLimit float64,
	priceHistory map[string][]float64,
) models.PolicyResult {
	result := models.PolicyResult{
		Approved:      true,
		BudgetOK:      dailySpent+amountEURQ <= dailyLimit,
		VendorAllowed: true,
	}

	if !result.BudgetOK {
		result.Approved = false
		result.BlockReason = "Daily spend limit exceeded"
		return result
	}

	if prices, ok := priceHistory[chosen.ID]; ok && len(prices) >= 2 {
		median := medianPrice(prices[:len(prices)-1])
		current := prices[len(prices)-1]
		if median > 0 && current/median >= anomalyMultiplier {
			result.PriceAnomaly = true
			result.Approved = false
			result.BlockReason = "Price anomaly: current quote exceeds 5× 7-day median"
		}
	}

	return result
}

func medianPrice(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	sorted := append([]float64(nil), prices...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// InjectAnomalyPrice returns a copy of price history with the last price multiplied.
func InjectAnomalyPrice(history map[string][]float64, vendorID string, multiplier float64) map[string][]float64 {
	out := make(map[string][]float64, len(history))
	for k, v := range history {
		copied := append([]float64(nil), v...)
		out[k] = copied
	}
	prices, ok := out[vendorID]
	if !ok || len(prices) == 0 {
		return out
	}
	last := prices[len(prices)-1]
	if last == 0 {
		last = 0.001
	}
	prices[len(prices)-1] = math.Round(last*multiplier*10000) / 10000
	out[vendorID] = prices
	return out
}
