package research

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// callHandler issues a GET request with the given company query param and decodes the JSON body.
func callHandler[T any](t *testing.T, h http.HandlerFunc, company string) ApiResponse[T] {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/?company="+company, nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	var resp ApiResponse[T]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestBasicInfoDeterministic(t *testing.T) {
	a := callHandler[BasicInfo](t, BasicInfoHandler, "Acme")
	b := callHandler[BasicInfo](t, BasicInfoHandler, "Acme")

	if a.Data == nil || b.Data == nil {
		t.Fatalf("expected data, got nil (a=%v b=%v)", a.Data, b.Data)
	}
	if a.Data.Jurisdiction != b.Data.Jurisdiction ||
		a.Data.CompanyNumber != b.Data.CompanyNumber ||
		a.Data.IncorporationDate != b.Data.IncorporationDate {
		t.Fatalf("expected identical data fields across calls, got %+v vs %+v", a.Data, b.Data)
	}
	if a.Confidence != b.Confidence || a.Confidence != confidenceByName["basic-info"] {
		t.Fatalf("unexpected confidence: %v vs %v", a.Confidence, b.Confidence)
	}
	if a.Metadata.Path != "/company/basic-info" {
		t.Fatalf("unexpected metadata path: %s", a.Metadata.Path)
	}
}

func TestRevenueEstimateDeterministic(t *testing.T) {
	a := callHandler[RevenueEstimate](t, RevenueEstimateHandler, "Globex")
	b := callHandler[RevenueEstimate](t, RevenueEstimateHandler, "Globex")

	if a.Data == nil || b.Data == nil {
		t.Fatalf("expected data, got nil")
	}
	if a.Data.Low != b.Data.Low || a.Data.High != b.Data.High || a.Data.Currency != b.Data.Currency {
		t.Fatalf("expected identical revenue fields across calls, got %+v vs %+v", a.Data, b.Data)
	}

	// A different company name should (overwhelmingly likely) produce different values.
	c := callHandler[RevenueEstimate](t, RevenueEstimateHandler, "Initech")
	if a.Data.Low == c.Data.Low && a.Data.High == c.Data.High {
		t.Fatalf("expected different companies to produce different mock data")
	}
}

func TestMissingCompanyReturnsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	IndustryHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp ApiResponse[Industry]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data != nil {
		t.Fatalf("expected nil data, got %+v", resp.Data)
	}
	if resp.Error == nil || resp.Error.Code != "missing_company" {
		t.Fatalf("expected missing_company error, got %+v", resp.Error)
	}
	if resp.Confidence != 0 {
		t.Fatalf("expected zero confidence on error, got %v", resp.Confidence)
	}
}

func TestPricingHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/pricing", nil)
	rec := httptest.NewRecorder()
	PricingHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Pricing []EndpointMeta `json:"pricing"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Pricing) != 10 {
		t.Fatalf("expected 10 priced endpoints, got %d", len(body.Pricing))
	}
}

func TestHandlersMapCoversAllPricedEndpoints(t *testing.T) {
	handlers := Handlers()
	if len(handlers) != len(Pricing) {
		t.Fatalf("expected %d handlers, got %d", len(Pricing), len(handlers))
	}
	for _, meta := range Pricing {
		if _, ok := handlers[meta.Path]; !ok {
			t.Fatalf("missing handler for path %s", meta.Path)
		}
	}
	if _, ok := handlers["/pricing"]; ok {
		t.Fatalf("expected /pricing to be excluded from Handlers()")
	}
}

func TestSelectRespectsBudget(t *testing.T) {
	confidence := confidenceByName
	sel := Select(Pricing, confidence, 0.30)

	var total float64
	for _, s := range sel {
		total += s.Endpoint.PriceUSD
	}
	if total > 0.30+1e-9 {
		t.Fatalf("selection exceeds budget: total=%.2f", total)
	}
	if len(sel) == 0 {
		t.Fatalf("expected at least one selection within budget 0.30")
	}
	for i := 1; i < len(sel); i++ {
		if sel[i-1].Score < sel[i].Score {
			t.Fatalf("expected selections sorted by Score desc, got %v then %v", sel[i-1].Score, sel[i].Score)
		}
	}
}
