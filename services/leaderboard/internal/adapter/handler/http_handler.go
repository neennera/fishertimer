package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/leaderboard/internal/domain"
	"github.com/neennera/fishertimer/services/leaderboard/internal/usecase"
)

type HTTPHandler struct {
	uc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
	return &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.withCORS(h.Health))
	mux.HandleFunc("/api/v1/leaderboard/status", h.withCORS(h.Status))
	// UC-08 ViewLeaderboard: GET /api/v1/leaderboard?period={weekly|monthly|all-time}
	mux.HandleFunc("/api/v1/leaderboard", h.withCORS(h.ViewLeaderboard))
	// GetRanking: returns only the rankings array (consumed by the frontend).
	mux.HandleFunc("/api/v1/leaderboard/rankings", h.withCORS(h.GetRanking))
}

// withCORS wraps a handler to add permissive CORS headers for development.
func (h *HTTPHandler) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// ViewLeaderboard handles GET /api/v1/leaderboard?period=...
// Implements UC-08: S-1 cache check → S-3 compute if stale → return CachedRanking.
func (h *HTTPHandler) ViewLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "all-time"
	}

	result, err := h.uc.ViewLeaderboard(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// E-1: empty state — still return 200 with an empty rankings array.
	if result.Rankings == nil {
		result.Rankings = []domain.RankEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetRanking handles GET /api/v1/leaderboard/rankings?period=...
// Returns only the []RankEntry array (lighter payload for UI polling).
func (h *HTTPHandler) GetRanking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "all-time"
	}

	result, err := h.uc.ViewLeaderboard(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rankings := result.Rankings
	if rankings == nil {
		rankings = []domain.RankEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rankings)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "leaderboard",
		"status":    "healthy",
		"port":      8086,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "leaderboard",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}
