# Study-timer Service

## Overview
Independent user timer execution, focus cycles, and rest intervals (UC-05).
Each participant has one timer per room. Remaining time is derived from
server timestamps, so pausing freezes it and a client reload never alters it.

Timers are stored in PostgreSQL `timer_db` (`repository.NewPostgres`); the
service refuses to start without it (`pnpm db:up`). Schemas live in
`database/schemas/` and run only when the database volume is first created,
so after pulling a new schema file run `pnpm db:reset` (wipes local data).
`002_add_timer_progress.sql` stores each timer's `phase`, `running_since` and
`elapsed_ms`, which remaining time is derived from. Session and user ids are
UUIDs.

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
grpcurl -plaintext -d '{"session_id":"00000000-0000-0000-0000-000000000001","user_id":"00000000-0000-0000-0000-000000000002"}' \
  localhost:50051 fishertimer.studytimer.v1.StudyTimerService/StartTimer
```

Swap `StartTimer` for `GetTimer`, `PauseTimer`, `ResumeTimer` or `ResetTimer`.

## HTTP API

| Method | Route | Purpose |
| :--- | :--- | :--- |
| `GET` | `/api/v1/study-timer/statistics?user_id=<id>` | Returns `{user_id, sessions_joined, cycles_completed, total_focus_minutes, last_active, daily_focus_minutes}` for a user, where `daily_focus_minutes` is a 30-entry `[{date, focus_minutes}]` array covering today and the previous 29 days (oldest first). Called by the Account service's `GET /api/v1/account/statistics?id=<user_id>` collaborator. |

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
