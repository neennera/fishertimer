package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/adapter/repository"
	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
	"github.com/neennera/fishertimer/services/study-timer/internal/usecase"
)

const (
	room  = "aaaaaaaa-0000-4000-8000-0000000000aa"
	alice = "aaaaaaaa-0000-4000-8000-000000000001"
)

// t0 is in the future of the real clock, so the in-memory repository's
// finalize (which uses time.Now) sees no cycle as due.
var t0 = time.Date(2030, 1, 1, 9, 0, 0, 0, time.UTC)

type fakeRewards struct {
	sent []domain.AwardRequest
	down bool
}

func (f *fakeRewards) AwardReward(ctx context.Context, req domain.AwardRequest) error {
	if f.down {
		return errors.New("reward unavailable")
	}
	f.sent = append(f.sent, req)
	return nil
}

type fakeSessions struct {
	count int
	err   error
}

func (f *fakeSessions) ParticipantCount(ctx context.Context, sessionID string) (int, error) {
	return f.count, f.err
}

type fixture struct {
	uc       usecase.Usecase
	repo     *repository.InMemoryRepository
	rewards  *fakeRewards
	sessions *fakeSessions
	now      time.Time
}

func newFixture() *fixture {
	f := &fixture{repo: repository.NewInMemory(), rewards: &fakeRewards{}, sessions: &fakeSessions{count: 3}, now: t0}
	f.uc = usecase.New(f.repo, f.rewards,
		usecase.WithClock(func() time.Time { return f.now }),
		usecase.WithSessionClient(f.sessions))
	return f
}

func (f *fixture) advance(d time.Duration) { f.now = f.now.Add(d) }

func (f *fixture) state(t *testing.T) *domain.View {
	t.Helper()
	v, err := f.uc.GetTimer(context.Background(), room, alice)
	if err != nil {
		t.Fatalf("GetTimer: %v", err)
	}
	return v
}

func (f *fixture) start(t *testing.T, phase domain.TimerPhase, minutes int) *domain.View {
	t.Helper()
	v, err := f.uc.StartTimer(context.Background(), room, alice, phase, minutes)
	if err != nil {
		t.Fatalf("StartTimer(%s, %d): %v", phase, minutes, err)
	}
	return v
}

func TestNewTimerIsReadyWithDefaults(t *testing.T) {
	f := newFixture()
	v := f.state(t)
	if v.State != domain.StateReady || v.DurationSeconds != domain.DefaultWorkMinutes*60 {
		t.Fatalf("view = %+v", v)
	}
}

func TestStart_ValidatesDurationAndIgnoresDuplicates(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	if _, err := f.uc.StartTimer(ctx, room, alice, domain.PhaseWork, 0+500); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("too long: %v, want ErrInvalid (E-1)", err)
	}
	if _, err := f.uc.StartTimer(ctx, room, alice, domain.PhaseRest, 5); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("rest before any work: %v", err)
	}

	v := f.start(t, domain.PhaseWork, 45)
	if v.State != domain.StateWorkRunning || v.DurationSeconds != 45*60 {
		t.Fatalf("started = %+v", v)
	}
	f.advance(time.Minute)
	again := f.start(t, domain.PhaseWork, 10)
	if again.CycleID != v.CycleID || again.DurationSeconds != 45*60 {
		t.Fatalf("duplicate start replaced the cycle (E-4): %+v", again)
	}
}

func TestFullCycle_WorkRewardsThenRest(t *testing.T) {
	f := newFixture()
	work := f.start(t, domain.PhaseWork, 25)

	f.advance(10 * time.Minute)
	if v := f.state(t); v.RemainingSeconds != 15*60 {
		t.Fatalf("remaining = %d", v.RemainingSeconds)
	}

	// Time runs out with no client call: the next read settles it.
	f.advance(20 * time.Minute)
	v := f.state(t)
	if v.State != domain.StateReadyForRest || v.CompletedWork != 1 || v.FocusSeconds != 25*60 {
		t.Fatalf("after work = %+v", v)
	}
	if v.LastCompletedID != work.CycleID || !v.LastCompletedAt.Equal(t0.Add(25*time.Minute)) {
		t.Fatalf("last completed = %s at %v", v.LastCompletedID, v.LastCompletedAt)
	}
	want := domain.AwardRequest{UserID: alice, SessionID: room, CycleID: work.CycleID, WorkMinutes: 25, ParticipantCount: 3}
	if len(f.rewards.sent) != 1 || f.rewards.sent[0] != want {
		t.Fatalf("rewards = %+v, want %+v", f.rewards.sent, want)
	}

	rest := f.start(t, domain.PhaseRest, 5)
	if rest.State != domain.StateRestRunning || rest.Phase != domain.PhaseRest {
		t.Fatalf("rest = %+v", rest)
	}
	f.advance(6 * time.Minute)
	if v := f.state(t); v.State != domain.StateReady {
		t.Fatalf("after rest = %s", v.State)
	}
	if len(f.rewards.sent) != 1 {
		t.Fatalf("rest earned a reward")
	}
}

