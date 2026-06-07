package vendor

import (
	"rationalgo/internal/catalog"
	"rationalgo/internal/models"
)

// Service exposes catalog vendors for agent discovery and policy.
type Service struct {
	reg *catalog.Registry
}

// NewService returns a vendor service backed by the local catalog. baseURL anchors each
// vendor's EndpointURL — pass cfg.PublicBaseURL() so x402 seller routes resolve locally.
func NewService(baseURL string) *Service {
	return &Service{reg: catalog.NewRegistry(baseURL)}
}

// GetAll returns all catalog vendors as VendorOption values.
func (s *Service) GetAll() []models.VendorOption {
	all := s.reg.GetAll()
	out := make([]models.VendorOption, len(all))
	for i, v := range all {
		out[i] = toOption(v)
	}
	return out
}

// GetByCategory returns vendors in a category.
func (s *Service) GetByCategory(category string) []models.VendorOption {
	all := s.reg.GetByCategory(category)
	out := make([]models.VendorOption, len(all))
	for i, v := range all {
		out[i] = toOption(v)
	}
	return out
}

// GetResearchEndpoints returns RationAlgo's own x402-protected /company/* research vendors.
func (s *Service) GetResearchEndpoints() []models.VendorOption {
	return s.GetByCategory(catalog.CategoryCompanyResearch)
}

// GetPriceHistory returns recent prices per vendor ID for anomaly detection.
func (s *Service) GetPriceHistory() map[string][]float64 {
	history := make(map[string][]float64)
	for _, v := range s.reg.GetAll() {
		history[v.ID] = defaultPriceHistory(v.PriceEURQ)
	}
	return history
}

// defaultPriceHistory returns 7 stable observations ending in current — the flat
// baseline against which policy.InjectAnomalyPrice can spike a single vendor's price.
func defaultPriceHistory(current float64) []float64 {
	base := current
	if base == 0 {
		base = 0.01
	}
	return []float64{base, base, base, base, base, base, current}
}

func toOption(v catalog.Vendor) models.VendorOption {
	return models.VendorOption{
		ID:           v.ID,
		Name:         v.Name,
		Category:     v.Category,
		URL:          v.EndpointURL,
		PriceEURQ:    v.PriceEURQ,
		TrustScore:   v.TrustScore,
		SuccessRate:  v.SuccessRate,
		AvgLatencyMs: v.AvgLatencyMs,
		Allowed:      v.Allowed,
		Description:  v.Description,
	}
}
