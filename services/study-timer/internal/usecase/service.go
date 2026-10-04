package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

// collaboratorTimeout bounds each call to Study Session or Reward made while
// delivering a reward.
const collaboratorTimeout = 3 * time.Second

type Usecase interface {
	GetTimer(ctx context.Context, sessionID, userID string) (*domain.View, error)
	// GetRoomTimers returns every open timer in the room (UC-02: others'
	// timer states are visible).
	GetRoomTimers(ctx context.Context, sessionID string) ([]domain.View, error)
	// StartTimer starts a work cycle or, after a completed work cycle, a
	// rest period. minutes 0 means the user's saved default.
	StartTimer(ctx context.Context, sessionID, userID string, phase domain.TimerPhase, minutes int) (*domain.View, error)
	PauseTimer(ctx context.Context, sessionID, userID string) (*domain.View, error)
	ResumeTimer(ctx context.Context, sessionID, userID string) (*domain.View, error)
	// StopTimer discards the active cycle; a work cycle earns nothing (S-4).
	StopTimer(ctx context.Context, sessionID, userID string) (*domain.View, error)
	// ResetTimer restarts the active cycle from its full length.
	ResetTimer(ctx context.Context, sessionID, userID string) (*domain.View, error)
	// CompleteCycle completes the active cycle if its time is really up.
	CompleteCycle(ctx context.Context, sessionID, userID string) (*domain.View, error)
	// SkipRest ends a running rest period, or skips one not yet started.
	SkipRest(ctx context.Context, sessionID, userID string) (*domain.View, error)
	// UpdateTimerSetting saves the user's default lengths. sessionID is
	// optional and only picks which timer to return.
	UpdateTimerSetting(ctx context.Context, sessionID, userID string, workMin, restMin int) (*domain.View, error)
	TimerStatistics(ctx context.Context, userID string) (*domain.TimerHistory, error)

	// RabbitMQ event handlers (Study Session -> Study Timer).
	OpenParticipantTimer(ctx context.Context, sessionID, userID, eventID string) error
	FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, reason string) error
	FinalizeSessionTimers(ctx context.Context, sessionID, eventID, reason string) error

	// Sweep completes cycles whose time is up even if no client is watching,
	// discards cycles paused too long (E-3) and retries undelivered rewards
	// (E-7). Run by a background ticker.
	Sweep(ctx context.Context)
}

type service struct {
	repo     domain.Repository
	rewards  domain.RewardClient
	sessions domain.SessionClient // optional; nil counts the room as 1
	limits   domain.Limits
	now      func() time.Time
}

// Option customises the usecase service.
type Option func(*service)

// WithClock replaces the wall clock, so tests can control elapsed time.
func WithClock(now func() time.Time) Option {
	return func(s *service) { s.now = now }
}

// WithLimits replaces the default duration limits.
func WithLimits(l domain.Limits) Option {
	return func(s *service) { s.limits = l }
}

// WithSessionClient sets where participant counts come from.
func WithSessionClient(c domain.SessionClient) Option {
	return func(s *service) { s.sessions = c }
}

