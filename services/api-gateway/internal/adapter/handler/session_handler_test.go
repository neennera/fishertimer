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

	sessionv1 "github.com/neennera/fishertimer/proto/studysession/v1"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/handler"
	"github.com/neennera/fishertimer/services/api-gateway/internal/adapter/middleware"
)

// fakeSessionClient records the last request and returns a canned result.
// RPCs the tests do not use fall through to the nil interface and panic.
type fakeSessionClient struct {
	sessionv1.StudySessionServiceClient
	lastJoin  *sessionv1.JoinSessionRequest
	lastLeave *sessionv1.LeaveSessionRequest
	err       error
}

func (f *fakeSessionClient) JoinSession(ctx context.Context, in *sessionv1.JoinSessionRequest, _ ...grpc.CallOption) (*sessionv1.JoinSessionResponse, error) {
	f.lastJoin = in
	if f.err != nil {
		return nil, f.err
	}
	return &sessionv1.JoinSessionResponse{Success: true, Session: &sessionv1.StudySessionResponse{
		Id: in.GetSessionId(), Name: "Deep Work", ParticipantLimit: 5, ParticipantCount: 2, Status: "ACTIVE",
	}}, nil
}

func (f *fakeSessionClient) LeaveSession(ctx context.Context, in *sessionv1.LeaveSessionRequest, _ ...grpc.CallOption) (*sessionv1.LeaveSessionResponse, error) {
	f.lastLeave = in
	return &sessionv1.LeaveSessionResponse{Success: true, Left: true, SessionId: in.GetSessionId()}, nil
}

func (f *fakeSessionClient) ListActiveSession(ctx context.Context, in *sessionv1.ListActiveSessionRequest, _ ...grpc.CallOption) (*sessionv1.ListActiveSessionResponse, error) {
	return &sessionv1.ListActiveSessionResponse{Sessions: []*sessionv1.StudySessionResponse{{Id: "s1", Name: "Room", ParticipantCount: 1, ParticipantLimit: 3}}}, nil
}

func serveSession(h *handler.SessionHandler, method, target, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSessionHandler_UsesVerifiedIdentityOverBody(t *testing.T) {
	client := &fakeSessionClient{}
	h := handler.NewSessionHandler(client, true)

	rec := serveSession(h, http.MethodPost, "/api/session/join",
		`{"session_id":"s1","user_id":"spoofed","display_name":"Spoof"}`,
		map[string]string{middleware.HeaderUserID: "u-real", middleware.HeaderDisplayName: "Mira%20K."})

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if client.lastJoin.GetUserId() != "u-real" || client.lastJoin.GetDisplayName() != "Mira K." {
		t.Errorf("join sent as %q/%q, want the verified user", client.lastJoin.GetUserId(), client.lastJoin.GetDisplayName())
	}
	var got map[string]any
	json.NewDecoder(rec.Body).Decode(&got)
	if got["participant_count"] != float64(2) || got["participant_limit"] != float64(5) {
		t.Errorf("body = %v", got)
	}
}

func TestSessionHandler_DevIdentityOnlyOutsideProduction(t *testing.T) {
	body := `{"session_id":"s1","user_id":"u-dev"}`

	client := &fakeSessionClient{}
	rec := serveSession(handler.NewSessionHandler(client, true), http.MethodPost, "/api/session/leave", body, nil)
	if rec.Code != http.StatusOK || client.lastLeave.GetUserId() != "u-dev" || client.lastLeave.GetReason() != "LEFT" {
		t.Errorf("dev leave = %d, %+v", rec.Code, client.lastLeave)
	}

	rec = serveSession(handler.NewSessionHandler(&fakeSessionClient{}, false), http.MethodPost, "/api/session/leave", body, nil)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "SIGNED_OUT") {
		t.Errorf("production without session = %d %s, want 401 SIGNED_OUT", rec.Code, rec.Body)
	}
}

func TestSessionHandler_MapsGRPCErrorsToCodes(t *testing.T) {
	cases := []struct {
		grpc     codes.Code
		wantHTTP int
		wantCode string
	}{
		{codes.InvalidArgument, http.StatusBadRequest, "INVALID"},
		{codes.NotFound, http.StatusNotFound, "ROOM_NOT_FOUND"},
		{codes.FailedPrecondition, http.StatusConflict, "ROOM_ENDED"},
		{codes.ResourceExhausted, http.StatusConflict, "ROOM_FULL"},
		{codes.AlreadyExists, http.StatusConflict, "ALREADY_IN_ROOM"},
		{codes.Unavailable, http.StatusServiceUnavailable, "UNAVAILABLE"},
	}
	for _, c := range cases {
		h := handler.NewSessionHandler(&fakeSessionClient{err: status.Error(c.grpc, "x")}, true)
		rec := serveSession(h, http.MethodPost, "/api/session/join", `{"session_id":"s1","user_id":"u"}`, nil)
		var got map[string]string
		json.NewDecoder(rec.Body).Decode(&got)
		if rec.Code != c.wantHTTP || got["code"] != c.wantCode {
			t.Errorf("%v -> %d %q, want %d %q", c.grpc, rec.Code, got["code"], c.wantHTTP, c.wantCode)
		}
	}
}

func TestSessionHandler_RoutesAndMethods(t *testing.T) {
	h := handler.NewSessionHandler(&fakeSessionClient{}, true)

	rec := serveSession(h, http.MethodGet, "/api/session/active", "", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"sessions":[{"id":"s1"`) {
		t.Errorf("active = %d %s", rec.Code, rec.Body)
	}
	if rec := serveSession(h, http.MethodGet, "/api/session/join", "", nil); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET join = %d", rec.Code)
	}
	if rec := serveSession(h, http.MethodPost, "/api/session/end", `{}`, nil); rec.Code != http.StatusNotFound {
		t.Errorf("browser end = %d, want 404 (admin only)", rec.Code)
	}
}
