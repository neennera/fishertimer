package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/middleware"
)

// sessionCallTimeout bounds each gRPC call to Study Session.
const sessionCallTimeout = 5 * time.Second

// SessionHandler translates the browser's REST /api/session/* API into gRPC
// calls to the Study Session service (UC-01..03).
//
//	GET  /api/session/active                 -> ListActiveSession
//	GET  /api/session/me                     -> GetMySession
//	GET  /api/session/room?session_id=       -> GetSession + GetParticipants
//	POST /api/session/create    {name, participant_limit} -> CreateSession
//	POST /api/session/join      {session_id}  -> JoinSession
//	POST /api/session/leave     {session_id}  -> LeaveSession (reason LEFT)
//	POST /api/session/heartbeat {session_id}  -> Heartbeat
//
// The acting user is the signed-in one: X-User-Id / X-Display-Name, set by
// middleware.Verifier from the session JWT. Outside production, a request
// without a session may name its user in the body or query instead
// (user_id, display_name), so the room pages also work with mock auth.
//
// Errors answer {"error": "...", "code": "..."} with a code the page can
// branch on: INVALID, SIGNED_OUT, ROOM_NOT_FOUND, ROOM_ENDED, ROOM_FULL,
// ALREADY_IN_ROOM, UNAVAILABLE.
type SessionHandler struct {
	client       sessionv1.StudySessionServiceClient
	allowDevUser bool
}

func NewSessionHandler(client sessionv1.StudySessionServiceClient, allowDevUser bool) *SessionHandler {
	return &SessionHandler{client: client, allowDevUser: allowDevUser}
}

