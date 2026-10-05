package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/reward/internal/domain"
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

// awardRequest is POST /api/v1/reward/award. The cycle length in minutes
// comes as work_minutes (preferred) or work_duration (what Study Timer
// sends). Both are pointers so a missing value is told apart from 0. Unknown
// fields such as Study Timer's "reason" are ignored.
type awardRequest struct {
	UserID           string `json:"user_id"`
	SessionID        string `json:"session_id"`
	CycleID          string `json:"cycle_id"`
	WorkMinutes      *int   `json:"work_minutes"`
	WorkDuration     *int   `json:"work_duration"`
	ParticipantCount int    `json:"participant_count"`
	DisplayName      string `json:"display_name"`
}

// workMinutes returns the cycle length from whichever field was sent. ok is
// false when neither was sent or both were sent with different values.
func (req awardRequest) workMinutes() (minutes int, ok bool, conflict bool) {
	switch {
	case req.WorkMinutes != nil && req.WorkDuration != nil:
		if *req.WorkMinutes != *req.WorkDuration {
			return 0, false, true
		}
		return *req.WorkMinutes, true, false
	case req.WorkMinutes != nil:
		return *req.WorkMinutes, true, false
	case req.WorkDuration != nil:
		return *req.WorkDuration, true, false
	}
	return 0, false, false
}

type awardResponse struct {
	CycleID        string                  `json:"cycle_id"`
	AlreadyAwarded bool                    `json:"already_awarded"`
	Rewards        []domain.UnlockedReward `json:"rewards"`
}

const (
	awardRequiredFields = "user_id, cycle_id and work_minutes (or work_duration) >= 0 are required"
	awardConflictingLen = "work_minutes and work_duration differ; send one, or the same value in both"
)

// AwardReward draws the rewards for a completed work cycle. 201 when rewards
// were stored, 200 when the cycle was already awarded or earned nothing.
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

	minutes, ok, conflict := req.workMinutes()
	if conflict {
		http.Error(w, awardConflictingLen, http.StatusBadRequest)
		return
	}
	if req.UserID == "" || req.CycleID == "" || !ok || minutes < 0 {
		http.Error(w, awardRequiredFields, http.StatusBadRequest)
		return
	}

	result, err := h.uc.AwardReward(r.Context(), usecase.AwardInput{
		UserID:           req.UserID,
		SessionID:        req.SessionID,
		CycleID:          req.CycleID,
		WorkMinutes:      minutes,
		ParticipantCount: req.ParticipantCount,
		DisplayName:      req.DisplayName,
	})
	if errors.Is(err, domain.ErrInvalid) {
		http.Error(w, awardRequiredFields, http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if !result.AlreadyAwarded && len(result.Rewards) > 0 {
		status = http.StatusCreated
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(awardResponse{
		CycleID:        result.CycleID,
		AlreadyAwarded: result.AlreadyAwarded,
		Rewards:        result.Rewards,
	})
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
