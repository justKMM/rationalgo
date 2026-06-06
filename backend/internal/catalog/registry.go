package catalog

// Registry is the local, hard-coded vendor catalog for MVP agent discovery.
type Registry struct {
	vendors []Vendor
}

// NewRegistry returns a registry seeded with the default vendor catalog.
func NewRegistry() *Registry {
	vendors := defaultVendors()
	return &Registry{vendors: vendors}
}

// GetAll returns every vendor in the catalog.
func (r *Registry) GetAll() []Vendor {
	out := make([]Vendor, len(r.vendors))
	copy(out, r.vendors)
	return out
}

// GetByCategory returns vendors matching category (e.g. "weather", "routing").
func (r *Registry) GetByCategory(category string) []Vendor {
	var out []Vendor
	for _, v := range r.vendors {
		if v.Category == category {
			out = append(out, v)
		}
	}
	return out
}

// GetByID returns a vendor by stable catalog ID.
func (r *Registry) GetByID(id string) (*Vendor, bool) {
	for i := range r.vendors {
		if r.vendors[i].ID == id {
			v := r.vendors[i]
			return &v, true
		}
	}
	return nil, false
}

// AllowedVendors returns vendors marked Allowed=true (policy allowlist input).
func (r *Registry) AllowedVendors() []Vendor {
	var out []Vendor
	for _, v := range r.vendors {
		if v.Allowed {
			out = append(out, v)
		}
	}
	return out
}

// Future integration (store/seed.go, agent discovery, hero demo):
//
//	reg := catalog.NewRegistry()
//	weather := reg.GetByCategory("weather")
//	chosen, _ := reg.GetByID("weather-pro")
//	allowed := reg.AllowedVendors()
//
// Dashboard seed data can derive vendor names, prices, and trust scores from the
// registry instead of duplicating literals — e.g. FuelPriceAPI row from
// reg.GetByID("fuel-price-api").
