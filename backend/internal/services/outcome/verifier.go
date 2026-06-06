package outcome

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"rationalgo/internal/models"
)

const openMeteoURL = "https://api.open-meteo.com/v1/forecast" +
	"?latitude=50.11&longitude=8.68" +
	"&hourly=precipitation_probability" +
	"&past_hours=1" +
	"&forecast_days=0"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Verify fetches near-real-time ground truth from OpenMeteo and computes an outcome score.
// paidForecastPrecipPct is the precipitation probability (0–100) returned by the paid API.
func Verify(ctx context.Context, paidForecastPrecipPct float64) (*models.OutcomeRecord, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", openMeteoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("outcome: verify: build request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("outcome: verify: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("outcome: verify: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("outcome: verify: open-meteo returned %d: %s", resp.StatusCode, body)
	}

	var apiResp struct {
		Hourly struct {
			PrecipitationProbability []float64 `json:"precipitation_probability"`
		} `json:"hourly"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("outcome: verify: parse response: %w", err)
	}
	if len(apiResp.Hourly.PrecipitationProbability) == 0 {
		return nil, fmt.Errorf("outcome: verify: no precipitation_probability data in response")
	}

	actual := apiResp.Hourly.PrecipitationProbability[0]
	score := 1.0 - math.Abs(paidForecastPrecipPct-actual)/100.0
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}

	return &models.OutcomeRecord{
		Predicted:   fmt.Sprintf("%.0f%% precipitation probability", paidForecastPrecipPct),
		Actual:      fmt.Sprintf("%.0f%% precipitation probability (observed)", actual),
		Score:       score,
		GroundTruth: "OpenMeteo historical (past_hours=1)",
	}, nil
}
