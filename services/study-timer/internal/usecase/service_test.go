package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

type fakeRepository struct {
	timers map[string]domain.TimerState
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{timers: make(map[string]domain.TimerState)}
}

func (r *fakeRepository) GetTimer(ctx context.Context, sessionID, userID string) (*domain.TimerState, error) {
	t, ok := r.timers[sessionID+":"+userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &t, nil
}

func (r *fakeRepository) SaveTimer(ctx context.Context, t *domain.TimerState) error {
	r.timers[t.SessionID+":"+t.UserID] = *t
	return nil
}

func (r *fakeRepository) GetHistory(ctx context.Context, userID string) (*domain.TimerHistory, error) {
	return &domain.TimerHistory{UserID: userID}, nil
}

type fakeRewardClient struct {
	awards int
}

func (m *fakeRewardClient) AwardReward(ctx context.Context, userID, reason string) error {
	m.awards++
	return nil
}

// fakeClock is a wall clock the test moves forward by hand.
type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func newService() (usecase.Usecase, *fakeClock, *fakeRewardClient) {
	clock := &fakeClock{now: time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)}
	reward := &fakeRewardClient{}
	svc := usecase.New(newFakeRepository(), reward, usecase.WithClock(clock.Now))
	return svc, clock, reward
}

func TestGetTimer_UnknownTimerIsStopped(t *testing.T) {
	svc, _, _ := newService()

	timer, err := svc.GetTimer(context.Background(), "room-1", "user-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if timer.Status != domain.StatusStopped {
		t.Fatalf("status = %s, want STOPPED", timer.Status)
	}
}

func TestTimerLifecycle_IsPersistedBetweenCalls(t *testing.T) {
	svc, clock, _ := newService()
	ctx := context.Background()

	if _, err := svc.StartTimer(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	clock.Advance(10 * time.Minute)
	if _, err := svc.PauseTimer(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("pause: %v", err)
	}
	clock.Advance(time.Hour)
	if _, err := svc.ResumeTimer(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("resume: %v", err)
	}
	clock.Advance(5 * time.Minute)

	timer, err := svc.GetTimer(ctx, "room-1", "user-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got, want := timer.Remaining(clock.Now()), 10*time.Minute; got != want {
		t.Fatalf("remaining = %v, want %v", got, want)
	}

	timer, err = svc.ResetTimer(ctx, "room-1", "user-1")
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if timer.Status != domain.StatusStopped {
		t.Fatalf("status after reset = %s, want STOPPED", timer.Status)
	}
}

func TestPauseTimer_NotRunningIsInvalidState(t *testing.T) {
	svc, _, _ := newService()

	_, err := svc.PauseTimer(context.Background(), "room-1", "user-1")
	if !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("err = %v, want ErrInvalidState", err)
	}
}

func TestTimer_MissingIDsAreInvalid(t *testing.T) {
	svc, _, _ := newService()

	_, err := svc.StartTimer(context.Background(), "", "user-1")
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}

func TestCompleteCycle_AwardsOnlyForWork(t *testing.T) {
	svc, _, reward := newService()
	ctx := context.Background()

	if _, err := svc.StartTimer(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := svc.CompleteCycle(ctx, "room-1", "user-1"); err != nil { // work -> rest
		t.Fatalf("complete work: %v", err)
	}
	if _, err := svc.CompleteCycle(ctx, "room-1", "user-1"); err != nil { // rest -> work
		t.Fatalf("complete rest: %v", err)
	}
	if reward.awards != 1 {
		t.Fatalf("awards = %d, want 1", reward.awards)
	}
}
