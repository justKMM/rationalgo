package catalog

import (
	"fmt"
	"strings"

	"rationalgo/internal/services/research"
)

// Vendor is a curated API offering in the local registry (MVP source of truth).
type Vendor struct {
	ID           string
	Name         string
	Category     string
	Description  string
	PriceEURQ    float64
	TrustScore   float64
	SuccessRate  float64
	AvgLatencyMs int
	Allowed      bool
	EndpointURL  string
	Tags         []string
}

// CategoryCompanyResearch groups RationAlgo's own x402-protected company-research endpoints.
const CategoryCompanyResearch = "company-research"

// displayNames gives each priced research endpoint a human-friendly catalog name
// (research.Pricing carries machine-oriented short names like "basic-info").
var displayNames = map[string]string{
	"basic-info":         "Company Basic Info",
	"industry":           "Industry Classification",
	"top-products":       "Top Products",
	"reviews-summary":    "Reviews Summary",
	"competitors":        "Competitor Landscape",
	"news-sentiment":     "News Sentiment",
	"growth-rate":        "Growth Rate Estimate",
	"revenue-estimate":   "Revenue Estimate",
	"security-incidents": "Security Incidents",
	"legal-issues":       "Legal Issues",
}

// defaultVendors derives the MVP catalog from research.Pricing — RationAlgo's own
// x402-protected /company/* endpoints — so price and confidence stay in sync with
// what the seller actually advertises and serves. baseURL anchors EndpointURL at the
// locally-reachable API address (see config.Config.PublicBaseURL).
func defaultVendors(baseURL string) []Vendor {
	confidence := research.ConfidenceMap()
	out := make([]Vendor, 0, len(research.Pricing))
	for _, meta := range research.Pricing {
		name := strings.TrimPrefix(meta.Path, "/company/")
		conf := confidence[name]
		display, ok := displayNames[name]
		if !ok {
			display = meta.Description
		}
		out = append(out, Vendor{
			ID:           "company-" + name,
			Name:         display,
			Category:     CategoryCompanyResearch,
			Description:  fmt.Sprintf("%s — x402-protected, served by RationAlgo's own research marketplace", meta.Description),
			PriceEURQ:    meta.PriceUSD,
			TrustScore:   conf * 100,
			SuccessRate:  conf,
			AvgLatencyMs: meta.LatencyEstMs,
			Allowed:      true,
			EndpointURL:  baseURL + meta.Path,
			Tags:         []string{CategoryCompanyResearch, "x402", "paid", name},
		})
	}
	return out
}
