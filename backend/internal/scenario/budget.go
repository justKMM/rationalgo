package scenario

// Budget tier names for the hero demo (query param ?budget=).
const (
	BudgetTierCheapass = "cheapass"
	BudgetTierMid      = "mid"
	BudgetTierLuxury   = "luxury"
)

// BudgetEURQ returns the research envelope for a tier name, or (0, false) if unknown.
func BudgetEURQ(tier string) (float64, bool) {
	switch tier {
	case BudgetTierCheapass:
		return 5.0, true
	case BudgetTierMid:
		return 10.0, true
	case BudgetTierLuxury:
		return 15.0, true
	default:
		return 0, false
	}
}