type sessionJSON struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	CreatorID        string `json:"creator_id"`
	ParticipantLimit int32  `json:"participant_limit"`
	ParticipantCount int32  `json:"participant_count"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	EndedAt          string `json:"ended_at,omitempty"`
	EndReason        string `json:"end_reason,omitempty"`
}

type participantJSON struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	JoinedAt    string `json:"joined_at"`
	LastSeenAt  string `json:"last_seen_at"`
}

type sessionBody struct {
	SessionID        string `json:"session_id"`
	Name             string `json:"name"`
	ParticipantLimit int32  `json:"participant_limit"`
	// Dev-only identity, ignored when a verified session is present.
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

func (h *SessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	action := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/session"), "/")

	wantMethod := http.MethodPost
	switch action {
	case "active", "me", "room":
		wantMethod = http.MethodGet
	case "create", "join", "leave", "heartbeat":
	default:
		writeSessionError(w, http.StatusNotFound, "NOT_FOUND", "unknown session endpoint")
		return
	}
	if r.Method != wantMethod {
		writeSessionError(w, http.StatusMethodNotAllowed, "INVALID", "use "+wantMethod)
		return
	}

	var body sessionBody
	if wantMethod == http.MethodPost {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&body); err != nil {
			writeSessionError(w, http.StatusBadRequest, "INVALID", "body must be JSON")
			return
		}
	} else {
		q := r.URL.Query()
		body.SessionID = q.Get("session_id")
		body.UserID = q.Get("user_id")
	}

	ctx, cancel := context.WithTimeout(r.Context(), sessionCallTimeout)
	defer cancel()

	if action == "active" {
		resp, err := h.client.ListActiveSession(ctx, &sessionv1.ListActiveSessionRequest{})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		sessions := make([]sessionJSON, 0, len(resp.GetSessions()))
		for _, s := range resp.GetSessions() {
			sessions = append(sessions, toSessionJSON(s))
		}
		writeSessionJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
		return
	}

	userID, displayName, ok := h.identity(r, body)
	if !ok {
		writeSessionError(w, http.StatusUnauthorized, "SIGNED_OUT", "sign in to use study rooms")
		return
	}

	switch action {
	case "me":
		resp, err := h.client.GetMySession(ctx, &sessionv1.GetMySessionRequest{UserId: userID})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		out := map[string]any{"in_session": resp.GetInSession()}
		if resp.GetInSession() {
			out["session"] = toSessionJSON(resp.GetSession())
			out["participant"] = toParticipantJSON(resp.GetParticipant())
		}
		writeSessionJSON(w, http.StatusOK, out)

	case "room":
		sess, err := h.client.GetSession(ctx, &sessionv1.GetSessionRequest{SessionId: body.SessionID})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		ps, err := h.client.GetParticipants(ctx, &sessionv1.GetParticipantsRequest{SessionId: body.SessionID})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		participants := make([]participantJSON, 0, len(ps.GetParticipants()))
		for _, p := range ps.GetParticipants() {
			participants = append(participants, toParticipantJSON(p))
		}
		writeSessionJSON(w, http.StatusOK, map[string]any{
			"session":      toSessionJSON(sess),
			"participants": participants,
		})

	case "create":
		sess, err := h.client.CreateSession(ctx, &sessionv1.CreateSessionRequest{
			Name:               body.Name,
			CreatorId:          userID,
			ParticipantLimit:   body.ParticipantLimit,
			CreatorDisplayName: displayName,
		})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		writeSessionJSON(w, http.StatusCreated, toSessionJSON(sess))

	case "join":
		resp, err := h.client.JoinSession(ctx, &sessionv1.JoinSessionRequest{
			SessionId:   body.SessionID,
			UserId:      userID,
			DisplayName: displayName,
		})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		writeSessionJSON(w, http.StatusOK, toSessionJSON(resp.GetSession()))

	case "leave":
		resp, err := h.client.LeaveSession(ctx, &sessionv1.LeaveSessionRequest{
			SessionId: body.SessionID,
			UserId:    userID,
			Reason:    "LEFT", // a browser can only leave on its own behalf
		})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		writeSessionJSON(w, http.StatusOK, map[string]any{
			"left":          resp.GetLeft(),
			"session_ended": resp.GetSessionEnded(),
			"session_id":    resp.GetSessionId(),
			"joined_at":     resp.GetJoinedAt(),
			"left_at":       resp.GetLeftAt(),
		})

	case "heartbeat":
		resp, err := h.client.Heartbeat(ctx, &sessionv1.HeartbeatRequest{SessionId: body.SessionID, UserId: userID})
		if err != nil {
			writeSessionRPCError(w, err)
			return
		}
		writeSessionJSON(w, http.StatusOK, map[string]any{"active": resp.GetActive(), "reason": resp.GetReason()})
	}
}

// identity returns the acting user: the verified session, or outside
// production the user named in the request.
func (h *SessionHandler) identity(r *http.Request, body sessionBody) (userID, displayName string, ok bool) {
	if id := r.Header.Get(middleware.HeaderUserID); id != "" {
		name, err := url.QueryUnescape(r.Header.Get(middleware.HeaderDisplayName))
		if err != nil {
			name = ""
		}
		return id, name, true
	}
	if h.allowDevUser && body.UserID != "" {
		return body.UserID, body.DisplayName, true
	}
	return "", "", false
}

func toSessionJSON(s *sessionv1.StudySessionResponse) sessionJSON {
	return sessionJSON{
		ID:               s.GetId(),
		Name:             s.GetName(),
		CreatorID:        s.GetCreatorId(),
		ParticipantLimit: s.GetParticipantLimit(),
		ParticipantCount: s.GetParticipantCount(),
		Status:           s.GetStatus(),
		CreatedAt:        s.GetCreatedAt(),
		EndedAt:          s.GetEndedAt(),
		EndReason:        s.GetEndReason(),
	}
}

func toParticipantJSON(p *sessionv1.ParticipantInfo) participantJSON {
	return participantJSON{
		UserID:      p.GetUserId(),
		DisplayName: p.GetDisplayName(),
		JoinedAt:    p.GetJoinedAt(),
		LastSeenAt:  p.GetLastSeenAt(),
	}
}

// writeSessionRPCError maps Study Session's gRPC codes (documented in
// session.proto) to an HTTP status and a stable error code.
func writeSessionRPCError(w http.ResponseWriter, err error) {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.InvalidArgument:
		writeSessionError(w, http.StatusBadRequest, "INVALID", st.Message())
	case codes.NotFound:
		writeSessionError(w, http.StatusNotFound, "ROOM_NOT_FOUND", st.Message())
	case codes.FailedPrecondition:
		writeSessionError(w, http.StatusConflict, "ROOM_ENDED", st.Message())
	case codes.ResourceExhausted:
		writeSessionError(w, http.StatusConflict, "ROOM_FULL", st.Message())
	case codes.AlreadyExists:
		writeSessionError(w, http.StatusConflict, "ALREADY_IN_ROOM", st.Message())
	case codes.DeadlineExceeded:
		log.Printf("study-session gRPC call timed out: %v", err)
		writeSessionError(w, http.StatusGatewayTimeout, "UNAVAILABLE", "study session service timed out")
	default:
		// Log the detail, but never show the browser internal addresses.
		log.Printf("study-session gRPC call failed: %v", err)
		writeSessionError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "study session service is unavailable")
	}
}

func writeSessionError(w http.ResponseWriter, httpStatus int, code, msg string) {
	writeSessionJSON(w, httpStatus, map[string]string{"error": msg, "code": code})
}

func writeSessionJSON(w http.ResponseWriter, httpStatus int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(v)
}
