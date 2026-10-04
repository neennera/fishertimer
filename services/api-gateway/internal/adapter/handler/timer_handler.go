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
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/middleware"
)

// timerCallTimeout bounds each gRPC call so a stuck Study Timer cannot hold
// the browser's request open.
const timerCallTimeout = 5 * time.Second

// TimerHandler translates the REST /api/timer/* API used by the browser into
// gRPC calls to the Study Timer service (UC-05).
//
//	GET  /api/timer/state?session_id=       -> GetTimer
//	GET  /api/timer/room?session_id=        -> GetRoomTimers ({"timers": [...]})
//	POST /api/timer/start     {session_id, phase?, duration_minutes?} -> StartTimer
//	POST /api/timer/pause     {session_id}  -> PauseTimer
//	POST /api/timer/resume    {session_id}  -> ResumeTimer
//	POST /api/timer/stop      {session_id}  -> StopTimer
//	POST /api/timer/reset     {session_id}  -> ResetTimer
//	POST /api/timer/complete  {session_id}  -> CompleteCycle
//	POST /api/timer/skip-rest {session_id}  -> SkipRest
//	POST /api/timer/settings  {session_id?, work_minutes, rest_minutes} -> UpdateTimerSetting
//
// The acting user is the signed-in one (X-User-Id from the session JWT).
// Outside production a request without a session may name its user_id
// (mock auth), like /api/session/*.
type TimerHandler struct {
	client       timerv1.StudyTimerServiceClient
	allowDevUser bool
}

func NewTimerHandler(client timerv1.StudyTimerServiceClient, allowDevUser bool) *TimerHandler {
	return &TimerHandler{client: client, allowDevUser: allowDevUser}
}

type timerRequest struct {
	SessionID       string `json:"session_id"`
	UserID          string `json:"user_id"`
	Phase           string `json:"phase"`
	DurationMinutes int32  `json:"duration_minutes"`
	WorkMinutes     int32  `json:"work_minutes"`
	RestMinutes     int32  `json:"rest_minutes"`
}

type timerResponse struct {
	SessionID            string `json:"session_id"`
	UserID               string `json:"user_id"`
	State                string `json:"state"`
	Status               string `json:"status"`
	Phase                string `json:"phase"`
	WorkMinutes          int32  `json:"work_minutes"`
	RestMinutes          int32  `json:"rest_minutes"`
	CurrentCycle         int32  `json:"current_cycle"`
	FocusSeconds         int32  `json:"focus_seconds"`
	LastUpdated          string `json:"last_updated"`
	DurationSeconds      int32  `json:"duration_seconds"`
	RemainingSeconds     int32  `json:"remaining_seconds"`
	CycleID              string `json:"cycle_id"`
	StartedAt            string `json:"started_at"`
	PausedTotalSeconds   int32  `json:"paused_total_seconds"`
	LastCompletedCycleID string `json:"last_completed_cycle_id"`
	LastCompletedAt      string `json:"last_completed_at"`
	MinWorkMinutes       int32  `json:"min_work_minutes"`
	MaxWorkMinutes       int32  `json:"max_work_minutes"`
	MinRestMinutes       int32  `json:"min_rest_minutes"`
	MaxRestMinutes       int32  `json:"max_rest_minutes"`
	MaxPauseMinutes      int32  `json:"max_pause_minutes"`
}

// timerAction is one POST endpoint: it calls a single RPC.
type timerAction func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error)

