package research

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"time"
)

// confidenceByName holds the fixed, per-endpoint confidence reflecting realistic source quality.
var confidenceByName = map[string]float64{
	"basic-info":         0.92,
	"industry":           0.85,
	"top-products":       0.78,
	"reviews-summary":    0.80,
	"competitors":        0.75,
	"news-sentiment":     0.70,
	"growth-rate":        0.65,
	"revenue-estimate":   0.60,
	"security-incidents": 0.58,
	"legal-issues":       0.55,
}

// Confidence returns the fixed advertised confidence for an endpoint by its short name
// (e.g. "basic-info"), defaulting to 0.5 for unknown names.
func Confidence(name string) float64 {
	if c, ok := confidenceByName[name]; ok {
		return c
	}
	return 0.5
}

// ConfidenceMap returns a defensive copy of the per-endpoint advertised confidence values,
// keyed by short endpoint name — the canonical input to research.Select's value scoring.
func ConfidenceMap() map[string]float64 {
	out := make(map[string]float64, len(confidenceByName))
	for k, v := range confidenceByName {
		out[k] = v
	}
	return out
}

// seedFor derives a stable 64-bit seed from a company name and a field-specific salt.
func seedFor(company, salt string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(strings.ToLower(strings.TrimSpace(company)) + "|" + salt))
	return h.Sum64()
}

// rngFor returns a deterministic PRNG seeded from the company name and salt.
func rngFor(company, salt string) *rand.Rand {
	return rand.New(rand.NewSource(int64(seedFor(company, salt))))
}

// retrievedAtNow formats the current instant as the provenance retrieval timestamp.
func retrievedAtNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func provenanceFor() Provenance {
	return Provenance{Source: "mock", RetrievedAt: retrievedAtNow()}
}

// pick returns a deterministic pseudo-random element of options using rng.
func pick(rng *rand.Rand, options []string) string {
	return options[rng.Intn(len(options))]
}

// pickN returns n distinct deterministic pseudo-random elements of options (n <= len(options)).
func pickN(rng *rand.Rand, options []string, n int) []string {
	pool := make([]string, len(options))
	copy(pool, options)
	rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	if n > len(pool) {
		n = len(pool)
	}
	return pool[:n]
}

// BasicInfo is the mock payload for the basic-info endpoint.
type BasicInfo struct {
	Provenance
	Company           string `json:"company"`
	Jurisdiction      string `json:"jurisdiction"`
	CompanyNumber     string `json:"company_number"`
	IncorporationDate string `json:"incorporation_date"`
}

var jurisdictions = []string{"Delaware, USA", "England & Wales, UK", "Singapore", "Cayman Islands", "Ireland", "Estonia"}

func basicInfoFor(company string) (BasicInfo, float64) {
	rng := rngFor(company, "basic-info")
	jurisdiction := pick(rng, jurisdictions)
	number := fmt.Sprintf("%08d", rng.Intn(100_000_000))
	year := 1995 + rng.Intn(30)
	month := 1 + rng.Intn(12)
	day := 1 + rng.Intn(28)
	payload := BasicInfo{
		Provenance:        provenanceFor(),
		Company:           company,
		Jurisdiction:      jurisdiction,
		CompanyNumber:     number,
		IncorporationDate: fmt.Sprintf("%04d-%02d-%02d", year, month, day),
	}
	return payload, confidenceByName["basic-info"]
}

// Industry is the mock payload for the industry endpoint.
type Industry struct {
	Provenance
	Company string   `json:"company"`
	Tags    []string `json:"tags"`
	NAICS   string   `json:"naics_code"`
}

var industryTags = []string{
	"software", "fintech", "e-commerce", "healthcare", "logistics",
	"energy", "media", "telecom", "manufacturing", "biotech",
	"cybersecurity", "real-estate", "education", "gaming", "agriculture",
}

func industryFor(company string) (Industry, float64) {
	rng := rngFor(company, "industry")
	tags := pickN(rng, industryTags, 2+rng.Intn(2))
	naics := fmt.Sprintf("%06d", 110000+rng.Intn(890000))
	payload := Industry{
		Provenance: provenanceFor(),
		Company:    company,
		Tags:       tags,
		NAICS:      naics,
	}
	return payload, confidenceByName["industry"]
}

// TopProducts is the mock payload for the top-products endpoint.
type TopProducts struct {
	Provenance
	Company  string   `json:"company"`
	Products []string `json:"products"`
}

