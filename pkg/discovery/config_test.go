package discovery

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("DISCOVERY_MODE", "")
	t.Setenv("CONSUL_HTTP_ADDR", "")
	t.Setenv("SERVICE_ADVERTISE_ADDR", "")
	c := ConfigFromEnv()
	if c.Enabled() || c.Mode != ModeStatic || c.ConsulAddr != "http://localhost:8500" || c.AdvertiseHost == "" {
		t.Fatalf("defaults: %+v", c)
	}

	t.Setenv("DISCOVERY_MODE", " Consul ")
	t.Setenv("CONSUL_HTTP_ADDR", "http://consul:8500")
	t.Setenv("SERVICE_ADVERTISE_ADDR", "account")
	c = ConfigFromEnv()
	if !c.Enabled() || c.ConsulAddr != "http://consul:8500" || c.AdvertiseHost != "account" {
		t.Fatalf("consul: %+v", c)
	}

	t.Setenv("DISCOVERY_MODE", "etcd")
	if ConfigFromEnv().Enabled() {
		t.Fatal("an unknown mode must fall back to static")
	}
}

func TestFromEnvStaticIsInert(t *testing.T) {
	t.Setenv("DISCOVERY_MODE", "static")
	d := FromEnv(map[string]string{Account: "http://account:8082"})
	if d.Locator.Enabled() {
		t.Fatal("static mode has no registry")
	}
	if got := d.Locator.URL(Account); got != "http://account:8082" {
		t.Fatalf("got %q", got)
	}
	stop := d.Announce(d.HTTP(Account, 8082)) // must not touch the network
	stop()
}

func TestDetectHost(t *testing.T) {
	ip := func(s string) net.Addr { return &net.IPNet{IP: net.ParseIP(s), Mask: net.CIDRMask(24, 32)} }
	list := func(a ...net.Addr) func() ([]net.Addr, error) { return func() ([]net.Addr, error) { return a, nil } }
	host := func(h string) func() (string, error) { return func() (string, error) { return h, nil } }

	if got := detectHost(list(ip("127.0.0.1"), ip("172.18.0.4")), host("abc")); got != "172.18.0.4" {
		t.Fatalf("skips loopback: %q", got)
	}
	if got := detectHost(list(ip("127.0.0.1")), host("abc")); got != "abc" {
		t.Fatalf("falls back to the hostname: %q", got)
	}
	failing := func() ([]net.Addr, error) { return nil, errors.New("no interfaces") }
	if got := detectHost(failing, func() (string, error) { return "", errors.New("no host") }); got != "localhost" {
		t.Fatalf("last resort: %q", got)
	}
}

type fakeRegistrar struct {
	mu         sync.Mutex
	registered map[string]int
	removed    []string
	fail       bool
}

func (f *fakeRegistrar) Register(_ context.Context, r Registration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return errors.New("registry down")
	}
	if f.registered == nil {
		f.registered = map[string]int{}
	}
	f.registered[r.ID]++
	return nil
}

func (f *fakeRegistrar) Deregister(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removed = append(f.removed, id)
	return nil
}

func (f *fakeRegistrar) count(id string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.registered[id]
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestKeepRegisteredRegistersRefreshesAndDeregisters(t *testing.T) {
	f := &fakeRegistrar{}
	reg := HTTPService(Account, "10.0.0.5", 8082)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		keepRegistered(ctx, f, []Registration{reg}, 10*time.Millisecond, 10*time.Millisecond, func(string, ...any) {})
	}()

	eventually(t, "re-registration", func() bool { return f.count(reg.ID) >= 3 })
	cancel()
	<-done
	if len(f.removed) != 1 || f.removed[0] != reg.ID {
		t.Fatalf("deregistered %v", f.removed)
	}
}

func TestKeepRegisteredRetriesUntilTheRegistryIsBack(t *testing.T) {
	f := &fakeRegistrar{fail: true}
	reg := HTTPService(Account, "10.0.0.5", 8082)
	var logs int
	var mu sync.Mutex
	logf := func(string, ...any) { mu.Lock(); logs++; mu.Unlock() }

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		keepRegistered(ctx, f, []Registration{reg}, 10*time.Millisecond, 5*time.Millisecond, logf)
	}()

	time.Sleep(60 * time.Millisecond) // many failed rounds
	f.mu.Lock()
	f.fail = false
	f.mu.Unlock()
	eventually(t, "registration after recovery", func() bool { return f.count(reg.ID) >= 1 })
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if logs != 2 { // one "failed", one "registered": state changes only
		t.Fatalf("expected 2 log lines (failure, recovery), got %d", logs)
	}
}
