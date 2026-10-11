package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

// HTTPHandler is the internal REST adapter. Browsers reach Study Session
// through the API Gateway over gRPC; these routes remain for service-to-
// service callers (the Admin service's kick / close) and for curl-friendly
// debugging. Shapes are unchanged from Phase 1.
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
	DisplayName      string `json:"display_name"`
	ParticipantLimit int    `json:"participant_limit"`
}

type participantRequest struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Reason      string `json:"reason"`
}

func (h *HTTPHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if !decodePost(w, r, &req) {
		return
	}
	sess, err := h.uc.CreateSession(r.Context(), req.Name, req.CreatorID, req.DisplayName, req.ParticipantLimit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sess)
}

func (h *HTTPHandler) JoinSession(w http.ResponseWriter, r *http.Request) {
	var req participantRequest
	if !decodePost(w, r, &req) {
		return
	}
	sess, err := h.uc.JoinSession(r.Context(), req.SessionID, req.UserID, req.DisplayName)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "joined", "session": sess})
}

// LeaveSession removes a participant. Its caller is the Admin service's
// kick (UC-04 S-2), so a missing reason means KICKED.
func (h *HTTPHandler) LeaveSession(w http.ResponseWriter, r *http.Request) {
	var req participantRequest
	if !decodePost(w, r, &req) {
		return
	}
	if req.Reason == "" {
		req.Reason = events.ReasonKicked
	}
	res, err := h.uc.LeaveSession(r.Context(), req.SessionID, req.UserID, req.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	status := "left"
	if !res.Left {
		status = "already_left"
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": status, "session_ended": res.SessionEnded})
}

// EndSession closes a room. Its caller is the Admin service's close
// (UC-04 S-3), so a missing reason means ADMIN_CLOSED.
func (h *HTTPHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	var req participantRequest
	if !decodePost(w, r, &req) {
		return
	}
	sess, err := h.uc.EndSession(r.Context(), req.SessionID, req.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": sess, "message": "Session closed"})
}

func (h *HTTPHandler) ListActiveSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.uc.ListActiveSession(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *HTTPHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	participants, err := h.uc.GetParticipants(r.Context(), r.URL.Query().Get("session_id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, participants)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service":   "study-session",
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "study-session",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}

func decodePost(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return false
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	msg := "internal error"
	switch {
	case errors.Is(err, domain.ErrInvalid):
		code, msg = http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrNotFound):
		code, msg = http.StatusNotFound, "room not found"
	case errors.Is(err, domain.ErrSessionEnded):
		code, msg = http.StatusConflict, "this room has ended"
	case errors.Is(err, domain.ErrSessionFull):
		code, msg = http.StatusConflict, "this room is full"
	case errors.Is(err, domain.ErrAlreadyInSession):
		code, msg = http.StatusConflict, "you are already in a room"
	default:
		log.Printf("study-session: internal error: %v", err)
	}
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
