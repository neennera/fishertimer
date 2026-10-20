package amqp

import (
	"testing"
	"time"
)

// A transient failure (database or cache down) must slow the redelivery down,
// not spin: the pause grows and is capped.
func TestBackoffFor_GrowsAndIsCapped(t *testing.T) {
	cases := []struct {
		failures int
		want     time.Duration
	}{
		{0, 0},
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second},
		{7, 30 * time.Second},
		{50, 30 * time.Second},
	}
	for _, c := range cases {
		if got := backoffFor(c.failures); got != c.want {
			t.Errorf("backoffFor(%d) = %s, want %s", c.failures, got, c.want)
		}
	}
}
