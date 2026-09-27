package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
)

// timerCallTimeout bounds each gRPC call so a stuck Study Timer cannot hold
// the browser's request open.
const timerCallTimeout = 5 * time.Second

// TimerHandler translates the REST /api/timer/* API used by the browser into
// gRPC calls to the Study Timer service.
//
//	GET  /api/timer/state?session_id=&user_id=  -> GetTimer
//	POST /api/timer/start   {session_id, user_id} -> StartTimer
//	POST /api/timer/pause   {session_id, user_id} -> PauseTimer
//	POST /api/timer/resume  {session_id, user_id} -> ResumeTimer
//	POST /api/timer/reset   {session_id, user_id} -> ResetTimer
type TimerHandler struct {
	client timerv1.StudyTimerServiceClient
}

func NewTimerHandler(client timerv1.StudyTimerServiceClient) *TimerHandler {
	return &TimerHandler{client: client}
}

type timerRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
}

type timerResponse struct {
	SessionID        string `json:"session_id"`
	UserID           string `json:"user_id"`
	Status           string `json:"status"`
	Phase            string `json:"phase"`
	WorkMinutes      int32  `json:"work_minutes"`
	RestMinutes      int32  `json:"rest_minutes"`
	CurrentCycle     int32  `json:"current_cycle"`
	LastUpdated      string `json:"last_updated"`
	DurationSeconds  int32  `json:"duration_seconds"`
	RemainingSeconds int32  `json:"remaining_seconds"`
}

// timerAction is one POST endpoint: it calls a single RPC with the ids.
type timerAction func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error)

var timerActions = map[string]timerAction{
	"start": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.StartTimer(ctx, &timerv1.StartTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"pause": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.PauseTimer(ctx, &timerv1.PauseTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"resume": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.ResumeTimer(ctx, &timerv1.ResumeTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"reset": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.ResetTimer(ctx, &timerv1.ResetTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
}

func (h *TimerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	action := strings.TrimPrefix(r.URL.Path, "/api/timer/")

	if action == "state" {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "use GET")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), timerCallTimeout)
		defer cancel()
		q := r.URL.Query()
		resp, err := h.client.GetTimer(ctx, &timerv1.GetTimerRequest{
			SessionId: q.Get("session_id"),
			UserId:    q.Get("user_id"),
		})
		writeTimerResult(w, resp, err)
		return
	}

	call, ok := timerActions[action]
	if !ok {
		writeJSONError(w, http.StatusNotFound, "unknown timer endpoint")
		return
	}
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}

	var req timerRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "body must be JSON with session_id and user_id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), timerCallTimeout)
	defer cancel()
	resp, err := call(ctx, h.client, req)
	writeTimerResult(w, resp, err)
}

func writeTimerResult(w http.ResponseWriter, resp *timerv1.TimerStateResponse, err error) {
	if err != nil {
		st, _ := status.FromError(err)
		code := httpStatusFromGRPC(st.Code())
		msg := st.Message()
		if code >= http.StatusInternalServerError {
			// Log the detail, but never show the browser internal addresses.
			log.Printf("study-timer gRPC call failed: %v", err)
			msg = "study timer service is unavailable"
		}
		writeJSONError(w, code, msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timerResponse{
		SessionID:        resp.GetSessionId(),
		UserID:           resp.GetUserId(),
		Status:           strings.TrimPrefix(resp.GetStatus().String(), "TIMER_STATUS_"),
		Phase:            strings.TrimPrefix(resp.GetPhase().String(), "TIMER_PHASE_"),
		WorkMinutes:      resp.GetWorkMinutes(),
		RestMinutes:      resp.GetRestMinutes(),
		CurrentCycle:     resp.GetCurrentCycle(),
		LastUpdated:      resp.GetLastUpdated(),
		DurationSeconds:  resp.GetDurationSeconds(),
		RemainingSeconds: resp.GetRemainingSeconds(),
	})
}

// httpStatusFromGRPC maps the Study Timer's gRPC status to what the browser
// should see. Failing to reach the service at all is a gateway problem (5xx),
// not the caller's.
func httpStatusFromGRPC(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.FailedPrecondition:
		return http.StatusConflict
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusBadGateway
	}
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
