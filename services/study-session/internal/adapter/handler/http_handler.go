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
	mux.HandleFunc("/api/v1/study-session/create", h.CreateSession)
	mux.HandleFunc("/api/v1/study-session/join", h.JoinSession)
	mux.HandleFunc("/api/v1/study-session/leave", h.LeaveSession)
	mux.HandleFunc("/api/v1/study-session/end", h.EndSession)
	mux.HandleFunc("/api/v1/study-session/active", h.ListActiveSessions)
	mux.HandleFunc("/api/v1/study-session/participants", h.GetParticipants)
}

type createSessionRequest struct {
	Name             string `json:"name"`
	CreatorID        string `json:"creator_id"`
	ParticipantLimit int    `json:"participant_limit"`
}

type participantRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

func (h *HTTPHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	sess, err := h.uc.CreateSession(r.Context(), req.Name, req.CreatorID, req.ParticipantLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

func (h *HTTPHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req participantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.uc.JoinSession(r.Context(), req.SessionID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "joined"})
}

func (h *HTTPHandler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req participantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.uc.LeaveSession(r.Context(), req.SessionID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "left"})
}

func (h *HTTPHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req participantRequest
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
		"message": "Session closed",
	})
}

func (h *HTTPHandler) ListActiveSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.uc.ListActiveSession(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (h *HTTPHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	participants, err := h.uc.GetParticipants(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
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
