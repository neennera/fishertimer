package domain

import (
	"errors"
	"testing"
	"time"
)

var t0 = time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)

func running(minutes int) *Cycle {
	return &Cycle{Type: PhaseWork, Status: CycleRunning, DurationSec: minutes * 60, StartedAt: t0, RewardStatus: RewardNone}
}

func TestCycle_PauseFreezesAndResumeContinues(t *testing.T) {
	c := running(25)
	if err := c.Pause(t0.Add(10 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	// Paused time does not count, however long it lasts.
	if got := c.Remaining(t0.Add(40 * time.Minute)); got != 15*time.Minute {
		t.Fatalf("paused remaining = %v, want 15m", got)
	}
	if err := c.Resume(t0.Add(40 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	if c.PausedTotalSec != 30*60 {
		t.Fatalf("paused total = %d, want 1800", c.PausedTotalSec)
	}
	if got := c.Remaining(t0.Add(45 * time.Minute)); got != 10*time.Minute {
		t.Fatalf("remaining after resume = %v, want 10m", got)
	}
	if !c.EndsAt().Equal(t0.Add(55 * time.Minute)) {
		t.Fatalf("ends at %v", c.EndsAt())
	}
}

func TestCycle_CompleteOnlyWhenDue(t *testing.T) {
	c := running(25)
	if err := c.Complete(t0.Add(24 * time.Minute)); !errors.Is(err, ErrNotFinished) {
		t.Fatalf("early complete: %v, want ErrNotFinished", err)
	}
	if err := c.Complete(t0.Add(40 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	// Ends when it ran out, not when the completion was noticed.
	if !c.EndedAt.Equal(t0.Add(25*time.Minute)) || c.RewardStatus != RewardPending {
		t.Fatalf("completed = %+v", c)
	}

	rest := &Cycle{Type: PhaseRest, Status: CycleRunning, DurationSec: 300, StartedAt: t0}
	_ = rest.Complete(t0.Add(5 * time.Minute))
	if rest.RewardStatus == RewardPending {
		t.Fatalf("rest earned a reward")
	}
}

func TestCycle_PauseRejectedOnceDue(t *testing.T) {
	c := running(1)
	if err := c.Pause(t0.Add(2 * time.Minute)); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("pause after time is up: %v", err)
	}
}

func TestTimer_StateMachine(t *testing.T) {
	tm := NewTimer("s", "u", DefaultSettings, t0)
	if tm.State() != StateReady {
		t.Fatalf("new timer = %s", tm.State())
	}
	tm.Active = running(25)
	if tm.State() != StateWorkRunning {
		t.Fatalf("= %s", tm.State())
	}
	_ = tm.Active.Pause(t0.Add(time.Minute))
	if tm.State() != StateWorkPaused {
		t.Fatalf("= %s", tm.State())
	}

	done := running(25)
	_ = done.Complete(t0.Add(30 * time.Minute))
	tm.Active, tm.Last = nil, done
	if tm.State() != StateReadyForRest {
		t.Fatalf("after work = %s", tm.State())
	}
	v := tm.View(t0.Add(30*time.Minute), DefaultLimits)
	if v.Phase != PhaseRest || v.Status != StatusStopped || v.DurationSeconds != DefaultRestMinutes*60 {
		t.Fatalf("ready-for-rest view = %+v", v)
	}

	tm.Active = &Cycle{Type: PhaseRest, Status: CycleRunning, DurationSec: 300, StartedAt: t0}
	if tm.State() != StateRestRunning {
		t.Fatalf("= %s", tm.State())
	}
	skipped := &Cycle{Type: PhaseRest, Status: CycleSkipped}
	tm.Active, tm.Last = nil, skipped
	if tm.State() != StateReady {
		t.Fatalf("after rest = %s", tm.State())
	}

	tm.Status = SessionTimerFinalized
	if tm.State() != StateFinalized {
		t.Fatalf("= %s", tm.State())
	}
}

func TestLimits_Check(t *testing.T) {
	l := Limits{MinWorkMinutes: 5, MaxWorkMinutes: 60, MinRestMinutes: 1, MaxRestMinutes: 30}
	cases := []struct {
		phase TimerPhase
		min   int
		ok    bool
	}{
		{PhaseWork, 4, false}, {PhaseWork, 5, true}, {PhaseWork, 60, true}, {PhaseWork, 61, false},
		{PhaseRest, 0, false}, {PhaseRest, 30, true}, {PhaseRest, 31, false},
	}
	for _, c := range cases {
		if err := l.Check(c.phase, c.min); (err == nil) != c.ok {
			t.Errorf("%s %d: err = %v", c.phase, c.min, err)
		}
	}
}
