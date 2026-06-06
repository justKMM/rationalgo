package vendor

import (
	"rationalgo/internal/catalog"
	"rationalgo/internal/models"
)

const demoTaskCategory = "weather"

// Service exposes catalog vendors for agent discovery and policy.
type Service struct {
	reg *catalog.Registry
}

// NewService returns a vendor service backed by the local catalog.
func NewService() *Service {
	return &Service{reg: catalog.NewRegistry()}
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

// GetDemoWeatherVendors returns weather vendors for the Frankfurt drone demo.
func (s *Service) GetDemoWeatherVendors() []models.VendorOption {
	return s.GetByCategory(demoTaskCategory)
}

// GetPriceHistory returns recent prices per vendor ID for anomaly detection.
func (s *Service) GetPriceHistory() map[string][]float64 {
	history := make(map[string][]float64)
	for _, v := range s.reg.GetAll() {
		history[v.ID] = defaultPriceHistory(v.ID, v.PriceEURQ)
	}
	return history
}

func defaultPriceHistory(id string, current float64) []float64 {
	// Last 7 observations; current price appended by caller for anomaly injection.
	switch id {
	case "goplausible-weather":
		return []float64{0.001, 0.001, 0.001, 0.001, 0.001, 0.001, current}
	case "open-meteo":
		return []float64{0, 0, 0, 0, 0, 0, current}
	default:
		base := current
		if base == 0 {
			base = 0.01
		}
		return []float64{base, base, base, base, base, base, current}
	}
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
