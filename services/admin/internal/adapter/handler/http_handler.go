package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/usecase"
)

type HTTPHandler struct {
	uc usecase.Usecase
}

func New(uc usecase.Usecase) *HTTPHandler {
	return &HTTPHandler{uc: uc}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/admin/status", h.Status)
	mux.HandleFunc("/api/v1/admin/sessions", h.ViewActiveSessions)
	mux.HandleFunc("/api/v1/admin/participants", h.ViewParticipants)
	mux.HandleFunc("/api/v1/admin/session-details", h.ViewSessionDetails)
	mux.HandleFunc("/api/v1/admin/kick", h.KickUser)
	mux.HandleFunc("/api/v1/admin/end-session", h.EndSession)
}

type adminActionRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

func (h *HTTPHandler) ViewActiveSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.uc.ViewActiveSessions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (h *HTTPHandler) ViewParticipants(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	parts, err := h.uc.ViewParticipants(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parts)
}

func (h *HTTPHandler) ViewSessionDetails(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	sess, err := h.uc.ViewSessionDetails(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

func (h *HTTPHandler) KickUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req adminActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.uc.KickUser(r.Context(), req.SessionID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "kicked"})
}

func (h *HTTPHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req adminActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.uc.EndSession(r.Context(), req.SessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "ended"})
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service":   "admin",
		"status":    "healthy",
		"port":      8087,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"service": "admin",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}
