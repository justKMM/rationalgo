package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"rationalgo/internal/config"
	"rationalgo/internal/repository"
	"rationalgo/internal/scenario"
	"rationalgo/internal/services/algorand"
	"rationalgo/internal/services/outcome"
	"rationalgo/internal/services/policy"
	"rationalgo/internal/services/reasoning"
	"rationalgo/internal/services/vendor"
	"rationalgo/internal/services/x402"
)

// Server is the HTTP API for the audit dashboard.
type Server struct {
	cfg          config.Config
	store        *repository.Store
	deps         scenario.Deps
	reasoningSvc *reasoning.Service
	mux          *http.ServeMux
}

// NewServer creates an API server with seeded in-memory state.
func NewServer(cfg config.Config, reasoningSvc *reasoning.Service) *Server {
	store := repository.NewStore()
	s := &Server{
		cfg:          cfg,
		store:        store,
		deps:         buildDeps(cfg, store, reasoningSvc),
		reasoningSvc: reasoningSvc,
		mux:          http.NewServeMux(),
	}
	s.routes()
	return s
}

func buildDeps(cfg config.Config, store *repository.Store, reasoningSvc *reasoning.Service) scenario.Deps {
	vendors := vendor.NewService()
	state := store.State()

	var algClient scenario.AlgorandCommitter
	if client, err := algorand.NewClient(cfg); err == nil {
		algClient = client
	}

	return scenario.Deps{
		Reasoning:  reasoningSvc,
		Outcome:    outcome.NewService(),
		Algorand:   algClient,
		X402:       x402.NewService(cfg),
		Store:      store,
		Vendors:    vendors.GetDemoWeatherVendors,
		Policy:     policy.Evaluate,
		Allowed:    func() []string { return state.AllowedVendors },
		PriceHist:  vendors.GetPriceHistory,
		Inject:     policy.InjectAnomalyPrice,
		AgentID:    state.Agent,
		DailySpent: state.Spent,
		DailyLimit: state.DailyLimit,
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/state", s.handleState)
	s.mux.HandleFunc("GET /api/decisions", s.handleDecisions)
	s.mux.HandleFunc("POST /api/state/reset", s.handleReset)
	s.mux.HandleFunc("POST /api/decide", s.handleDecide)
	s.mux.HandleFunc("POST /api/scenario/run", s.handleScenarioRun)
}

// ListenAndServe starts the HTTP server with CORS for local frontend dev.
func (s *Server) ListenAndServe() error {
	addr := s.cfg.HTTPAddr
	log.Printf("RationAlgo API listening on %s", addr)
	return http.ListenAndServe(addr, s.withCORS(s.mux))
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"phase":  "2",
	})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.State())
}

func (s *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.State().Decisions)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.store.Reset()
	s.deps = buildDeps(s.cfg, s.store, s.reasoningSvc)
	writeJSON(w, http.StatusOK, s.store.State())
}

func (s *Server) handleScenarioRun(w http.ResponseWriter, r *http.Request) {
	scenarioType := scenario.ScenarioNormal
	if r.URL.Query().Get("scenario") == "anomaly" {
		scenarioType = scenario.ScenarioAnomaly
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	events, err := scenario.Run(r.Context(), scenarioType, s.deps)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	for event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
		flusher.Flush()
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintf(w, `{"error":%q}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
}

// decideRequest is the body for POST /api/decide.
type decideRequest struct {
	Intent    string `json:"intent"`
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
}

// handleDecide runs the full reasoning pipeline and returns a DecisionRecord.
func (s *Server) handleDecide(w http.ResponseWriter, r *http.Request) {
	var req decideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Intent == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "intent is required"})
		return
	}
	if req.AgentID == "" {
		req.AgentID = "rationalgo-agent-01"
	}
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("sess-%d", time.Now().UnixNano())
	}

	vendorSvc := vendor.NewService()
	vendors := vendorSvc.GetDemoWeatherVendors()
	state := s.store.State()

	// Evaluate policy against the primary (paid) vendor with default budget parameters.
	pol := policy.Evaluate(
		vendors[0],
		vendors[0].PriceEURQ,
		0.0,
		10.0,
		state.AllowedVendors,
		vendorSvc.GetPriceHistory(),
	)

	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()

	record, err := s.reasoningSvc.GenerateDecision(ctx, req.AgentID, req.SessionID, req.Intent, vendors, pol)
	if err != nil {
		log.Printf("reasoning: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, record)
}