func New(repo domain.Repository, rewards domain.RewardClient, opts ...Option) Usecase {
	s := &service{
		repo:    repo,
		rewards: rewards,
		limits:  domain.DefaultLimits,
		now:     func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// load returns the user's timer in the room with any due work settled, or a
// fresh ready timer (not stored) if they have none yet.
func (s *service) load(ctx context.Context, sessionID, userID string) (*domain.Timer, error) {
	// No room: just the user's settings and limits (the settings page).
	// Every action on such a timer is refused (no room to run it in).
	if sessionID == "" && isUUID(userID) {
		settings, err := s.repo.GetSettings(ctx, userID)
		if err != nil {
			return nil, err
		}
		return domain.NewTimer("", userID, settings, s.now()), nil
	}
	if !isUUID(sessionID) || !isUUID(userID) {
		return nil, domain.ErrInvalid
	}
	t, err := s.repo.GetTimer(ctx, sessionID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		settings, err := s.repo.GetSettings(ctx, userID)
		if err != nil {
			return nil, err
		}
		return domain.NewTimer(sessionID, userID, settings, s.now()), nil
	}
	if err != nil {
		return nil, err
	}
	return s.settle(ctx, t)
}

// settle applies what time has already decided: a cycle that ran out is
// completed, one paused too long is discarded. Reads stay correct even
// between sweeps.
func (s *service) settle(ctx context.Context, t *domain.Timer) (*domain.Timer, error) {
	c, now := t.Active, s.now()
	if c == nil || t.Status != domain.SessionTimerOpen {
		return t, nil
	}
	switch {
	case c.Due(now):
		if err := s.complete(ctx, domain.CycleRef{Cycle: *c, SessionID: t.SessionID, UserID: t.UserID}, now); err != nil {
			return nil, err
		}
	case s.pausedTooLong(c, now):
		if err := s.discard(ctx, *c, now); err != nil {
			return nil, err
		}
	default:
		return t, nil
	}
	return s.repo.GetTimer(ctx, t.SessionID, t.UserID)
}

func (s *service) pausedTooLong(c *domain.Cycle, now time.Time) bool {
	return c.Status == domain.CyclePaused && c.PausedAt != nil &&
		now.Sub(*c.PausedAt) > time.Duration(s.limits.MaxPauseMinutes)*time.Minute
}

// complete marks a due cycle completed and, for work, delivers its reward.
func (s *service) complete(ctx context.Context, ref domain.CycleRef, now time.Time) error {
	c := ref.Cycle
	if err := c.Complete(now); err != nil {
		return err
	}
	if err := s.repo.UpdateCycle(ctx, &c, domain.CycleRunning); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil // someone else completed or stopped it first
		}
		return err
	}
	if c.Type == domain.PhaseWork {
		ref.Cycle = c
		s.deliver(ctx, ref)
	}
	return nil
}

func (s *service) discard(ctx context.Context, c domain.Cycle, now time.Time) error {
	if err := c.End(domain.CycleDiscarded, now); err != nil {
		return err
	}
	err := s.repo.UpdateCycle(ctx, &c, domain.CycleRunning, domain.CyclePaused)
	if errors.Is(err, domain.ErrConflict) {
		return nil
	}
	return err
}

// deliver sends one completed work cycle to Reward with the room's current
// participant count (UC-09 steps 2-3). Failures leave it PENDING for the
// sweeper to retry; Reward ignores a repeated cycle_id.
func (s *service) deliver(ctx context.Context, ref domain.CycleRef) {
	if s.rewards == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, collaboratorTimeout)
	defer cancel()

	count := 1
	if s.sessions != nil {
		n, err := s.sessions.ParticipantCount(ctx, ref.SessionID)
		if err != nil {
			log.Printf("study-timer: reward for cycle %s waits: participant count: %v", ref.CycleID, err)
			return
		}
		count = max(n, 1)
	}
	err := s.rewards.AwardReward(ctx, domain.AwardRequest{
		UserID:           ref.UserID,
		SessionID:        ref.SessionID,
		CycleID:          ref.CycleID,
		WorkMinutes:      ref.DurationSec / 60,
		ParticipantCount: count,
	})
	if err != nil {
		log.Printf("study-timer: reward for cycle %s waits: %v", ref.CycleID, err)
		return
	}
	if err := s.repo.MarkRewardSent(ctx, ref.CycleID); err != nil {
		log.Printf("study-timer: mark reward of cycle %s sent: %v", ref.CycleID, err)
	}
}

func (s *service) view(t *domain.Timer) *domain.View {
	v := t.View(s.now(), s.limits)
	return &v
}

