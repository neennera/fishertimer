package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/adapter/middleware"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
	"github.com/neennera/fishertimer/services/admin/internal/usecase"
)

type HTTPHandler struct {
	uc   usecase.Usecase
	auth *middleware.Verifier
}

func New(uc usecase.Usecase, auth *middleware.Verifier) *HTTPHandler {
	return &HTTPHandler{uc: uc, auth: auth}
}

// RegisterRoutes mounts the admin API. Everything except /health and
// /status requires a session JWT with role = ADMIN.
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/v1/admin/status", h.Status)

	admin := func(pattern string, fn http.HandlerFunc) {
		mux.Handle(pattern, h.auth.RequireAdmin(fn))
	}
	// Monitoring (UC-04)
	admin("GET /api/v1/admin/sessions", h.ViewActiveSessions)
	admin("GET /api/v1/admin/sessions/{id}", h.ViewSessionDetails)
	admin("GET /api/v1/admin/sessions/{id}/participants", h.ViewParticipants)
	// Moderation (UC-10)
	admin("POST /api/v1/admin/sessions/{id}/participants/{userId}/kick", h.KickUser)
	admin("POST /api/v1/admin/sessions/{id}/end", h.EndSession)
	admin("GET /api/v1/admin/logs", h.ViewLogs)
}

// moderationRequest is the optional body of kick / end; the room and user
// come from the path.
type moderationRequest struct {
	Reason string `json:"reason"`
}

func (h *HTTPHandler) ViewActiveSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.uc.ViewActiveSessions(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	if sessions == nil {
		sessions = []domain.SessionOverview{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *HTTPHandler) ViewParticipants(w http.ResponseWriter, r *http.Request) {
	parts, err := h.uc.ViewParticipants(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if parts == nil {
		parts = []domain.SessionParticipantOverview{}
	}
	writeJSON(w, http.StatusOK, parts)
}

func (h *HTTPHandler) ViewSessionDetails(w http.ResponseWriter, r *http.Request) {
	detail, err := h.uc.ViewSessionDetails(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if detail.Participants == nil {
		detail.Participants = []domain.SessionParticipantOverview{}
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *HTTPHandler) KickUser(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeModeration(w, r)
	if !ok {
		return
	}
	res, err := h.uc.KickUser(r.Context(), adminID(r), r.PathValue("id"), r.PathValue("userId"), req.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "kicked",
		"session_ended": res.SessionEnded,
		"log":           res.Log,
	})
}

func (h *HTTPHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeModeration(w, r)
	if !ok {
		return
	}
	res, err := h.uc.EndSession(r.Context(), adminID(r), r.PathValue("id"), req.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ended",
		"session_ended": res.SessionEnded,
		"log":           res.Log,
	})
}

func (h *HTTPHandler) ViewLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.uc.ViewLogs(r.Context(), r.URL.Query().Get("action"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service":   "admin",
		"status":    "healthy",
		"port":      8087,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HTTPHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "admin",
		"layer":   "adapter.handler",
		"ready":   true,
	})
}

func adminID(r *http.Request) string {
	if c, ok := middleware.AdminFrom(r.Context()); ok {
		return c.UserID
	}
	return ""
}

// decodeModeration reads the optional JSON body; an empty body is allowed
// because the reason is optional.
func decodeModeration(w http.ResponseWriter, r *http.Request) (moderationRequest, bool) {
	var req moderationRequest
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req)
	if err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return req, false
	}
	return req, true
}

func writeError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	msg := "internal error"
	switch {
	case errors.Is(err, domain.ErrInvalid):
		code, msg = http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrNotFound):
		code, msg = http.StatusNotFound, "room not found or no longer active"
	case errors.Is(err, domain.ErrStateChanged):
		code, msg = http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrUpstream):
		code, msg = http.StatusBadGateway, "study session service is unavailable, please retry"
		log.Printf("admin: upstream error: %v", err)
	default:
		log.Printf("admin: internal error: %v", err)
	}
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
