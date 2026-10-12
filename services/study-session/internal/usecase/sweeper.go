package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/neennera/fishertimer/pkg/events"
	"github.com/neennera/fishertimer/services/study-session/internal/domain"
)

// SweeperConfig sets the automatic removals. A zero duration turns that
// check off.
type SweeperConfig struct {
	// Interval between sweeps.
	Interval time.Duration
	// MaxSessionAge ends rooms older than this (UC-03 E-6, 24 hours).
	MaxSessionAge time.Duration
	// DisconnectTimeout removes participants whose room page stopped sending
	// heartbeats (UC-03 E-3).
	DisconnectTimeout time.Duration
	// IdleTimeout removes participants with no running work cycle for this
	// long (UC-03 E-7 / UC-05 E-8, IDLE_TIMEOUT_MINUTES).
	IdleTimeout time.Duration
}

// Sweeper runs the time-based room rules. Every removal goes through the
// normal LeaveSession / EndSession path, so it publishes the same events and
// Study Timer finalizes the timers exactly as for a manual leave.
type Sweeper struct {
	uc     Usecase
	repo   domain.Repository
	timers domain.TimerReader // nil disables idle detection
	cfg    SweeperConfig
}

func NewSweeper(uc Usecase, repo domain.Repository, timers domain.TimerReader, cfg SweeperConfig) *Sweeper {
	return &Sweeper{uc: uc, repo: repo, timers: timers, cfg: cfg}
}

// Run sweeps every Interval until ctx is cancelled.
func (s *Sweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Sweep(ctx, time.Now().UTC())
		}
	}
}

// Sweep applies every enabled rule once, as of now.
func (s *Sweeper) Sweep(ctx context.Context, now time.Time) {
	if s.cfg.MaxSessionAge > 0 {
		s.endExpiredSessions(ctx, now)
	}
	if s.cfg.DisconnectTimeout > 0 {
		s.removeDisconnected(ctx, now)
	}
	if s.cfg.IdleTimeout > 0 && s.timers != nil {
		s.removeIdle(ctx, now)
	}
}

func (s *Sweeper) endExpiredSessions(ctx context.Context, now time.Time) {
	expired, err := s.repo.ListSessionsCreatedBefore(ctx, now.Add(-s.cfg.MaxSessionAge))
	if err != nil {
		log.Printf("study-session: sweeper: list expired rooms: %v", err)
		return
	}
	for _, sess := range expired {
		if _, err := s.uc.EndSession(ctx, sess.ID, events.ReasonTimeout24H); err != nil {
			log.Printf("study-session: sweeper: end expired room %s: %v", sess.ID, err)
			continue
		}
		log.Printf("study-session: ended room %s after %s", sess.ID, s.cfg.MaxSessionAge)
	}
}

func (s *Sweeper) removeDisconnected(ctx context.Context, now time.Time) {
	stale, err := s.repo.ListStaleParticipants(ctx, now.Add(-s.cfg.DisconnectTimeout))
	if err != nil {
		log.Printf("study-session: sweeper: list stale participants: %v", err)
		return
	}
	for _, p := range stale {
		s.remove(ctx, p, events.ReasonDisconnectTimeout)
	}
}

func (s *Sweeper) removeIdle(ctx context.Context, now time.Time) {
	rooms, err := s.repo.ListActiveSessions(ctx)
	if err != nil {
		log.Printf("study-session: sweeper: list rooms: %v", err)
		return
	}
	for _, room := range rooms {
		participants, err := s.repo.GetParticipants(ctx, room.ID)
		if err != nil || len(participants) == 0 {
			continue
		}
		timers, err := s.timers.RoomTimers(ctx, room.ID)
		if err != nil {
			// Without the timers we cannot tell who is idle; removing people
			// because Study Timer is down would be wrong, so wait.
			log.Printf("study-session: sweeper: read timers of room %s: %v", room.ID, err)
			continue
		}
		byUser := make(map[string]domain.TimerActivity, len(timers))
		for _, t := range timers {
			byUser[t.UserID] = t
		}
		for _, p := range participants {
			t, hasTimer := byUser[p.UserID]
			if IsIdle(p, t, hasTimer, now, s.cfg.IdleTimeout) {
				s.remove(ctx, p, events.ReasonIdleTimeout)
			}
		}
	}
}

func (s *Sweeper) remove(ctx context.Context, p domain.Participant, reason string) {
	res, err := s.uc.LeaveSession(ctx, p.SessionID, p.UserID, reason)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			log.Printf("study-session: sweeper: remove %s from %s (%s): %v", p.UserID, p.SessionID, reason, err)
		}
		return
	}
	if res.Left {
		log.Printf("study-session: removed %s from room %s (%s)", p.UserID, p.SessionID, reason)
	}
}

// IsIdle applies UC-05 E-8: a participant is idle when no work cycle is
// running. A paused work cycle and a rest period are not idle. The idle
// clock starts when the user joined, or when their last cycle ended.
func IsIdle(p domain.Participant, t domain.TimerActivity, hasTimer bool, now time.Time, timeout time.Duration) bool {
	idleSince := p.JoinedAt
	if hasTimer {
		switch t.Status {
		case "PAUSED":
			return false
		case "RUNNING":
			if t.RemainingSeconds > 0 {
				return false
			}
			// The phase ran out but nobody completed it. It ended no later
			// than last start/resume + the phase length, so count from there.
			idleSince = latest(idleSince, t.LastUpdated.Add(time.Duration(t.DurationSeconds)*time.Second))
		default: // STOPPED
			idleSince = latest(idleSince, t.LastUpdated)
		}
	}
	return now.Sub(idleSince) >= timeout
}

func latest(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}
