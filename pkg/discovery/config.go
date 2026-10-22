package discovery

import (
	"context"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// Modes.
const (
	ModeStatic = "static"
	ModeConsul = "consul"
)

// Config is read from the environment:
//
//	DISCOVERY_MODE         static (default) or consul
//	CONSUL_HTTP_ADDR       the Consul agent (default http://localhost:8500)
//	SERVICE_ADVERTISE_ADDR the host other services should use to reach this one
//	                       (default: this machine's first non-loopback IPv4)
type Config struct {
	Mode          string
	ConsulAddr    string
	AdvertiseHost string
}

// ConfigFromEnv reads the configuration. An unknown mode is treated as static.
func ConfigFromEnv() Config {
	c := Config{
		Mode:          strings.ToLower(strings.TrimSpace(os.Getenv("DISCOVERY_MODE"))),
		ConsulAddr:    strings.TrimSpace(os.Getenv("CONSUL_HTTP_ADDR")),
		AdvertiseHost: strings.TrimSpace(os.Getenv("SERVICE_ADVERTISE_ADDR")),
	}
	switch c.Mode {
	case "", ModeStatic:
		c.Mode = ModeStatic
	case ModeConsul:
	default:
		log.Printf("discovery: unknown DISCOVERY_MODE %q, using %q", c.Mode, ModeStatic)
		c.Mode = ModeStatic
	}
	if c.ConsulAddr == "" {
		c.ConsulAddr = "http://localhost:8500"
	}
	if c.AdvertiseHost == "" {
		c.AdvertiseHost = DetectHost()
	}
	return c
}

// Enabled reports whether a registry is in use.
func (c Config) Enabled() bool { return c.Mode == ModeConsul }

// Discovery bundles what a service needs: a Locator to find peers and a
// Registrar to announce itself. In static mode both are inert.
type Discovery struct {
	Config    Config
	Locator   *Locator
	registrar Registrar
}

// FromEnv builds a Discovery from the environment. fallback maps each peer's
// logical name to the address the service was configured with, e.g.
// {discovery.Account: cfg.AccountServiceURL, discovery.GRPC(discovery.StudyTimer): cfg.TimerGRPCTarget}.
func FromEnv(fallback map[string]string, opts ...Option) *Discovery {
	cfg := ConfigFromEnv()
	d := &Discovery{Config: cfg, registrar: noop{}}
	if cfg.Enabled() {
		c := NewConsul(cfg.ConsulAddr)
		d.registrar = c
		d.Locator = NewLocator(c, fallback, opts...)
		log.Printf("discovery: consul at %s, advertising as %s", cfg.ConsulAddr, cfg.AdvertiseHost)
	} else {
		d.Locator = NewLocator(nil, fallback, opts...)
	}
	return d
}

// New returns a Discovery over explicit parts (used in tests and by callers
// that don't read the environment).
func New(cfg Config, src Source, reg Registrar, fallback map[string]string, opts ...Option) *Discovery {
	if reg == nil {
		reg = noop{}
	}
	return &Discovery{Config: cfg, Locator: NewLocator(src, fallback, opts...), registrar: reg}
}

// HTTP is shorthand for the registration of this service's HTTP port, at the
// advertised host.
func (d *Discovery) HTTP(name string, port int) Registration {
	return HTTPService(name, d.Config.AdvertiseHost, port)
}

// GRPC is shorthand for the registration of this service's gRPC port.
func (d *Discovery) GRPC(name string, port int) Registration {
	return GRPCService(name, d.Config.AdvertiseHost, port)
}

// Announce registers regs and keeps them registered (re-registering if the
// registry forgets them, for instance after a restart) until the returned stop
// function is called, which also deregisters them. It does not block: when the
// registry is down, the service still starts and registers once it is back.
// In static mode it does nothing.
func (d *Discovery) Announce(regs ...Registration) (stop func()) {
	if !d.Config.Enabled() || len(regs) == 0 {
		return func() {}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		keepRegistered(ctx, d.registrar, regs, 15*time.Second, 3*time.Second, log.Printf)
	}()
	return func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}
}

// keepRegistered registers regs, repeats it every interval (retryEvery while
// failing) and deregisters them when ctx ends. It logs only when the state
// changes, not on every round.
func keepRegistered(ctx context.Context, r Registrar, regs []Registration, interval, retryEvery time.Duration, logf func(string, ...any)) {
	const (
		unknown = iota
		healthy
		failing
	)
	state := unknown
	var wait time.Duration
	for {
		select {
		case <-ctx.Done():
			dctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			for _, reg := range regs {
				if err := r.Deregister(dctx, reg.ID); err != nil {
					logf("discovery: deregister %s: %v", reg.ID, err)
				}
			}
			cancel()
			return
		case <-time.After(wait):
		}

		var firstErr error
		for _, reg := range regs {
			rctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			if err := r.Register(rctx, reg); err != nil && firstErr == nil {
				firstErr = err
			}
			cancel()
		}
		switch {
		case firstErr == nil && state != healthy:
			logf("discovery: registered %d entries", len(regs))
			state = healthy
		case firstErr != nil && state != failing:
			logf("discovery: register failed: %v (will keep retrying)", firstErr)
			state = failing
		}
		if firstErr == nil {
			wait = interval
		} else {
			wait = retryEvery
		}
	}
}

type noop struct{}

func (noop) Register(context.Context, Registration) error { return nil }
func (noop) Deregister(context.Context, string) error     { return nil }

// DetectHost returns the first non-loopback IPv4 address of this machine, the
// address other containers can reach it on, or its hostname, or "localhost".
func DetectHost() string {
	return detectHost(net.InterfaceAddrs, os.Hostname)
}

func detectHost(addrs func() ([]net.Addr, error), hostname func() (string, error)) string {
	if list, err := addrs(); err == nil {
		for _, a := range list {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	if h, err := hostname(); err == nil && h != "" {
		return h
	}
	return "localhost"
}