var timerActions = map[string]timerAction{
	"start": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		phase := timerv1.TimerPhase_TIMER_PHASE_WORK
		if strings.EqualFold(req.Phase, "REST") {
			phase = timerv1.TimerPhase_TIMER_PHASE_REST
		}
		return c.StartTimer(ctx, &timerv1.StartTimerRequest{SessionId: req.SessionID, UserId: req.UserID, Phase: phase, DurationMinutes: req.DurationMinutes})
	},
	"pause": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.PauseTimer(ctx, &timerv1.PauseTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"resume": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.ResumeTimer(ctx, &timerv1.ResumeTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"stop": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.StopTimer(ctx, &timerv1.StopTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"reset": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.ResetTimer(ctx, &timerv1.ResetTimerRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"complete": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.CompleteCycle(ctx, &timerv1.CompleteCycleRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"skip-rest": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.SkipRest(ctx, &timerv1.SkipRestRequest{SessionId: req.SessionID, UserId: req.UserID})
	},
	"settings": func(ctx context.Context, c timerv1.StudyTimerServiceClient, req timerRequest) (*timerv1.TimerStateResponse, error) {
		return c.UpdateTimerSetting(ctx, &timerv1.UpdateTimerSettingRequest{
			SessionId: req.SessionID, UserId: req.UserID, WorkMinutes: req.WorkMinutes, RestMinutes: req.RestMinutes,
		})
	},
}

func (h *TimerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	action := strings.TrimPrefix(r.URL.Path, "/api/timer/")

	if action == "state" || action == "room" {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "use GET")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), timerCallTimeout)
		defer cancel()
		q := r.URL.Query()

		if action == "room" {
			resp, err := h.client.GetRoomTimers(ctx, &timerv1.GetRoomTimersRequest{SessionId: q.Get("session_id")})
			if err != nil {
				writeTimerError(w, err)
				return
			}
			timers := make([]timerResponse, 0, len(resp.GetTimers()))
			for _, t := range resp.GetTimers() {
				timers = append(timers, toTimerJSON(t))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"timers": timers})
			return
		}

		userID, ok := h.identity(r, q.Get("user_id"))
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "sign in to use the timer")
			return
		}
		resp, err := h.client.GetTimer(ctx, &timerv1.GetTimerRequest{SessionId: q.Get("session_id"), UserId: userID})
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
		writeJSONError(w, http.StatusBadRequest, "body must be JSON with session_id")
		return
	}
	userID, ok := h.identity(r, req.UserID)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "sign in to use the timer")
		return
	}
	req.UserID = userID

	ctx, cancel := context.WithTimeout(r.Context(), timerCallTimeout)
	defer cancel()
	resp, err := call(ctx, h.client, req)
	writeTimerResult(w, resp, err)
}

// identity returns the verified user, or outside production the user the
// request names.
func (h *TimerHandler) identity(r *http.Request, named string) (string, bool) {
	if id := r.Header.Get(middleware.HeaderUserID); id != "" {
		return id, true
	}
	if h.allowDevUser && named != "" {
		return named, true
	}
	return "", false
}

func writeTimerResult(w http.ResponseWriter, resp *timerv1.TimerStateResponse, err error) {
	if err != nil {
		writeTimerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toTimerJSON(resp))
}

func writeTimerError(w http.ResponseWriter, err error) {
	st, _ := status.FromError(err)
	code := httpStatusFromGRPC(st.Code())
	msg := st.Message()
	if code >= http.StatusInternalServerError {
		// Log the detail, but never show the browser internal addresses.
		log.Printf("study-timer gRPC call failed: %v", err)
		msg = "study timer service is unavailable"
	}
	writeJSONError(w, code, msg)
}

func toTimerJSON(resp *timerv1.TimerStateResponse) timerResponse {
	return timerResponse{
		SessionID:            resp.GetSessionId(),
		UserID:               resp.GetUserId(),
		State:                resp.GetState(),
		Status:               strings.TrimPrefix(resp.GetStatus().String(), "TIMER_STATUS_"),
		Phase:                strings.TrimPrefix(resp.GetPhase().String(), "TIMER_PHASE_"),
		WorkMinutes:          resp.GetWorkMinutes(),
		RestMinutes:          resp.GetRestMinutes(),
		CurrentCycle:         resp.GetCurrentCycle(),
		FocusSeconds:         resp.GetFocusSeconds(),
		LastUpdated:          resp.GetLastUpdated(),
		DurationSeconds:      resp.GetDurationSeconds(),
		RemainingSeconds:     resp.GetRemainingSeconds(),
		CycleID:              resp.GetCycleId(),
		StartedAt:            resp.GetStartedAt(),
		PausedTotalSeconds:   resp.GetPausedTotalSeconds(),
		LastCompletedCycleID: resp.GetLastCompletedCycleId(),
		LastCompletedAt:      resp.GetLastCompletedAt(),
		MinWorkMinutes:       resp.GetMinWorkMinutes(),
		MaxWorkMinutes:       resp.GetMaxWorkMinutes(),
		MinRestMinutes:       resp.GetMinRestMinutes(),
		MaxRestMinutes:       resp.GetMaxRestMinutes(),
		MaxPauseMinutes:      resp.GetMaxPauseMinutes(),
	}
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
	case codes.PermissionDenied:
		return http.StatusForbidden
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
