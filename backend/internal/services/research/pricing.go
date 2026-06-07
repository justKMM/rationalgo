package research

import "strings"

// algoUSDRate is an illustrative ALGO→USD rate used only to populate price_microalgo for display/spec-compliance.
const algoUSDRate = 0.20

// EndpointMeta describes a priced research endpoint.
type EndpointMeta struct {
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	PriceUSD       float64 `json:"price_usd"`
	PriceMicroAlgo int64   `json:"price_microalgo"`
	Description    string  `json:"description"`
	LatencyEstMs   int     `json:"latency_est_ms"`
}

// Pricing lists all priced research endpoints in canonical order.
//
// PriceMicroAlgo is informational only (spec-compliance / display); actual on-chain
// settlement uses a stablecoin ASA and a separate package handles that conversion.
var Pricing = []EndpointMeta{
	newMeta("basic-info", "/company/basic-info", 0.01, "Basic company facts", 50),
	newMeta("industry", "/company/industry", 0.01, "Industry classification", 50),
	newMeta("top-products", "/company/top-products", 0.02, "Top products", 100),
	newMeta("reviews-summary", "/company/reviews-summary", 0.10, "Aggregated reviews", 200),
	newMeta("competitors", "/company/competitors", 0.10, "Top competitors", 150),
	newMeta("news-sentiment", "/company/news-sentiment", 0.15, "30-day news sentiment", 300),
	newMeta("growth-rate", "/company/growth-rate", 0.20, "YoY growth estimate", 250),
	newMeta("revenue-estimate", "/company/revenue-estimate", 0.50, "Revenue range", 400),
	newMeta("security-incidents", "/company/security-incidents", 0.50, "Security incidents", 400),
	newMeta("legal-issues", "/company/legal-issues", 1.00, "Legal issues", 500),
}

func newMeta(name, path string, priceUSD float64, description string, latencyEstMs int) EndpointMeta {
	return EndpointMeta{
		Name:           name,
		Path:           path,
		PriceUSD:       priceUSD,
		PriceMicroAlgo: int64(priceUSD / algoUSDRate * 1_000_000),
		Description:    description,
		LatencyEstMs:   latencyEstMs,
	}
}

// Find looks up an endpoint's metadata by its short name (e.g. "basic-info").
func Find(name string) (EndpointMeta, bool) {
	for _, m := range Pricing {
		if m.Name == name {
			return m, true
		}
	}
	return EndpointMeta{}, false
}

// nameFromPath derives the short endpoint name from its path, e.g. "/company/basic-info" -> "basic-info".
func nameFromPath(path string) string {
	return strings.TrimPrefix(path, "/company/")
}