func TestStop_DiscardsWithoutReward(t *testing.T) {
	f := newFixture()
	f.start(t, domain.PhaseWork, 25)
	f.advance(20 * time.Minute)
	v, err := f.uc.StopTimer(context.Background(), room, alice)
	if err != nil || v.State != domain.StateReady || v.CompletedWork != 0 {
		t.Fatalf("stop = %+v, %v", v, err)
	}
	f.advance(time.Hour)
	f.uc.Sweep(context.Background())
	if len(f.rewards.sent) != 0 {
		t.Fatalf("stopped cycle was rewarded")
	}
}

func TestPauseResumeAndPauseLimit(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	f.start(t, domain.PhaseWork, 25)
	f.advance(5 * time.Minute)
	if _, err := f.uc.PauseTimer(ctx, room, alice); err != nil {
		t.Fatal(err)
	}
	f.advance(10 * time.Minute)
	v, err := f.uc.ResumeTimer(ctx, room, alice)
	if err != nil || v.RemainingSeconds != 20*60 || v.PausedTotalSeconds != 600 {
		t.Fatalf("resume = %+v, %v", v, err)
	}

	// Paused longer than MAX_PAUSE_MINUTES: discarded by the sweeper (E-3).
	f.uc.PauseTimer(ctx, room, alice)
	f.advance(16 * time.Minute)
	f.uc.Sweep(ctx)
	if v := f.state(t); v.State != domain.StateReady || v.CompletedWork != 0 {
		t.Fatalf("after long pause = %+v", v)
	}
	if _, err := f.uc.PauseTimer(ctx, room, alice); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("pause with nothing running: %v", err)
	}
}

func TestReset_RewindsActiveCycleAndHoldsIt(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	f.start(t, domain.PhaseWork, 25)
	f.advance(20 * time.Minute)
	v, err := f.uc.ResetTimer(ctx, room, alice)
	if err != nil || v.RemainingSeconds != 25*60 || v.State != domain.StateWorkPaused {
		t.Fatalf("reset = %+v, %v", v, err)
	}
	f.advance(5 * time.Minute)
	if v, _ = f.uc.GetTimer(ctx, room, alice); v.RemainingSeconds != 25*60 {
		t.Fatalf("reset timer kept counting: %+v", v)
	}
	if v, err = f.uc.ResumeTimer(ctx, room, alice); err != nil || v.State != domain.StateWorkRunning {
		t.Fatalf("resume after reset = %+v, %v", v, err)
	}
	f.advance(10 * time.Minute)
	if v, _ = f.uc.GetTimer(ctx, room, alice); v.RemainingSeconds != 15*60 {
		t.Fatalf("after resume = %+v", v)
	}
}

func TestComplete_IsServerCheckedAndIdempotent(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	f.start(t, domain.PhaseWork, 25)
	f.advance(24 * time.Minute)
	if _, err := f.uc.CompleteCycle(ctx, room, alice); !errors.Is(err, domain.ErrNotFinished) {
		t.Fatalf("early complete: %v", err)
	}
	f.advance(time.Minute)
	for i := 0; i < 3; i++ {
		v, err := f.uc.CompleteCycle(ctx, room, alice)
		if err != nil || v.CompletedWork != 1 {
			t.Fatalf("complete #%d = %+v, %v", i, v, err)
		}
	}
	if len(f.rewards.sent) != 1 {
		t.Fatalf("%d rewards for one cycle", len(f.rewards.sent))
	}
}

