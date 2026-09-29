package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/usecase"
)

type HTTPHandler struct {
	uc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
	return &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/reward/status", h.Status)
	mux.HandleFunc("/api/v1/reward/award", h.AwardReward)
	mux.HandleFunc("/api/v1/reward/rewards", h.ViewRewards)
	mux.HandleFunc("/api/v1/reward/all-rewards", h.AllRewards)
	mux.HandleFunc("/api/v1/reward/last-update", h.GetLastUpdate)
}

// ViewRewards returns rewards for a specific user (or all users if user_id is omitted).
func (h *HTTPHandler) ViewRewards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	rewards, err := h.uc.GetUserInventory(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rewards)
}

// AllRewards returns all rewards across all users (used by leaderboard for ranking).
func (h *HTTPHandler) AllRewards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rewards, err := h.uc.ListAllRewards(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rewards)
}

// GetLastUpdate returns the timestamp of the most recent reward change.
// Leaderboard uses this to validate its in-memory cache without fetching all data.
func (h *HTTPHandler) GetLastUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ts, err := h.uc.GetLastUpdate(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"reward_last_update": ts.UTC().Format(time.RFC3339Nano),
	})
}

type awardRequest struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

func (h *HTTPHandler) AwardReward(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req awardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	reward, err := h.uc.AwardReward(r.Context(), req.UserID, req.Reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reward)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "reward",
		"status":    "healthy",
		"port":      8085,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "reward",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}
