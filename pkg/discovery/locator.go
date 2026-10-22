package discovery

import (
	"context"
	"log"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Locator turns a logical service name into an address a caller can use.
//
// With no Source (static mode) it just returns the fallback the service was
// configured with. With a Source it asks the registry, caches the answer for a
// short time, spreads calls over the healthy instances (round-robin) and falls
// back to the configured address when the registry has nothing to offer. The
// fallback is a URL ("http://account:8082") or a "host:port" target.
type Locator struct {
	src      Source
	fallback map[string]string
	ttl      time.Duration
	now      func() time.Time
	logf     func(format string, args ...any)

	next atomic.Uint64

	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	instances []Instance
	fetched   time.Time
}

// Option customises a Locator.
type Option func(*Locator)

// WithTTL sets how long a registry answer is reused (default 3s).
func WithTTL(d time.Duration) Option { return func(l *Locator) { l.ttl = d } }

// WithLogf sets where the Locator reports registry problems (default log.Printf).
func WithLogf(f func(string, ...any)) Option { return func(l *Locator) { l.logf = f } }

func withClock(now func() time.Time) Option { return func(l *Locator) { l.now = now } }

// NewLocator returns a Locator. src may be nil (static mode); fallback maps a
// service name to its configured address.
func NewLocator(src Source, fallback map[string]string, opts ...Option) *Locator {
	l := &Locator{
		src:      src,
		fallback: fallback,
		ttl:      3 * time.Second,
		now:      time.Now,
		logf:     log.Printf,
		cache:    make(map[string]cacheEntry),
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Enabled reports whether a registry is in use (false in static mode).
func (l *Locator) Enabled() bool { return l.src != nil }

// Targets returns the "host:port" of every healthy instance of service, or the
// configured fallback (a single entry) when the registry has none. It is empty
// only when nothing is configured either.
func (l *Locator) Targets(service string) []string {
	if inst := l.instances(service); len(inst) > 0 {
		out := make([]string, len(inst))
		for i, in := range inst {
			out[i] = in.HostPort()
		}
		return out
	}
	if fb := l.fallbackHostPort(service); fb != "" {
		return []string{fb}
	}
	return nil
}

// Target returns one "host:port" for service, rotating over healthy instances.
func (l *Locator) Target(service string) string {
	t := l.Targets(service)
	if len(t) == 0 {
		return ""
	}
	return t[int((l.next.Add(1)-1)%uint64(len(t)))]
}

// URL returns a base URL for service, e.g. "http://10.0.0.5:8082". The scheme
// of the configured fallback is kept (http when none).
func (l *Locator) URL(service string) string {
	fb := l.fallback[service]
	if inst := l.instances(service); len(inst) > 0 {
		pick := inst[int((l.next.Add(1)-1)%uint64(len(inst)))]
		return schemeOf(fb) + "://" + pick.HostPort()
	}
	return fb
}

func (l *Locator) instances(service string) []Instance {
	if l.src == nil {
		return nil
	}
	l.mu.Lock()
	e, ok := l.cache[service]
	l.mu.Unlock()
	if ok && l.now().Sub(e.fetched) < l.ttl {
		return e.instances
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	inst, err := l.src.Resolve(ctx, service)
	if err != nil {
		if ok {
			l.logf("discovery: %s: registry unavailable (%v), using last known instances", service, err)
			return e.instances
		}
		l.logf("discovery: %s: registry unavailable (%v), using configured address", service, err)
		return nil
	}
	l.mu.Lock()
	l.cache[service] = cacheEntry{instances: inst, fetched: l.now()}
	l.mu.Unlock()
	return inst
}

func (l *Locator) fallbackHostPort(service string) string {
	fb := l.fallback[service]
	if fb == "" {
		return ""
	}
	if strings.Contains(fb, "://") {
		if u, err := url.Parse(fb); err == nil {
			return u.Host
		}
	}
	return fb
}

func schemeOf(addr string) string {
	if i := strings.Index(addr, "://"); i > 0 {
		return addr[:i]
	}
	return "http"
}
