package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

type HTTPHandler struct {
	uc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
	return &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/study-session/status", h.Status)
	mux.HandleFunc("/api/v1/study-session/end", h.EndSession)
}

type endSessionRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

func (h *HTTPHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req endSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sess, err := h.uc.EndSession(r.Context(), req.SessionID, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"session": sess,
		"message": "Session ended and AwardReward triggered",
	})
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "study-session",
		"status":    "healthy",
		"port":      8083,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "study-session",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}
