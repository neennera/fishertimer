package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-session/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

const (
	alice = "aaaaaaaa-0000-4000-8000-000000000001"
	bob   = "aaaaaaaa-0000-4000-8000-000000000002"
	carol = "aaaaaaaa-0000-4000-8000-000000000003"
)

var t0 = time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)

type fixture struct {
	repo *repository.InMemoryRepository
	uc   usecase.Usecase
	now  time.Time
}

func newFixture() *fixture {
	f := &fixture{repo: repository.NewInMemory(), now: t0}
	f.uc = usecase.New(f.repo, usecase.WithClock(func() time.Time { return f.now }))
	return f
}

func (f *fixture) createRoom(t *testing.T, creator string, limit int) *domain.StudySession {
	t.Helper()
	sess, err := f.uc.CreateSession(context.Background(), "Deep Work", creator, "Creator", limit)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return sess
}

// pending returns the routing keys of the unpublished outbox events, in order.
func (f *fixture) pending(t *testing.T) []string {
	t.Helper()
	evs, err := f.repo.PendingEvents(context.Background(), 100)
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, len(evs))
	for i, ev := range evs {
		keys[i] = ev.RoutingKey
	}
	return keys
}

func equal(a, b []string) bool {
	return strings.Join(a, ",") == strings.Join(b, ",")
}

func TestCreateSession_ValidatesInput(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	cases := []struct {
		name  string
		room  string
		user  string
		limit int
	}{
		{"empty name", "   ", alice, 3},
		{"name too long", strings.Repeat("x", domain.MaxNameLength+1), alice, 3},
		{"limit zero", "Room", alice, 0},
		{"limit above max", "Room", alice, domain.MaxParticipantLimit + 1},
		{"creator not a uuid", "Room", "user1", 3},
	}
	for _, c := range cases {
		if _, err := f.uc.CreateSession(ctx, c.room, c.user, "", c.limit); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", c.name, err)
		}
	}
	if got := f.pending(t); len(got) != 0 {
		t.Errorf("invalid creates queued events %v", got)
	}
}

func TestCreateSession_EnrollsCreatorAndPublishesJoined(t *testing.T) {
	f := newFixture()
	sess := f.createRoom(t, alice, 1) // limit 1 = solo room

	if sess.Status != domain.StatusActive || sess.ParticipantCount != 1 || sess.Name != "Deep Work" {
		t.Errorf("session = %+v", sess)
	}
	ps, _ := f.uc.GetParticipants(context.Background(), sess.ID)
	if len(ps) != 1 || ps[0].UserID != alice || ps[0].DisplayName != "Creator" {
		t.Errorf("participants = %+v", ps)
	}
	if got := f.pending(t); !equal(got, []string{events.RoutingKeyJoined}) {
		t.Errorf("events = %v", got)
	}

	// A solo room accepts nobody else.
	if _, err := f.uc.JoinSession(context.Background(), sess.ID, bob, "Bob"); !errors.Is(err, domain.ErrSessionFull) {
		t.Errorf("join solo room: err = %v, want ErrSessionFull", err)
	}
}

func TestCreateSession_RejectsUserAlreadyInRoom(t *testing.T) {
	f := newFixture()
	f.createRoom(t, alice, 3)
	if _, err := f.uc.CreateSession(context.Background(), "Second", alice, "", 3); !errors.Is(err, domain.ErrAlreadyInSession) {
		t.Errorf("err = %v, want ErrAlreadyInSession", err)
	}
}

func TestJoinSession_Errors(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 2)

	if _, err := f.uc.JoinSession(ctx, "aaaaaaaa-0000-4000-8000-0000000000ff", bob, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown room: err = %v", err)
	}
	if _, err := f.uc.JoinSession(ctx, room.ID, bob, "Bob"); err != nil {
		t.Fatalf("join: %v", err)
	}
	if _, err := f.uc.JoinSession(ctx, room.ID, carol, "Carol"); !errors.Is(err, domain.ErrSessionFull) {
		t.Errorf("full room: err = %v, want ErrSessionFull", err)
	}

	other := f.createRoom(t, carol, 3)
	if _, err := f.uc.JoinSession(ctx, other.ID, bob, "Bob"); !errors.Is(err, domain.ErrAlreadyInSession) {
		t.Errorf("second room: err = %v, want ErrAlreadyInSession", err)
	}

	if _, err := f.uc.EndSession(ctx, other.ID, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := f.uc.JoinSession(ctx, other.ID, carol, ""); !errors.Is(err, domain.ErrSessionEnded) {
		t.Errorf("ended room: err = %v, want ErrSessionEnded", err)
	}
}

func TestJoinSession_RejoiningSameRoomIsANoOp(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)

	sess, err := f.uc.JoinSession(ctx, room.ID, alice, "Alice")
	if err != nil {
		t.Fatalf("rejoin: %v", err)
	}
	if sess.ParticipantCount != 1 {
		t.Errorf("count = %d, want 1 (no duplicate row, UC-02 E-5)", sess.ParticipantCount)
	}
	if got := f.pending(t); len(got) != 1 {
		t.Errorf("events = %v, want only the create's joined", got)
	}
}

