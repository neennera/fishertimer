package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/neennera/fishertimer/services/study-timer/internal/domain"
)

var t0 = time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)

func newTimer() *domain.TimerState {
	return domain.NewTimer("room-1", "user-1")
}

func TestNewTimer_IsStoppedWithFullWorkPhase(t *testing.T) {
	timer := newTimer()

	if timer.Status != domain.StatusStopped || timer.Phase != domain.PhaseWork {
		t.Fatalf("got status %s phase %s, want STOPPED WORK", timer.Status, timer.Phase)
	}
	if got, want := timer.Remaining(t0), domain.DefaultWorkMinutes*time.Minute; got != want {
		t.Fatalf("remaining = %v, want %v", got, want)
	}
}

func TestStart_CountsDownFromServerTime(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)

	if timer.Status != domain.StatusRunning {
		t.Fatalf("status = %s, want RUNNING", timer.Status)
	}
	if got, want := timer.Remaining(t0.Add(10*time.Minute)), 15*time.Minute; got != want {
		t.Fatalf("remaining after 10m = %v, want %v", got, want)
	}
}

func TestRemainingSeconds_RoundsUp(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)

	if got, want := timer.RemainingSeconds(t0.Add(400*time.Millisecond)), 25*60; got != want {
		t.Fatalf("remaining seconds = %d, want %d", got, want)
	}
	if got := timer.RemainingSeconds(t0.Add(time.Hour)); got != 0 {
		t.Fatalf("remaining seconds after time ran out = %d, want 0", got)
	}
}

func TestStart_WhileActiveIsIgnored(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)
	timer.Start(t0.Add(5 * time.Minute))

	if got, want := timer.Remaining(t0.Add(5*time.Minute)), 20*time.Minute; got != want {
		t.Fatalf("duplicate start restarted the cycle: remaining = %v, want %v", got, want)
	}
}

func TestStart_AfterTimeRunsOutBeginsNewWorkPhase(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)
	later := t0.Add(30 * time.Minute)
	timer.Start(later)

	if got, want := timer.Remaining(later), 25*time.Minute; got != want {
		t.Fatalf("remaining = %v, want %v", got, want)
	}
}

func TestPauseResume_FreezesAndContinuesRemainingTime(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)

	if err := timer.Pause(t0.Add(10 * time.Minute)); err != nil {
		t.Fatalf("pause: %v", err)
	}
	// Time spent paused must not count.
	if got, want := timer.Remaining(t0.Add(40*time.Minute)), 15*time.Minute; got != want {
		t.Fatalf("remaining while paused = %v, want %v", got, want)
	}

	resumedAt := t0.Add(40 * time.Minute)
	if err := timer.Resume(resumedAt); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if got, want := timer.Remaining(resumedAt.Add(5*time.Minute)), 10*time.Minute; got != want {
		t.Fatalf("remaining after resume = %v, want %v", got, want)
	}
}

func TestPause_RejectedUnlessRunningWithTimeLeft(t *testing.T) {
	cases := map[string]func() *domain.TimerState{
		"stopped": newTimer,
		"paused": func() *domain.TimerState {
			timer := newTimer()
			timer.Start(t0)
			_ = timer.Pause(t0)
			return timer
		},
		"time ran out": func() *domain.TimerState {
			timer := newTimer()
			timer.Start(t0.Add(-time.Hour))
			return timer
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			if err := build().Pause(t0); !errors.Is(err, domain.ErrInvalidState) {
				t.Fatalf("err = %v, want ErrInvalidState", err)
			}
		})
	}
}

func TestResume_RejectedUnlessPaused(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)

	if err := timer.Resume(t0); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("err = %v, want ErrInvalidState", err)
	}
}

func TestReset_ReturnsToInitialStateButKeepsSettings(t *testing.T) {
	timer := newTimer()
	_ = timer.UpdateSetting(50, 10, t0)
	timer.Start(t0)
	timer.CompleteCycle(t0.Add(50 * time.Minute))

	timer.Reset(t0.Add(51 * time.Minute))

	if timer.Status != domain.StatusStopped || timer.Phase != domain.PhaseWork || timer.CurrentCycle != 0 {
		t.Fatalf("got status %s phase %s cycle %d, want STOPPED WORK 0", timer.Status, timer.Phase, timer.CurrentCycle)
	}
	if got, want := timer.Remaining(t0.Add(time.Hour)), 50*time.Minute; got != want {
		t.Fatalf("remaining = %v, want %v (custom work length kept)", got, want)
	}
}

func TestCompleteCycle_OnlyWorkCounts(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)

	if !timer.CompleteCycle(t0) {
		t.Fatal("completing a work phase should report workCompleted")
	}
	if timer.Phase != domain.PhaseRest || timer.CurrentCycle != 1 {
		t.Fatalf("got phase %s cycle %d, want REST 1", timer.Phase, timer.CurrentCycle)
	}
	if timer.CompleteCycle(t0) {
		t.Fatal("completing a rest phase should not report workCompleted")
	}
	if timer.Phase != domain.PhaseWork || timer.CurrentCycle != 1 {
		t.Fatalf("got phase %s cycle %d, want WORK 1", timer.Phase, timer.CurrentCycle)
	}
}

func TestSkipRest_RejectedDuringWork(t *testing.T) {
	timer := newTimer()
	timer.Start(t0)

	if err := timer.SkipRest(t0); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("err = %v, want ErrInvalidState", err)
	}
}

func TestUpdateSetting_RejectsNonPositiveLengths(t *testing.T) {
	if err := newTimer().UpdateSetting(0, 5, t0); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}
