package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/admin/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/admin/internal/domain"
	"github.com/neennera/fishertimer/services/admin/internal/usecase"
)

const (
	adminID = "57f2060b-edf9-499e-a430-9d6565591b25"
	roomID  = "f2000000-0000-4000-8000-000000000001"
	hostID  = "e1000000-0000-4000-8000-000000000001"
	guestID = "e1000000-0000-4000-8000-000000000002"
)

func strPtr(s string) *string { return &s }

type mockSessions struct {
	active []domain.SessionOverview
	parts  []domain.SessionParticipantOverview
	err    error
}

func (m *mockSessions) ListActiveSessions(ctx context.Context) ([]domain.SessionOverview, error) {
	return append([]domain.SessionOverview(nil), m.active...), m.err
}

func (m *mockSessions) GetActiveSession(ctx context.Context, sessionID string) (*domain.SessionOverview, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, s := range m.active {
		if s.SessionID == sessionID {
			return &s, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockSessions) ListParticipants(ctx context.Context, sessionID string) ([]domain.SessionParticipantOverview, error) {
	return append([]domain.SessionParticipantOverview(nil), m.parts...), m.err
}

type mockSessionClient struct {
	kicked, closed bool
	sessionEnded   bool
	err            error
}

func (m *mockSessionClient) KickParticipant(ctx context.Context, sessionID, userID string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	m.kicked = true
	return m.sessionEnded, nil
}

func (m *mockSessionClient) CloseSession(ctx context.Context, sessionID string) error {
	if m.err != nil {
		return m.err
	}
	m.closed = true
	return nil
}

func liveRoom() *mockSessions {
	return &mockSessions{
		active: []domain.SessionOverview{
			{SessionID: "older", Title: "Old", CreatedAt: time.Now().Add(-time.Hour)},
			{SessionID: roomID, Title: "Room 1", HostID: hostID, IsActive: true, ParticipantCount: 2, CreatedAt: time.Now()},
		},
		parts: []domain.SessionParticipantOverview{
			{UserID: guestID, DisplayName: strPtr("Guest")},
			{UserID: hostID, DisplayName: strPtr("Host")},
		},
	}
}

func TestViewActiveSessions_NewestFirst(t *testing.T) {
	svc := usecase.New(liveRoom(), repository.NewInMemory(), &mockSessionClient{})
	got, err := svc.ViewActiveSessions(context.Background())
	if err != nil || len(got) != 2 || got[0].SessionID != roomID {
		t.Fatalf("unexpected sessions %+v, err %v", got, err)
	}
}

func TestViewSessionDetails_HostFirstWithName(t *testing.T) {
	svc := usecase.New(liveRoom(), repository.NewInMemory(), &mockSessionClient{})
	got, err := svc.ViewSessionDetails(context.Background(), roomID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Participants) != 2 || got.Participants[0].UserID != hostID || !got.Participants[0].IsHost ||
		got.Participants[1].IsHost || got.Participants[1].SessionID != roomID {
		t.Fatalf("unexpected participants: %+v", got.Participants)
	}
	if got.HostName == nil || *got.HostName != "Host" {
		t.Fatalf("expected host name from participant list, got %v", got.HostName)
	}
}

func TestViewSessionDetails_Errors(t *testing.T) {
	svc := usecase.New(liveRoom(), repository.NewInMemory(), &mockSessionClient{})
	if _, err := svc.ViewParticipants(context.Background(), "  "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
	if _, err := svc.ViewSessionDetails(context.Background(), "ended-room"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestKickUser_DispatchesAndAudits(t *testing.T) {
	logs := repository.NewInMemory()
	client := &mockSessionClient{sessionEnded: true}
	svc := usecase.New(liveRoom(), logs, client)

	res, err := svc.KickUser(context.Background(), adminID, roomID, guestID, "  spamming  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.kicked || !res.SessionEnded {
		t.Fatalf("expected kick dispatched and session_ended passed through: %+v", res)
	}
	l := res.Log
	if l.LogID == "" || l.AdminID != adminID || l.Action != domain.ActionKickUser || l.TargetID != guestID ||
		l.Reason == nil || *l.Reason != "spamming" || l.CreatedAt.IsZero() {
		t.Fatalf("unexpected audit log: %+v", l)
	}
	stored, _ := svc.ViewLogs(context.Background(), "KICK_USER")
	if len(stored) != 1 || stored[0].LogID != l.LogID {
		t.Fatalf("expected the kick to be stored, got %+v", stored)
	}
}

func TestEndSession_DispatchesAndAudits(t *testing.T) {
	client := &mockSessionClient{}
	svc := usecase.New(liveRoom(), repository.NewInMemory(), client)

	res, err := svc.EndSession(context.Background(), adminID, roomID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !client.closed || !res.SessionEnded || res.Log.Action != domain.ActionForceCloseSession ||
		res.Log.TargetID != roomID || res.Log.Reason != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestModeration_RoomNoLongerLiveIsStateChanged(t *testing.T) {
	logs := repository.NewInMemory()
	client := &mockSessionClient{}
	svc := usecase.New(&mockSessions{}, logs, client)

	if _, err := svc.KickUser(context.Background(), adminID, roomID, guestID, ""); !errors.Is(err, domain.ErrStateChanged) {
		t.Fatalf("kick: expected ErrStateChanged, got %v", err)
	}
	if _, err := svc.EndSession(context.Background(), adminID, roomID, ""); !errors.Is(err, domain.ErrStateChanged) {
		t.Fatalf("end: expected ErrStateChanged, got %v", err)
	}
	if client.kicked || client.closed {
		t.Fatal("no command may be sent for a room that is not live")
	}
	if stored, _ := logs.List(context.Background(), "", 10); len(stored) != 0 {
		t.Fatalf("nothing may be audited, got %+v", stored)
	}
}

func TestModeration_FailedCommandIsNotAudited(t *testing.T) {
	logs := repository.NewInMemory()
	svc := usecase.New(liveRoom(), logs, &mockSessionClient{err: domain.ErrUpstream})

	if _, err := svc.KickUser(context.Background(), adminID, roomID, guestID, ""); !errors.Is(err, domain.ErrUpstream) {
		t.Fatalf("expected ErrUpstream, got %v", err)
	}
	if stored, _ := logs.List(context.Background(), "", 10); len(stored) != 0 {
		t.Fatalf("a failed command must not be audited, got %+v", stored)
	}
}

func TestModeration_RejectsInvalidInput(t *testing.T) {
	svc := usecase.New(liveRoom(), repository.NewInMemory(), &mockSessionClient{})
	ctx := context.Background()
	long := make([]byte, domain.MaxReasonLength+1)
	for i := range long {
		long[i] = 'x'
	}

	cases := map[string]func() error{
		"non-uuid admin": func() error { _, err := svc.KickUser(ctx, "admin", roomID, guestID, ""); return err },
		"non-uuid user":  func() error { _, err := svc.KickUser(ctx, adminID, roomID, "u1", ""); return err },
		"non-uuid room":  func() error { _, err := svc.EndSession(ctx, adminID, "room", ""); return err },
		"long reason":    func() error { _, err := svc.EndSession(ctx, adminID, roomID, string(long)); return err },
		"unknown action": func() error { _, err := svc.ViewLogs(ctx, "BAN_USER"); return err },
	}
	for name, call := range cases {
		if err := call(); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: expected ErrInvalid, got %v", name, err)
		}
	}
}
