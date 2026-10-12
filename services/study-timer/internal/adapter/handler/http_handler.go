package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

// HTTPHandler is the internal REST adapter (service-to-service and curl).
// Browsers go through the API Gateway, which calls the gRPC handler; the
// Account service reads /statistics from here.
type HTTPHandler struct {
	uc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
	return &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/study-timer/status", h.Status)
	mux.HandleFunc("/api/v1/study-timer/state", h.State)
	mux.HandleFunc("/api/v1/study-timer/start", h.Start)
	mux.HandleFunc("/api/v1/study-timer/pause", h.action(h.uc.PauseTimer))
	mux.HandleFunc("/api/v1/study-timer/resume", h.action(h.uc.ResumeTimer))
	mux.HandleFunc("/api/v1/study-timer/stop", h.action(h.uc.StopTimer))
	mux.HandleFunc("/api/v1/study-timer/reset", h.action(h.uc.ResetTimer))
	mux.HandleFunc("/api/v1/study-timer/complete", h.action(h.uc.CompleteCycle))
	mux.HandleFunc("/api/v1/study-timer/skip-rest", h.action(h.uc.SkipRest))
	mux.HandleFunc("/api/v1/study-timer/setting", h.Setting)
	mux.HandleFunc("/api/v1/study-timer/statistics", h.Statistics)
}

type timerRequest struct {
	SessionID       string `json:"session_id"`
	UserID          string `json:"user_id"`
	Phase           string `json:"phase"`
	DurationMinutes int    `json:"duration_minutes"`
}

type timerSettingRequest struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	WorkMinutes int    `json:"work_minutes"`
	RestMinutes int    `json:"rest_minutes"`
}

// ViewJSON is a timer as JSON, field for field the same as the gateway's
// /api/timer/* answers.
type ViewJSON struct {
	SessionID            string `json:"session_id"`
	UserID               string `json:"user_id"`
	State                string `json:"state"`
	Status               string `json:"status"`
	Phase                string `json:"phase"`
	WorkMinutes          int    `json:"work_minutes"`
	RestMinutes          int    `json:"rest_minutes"`
	CurrentCycle         int    `json:"current_cycle"`
	FocusSeconds         int    `json:"focus_seconds"`
	LastUpdated          string `json:"last_updated"`
	DurationSeconds      int    `json:"duration_seconds"`
	RemainingSeconds     int    `json:"remaining_seconds"`
	CycleID              string `json:"cycle_id"`
	StartedAt            string `json:"started_at"`
	PausedTotalSeconds   int    `json:"paused_total_seconds"`
	LastCompletedCycleID string `json:"last_completed_cycle_id"`
	LastCompletedAt      string `json:"last_completed_at"`
	MinWorkMinutes       int    `json:"min_work_minutes"`
	MaxWorkMinutes       int    `json:"max_work_minutes"`
	MinRestMinutes       int    `json:"min_rest_minutes"`
	MaxRestMinutes       int    `json:"max_rest_minutes"`
	MaxPauseMinutes      int    `json:"max_pause_minutes"`
}

func toViewJSON(v *domain.View) ViewJSON {
	p := ToProto(v)
	return ViewJSON{
		SessionID: p.GetSessionId(), UserID: p.GetUserId(), State: p.GetState(),
		Status: string(v.Status), Phase: string(v.Phase),
		WorkMinutes: v.Settings.WorkMinutes, RestMinutes: v.Settings.RestMinutes,
		CurrentCycle: v.CompletedWork, FocusSeconds: v.FocusSeconds, LastUpdated: p.GetLastUpdated(),
		DurationSeconds: v.DurationSeconds, RemainingSeconds: v.RemainingSeconds,
		CycleID: v.CycleID, StartedAt: p.GetStartedAt(), PausedTotalSeconds: v.PausedTotalSeconds,
		LastCompletedCycleID: v.LastCompletedID, LastCompletedAt: p.GetLastCompletedAt(),
		MinWorkMinutes: v.Limits.MinWorkMinutes, MaxWorkMinutes: v.Limits.MaxWorkMinutes,
		MinRestMinutes: v.Limits.MinRestMinutes, MaxRestMinutes: v.Limits.MaxRestMinutes,
		MaxPauseMinutes: v.Limits.MaxPauseMinutes,
	}
}

// writeTimer sends the timer, or maps a usecase error to its HTTP status.
func writeTimer(w http.ResponseWriter, v *domain.View, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidState), errors.Is(err, domain.ErrNotFinished):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrTimerClosed):
		http.Error(w, err.Error(), http.StatusForbidden)
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toViewJSON(v))
	}
}

func (h *HTTPHandler) State(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	v, err := h.uc.GetTimer(r.Context(), q.Get("session_id"), q.Get("user_id"))
	writeTimer(w, v, err)
}

func (h *HTTPHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	v, err := h.uc.StartTimer(r.Context(), req.SessionID, req.UserID, domain.TimerPhase(req.Phase), req.DurationMinutes)
	writeTimer(w, v, err)
}

// action serves one POST {session_id, user_id} endpoint.
func (h *HTTPHandler) action(call func(ctx context.Context, sessionID, userID string) (*domain.View, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req timerRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		v, err := call(r.Context(), req.SessionID, req.UserID)
		writeTimer(w, v, err)
	}
}

func (h *HTTPHandler) Setting(w http.ResponseWriter, r *http.Request) {
	var req timerSettingRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	v, err := h.uc.UpdateTimerSetting(r.Context(), req.SessionID, req.UserID, req.WorkMinutes, req.RestMinutes)
	writeTimer(w, v, err)
}

func (h *HTTPHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.uc.TimerStatistics(r.Context(), r.URL.Query().Get("user_id"))
	if errors.Is(err, domain.ErrInvalid) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "study-timer",
		"status":    "healthy",
		"port":      8084,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "study-timer",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}
