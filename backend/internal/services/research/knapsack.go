package research

import (
	"math"
	"sort"
)

// Importance maps each endpoint name to its relative value in a research report.
var Importance = map[string]float64{
	"basic-info":         1.0,
	"industry":           0.8,
	"top-products":       1.2,
	"reviews-summary":    1.1,
	"competitors":        1.0,
	"news-sentiment":     1.3,
	"growth-rate":        1.4,
	"revenue-estimate":   1.6,
	"security-incidents": 1.5,
	"legal-issues":       1.5,
}

// Selection is one chosen endpoint with its computed value/price scoring.
type Selection struct {
	Endpoint EndpointMeta
	Value    float64 // importance * confidence
	Score    float64 // value / price_usd — value per dollar
}

// Select runs 0/1 knapsack over the given endpoints constrained by budgetUSD, using
// integer-cents weights, and returns the chosen subset ordered by Score descending.
func Select(items []EndpointMeta, confidence map[string]float64, budgetUSD float64) []Selection {
	n := len(items)
	budgetCents := int(math.Round(budgetUSD * 100))
	if n == 0 || budgetCents <= 0 {
		return []Selection{}
	}

	weights := make([]int, n)
	values := make([]float64, n)
	for i, item := range items {
		name := nameFromPath(item.Path)
		weights[i] = int(math.Round(item.PriceUSD * 100))

		conf, ok := confidence[name]
		if !ok {
			conf = 0.5
		}
		values[i] = Importance[name] * conf
	}

	// dp[i][c] = best total value achievable using items[:i] within budget c (cents).
	dp := make([][]float64, n+1)
	for i := range dp {
		dp[i] = make([]float64, budgetCents+1)
	}
	for i := 1; i <= n; i++ {
		w := weights[i-1]
		v := values[i-1]
		for c := 0; c <= budgetCents; c++ {
			dp[i][c] = dp[i-1][c]
			if w <= c {
				if alt := dp[i-1][c-w] + v; alt > dp[i][c] {
					dp[i][c] = alt
				}
			}
		}
	}

	// Reconstruct chosen subset by walking back through the DP table.
	chosen := make([]bool, n)
	c := budgetCents
	for i := n; i >= 1; i-- {
		if dp[i][c] != dp[i-1][c] {
			chosen[i-1] = true
			c -= weights[i-1]
		}
	}

	result := make([]Selection, 0, n)
	for i, item := range items {
		if !chosen[i] {
			continue
		}
		value := values[i]
		score := 0.0
		if item.PriceUSD > 0 {
			score = value / item.PriceUSD
		}
		result = append(result, Selection{
			Endpoint: item,
			Value:    value,
			Score:    score,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result
}
