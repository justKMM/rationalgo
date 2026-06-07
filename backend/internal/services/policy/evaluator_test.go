package policy

import (
	"testing"

	"rationalgo/internal/models"
)

func TestEvaluate(t *testing.T) {
	vendor := models.VendorOption{
		ID:        "goplausible-weather",
		Name:      "GoPlausible Weather",
		PriceEURQ: 0.001,
	}
	history := map[string][]float64{
		"goplausible-weather": {0.001, 0.001, 0.001, 0.0011, 0.001, 0.001, 0.001},
	}

	tests := []struct {
		name        string
		amount      float64
		dailySpent  float64
		dailyLimit  float64
		history     map[string][]float64
		wantBudget  bool
		wantAnomaly bool
	}{
		{
			name:        "all checks pass",
			amount:      0.001,
			dailySpent:  0.005,
			dailyLimit:  10.0,
			history:     history,
			wantBudget:  true,
			wantAnomaly: false,
		},
		{
			name:        "budget exceeded",
			amount:      5.0,
			dailySpent:  8.0,
			dailyLimit:  10.0,
			history:     history,
			wantBudget:  false,
			wantAnomaly: false,
		},
		{
			name:        "price anomaly triggered",
			amount:      0.001,
			dailySpent:  0.0,
			dailyLimit:  10.0,
			history: map[string][]float64{
				"goplausible-weather": {0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.01},
			},
			wantBudget:  true,
			wantAnomaly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(vendor, tt.amount, tt.dailySpent, tt.dailyLimit, tt.history)
			if got.BudgetOK != tt.wantBudget {
				t.Errorf("BudgetOK: got %v, want %v", got.BudgetOK, tt.wantBudget)
			}
			if !got.VendorAllowed {
				t.Error("VendorAllowed should always be true")
			}
			if got.PriceAnomaly != tt.wantAnomaly {
				t.Errorf("PriceAnomaly: got %v, want %v", got.PriceAnomaly, tt.wantAnomaly)
			}
			if !tt.wantBudget || tt.wantAnomaly {
				if got.BlockReason == "" {
					t.Error("BlockReason must be non-empty on failure/anomaly")
				}
			}
		})
	}
}
