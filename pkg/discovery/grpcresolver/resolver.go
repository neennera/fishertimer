// Package grpcresolver plugs the discovery Locator into gRPC, so a client
// connection follows the healthy instances of a service as they come and go
// and spreads calls over them (round-robin).
//
//	target, opts := grpcresolver.Dial(loc, discovery.GRPC(discovery.StudyTimer), cfg.TimerGRPCTarget)
//	conn, err := grpc.NewClient(target, append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
//
// In static mode Dial returns the configured target and no options, so the
// connection behaves exactly as it did before discovery existed.
package grpcresolver

import (
	"slices"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"

	"github.com/neennera/fishertimer/pkg/discovery"
)

// Scheme is the gRPC target scheme handled by this package:
// "discovery:///study-timer-grpc".
const Scheme = "discovery"

// ServiceConfig spreads calls over every address the resolver reports. gRPC's
// default (pick_first) would stick to the first one.
const ServiceConfig = `{"loadBalancingConfig":[{"round_robin":{}}]}`

// pollEvery is how often the resolver asks the Locator again; the Locator
// caches registry answers for a few seconds, so this is cheap.
const pollEvery = 3 * time.Second

// Dial returns the target and dial options for a client of service. fallback
// is the configured address ("study-timer:50051"); it is used as is when
// discovery is off.
func Dial(loc *discovery.Locator, service, fallback string) (string, []grpc.DialOption) {
	if loc == nil || !loc.Enabled() {
		return fallback, nil
	}
	return Scheme + ":///" + service, []grpc.DialOption{
		grpc.WithResolvers(NewBuilder(loc)),
		grpc.WithDefaultServiceConfig(ServiceConfig),
	}
}

// Builder is a gRPC resolver builder backed by a Locator.
type Builder struct {
	loc      *discovery.Locator
	interval time.Duration
}

// NewBuilder returns a resolver builder for the discovery scheme.
func NewBuilder(loc *discovery.Locator) *Builder {
	return &Builder{loc: loc, interval: pollEvery}
}

// Scheme implements resolver.Builder.
func (b *Builder) Scheme() string { return Scheme }

// Build implements resolver.Builder. The target's endpoint is the service
// name: "discovery:///study-timer-grpc" watches "study-timer-grpc".
func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	w := &watcher{
		loc:     b.loc,
		service: target.Endpoint(),
		cc:      cc,
		stop:    make(chan struct{}),
		nudge:   make(chan struct{}, 1),
	}
	w.update()
	go w.loop(b.interval)
	return w, nil
}

type watcher struct {
	loc     *discovery.Locator
	service string
	cc      resolver.ClientConn

	stop     chan struct{}
	stopOnce sync.Once
	nudge    chan struct{}

	last   []string
	synced bool // an update (even "nothing known") has been sent
}

func (w *watcher) loop(every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-t.C:
		case <-w.nudge:
		}
		w.update()
	}
}

// update pushes the current addresses to gRPC when they changed.
func (w *watcher) update() {
	targets := slices.Clone(w.loc.Targets(w.service))
	slices.Sort(targets)
	if w.synced && slices.Equal(targets, w.last) {
		return
	}
	if len(targets) == 0 {
		w.cc.ReportError(errNoInstances(w.service))
		w.last, w.synced = nil, true
		return
	}
	addrs := make([]resolver.Address, len(targets))
	for i, t := range targets {
		addrs[i] = resolver.Address{Addr: t}
	}
	if err := w.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		return // gRPC rejected the state; try again on the next round
	}
	w.last, w.synced = targets, true
}

// ResolveNow implements resolver.Resolver: gRPC asks for a refresh after a
// connection error.
func (w *watcher) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case w.nudge <- struct{}{}:
	default:
	}
}

// Close implements resolver.Resolver.
func (w *watcher) Close() { w.stopOnce.Do(func() { close(w.stop) }) }

type errNoInstances string

func (e errNoInstances) Error() string { return "discovery: no instance of " + string(e) }
