package research

import (
	"encoding/json"
	"net/http"
	"strings"
)

// maxCompanyLen caps the accepted length of the company query parameter.
const maxCompanyLen = 200

// sanitizeCompany trims whitespace and caps the length of the company input.
func sanitizeCompany(raw string) string {
	c := strings.TrimSpace(raw)
	if len(c) > maxCompanyLen {
		c = c[:maxCompanyLen]
	}
	return c
}

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// missingCompanyError is the standard error returned when the company parameter is absent.
func missingCompanyError() *ApiError {
	return &ApiError{Code: "missing_company", Message: "company parameter is required"}
}

// serve handles the common request lifecycle for a priced research endpoint: read/sanitize the
// company param, and either respond 400 with a missing_company error or build+respond with the payload.
func serve[T any](w http.ResponseWriter, r *http.Request, endpointName string, build func(company string) (T, float64)) {
	meta, _ := Find(endpointName)
	company := sanitizeCompany(r.URL.Query().Get("company"))

	if company == "" {
		writeJSON(w, http.StatusBadRequest, ApiResponse[T]{
			Data:       nil,
			Metadata:   meta,
			Confidence: 0,
			Error:      missingCompanyError(),
		})
		return
	}

	payload, confidence := build(company)
	writeJSON(w, http.StatusOK, ApiResponse[T]{
		Data:       &payload,
		Metadata:   meta,
		Confidence: confidence,
	})
}

// BasicInfoHandler serves the basic-info endpoint.
func BasicInfoHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "basic-info", basicInfoFor)
}

// IndustryHandler serves the industry endpoint.
func IndustryHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "industry", industryFor)
}

// TopProductsHandler serves the top-products endpoint.
func TopProductsHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "top-products", topProductsFor)
}

// ReviewsSummaryHandler serves the reviews-summary endpoint.
func ReviewsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "reviews-summary", reviewsSummaryFor)
}

// CompetitorsHandler serves the competitors endpoint.
func CompetitorsHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "competitors", competitorsFor)
}

// NewsSentimentHandler serves the news-sentiment endpoint.
func NewsSentimentHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "news-sentiment", newsSentimentFor)
}

// GrowthRateHandler serves the growth-rate endpoint.
func GrowthRateHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "growth-rate", growthRateFor)
}

// RevenueEstimateHandler serves the revenue-estimate endpoint.
func RevenueEstimateHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "revenue-estimate", revenueEstimateFor)
}

// SecurityIncidentsHandler serves the security-incidents endpoint.
func SecurityIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "security-incidents", securityIncidentsFor)
}

// LegalIssuesHandler serves the legal-issues endpoint.
func LegalIssuesHandler(w http.ResponseWriter, r *http.Request) {
	serve(w, r, "legal-issues", legalIssuesFor)
}

// PricingHandler serves the unprotected pricing-discovery endpoint.
func PricingHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"pricing": Pricing})
}

// Handlers maps each priced endpoint's path to its handler (excludes /pricing, wired separately).
func Handlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/company/basic-info":         BasicInfoHandler,
		"/company/industry":           IndustryHandler,
		"/company/top-products":       TopProductsHandler,
		"/company/reviews-summary":    ReviewsSummaryHandler,
		"/company/competitors":        CompetitorsHandler,
		"/company/news-sentiment":     NewsSentimentHandler,
		"/company/growth-rate":        GrowthRateHandler,
		"/company/revenue-estimate":   RevenueEstimateHandler,
		"/company/security-incidents": SecurityIncidentsHandler,
		"/company/legal-issues":       LegalIssuesHandler,
	}
}