func TestSkipRest(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	f.start(t, domain.PhaseWork, 1)
	f.advance(time.Minute)
	v, err := f.uc.SkipRest(ctx, room, alice)
	if err != nil || v.State != domain.StateReady {
		t.Fatalf("skip unstarted rest = %+v, %v", v, err)
	}
	if _, err := f.uc.SkipRest(ctx, room, alice); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("skip with no rest due: %v", err)
	}

	f.start(t, domain.PhaseWork, 1)
	f.advance(time.Minute)
	f.start(t, domain.PhaseRest, 10)
	f.advance(2 * time.Minute)
	if v, _ := f.uc.SkipRest(ctx, room, alice); v.State != domain.StateReady {
		t.Fatalf("skip running rest = %s", v.State)
	}
}

func TestRewardRetry_WhenRewardOrSessionIsDown(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	f.rewards.down = true
	work := f.start(t, domain.PhaseWork, 15)

	// Completed by the sweeper with nobody watching (S-2), reward fails (E-7).
	f.advance(15 * time.Minute)
	f.uc.Sweep(ctx)
	if v := f.state(t); v.CompletedWork != 1 {
		t.Fatalf("cycle not completed: %+v", v)
	}
	if len(f.rewards.sent) != 0 {
		t.Fatalf("sent while down")
	}

	f.rewards.down = false
	f.sessions.err = errors.New("session unreachable")
	f.uc.Sweep(ctx)
	if len(f.rewards.sent) != 0 {
		t.Fatalf("sent without a participant count")
	}

	f.sessions.err = nil
	f.uc.Sweep(ctx)
	f.uc.Sweep(ctx)
	if len(f.rewards.sent) != 1 || f.rewards.sent[0].CycleID != work.CycleID {
		t.Fatalf("rewards = %+v, want exactly one for %s", f.rewards.sent, work.CycleID)
	}
}

func TestEvents_FinalizeAndReopen(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	if err := f.uc.OpenParticipantTimer(ctx, room, alice, "ev-join-1"); err != nil {
		t.Fatal(err)
	}
	f.start(t, domain.PhaseWork, 1)
	f.advance(time.Minute)
	f.state(t) // completes one cycle
	f.start(t, domain.PhaseWork, 25)

	if err := f.uc.FinalizeParticipantTimer(ctx, room, alice, "ev-left-1", "LEFT"); err != nil {
		t.Fatal(err)
	}
	if v := f.state(t); v.State != domain.StateFinalized {
		t.Fatalf("after leave = %s", v.State)
	}
	if _, err := f.uc.StartTimer(ctx, room, alice, domain.PhaseWork, 25); !errors.Is(err, domain.ErrTimerClosed) {
		t.Fatalf("start on closed timer: %v, want ErrTimerClosed (E-2)", err)
	}
	// A redelivered event changes nothing.
	if err := f.uc.FinalizeParticipantTimer(ctx, room, alice, "ev-left-1", "LEFT"); err != nil {
		t.Fatal(err)
	}

	// Rejoining reopens the timer and starts the stay's counters from zero.
	f.advance(time.Minute)
	if err := f.uc.OpenParticipantTimer(ctx, room, alice, "ev-join-2"); err != nil {
		t.Fatal(err)
	}
	v := f.state(t)
	if v.State != domain.StateReady || v.CompletedWork != 0 {
		t.Fatalf("after rejoin = %+v", v)
	}

	if err := f.uc.OpenParticipantTimer(ctx, "not-a-uuid", alice, "ev-bad"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("bad ids: %v", err)
	}
}

func TestSettings(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	if _, err := f.uc.UpdateTimerSetting(ctx, "", alice, 50, 0); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("rest 0: %v", err)
	}
	if _, err := f.uc.UpdateTimerSetting(ctx, "", alice, 50, 10); err != nil {
		t.Fatal(err)
	}
	v, err := f.uc.GetTimer(ctx, "", alice)
	if err != nil || v.Settings.WorkMinutes != 50 || v.Settings.RestMinutes != 10 {
		t.Fatalf("settings view = %+v, %v", v, err)
	}
	if _, err := f.uc.PauseTimer(ctx, "", alice); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("action without a room: %v", err)
	}
	if v := f.start(t, domain.PhaseWork, 0); v.DurationSeconds != 50*60 {
		t.Fatalf("default work = %d", v.DurationSeconds)
	}
}

func TestStatistics(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	f.start(t, domain.PhaseWork, 30)
	f.advance(30 * time.Minute)
	f.state(t)
	h, err := f.uc.TimerStatistics(ctx, alice)
	if err != nil || h.CyclesCompleted != 1 || h.TotalFocusMinutes != 30 || h.SessionsJoined != 1 {
		t.Fatalf("history = %+v, %v", h, err)
	}
}