func TestJoinSession_LastSeatRaceAdmitsExactlyOne(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 2)

	users := []string{bob, carol, "aaaaaaaa-0000-4000-8000-000000000004", "aaaaaaaa-0000-4000-8000-000000000005"}
	var wg sync.WaitGroup
	results := make(chan error, len(users))
	for _, u := range users {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			_, err := f.uc.JoinSession(ctx, room.ID, u, "")
			results <- err
		}(u)
	}
	wg.Wait()
	close(results)

	joined := 0
	for err := range results {
		switch {
		case err == nil:
			joined++
		case !errors.Is(err, domain.ErrSessionFull):
			t.Errorf("unexpected error %v", err)
		}
	}
	if joined != 1 {
		t.Errorf("%d users took the last seat, want 1", joined)
	}
}

func TestLeaveSession_LastOneOutEndsRoom(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	if _, err := f.uc.JoinSession(ctx, room.ID, bob, "Bob"); err != nil {
		t.Fatal(err)
	}

	f.now = t0.Add(30 * time.Minute)
	res, err := f.uc.LeaveSession(ctx, room.ID, bob, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Left || res.SessionEnded || res.Participant.LeaveReason != events.ReasonLeft {
		t.Errorf("bob leave = %+v", res)
	}
	if !res.Participant.JoinedAt.Equal(t0) || !res.Participant.LeftAt.Equal(f.now) {
		t.Errorf("bob stay = %v..%v", res.Participant.JoinedAt, res.Participant.LeftAt)
	}

	res, err = f.uc.LeaveSession(ctx, room.ID, alice, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.SessionEnded {
		t.Errorf("last leave did not end the room")
	}
	sess, _ := f.uc.GetSession(ctx, room.ID)
	if sess.Status != domain.StatusEnded || sess.EndReason != events.ReasonEmpty || sess.EndedAt == nil {
		t.Errorf("room = %+v", sess)
	}

	want := []string{
		events.RoutingKeyJoined, events.RoutingKeyJoined,
		events.RoutingKeyLeft,
		events.RoutingKeyLeft, events.RoutingKeyEnded,
	}
	if got := f.pending(t); !equal(got, want) {
		t.Errorf("events = %v, want %v", got, want)
	}
}

func TestLeaveSession_AlreadyLeftIsANoOp(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	f.uc.JoinSession(ctx, room.ID, bob, "")
	f.uc.LeaveSession(ctx, room.ID, bob, events.ReasonKicked)
	before := len(f.pending(t))

	res, err := f.uc.LeaveSession(ctx, room.ID, bob, events.ReasonKicked)
	if err != nil || res.Left {
		t.Errorf("second leave = %+v, %v; want no-op", res, err)
	}
	if len(f.pending(t)) != before {
		t.Errorf("no-op leave queued events")
	}
	if _, err := f.uc.LeaveSession(ctx, room.ID, bob, "BORED"); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("unknown reason: err = %v", err)
	}
}

func TestLeaveSession_PayloadMatchesContract(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	f.uc.JoinSession(ctx, room.ID, bob, "")
	f.uc.LeaveSession(ctx, room.ID, bob, events.ReasonKicked)

	evs, _ := f.repo.PendingEvents(ctx, 100)
	last := evs[len(evs)-1]
	var left events.ParticipantLeft
	if err := json.Unmarshal(last.Payload, &left); err != nil {
		t.Fatal(err)
	}
	if left.EventID == "" || left.EventID != last.EventID || left.SessionID != room.ID ||
		left.UserID != bob || left.Reason != events.ReasonKicked || !left.OccurredAt.Equal(t0) {
		t.Errorf("payload = %+v", left)
	}
}

func TestEndSession_ClosesEveryoneOnce(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	f.uc.JoinSession(ctx, room.ID, bob, "")

	sess, err := f.uc.EndSession(ctx, room.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Status != domain.StatusEnded || sess.EndReason != events.ReasonAdminClosed || sess.ParticipantCount != 0 {
		t.Errorf("room = %+v", sess)
	}
	if ps, _ := f.uc.GetParticipants(ctx, room.ID); len(ps) != 0 {
		t.Errorf("participants left = %+v", ps)
	}
	if active, _ := f.uc.ListActiveSession(ctx); len(active) != 0 {
		t.Errorf("ended room still listed")
	}

	// Ending again changes nothing and queues no second event.
	if _, err := f.uc.EndSession(ctx, room.ID, ""); err != nil {
		t.Fatal(err)
	}
	ended := 0
	for _, k := range f.pending(t) {
		if k == events.RoutingKeyEnded {
			ended++
		}
	}
	if ended != 1 {
		t.Errorf("%d ended events, want 1", ended)
	}

	// Everyone is free to join another room.
	if _, err := f.uc.CreateSession(ctx, "Next", bob, "", 2); err != nil {
		t.Errorf("bob still stuck in ended room: %v", err)
	}
}

func TestGetMySessionAndHeartbeat(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	f.uc.JoinSession(ctx, room.ID, bob, "Bob")

	sess, p, err := f.uc.GetMySession(ctx, bob)
	if err != nil || sess == nil || sess.ID != room.ID || p.DisplayName != "Bob" {
		t.Errorf("GetMySession = %+v %+v %v", sess, p, err)
	}
	if sess, _, _ := f.uc.GetMySession(ctx, carol); sess != nil {
		t.Errorf("carol is in %+v", sess)
	}

	if active, _, err := f.uc.Heartbeat(ctx, room.ID, bob); !active || err != nil {
		t.Errorf("heartbeat = %v, %v", active, err)
	}
	f.uc.LeaveSession(ctx, room.ID, bob, events.ReasonKicked)
	active, reason, err := f.uc.Heartbeat(ctx, room.ID, bob)
	if active || reason != events.ReasonKicked || err != nil {
		t.Errorf("heartbeat after kick = %v, %q, %v", active, reason, err)
	}
}
