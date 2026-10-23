// Package discovery lets a service find its peers, and announce itself, at run
// time instead of from addresses fixed in its environment.
//
// Two modes share one API (DISCOVERY_MODE):
//
//   - static (default): the address of every peer is the value already given
//     by the service's environment (ACCOUNT_SERVICE_URL, TIMER_GRPC_TARGET...).
//     Nothing is registered and nothing is looked up. This is how `pnpm dev`
//     and Kubernetes (cluster DNS) behave, and it stays the default.
//   - consul: services register with a Consul agent (with a health check) and
//     callers ask Consul for the healthy instances. Whatever Consul cannot
//     answer falls back to the static value, so a service that is not
//     registered (for example one run on the host in hybrid mode) is still
//     reachable and a Consul outage never takes the system down.
//
// The package uses only the standard library, apart from grpcresolver, the
// gRPC integration, which lives in its own package.
package discovery

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Logical service names. HTTP and gRPC are registered under separate names
// because one service (study-timer, study-session) serves both on different
// ports.
const (
	Account      = "account"
	StudySession = "study-session"
	StudyTimer   = "study-timer"
	Reward       = "reward"
	Leaderboard  = "leaderboard"
	Admin        = "admin"
	APIGateway   = "api-gateway"

	// GRPCSuffix turns an HTTP service name into its gRPC counterpart.
	GRPCSuffix = "-grpc"
)

// GRPC returns the logical name under which a service's gRPC port is
// registered, e.g. GRPC(StudyTimer) == "study-timer-grpc".
func GRPC(service string) string { return service + GRPCSuffix }

// Instance is one running copy of a service.
type Instance struct {
	ID   string
	Name string
	Host string
	Port int
}

// HostPort is the instance's address as "host:port".
func (i Instance) HostPort() string {
	return net.JoinHostPort(i.Host, strconv.Itoa(i.Port))
}

// Source answers "which instances of this service are healthy right now?".
type Source interface {
	Resolve(ctx context.Context, service string) ([]Instance, error)
}

// Check says how the registry verifies an instance is alive. Set HTTP (a URL
// that must answer 2xx) or TCP (host:port that must accept a connection).
type Check struct {
	HTTP string
	TCP  string

	Interval        time.Duration // how often to check (default 5s)
	Timeout         time.Duration // per check (default 2s)
	DeregisterAfter time.Duration // drop an instance critical this long (default 1m)
}

func (c Check) withDefaults() Check {
	if c.Interval == 0 {
		c.Interval = 5 * time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = 2 * time.Second
	}
	if c.DeregisterAfter == 0 {
		c.DeregisterAfter = time.Minute
	}
	return c
}

// Registration is what a service tells the registry about itself.
type Registration struct {
	ID    string
	Name  string
	Host  string
	Port  int
	Check Check
}

// Registrar adds and removes instances.
type Registrar interface {
	Register(ctx context.Context, r Registration) error
	Deregister(ctx context.Context, id string) error
}

// HTTPService builds the registration of an HTTP service that answers GET
// /health on its own port.
func HTTPService(name, host string, port int) Registration {
	return Registration{
		ID:   instanceID(name, host, port),
		Name: name,
		Host: host,
		Port: port,
		Check: Check{
			HTTP: fmt.Sprintf("http://%s/health", net.JoinHostPort(host, strconv.Itoa(port))),
		},
	}
}

// GRPCService builds the registration of a gRPC port, health-checked by
// opening a TCP connection to it. It is registered as GRPC(name).
func GRPCService(name, host string, port int) Registration {
	n := GRPC(name)
	return Registration{
		ID:    instanceID(n, host, port),
		Name:  n,
		Host:  host,
		Port:  port,
		Check: Check{TCP: net.JoinHostPort(host, strconv.Itoa(port))},
	}
}

func instanceID(name, host string, port int) string {
	return fmt.Sprintf("%s-%s-%d", name, host, port)
}