// act loads an open timer, applies change to its active cycle and saves it.
func (s *service) act(ctx context.Context, sessionID, userID string, change func(t *domain.Timer, now time.Time) error) (*domain.View, error) {
	t, err := s.load(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.SessionTimerFinalized {
		return nil, domain.ErrTimerClosed
	}
	if t.SessionID == "" {
		return nil, domain.ErrInvalid
	}
	if err := change(t, s.now()); err != nil {
		return nil, err
	}
	return s.reload(ctx, sessionID, userID)
}

func (s *service) reload(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	t, err := s.load(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	return s.view(t), nil
}

// saveActive applies f to the active cycle and stores it, if the cycle is
// still in the status it was loaded in. With no active cycle (it just ran
// out, or another tab ended it) there is nothing to change: the caller gets
// the current state back instead of an error.
func (s *service) saveActive(ctx context.Context, t *domain.Timer, f func(c *domain.Cycle) error) error {
	if t.Active == nil {
		return nil
	}
	c := *t.Active
	from := c.Status
	if err := f(&c); err != nil {
		return err
	}
	err := s.repo.UpdateCycle(ctx, &c, from)
	if errors.Is(err, domain.ErrConflict) {
		return nil // a concurrent request already moved it; show the latest
	}
	return err
}

func (s *service) GetTimer(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	t, err := s.load(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	return s.view(t), nil
}

func (s *service) GetRoomTimers(ctx context.Context, sessionID string) ([]domain.View, error) {
	if !isUUID(sessionID) {
		return nil, domain.ErrInvalid
	}
	timers, err := s.repo.ListRoomTimers(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	views := make([]domain.View, 0, len(timers))
	for _, t := range timers {
		if t, err = s.settle(ctx, t); err != nil {
			return nil, err
		}
		views = append(views, *s.view(t))
	}
	return views, nil
}

func (s *service) StartTimer(ctx context.Context, sessionID, userID string, phase domain.TimerPhase, minutes int) (*domain.View, error) {
	if phase == "" {
		phase = domain.PhaseWork
	}
	if phase != domain.PhaseWork && phase != domain.PhaseRest {
		return nil, domain.ErrInvalid
	}
	if !isUUID(sessionID) || !isUUID(userID) {
		return nil, domain.ErrInvalid
	}
	// The participant.joined event normally opens the timer first; a press
	// that beats the event opens it here. A finalized timer stays closed.
	if err := s.repo.EnsureTimer(ctx, sessionID, userID, s.now()); err != nil {
		return nil, err
	}
	return s.act(ctx, sessionID, userID, func(t *domain.Timer, now time.Time) error {
		if minutes == 0 {
			minutes = t.Settings.WorkMinutes
			if phase == domain.PhaseRest {
				minutes = t.Settings.RestMinutes
			}
		}
		if err := s.limits.Check(phase, minutes); err != nil {
			return err
		}
		if t.Active != nil {
			return nil // duplicate Start: keep the single active cycle (E-4)
		}
		if phase == domain.PhaseRest && t.State() != domain.StateReadyForRest {
			return domain.ErrInvalidState
		}
		err := s.repo.InsertCycle(ctx, &domain.Cycle{
			TimerID:      t.TimerID,
			Type:         phase,
			Status:       domain.CycleRunning,
			DurationSec:  minutes * 60,
			StartedAt:    now,
			RewardStatus: domain.RewardNone,
		})
		if errors.Is(err, domain.ErrInvalidState) {
			return nil // a concurrent Start won
		}
		return err
	})
}

func (s *service) PauseTimer(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	return s.act(ctx, sessionID, userID, func(t *domain.Timer, now time.Time) error {
		if t.Active == nil {
			return domain.ErrInvalidState
		}
		return s.saveActive(ctx, t, func(c *domain.Cycle) error { return c.Pause(now) })
	})
}

func (s *service) ResumeTimer(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	return s.act(ctx, sessionID, userID, func(t *domain.Timer, now time.Time) error {
		if t.Active == nil {
			return domain.ErrInvalidState
		}
		return s.saveActive(ctx, t, func(c *domain.Cycle) error { return c.Resume(now) })
	})
}

func (s *service) StopTimer(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	return s.act(ctx, sessionID, userID, func(t *domain.Timer, now time.Time) error {
		if t.Active == nil {
			return nil // nothing to stop
		}
		return s.saveActive(ctx, t, func(c *domain.Cycle) error { return c.End(domain.CycleDiscarded, now) })
	})
}

func (s *service) ResetTimer(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	return s.act(ctx, sessionID, userID, func(t *domain.Timer, now time.Time) error {
		if t.Active == nil {
			return nil // nothing to reset
		}
		return s.saveActive(ctx, t, func(c *domain.Cycle) error { return c.Restart(now) })
	})
}

func (s *service) CompleteCycle(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	return s.act(ctx, sessionID, userID, func(t *domain.Timer, now time.Time) error {
		// load() already completed a cycle whose time is up, so an active
		// cycle here still has time left. No active cycle: already done.
		if t.Active == nil {
			return nil
		}
		if t.Active.Status == domain.CyclePaused {
			return domain.ErrInvalidState
		}
		return domain.ErrNotFinished
	})
}

func (s *service) SkipRest(ctx context.Context, sessionID, userID string) (*domain.View, error) {
	return s.act(ctx, sessionID, userID, func(t *domain.Timer, now time.Time) error {
		switch t.State() {
		case domain.StateRestRunning, domain.StateRestPaused:
			return s.saveActive(ctx, t, func(c *domain.Cycle) error { return c.End(domain.CycleSkipped, now) })
		case domain.StateReadyForRest:
			// Record the skipped break, which also returns the timer to READY.
			return s.repo.InsertCycle(ctx, &domain.Cycle{
				TimerID:      t.TimerID,
				Type:         domain.PhaseRest,
				Status:       domain.CycleSkipped,
				StartedAt:    now,
				EndedAt:      &now,
				RewardStatus: domain.RewardNone,
			})
		default:
			return domain.ErrInvalidState
		}
	})
}

func (s *service) UpdateTimerSetting(ctx context.Context, sessionID, userID string, workMin, restMin int) (*domain.View, error) {
	if !isUUID(userID) {
		return nil, domain.ErrInvalid
	}
	if err := s.limits.Check(domain.PhaseWork, workMin); err != nil {
		return nil, err
	}
	if err := s.limits.Check(domain.PhaseRest, restMin); err != nil {
		return nil, err
	}
	if err := s.repo.SaveSettings(ctx, userID, domain.Settings{WorkMinutes: workMin, RestMinutes: restMin}); err != nil {
		return nil, err
	}
	if sessionID == "" {
		return s.view(domain.NewTimer("", userID, domain.Settings{WorkMinutes: workMin, RestMinutes: restMin}, s.now())), nil
	}
	return s.GetTimer(ctx, sessionID, userID)
}

func (s *service) TimerStatistics(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	if !isUUID(userID) {
		return nil, domain.ErrInvalid
	}
	return s.repo.GetHistory(ctx, userID)
}

func (s *service) OpenParticipantTimer(ctx context.Context, sessionID, userID, eventID string) error {
	if !isUUID(sessionID) || !isUUID(userID) || eventID == "" {
		return domain.ErrInvalid
	}
	return s.repo.OpenParticipantTimer(ctx, sessionID, userID, eventID, s.now())
}

func (s *service) FinalizeParticipantTimer(ctx context.Context, sessionID, userID, eventID, reason string) error {
	if !isUUID(sessionID) || !isUUID(userID) || eventID == "" {
		return domain.ErrInvalid
	}
	return s.repo.FinalizeParticipantTimer(ctx, sessionID, userID, eventID, "session.participant.left")
}

func (s *service) FinalizeSessionTimers(ctx context.Context, sessionID, eventID, reason string) error {
	if !isUUID(sessionID) || eventID == "" {
		return domain.ErrInvalid
	}
	return s.repo.FinalizeSessionTimers(ctx, sessionID, eventID, "session.ended")
}

// sweepBatch caps how many cycles one sweep handles per job.
const sweepBatch = 100

func (s *service) Sweep(ctx context.Context) {
	now := s.now()

	due, err := s.repo.DueCycles(ctx, now, sweepBatch)
	if err != nil {
		log.Printf("study-timer: sweep: due cycles: %v", err)
	}
	for _, ref := range due {
		if err := s.complete(ctx, ref, now); err != nil {
			log.Printf("study-timer: sweep: complete cycle %s: %v", ref.CycleID, err)
		}
	}

	cutoff := now.Add(-time.Duration(s.limits.MaxPauseMinutes) * time.Minute)
	stale, err := s.repo.StalePausedCycles(ctx, cutoff, sweepBatch)
	if err != nil {
		log.Printf("study-timer: sweep: paused cycles: %v", err)
	}
	for _, ref := range stale {
		if err := s.discard(ctx, ref.Cycle, now); err != nil {
			log.Printf("study-timer: sweep: discard cycle %s: %v", ref.CycleID, err)
		}
	}

	pending, err := s.repo.PendingRewards(ctx, sweepBatch)
	if err != nil {
		log.Printf("study-timer: sweep: pending rewards: %v", err)
	}
	for _, ref := range pending {
		s.deliver(ctx, ref)
	}
}

// isUUID reports whether s is a canonical 8-4-4-4-12 hex UUID. Ids are UUID
// columns in timer_db, so anything else is rejected as invalid input rather
// than surfacing as a database error.
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
				return false
			}
		}
	}
	return true
}