var productCatalog = []string{
	"Cloud Platform", "Mobile App", "Analytics Suite", "Payments Gateway", "Developer SDK",
	"Customer Portal", "AI Assistant", "Data Warehouse", "Marketplace", "Subscription Plans",
	"Enterprise Edition", "Browser Extension", "API Gateway", "Logistics Tracker", "Security Suite",
}

func topProductsFor(company string) (TopProducts, float64) {
	rng := rngFor(company, "top-products")
	products := pickN(rng, productCatalog, 3+rng.Intn(3))
	payload := TopProducts{
		Provenance: provenanceFor(),
		Company:    company,
		Products:   products,
	}
	return payload, confidenceByName["top-products"]
}

// ReviewsSummary is the mock payload for the reviews-summary endpoint.
type ReviewsSummary struct {
	Provenance
	Company     string   `json:"company"`
	AvgRating   float64  `json:"avg_rating"`
	SampleCount int      `json:"sample_count"`
	Highlights  []string `json:"highlights"`
}

var reviewHighlights = []string{
	"Responsive customer support",
	"Intuitive user interface",
	"Occasional reliability issues",
	"Strong value for money",
	"Steep learning curve for new users",
	"Frequent feature updates",
	"Slow response times during peak hours",
	"Excellent onboarding documentation",
}

func reviewsSummaryFor(company string) (ReviewsSummary, float64) {
	rng := rngFor(company, "reviews-summary")
	avgRating := round2(2.5 + rng.Float64()*(4.8-2.5))
	sampleCount := 50 + rng.Intn(4951)
	highlights := pickN(rng, reviewHighlights, 3)
	payload := ReviewsSummary{
		Provenance:  provenanceFor(),
		Company:     company,
		AvgRating:   avgRating,
		SampleCount: sampleCount,
		Highlights:  highlights,
	}
	return payload, confidenceByName["reviews-summary"]
}

// Competitors is the mock payload for the competitors endpoint.
type Competitors struct {
	Provenance
	Company string   `json:"company"`
	Names   []string `json:"names"`
}

var competitorPool = []string{
	"Northwind Systems", "Acme Corp", "Globex Inc", "Initech", "Umbrella Group",
	"Stark Industries", "Wayne Enterprises", "Hooli", "Soylent Corp", "Massive Dynamic",
	"Cyberdyne Systems", "Pied Piper", "Aperture Labs", "Tyrell Corporation", "Wonka Industries",
}

func competitorsFor(company string) (Competitors, float64) {
	rng := rngFor(company, "competitors")
	names := pickN(rng, competitorPool, 3+rng.Intn(3))
	payload := Competitors{
		Provenance: provenanceFor(),
		Company:    company,
		Names:      names,
	}
	return payload, confidenceByName["competitors"]
}

// NewsSentiment is the mock payload for the news-sentiment endpoint.
type NewsSentiment struct {
	Provenance
	Company         string   `json:"company"`
	Score           float64  `json:"score"`
	SampleHeadlines []string `json:"sample_headlines"`
}

var newsHeadlineTemplates = []string{
	"%s announces quarterly results amid market speculation",
	"%s expands into new regional markets",
	"Analysts weigh in on %s's latest strategic shift",
	"%s faces scrutiny over recent operational changes",
	"%s unveils new product roadmap at industry event",
	"Investors react to %s's leadership announcement",
	"%s partners with major industry player",
	"%s under the spotlight after viral customer story",
}

func newsSentimentFor(company string) (NewsSentiment, float64) {
	rng := rngFor(company, "news-sentiment")
	score := round2(-1 + rng.Float64()*2)
	templates := pickN(rng, newsHeadlineTemplates, 3)
	headlines := make([]string, 0, 3)
	for _, t := range templates {
		headlines = append(headlines, fmt.Sprintf(t, company))
	}
	payload := NewsSentiment{
		Provenance:      provenanceFor(),
		Company:         company,
		Score:           score,
		SampleHeadlines: headlines,
	}
	return payload, confidenceByName["news-sentiment"]
}

// GrowthRate is the mock payload for the growth-rate endpoint.
type GrowthRate struct {
	Provenance
	Company string  `json:"company"`
	YoYPct  float64 `json:"yoy_pct"`
}

func growthRateFor(company string) (GrowthRate, float64) {
	rng := rngFor(company, "growth-rate")
	yoy := round2(-10 + rng.Float64()*50) // -10%..+40%
	payload := GrowthRate{
		Provenance: provenanceFor(),
		Company:    company,
		YoYPct:     yoy,
	}
	return payload, confidenceByName["growth-rate"]
}

