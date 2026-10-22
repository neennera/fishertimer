# pkg/discovery

Lets a service find its peers, and announce itself, at run time. Standard library
only (the gRPC integration in `grpcresolver/` also needs `google.golang.org/grpc`).

| `DISCOVERY_MODE` | Behaviour |
| :--- | :--- |
| `static` (default) | Peers are the addresses already in the environment (`ACCOUNT_SERVICE_URL`, `TIMER_GRPC_TARGET`, ...). Nothing is registered or looked up. Used by `pnpm dev` and Kubernetes. |
| `consul` | Services register with a Consul agent (with a health check); callers ask Consul for healthy instances and fall back to the static value when Consul has nothing. |

Other variables: `CONSUL_HTTP_ADDR` (default `http://localhost:8500`),
`SERVICE_ADVERTISE_ADDR` (host others use to reach this service; default: the
machine's first non-loopback IPv4).

## Use

```go
d := discovery.FromEnv(map[string]string{
    discovery.Account:                    cfg.AccountServiceURL,  // "http://account:8082"
    discovery.GRPC(discovery.StudyTimer): cfg.TimerGRPCTarget,    // "study-timer:50051"
})

// Announce this service (no-op in static mode); stop() deregisters.
stop := d.Announce(d.HTTP(discovery.Reward, cfg.Port))
defer stop()

// Find a peer, per request. Round-robins over healthy instances.
base := d.Locator.URL(discovery.Account)   // "http://172.18.0.5:8082"

// gRPC: the connection follows the healthy instances.
target, opts := grpcresolver.Dial(d.Locator, discovery.GRPC(discovery.StudyTimer), cfg.TimerGRPCTarget)
conn, err := grpc.NewClient(target, append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
```

HTTP and gRPC ports are separate logical services: `study-timer` and
`study-timer-grpc`.

## Failure behaviour

- Registry down at startup: the service still starts and registers when the registry is back.
- Registry down later: callers keep the last known instances.
- Service not registered (e.g. run on the host): the static address is used.
- A crashed instance fails its health check and drops out of `Resolve` within seconds.

Tests: `go test ./...` (from this directory, or `cd pkg/discovery`).
