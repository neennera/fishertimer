package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

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
	mux.HandleFunc("/api/v1/study-timer/pause", h.Pause)
	mux.HandleFunc("/api/v1/study-timer/resume", h.Resume)
	mux.HandleFunc("/api/v1/study-timer/stop", h.Stop)
	mux.HandleFunc("/api/v1/study-timer/reset", h.Reset)
	mux.HandleFunc("/api/v1/study-timer/complete", h.Complete)
	mux.HandleFunc("/api/v1/study-timer/skip-rest", h.SkipRest)
	mux.HandleFunc("/api/v1/study-timer/setting", h.Setting)
	mux.HandleFunc("/api/v1/study-timer/statistics", h.Statistics)
}

type timerRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

type timerSettingRequest struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	WorkMinutes int    `json:"work_minutes"`
	RestMinutes int    `json:"rest_minutes"`
}

type timerResponse struct {
	SessionID        string `json:"session_id"`
	UserID           string `json:"user_id"`
	Status           string `json:"status"`
	Phase            string `json:"phase"`
	WorkMinutes      int    `json:"work_minutes"`
	RestMinutes      int    `json:"rest_minutes"`
	CurrentCycle     int    `json:"current_cycle"`
	LastUpdated      string `json:"last_updated"`
	DurationSeconds  int    `json:"duration_seconds"`
	RemainingSeconds int    `json:"remaining_seconds"`
}

func newTimerResponse(t *domain.TimerState, now time.Time) timerResponse {
	resp := timerResponse{
		SessionID:        t.SessionID,
		UserID:           t.UserID,
		Status:           string(t.Status),
		Phase:            string(t.Phase),
		WorkMinutes:      t.WorkMinutes,
		RestMinutes:      t.RestMinutes,
		CurrentCycle:     t.CurrentCycle,
		DurationSeconds:  int(t.Duration().Seconds()),
		RemainingSeconds: t.RemainingSeconds(now),
	}
	if !t.LastUpdated.IsZero() {
		resp.LastUpdated = t.LastUpdated.Format(time.RFC3339)
	}
	return resp
}

// writeTimer sends the timer state, or maps a usecase error to its HTTP status.
func writeTimer(w http.ResponseWriter, t *domain.TimerState, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, domain.ErrInvalidState):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newTimerResponse(t, time.Now().UTC()))
}

func (h *HTTPHandler) State(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	st, err := h.uc.GetTimer(r.Context(), q.Get("session_id"), q.Get("user_id"))
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.StartTimer(r.Context(), req.SessionID, req.UserID)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Pause(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.PauseTimer(r.Context(), req.SessionID, req.UserID)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Resume(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.ResumeTimer(r.Context(), req.SessionID, req.UserID)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Stop(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.StopTimer(r.Context(), req.SessionID, req.UserID)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.ResetTimer(r.Context(), req.SessionID, req.UserID)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.CompleteCycle(r.Context(), req.SessionID, req.UserID)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) SkipRest(w http.ResponseWriter, r *http.Request) {
	var req timerRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.SkipRest(r.Context(), req.SessionID, req.UserID)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Setting(w http.ResponseWriter, r *http.Request) {
	var req timerSettingRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	st, err := h.uc.UpdateTimerSetting(r.Context(), req.SessionID, req.UserID, req.WorkMinutes, req.RestMinutes)
	writeTimer(w, st, err)
}

func (h *HTTPHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	stats, err := h.uc.TimerStatistics(r.Context(), userID)
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
