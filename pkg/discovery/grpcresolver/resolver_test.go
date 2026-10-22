package grpcresolver

import (
	"context"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"

	"github.com/neennera/fishertimer/pkg/discovery"
)

type fakeCC struct {
	mu     sync.Mutex
	states []resolver.State
	errs   []error
}

func (c *fakeCC) UpdateState(s resolver.State) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.states = append(c.states, s)
	return nil
}
func (c *fakeCC) ReportError(err error)                                { c.mu.Lock(); c.errs = append(c.errs, err); c.mu.Unlock() }
func (c *fakeCC) NewAddress([]resolver.Address)                        {}
func (c *fakeCC) ParseServiceConfig(string) *serviceconfig.ParseResult { return nil }

func (c *fakeCC) last() (resolver.State, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.states) == 0 {
		return resolver.State{}, false
	}
	return c.states[len(c.states)-1], true
}

type src struct {
	mu  sync.Mutex
	out []discovery.Instance
}

func (s *src) set(i ...discovery.Instance) { s.mu.Lock(); s.out = i; s.mu.Unlock() }
func (s *src) Resolve(context.Context, string) ([]discovery.Instance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.out, nil
}

func TestDialIsAPassThroughInStaticMode(t *testing.T) {
	loc := discovery.NewLocator(nil, map[string]string{"study-timer-grpc": "study-timer:50051"})
	target, opts := Dial(loc, "study-timer-grpc", "study-timer:50051")
	if target != "study-timer:50051" || len(opts) != 0 {
		t.Fatalf("static mode must not change the connection: %q %v", target, opts)
	}
	if target, opts = Dial(nil, "x", "y:1"); target != "y:1" || len(opts) != 0 {
		t.Fatalf("nil locator: %q %v", target, opts)
	}
}

func TestDialUsesTheDiscoveryScheme(t *testing.T) {
	loc := discovery.NewLocator(&src{}, nil)
	target, opts := Dial(loc, "study-timer-grpc", "study-timer:50051")
	if target != "discovery:///study-timer-grpc" || len(opts) != 2 {
		t.Fatalf("%q %d options", target, len(opts))
	}
}

func TestResolverFollowsInstances(t *testing.T) {
	s := &src{}
	s.set(discovery.Instance{Host: "10.0.0.2", Port: 50051}, discovery.Instance{Host: "10.0.0.1", Port: 50051})
	loc := discovery.NewLocator(s, nil, discovery.WithTTL(time.Millisecond), discovery.WithLogf(func(string, ...any) {}))

	b := NewBuilder(loc)
	b.interval = 5 * time.Millisecond
	cc := &fakeCC{}
	r, err := b.Build(resolver.Target{}, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	st, ok := cc.last()
	if !ok || len(st.Addresses) != 2 || st.Addresses[0].Addr != "10.0.0.1:50051" {
		t.Fatalf("initial state %+v", st)
	}

	// A third instance appears and the resolver notices on its own.
	s.set(discovery.Instance{Host: "10.0.0.1", Port: 50051}, discovery.Instance{Host: "10.0.0.2", Port: 50051}, discovery.Instance{Host: "10.0.0.3", Port: 50051})
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if st, _ := cc.last(); len(st.Addresses) == 3 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("resolver did not pick up the new instance")
}

func TestResolverReportsWhenNothingIsKnown(t *testing.T) {
	loc := discovery.NewLocator(&src{}, nil, discovery.WithLogf(func(string, ...any) {}))
	cc := &fakeCC{}
	r, _ := NewBuilder(loc).Build(resolver.Target{}, cc, resolver.BuildOptions{})
	defer r.Close()
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if len(cc.errs) != 1 || len(cc.states) != 0 {
		t.Fatalf("errs=%v states=%v", cc.errs, cc.states)
	}
}
