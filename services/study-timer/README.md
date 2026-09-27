# Study-timer Service

## Overview
Independent user timer execution, focus cycles, and rest intervals (UC-05).
Each participant has one timer per room. Remaining time is derived from
server timestamps, so pausing freezes it and a client reload never alters it.

Storage is in memory for now (`repository.NewInMemory`); timers reset when the
service restarts.

## Ports
- gRPC: `50051` (`TIMER_GRPC_PORT`) — `fishertimer.studytimer.v1.StudyTimerService`, see [`proto/studytimer/v1/timer.proto`](../../proto/studytimer/v1/timer.proto). This is what the API Gateway calls.
- HTTP: `8084` (`PORT` / `TIMER_SERVICE_PORT`) — same operations over REST, plus `/health`.

Both transports are driving adapters over the same usecase
(`internal/adapter/handler/grpc_handler.go` and `http_handler.go`).

## gRPC operations

| RPC | CRUD | What it does |
| --- | --- | --- |
| `StartTimer` | Create | Starts a work phase. Ignored while a phase is already active (UC-05 E-4). |
| `GetTimer` | Read | Current state and `remaining_seconds`. An unknown timer reads as stopped. |
| `PauseTimer` / `ResumeTimer` | Update | Freezes / continues the remaining time. |
| `ResetTimer` | Delete | Returns the timer to its initial stopped state (keeps work/rest lengths). |
| `StopTimer`, `CompleteCycle`, `SkipRest`, `UpdateTimerSetting`, `TimerStatistics` | | Cycle flow and dashboard stats. |

Errors map to gRPC codes: invalid input -> `InvalidArgument`, action not
allowed in the current state -> `FailedPrecondition`.

### Try it with grpcurl
Server reflection is on outside production, so no `.proto` file is needed:

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{"session_id":"demo","user_id":"demo"}' \
  localhost:50051 fishertimer.studytimer.v1.StudyTimerService/StartTimer
```

Swap `StartTimer` for `GetTimer`, `PauseTimer`, `ResumeTimer` or `ResetTimer`.

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
