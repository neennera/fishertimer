package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
	"github.com/neennera/fishertimer/services/study-session/internal/usecase"
)

type fakePublisher struct {
	sent []string // routing keys, in publish order
	down bool
}

func (p *fakePublisher) Publish(ctx context.Context, routingKey, eventID string, payload []byte) error {
	if p.down {
		return errors.New("connection refused")
	}
	p.sent = append(p.sent, routingKey)
	return nil
}

func TestOutboxRelay_HoldsEventsWhileBrokerIsDown(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	f.uc.LeaveSession(ctx, room.ID, alice, "")

	pub := &fakePublisher{down: true}
	relay := usecase.NewOutboxRelay(f.repo, pub, time.Second)

	if n, err := relay.Drain(ctx); err == nil || n != 0 {
		t.Fatalf("drain with broker down = %d, %v; want an error", n, err)
	}
	if got := len(f.pending(t)); got != 3 {
		t.Fatalf("%d events pending, want all 3 kept", got)
	}

	pub.down = false
	if n, err := relay.Drain(ctx); err != nil || n != 3 {
		t.Fatalf("drain = %d, %v", n, err)
	}
	want := []string{events.RoutingKeyJoined, events.RoutingKeyLeft, events.RoutingKeyEnded}
	if !equal(pub.sent, want) {
		t.Errorf("published %v, want %v (in order)", pub.sent, want)
	}
	if got := f.pending(t); len(got) != 0 {
		t.Errorf("still pending %v", got)
	}
}

type fakeTimers struct {
	activity []domain.TimerActivity
	err      error
}

func (f *fakeTimers) RoomTimers(ctx context.Context, sessionID string) ([]domain.TimerActivity, error) {
	return f.activity, f.err
}

func TestSweeper_RemovesDisconnectedAndEndsOldRooms(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	old := f.createRoom(t, alice, 3)

	f.now = t0.Add(time.Hour)
	fresh := f.createRoom(t, bob, 3)
	f.uc.JoinSession(ctx, fresh.ID, carol, "")
	f.now = t0.Add(time.Hour + 50*time.Second)
	f.uc.Heartbeat(ctx, fresh.ID, carol) // carol's and alice's pages are open, bob's is not
	f.uc.Heartbeat(ctx, old.ID, alice)

	sweeper := usecase.NewSweeper(f.uc, f.repo, nil, usecase.SweeperConfig{
		MaxSessionAge:     24 * time.Hour,
		DisconnectTimeout: time.Minute,
	})

	sweeper.Sweep(ctx, t0.Add(time.Hour+90*time.Second))
	if ps, _ := f.uc.GetParticipants(ctx, fresh.ID); len(ps) != 1 || ps[0].UserID != carol {
		t.Errorf("after disconnect sweep participants = %+v, want only carol", ps)
	}
	last, _ := f.repo.GetLastParticipation(ctx, fresh.ID, bob)
	if last.LeaveReason != events.ReasonDisconnectTimeout {
		t.Errorf("bob reason = %q", last.LeaveReason)
	}

	f.uc.Heartbeat(ctx, fresh.ID, carol)
	f.uc.Heartbeat(ctx, old.ID, alice)
	f.now = t0.Add(24*time.Hour + time.Minute)
	f.uc.Heartbeat(ctx, fresh.ID, carol)
	f.uc.Heartbeat(ctx, old.ID, alice)
	sweeper.Sweep(ctx, f.now)
	if s, _ := f.uc.GetSession(ctx, old.ID); s.Status != domain.StatusEnded || s.EndReason != events.ReasonTimeout24H {
		t.Errorf("24h room = %+v", s)
	}
	if s, _ := f.uc.GetSession(ctx, fresh.ID); s.Status != domain.StatusActive {
		t.Errorf("younger room ended too")
	}
}

func TestSweeper_RemovesIdleButNotWhenTimerUnreachable(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	room := f.createRoom(t, alice, 3)
	f.uc.JoinSession(ctx, room.ID, bob, "")
	cfg := usecase.SweeperConfig{IdleTimeout: 10 * time.Minute}
	later := t0.Add(11 * time.Minute)

	down := usecase.NewSweeper(f.uc, f.repo, &fakeTimers{err: errors.New("unavailable")}, cfg)
	down.Sweep(ctx, later)
	if ps, _ := f.uc.GetParticipants(ctx, room.ID); len(ps) != 2 {
		t.Fatalf("removed participants while Study Timer was down")
	}

	// Alice is focusing; Bob never started a cycle.
	timers := &fakeTimers{activity: []domain.TimerActivity{
		{UserID: alice, Status: "RUNNING", Phase: "WORK", LastUpdated: t0.Add(5 * time.Minute), DurationSeconds: 1500, RemainingSeconds: 1200},
	}}
	usecase.NewSweeper(f.uc, f.repo, timers, cfg).Sweep(ctx, later)
	ps, _ := f.uc.GetParticipants(ctx, room.ID)
	if len(ps) != 1 || ps[0].UserID != alice {
		t.Errorf("participants = %+v, want only alice", ps)
	}
	last, _ := f.repo.GetLastParticipation(ctx, room.ID, bob)
	if last.LeaveReason != events.ReasonIdleTimeout {
		t.Errorf("bob reason = %q", last.LeaveReason)
	}
}

func TestIsIdle(t *testing.T) {
	p := domain.Participant{JoinedAt: t0}
	timeout := 10 * time.Minute
	cases := []struct {
		name     string
		timer    domain.TimerActivity
		hasTimer bool
		at       time.Duration // since t0
		want     bool
	}{
		{"no timer, under timeout", domain.TimerActivity{}, false, 9 * time.Minute, false},
		{"no timer, timed out", domain.TimerActivity{}, false, 10 * time.Minute, true},
		{"work running", domain.TimerActivity{Status: "RUNNING", Phase: "WORK", RemainingSeconds: 60}, true, time.Hour, false},
		{"rest running is not idle", domain.TimerActivity{Status: "RUNNING", Phase: "REST", RemainingSeconds: 60}, true, time.Hour, false},
		{"paused is not idle", domain.TimerActivity{Status: "PAUSED"}, true, time.Hour, false},
		{"stopped recently", domain.TimerActivity{Status: "STOPPED", LastUpdated: t0.Add(55 * time.Minute)}, true, time.Hour, false},
		{"stopped long ago", domain.TimerActivity{Status: "STOPPED", LastUpdated: t0.Add(45 * time.Minute)}, true, time.Hour, true},
		{"ran out, counted from its end", domain.TimerActivity{Status: "RUNNING", LastUpdated: t0, DurationSeconds: 3300, RemainingSeconds: 0}, true, time.Hour, false},
		{"ran out long ago", domain.TimerActivity{Status: "RUNNING", LastUpdated: t0, DurationSeconds: 600, RemainingSeconds: 0}, true, time.Hour, true},
	}
	for _, c := range cases {
		if got := usecase.IsIdle(p, c.timer, c.hasTimer, t0.Add(c.at), timeout); got != c.want {
			t.Errorf("%s: IsIdle = %v, want %v", c.name, got, c.want)
		}
	}
}
