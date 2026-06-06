package catalog

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

// defaultVendors returns the hard-coded MVP catalog.
// TrustScore is 0–100; SuccessRate is 0–1.
func defaultVendors() []Vendor {
	return []Vendor{
		// Frankfurt drone demo (hero task)
		{
			ID: "goplausible-weather", Name: "GoPlausible Weather", Category: "weather",
			Description: "x402-protected Frankfurt precipitation forecast; 91% accuracy vs historical ground truth",
			PriceEURQ: 0.001, TrustScore: 91, SuccessRate: 0.91, AvgLatencyMs: 200,
			Allowed: true, EndpointURL: "https://example.x402.goplausible.xyz/avm/weather",
			Tags: []string{"weather", "x402", "frankfurt", "paid", "demo"},
		},
		{
			ID: "open-meteo", Name: "OpenMeteo", Category: "weather",
			Description: "Free Open-Meteo endpoint; 64% accuracy for localized drone ops windows",
			PriceEURQ: 0, TrustScore: 64, SuccessRate: 0.64, AvgLatencyMs: 350,
			Allowed: true, EndpointURL: "https://api.open-meteo.com/v1/forecast",
			Tags: []string{"weather", "free", "frankfurt", "demo"},
		},
		// Weather
		{
			ID: "weather-pro", Name: "WeatherPro", Category: "weather",
			Description: "Enterprise 24h forecast with sub-km grid, EU corridor coverage, SLA-backed uptime",
			PriceEURQ: 0.02, TrustScore: 92, SuccessRate: 0.98, AvgLatencyMs: 180,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/weather/weather-pro",
			Tags: []string{"weather", "forecast", "24h", "eu", "premium", "sla"},
		},
		{
			ID: "weather-cheap", Name: "WeatherCheap", Category: "weather",
			Description: "Budget hourly forecast; adequate for low-stakes routing, occasional gaps in BE/NL border cells",
			PriceEURQ: 0.005, TrustScore: 68, SuccessRate: 0.84, AvgLatencyMs: 420,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/weather/weather-cheap",
			Tags: []string{"weather", "forecast", "budget", "hourly"},
		},
		{
			ID: "free-weather", Name: "FreeWeather", Category: "weather",
			Description: "Community-maintained OpenMeteo mirror; no SLA, rate-limited, stale during peak hours",
			PriceEURQ: 0.001, TrustScore: 51, SuccessRate: 0.71, AvgLatencyMs: 890,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/weather/free-weather",
			Tags: []string{"weather", "free", "community", "opportunistic"},
		},
		// Routing
		{
			ID: "route-max", Name: "RouteMax", Category: "routing",
			Description: "Live-traffic-aware truck routing N1–N3 classes; toll and weight restrictions included",
			PriceEURQ: 0.08, TrustScore: 94, SuccessRate: 0.97, AvgLatencyMs: 210,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/routing/route-max",
			Tags: []string{"routing", "traffic", "truck", "toll", "live"},
		},
		{
			ID: "fleet-route-ai", Name: "FleetRouteAI", Category: "routing",
			Description: "ML-optimized fleet corridors; strong ETA accuracy, higher per-request cost",
			PriceEURQ: 0.12, TrustScore: 88, SuccessRate: 0.95, AvgLatencyMs: 340,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/routing/fleet-route-ai",
			Tags: []string{"routing", "ml", "fleet", "eta"},
		},
		// Fuel
		{
			ID: "fuel-price-api", Name: "FuelPriceAPI", Category: "fuel",
			Description: "Daily diesel index BE/NL/DE with 15-minute refresh on major hubs",
			PriceEURQ: 0.015, TrustScore: 86, SuccessRate: 0.96, AvgLatencyMs: 150,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/fuel/fuel-price-api",
			Tags: []string{"fuel", "diesel", "index", "eu"},
		},
		{
			ID: "energy-index-eu", Name: "EnergyIndexEU", Category: "fuel",
			Description: "Wholesale energy composite; broader coverage but 2h aggregation lag vs spot diesel",
			PriceEURQ: 0.006, TrustScore: 74, SuccessRate: 0.89, AvgLatencyMs: 280,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/fuel/energy-index-eu",
			Tags: []string{"fuel", "energy", "wholesale", "lagged"},
		},
		// Traffic
		{
			ID: "traffic-live", Name: "TrafficLive", Category: "traffic",
			Description: "Real-time incident and congestion feed for Benelux motorways",
			PriceEURQ: 0.025, TrustScore: 90, SuccessRate: 0.97, AvgLatencyMs: 120,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/traffic/traffic-live",
			Tags: []string{"traffic", "incidents", "realtime", "benelux"},
		},
		{
			ID: "road-pulse", Name: "RoadPulse", Category: "traffic",
			Description: "Crowd-sourced slowdown alerts; cheap but noisy signal, false positives on weekends",
			PriceEURQ: 0.004, TrustScore: 62, SuccessRate: 0.78, AvgLatencyMs: 510,
			Allowed: true, EndpointURL: "https://catalog.rationalgo.local/traffic/road-pulse",
			Tags: []string{"traffic", "crowd", "budget", "noisy"},
		},
	}
}
