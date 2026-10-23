package discovery

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSource struct {
	calls int
	out   map[string][]Instance
	err   error
}

func (f *fakeSource) Resolve(_ context.Context, service string) ([]Instance, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.out[service], nil
}

func quiet() Option { return WithLogf(func(string, ...any) {}) }

func TestStaticLocatorReturnsConfiguredAddresses(t *testing.T) {
	l := NewLocator(nil, map[string]string{
		Account:          "http://account:8082",
		GRPC(StudyTimer): "study-timer:50051",
	}, quiet())

	if l.Enabled() {
		t.Fatal("no source means static mode")
	}
	if got := l.URL(Account); got != "http://account:8082" {
		t.Fatalf("URL = %q", got)
	}
	if got := l.Target(GRPC(StudyTimer)); got != "study-timer:50051" {
		t.Fatalf("Target = %q", got)
	}
	if got := l.Target(Account); got != "account:8082" {
		t.Fatalf("host:port taken from a URL, got %q", got)
	}
	if l.Target("nope") != "" || l.URL("nope") != "" || l.Targets("nope") != nil {
		t.Fatal("an unconfigured service has no address")
	}
}

func TestLocatorSpreadsCallsAndKeepsScheme(t *testing.T) {
	src := &fakeSource{out: map[string][]Instance{
		Account: {{Host: "10.0.0.1", Port: 8082}, {Host: "10.0.0.2", Port: 8082}},
	}}
	l := NewLocator(src, map[string]string{Account: "https://account:8082"}, quiet())

	a, b, c := l.URL(Account), l.URL(Account), l.URL(Account)
	if a != "https://10.0.0.1:8082" || b != "https://10.0.0.2:8082" || c != a {
		t.Fatalf("round robin: %s %s %s", a, b, c)
	}
	if got := l.Targets(Account); len(got) != 2 {
		t.Fatalf("targets %v", got)
	}
}

func TestLocatorCachesForTTL(t *testing.T) {
	now := time.Unix(1000, 0)
	src := &fakeSource{out: map[string][]Instance{Account: {{Host: "h", Port: 1}}}}
	l := NewLocator(src, nil, quiet(), WithTTL(3*time.Second), withClock(func() time.Time { return now }))

	l.Target(Account)
	l.Target(Account)
	if src.calls != 1 {
		t.Fatalf("within the TTL the registry is asked once, got %d", src.calls)
	}
	now = now.Add(4 * time.Second)
	l.Target(Account)
	if src.calls != 2 {
		t.Fatalf("after the TTL it is asked again, got %d", src.calls)
	}
}

func TestLocatorFallsBackWhenRegistryHasNoInstances(t *testing.T) {
	src := &fakeSource{out: map[string][]Instance{}} // registry up, service unknown
	l := NewLocator(src, map[string]string{Reward: "http://reward:8085"}, quiet())
	if got := l.URL(Reward); got != "http://reward:8085" {
		t.Fatalf("fallback expected, got %q", got)
	}
}

func TestLocatorKeepsLastKnownInstancesWhenRegistryFails(t *testing.T) {
	now := time.Unix(1000, 0)
	src := &fakeSource{out: map[string][]Instance{Account: {{Host: "10.0.0.1", Port: 8082}}}}
	l := NewLocator(src, map[string]string{Account: "http://account:8082"}, quiet(), WithTTL(time.Second), withClock(func() time.Time { return now }))

	if got := l.Target(Account); got != "10.0.0.1:8082" {
		t.Fatalf("got %q", got)
	}
	src.err = errors.New("consul is down")
	now = now.Add(10 * time.Second)
	if got := l.Target(Account); got != "10.0.0.1:8082" {
		t.Fatalf("stale answer expected while the registry is down, got %q", got)
	}
}

func TestLocatorFallsBackWhenRegistryFailsFromTheStart(t *testing.T) {
	src := &fakeSource{err: errors.New("consul is down")}
	l := NewLocator(src, map[string]string{Account: "http://account:8082"}, quiet())
	if got := l.URL(Account); got != "http://account:8082" {
		t.Fatalf("got %q", got)
	}
}
