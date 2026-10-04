package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	timerv1 "github.com/neennera/fishertimer/proto/studytimer/v1"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/middleware"
)

// fakeTimerClient records the last call and returns a canned result. RPCs the
// tests do not use fall through to the embedded nil interface and panic.
type fakeTimerClient struct {
	timerv1.StudyTimerServiceClient
	lastCall  string
	lastIDs   [2]string
	lastStart *timerv1.StartTimerRequest
	err       error
}

func (f *fakeTimerClient) reply(call, sessionID, userID string) (*timerv1.TimerStateResponse, error) {
	f.lastCall, f.lastIDs = call, [2]string{sessionID, userID}
	if f.err != nil {
		return nil, f.err
	}
	return &timerv1.TimerStateResponse{
		SessionId:        sessionID,
		UserId:           userID,
		Status:           timerv1.TimerStatus_TIMER_STATUS_RUNNING,
		Phase:            timerv1.TimerPhase_TIMER_PHASE_WORK,
		DurationSeconds:  1500,
		RemainingSeconds: 1200,
	}, nil
}

func (f *fakeTimerClient) StartTimer(ctx context.Context, in *timerv1.StartTimerRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	f.lastStart = in
	return f.reply("StartTimer", in.GetSessionId(), in.GetUserId())
}

func (f *fakeTimerClient) GetTimer(ctx context.Context, in *timerv1.GetTimerRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	return f.reply("GetTimer", in.GetSessionId(), in.GetUserId())
}

func (f *fakeTimerClient) PauseTimer(ctx context.Context, in *timerv1.PauseTimerRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	return f.reply("PauseTimer", in.GetSessionId(), in.GetUserId())
}

func (f *fakeTimerClient) ResumeTimer(ctx context.Context, in *timerv1.ResumeTimerRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	return f.reply("ResumeTimer", in.GetSessionId(), in.GetUserId())
}

func (f *fakeTimerClient) ResetTimer(ctx context.Context, in *timerv1.ResetTimerRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	return f.reply("ResetTimer", in.GetSessionId(), in.GetUserId())
}

func (f *fakeTimerClient) StopTimer(ctx context.Context, in *timerv1.StopTimerRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	return f.reply("StopTimer", in.GetSessionId(), in.GetUserId())
}

func (f *fakeTimerClient) CompleteCycle(ctx context.Context, in *timerv1.CompleteCycleRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	return f.reply("CompleteCycle", in.GetSessionId(), in.GetUserId())
}

func (f *fakeTimerClient) SkipRest(ctx context.Context, in *timerv1.SkipRestRequest, _ ...grpc.CallOption) (*timerv1.TimerStateResponse, error) {
	return f.reply("SkipRest", in.GetSessionId(), in.GetUserId())
}

func serve(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestTimerHandler_ActionsCallMatchingRPC(t *testing.T) {
	for action, rpc := range map[string]string{
		"start":     "StartTimer",
		"pause":     "PauseTimer",
		"resume":    "ResumeTimer",
		"reset":     "ResetTimer",
		"stop":      "StopTimer",
		"complete":  "CompleteCycle",
		"skip-rest": "SkipRest",
	} {
		t.Run(action, func(t *testing.T) {
			client := &fakeTimerClient{}
			rec := serve(handler.NewTimerHandler(client, true), http.MethodPost, "/api/timer/"+action,
				`{"session_id":"demo","user_id":"u1"}`)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
			}
			if client.lastCall != rpc || client.lastIDs != [2]string{"demo", "u1"} {
				t.Fatalf("called %s with %v, want %s with [demo u1]", client.lastCall, client.lastIDs, rpc)
			}
		})
	}
}

func TestTimerHandler_StateReadsQueryAndSimplifiesEnums(t *testing.T) {
	client := &fakeTimerClient{}
	rec := serve(handler.NewTimerHandler(client, true), http.MethodGet, "/api/timer/state?session_id=demo&user_id=u1", "")

	if client.lastCall != "GetTimer" || client.lastIDs != [2]string{"demo", "u1"} {
		t.Fatalf("called %s with %v", client.lastCall, client.lastIDs)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "RUNNING" || body["phase"] != "WORK" || body["remaining_seconds"] != float64(1200) {
		t.Fatalf("unexpected body %v", body)
	}
}

func TestTimerHandler_MapsGRPCErrorsToHTTP(t *testing.T) {
	cases := []struct {
		code     codes.Code
		want     int
		wantBody string
	}{
		{codes.InvalidArgument, http.StatusBadRequest, `"error":"boom"`},
		{codes.FailedPrecondition, http.StatusConflict, `"error":"boom"`},
		// Server-side failures hide the internal message from the browser.
		{codes.Unavailable, http.StatusServiceUnavailable, `"error":"study timer service is unavailable"`},
		{codes.Internal, http.StatusBadGateway, `"error":"study timer service is unavailable"`},
	}
	for _, tc := range cases {
		t.Run(tc.code.String(), func(t *testing.T) {
			client := &fakeTimerClient{err: status.Error(tc.code, "boom")}
			rec := serve(handler.NewTimerHandler(client, true), http.MethodPost, "/api/timer/pause", `{"session_id":"s","user_id":"u"}`)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
			if !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("body = %s, want %s", rec.Body, tc.wantBody)
			}
		})
	}
}

func TestTimerHandler_RejectsBadRequests(t *testing.T) {
	cases := []struct {
		name, method, target, body string
		want                       int
	}{
		{"unknown endpoint", http.MethodPost, "/api/timer/explode", `{}`, http.StatusNotFound},
		{"GET on action", http.MethodGet, "/api/timer/start", "", http.StatusMethodNotAllowed},
		{"POST on state", http.MethodPost, "/api/timer/state", "", http.StatusMethodNotAllowed},
		{"invalid JSON", http.MethodPost, "/api/timer/start", `not json`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := serve(handler.NewTimerHandler(&fakeTimerClient{}, true), tc.method, tc.target, tc.body)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestTimerHandler_IdentityAndStartOptions(t *testing.T) {
	client := &fakeTimerClient{}
	h := handler.NewTimerHandler(client, false)

	// Production without a session: refused before reaching the service.
	if rec := serve(h, http.MethodPost, "/api/timer/start", `{"session_id":"s","user_id":"spoof"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no session = %d, want 401", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/timer/start",
		strings.NewReader(`{"session_id":"s","user_id":"spoof","phase":"rest","duration_minutes":10}`))
	req.Header.Set(middleware.HeaderUserID, "u-real")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || client.lastIDs != [2]string{"s", "u-real"} {
		t.Fatalf("start = %d as %v, want the verified user", rec.Code, client.lastIDs)
	}
	if client.lastStart.GetPhase() != timerv1.TimerPhase_TIMER_PHASE_REST || client.lastStart.GetDurationMinutes() != 10 {
		t.Fatalf("start request = %+v", client.lastStart)
	}
}