// RevenueEstimate is the mock payload for the revenue-estimate endpoint.
type RevenueEstimate struct {
	Provenance
	Company  string  `json:"company"`
	Low      float64 `json:"low"`
	High     float64 `json:"high"`
	Currency string  `json:"currency"`
}

func revenueEstimateFor(company string) (RevenueEstimate, float64) {
	rng := rngFor(company, "revenue-estimate")
	// Pick a base order of magnitude (in USD), then a low/high band around it.
	magnitude := []float64{1e6, 5e6, 2.5e7, 1e8, 5e8, 1e9}[rng.Intn(6)]
	low := round2(magnitude * (0.5 + rng.Float64()*0.5))
	high := round2(low * (1.2 + rng.Float64()*0.8))
	payload := RevenueEstimate{
		Provenance: provenanceFor(),
		Company:    company,
		Low:        low,
		High:       high,
		Currency:   "USD",
	}
	return payload, confidenceByName["revenue-estimate"]
}

// SecurityIncidentRecord is a single mock security incident entry.
type SecurityIncidentRecord struct {
	Date     string `json:"date"`
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
}

// SecurityIncidents is the mock payload for the security-incidents endpoint.
type SecurityIncidents struct {
	Provenance
	Company   string                   `json:"company"`
	Incidents []SecurityIncidentRecord `json:"incidents"`
}

var securitySummaries = []string{
	"Unauthorized access to internal admin panel detected and contained",
	"Phishing campaign targeted employee credentials",
	"Misconfigured cloud storage bucket briefly exposed customer metadata",
	"Distributed denial-of-service attack disrupted service for several hours",
	"Third-party vendor breach exposed limited customer contact information",
	"Vulnerability in API authentication patched after responsible disclosure",
}

var severities = []string{"low", "medium", "high", "critical"}

func securityIncidentsFor(company string) (SecurityIncidents, float64) {
	rng := rngFor(company, "security-incidents")
	count := rng.Intn(4) // 0..3 incidents
	incidents := make([]SecurityIncidentRecord, 0, count)
	for i := 0; i < count; i++ {
		incidents = append(incidents, SecurityIncidentRecord{
			Date:     randomPastDate(rng),
			Summary:  pick(rng, securitySummaries),
			Severity: pick(rng, severities),
		})
	}
	payload := SecurityIncidents{
		Provenance: provenanceFor(),
		Company:    company,
		Incidents:  incidents,
	}
	return payload, confidenceByName["security-incidents"]
}

// LegalIssueRecord is a single mock legal-issue entry.
type LegalIssueRecord struct {
	Date    string `json:"date"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

// LegalIssues is the mock payload for the legal-issues endpoint.
type LegalIssues struct {
	Provenance
	Company string             `json:"company"`
	Issues  []LegalIssueRecord `json:"issues"`
}

var legalSummaries = []string{
	"Class-action lawsuit alleging deceptive marketing practices",
	"Patent infringement dispute with industry competitor",
	"Regulatory inquiry into data-handling practices",
	"Employment dispute regarding workplace policy",
	"Antitrust investigation into market conduct",
	"Contract dispute with former business partner",
}

var legalStatuses = []string{"filed", "ongoing", "settled", "dismissed", "appealed"}

func legalIssuesFor(company string) (LegalIssues, float64) {
	rng := rngFor(company, "legal-issues")
	count := rng.Intn(4) // 0..3 issues
	issues := make([]LegalIssueRecord, 0, count)
	for i := 0; i < count; i++ {
		issues = append(issues, LegalIssueRecord{
			Date:    randomPastDate(rng),
			Summary: pick(rng, legalSummaries),
			Status:  pick(rng, legalStatuses),
		})
	}
	payload := LegalIssues{
		Provenance: provenanceFor(),
		Company:    company,
		Issues:     issues,
	}
	return payload, confidenceByName["legal-issues"]
}

// randomPastDate returns a deterministic ISO date within the last ~5 years.
func randomPastDate(rng *rand.Rand) string {
	daysAgo := rng.Intn(5 * 365)
	t := time.Now().UTC().AddDate(0, 0, -daysAgo)
	return t.Format("2006-01-02")
}

// round2 rounds a float64 to 2 decimal places.
func round2(v float64) float64 {
	return float64(int64(v*100+sign(v)*0.5)) / 100
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}
