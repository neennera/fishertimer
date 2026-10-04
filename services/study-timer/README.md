# Study-timer Service

## Overview
Independent user timers (UC-05): one timer per participant per room, work cycles that earn rewards and rest periods that don't. Remaining time is derived only from server timestamps (`started_at`, `paused_at`, `paused_total_sec`), so a reload or disconnect never alters it (E-5).

```
READY -> WORK_RUNNING <-> WORK_PAUSED -> (time up) -> READY_FOR_REST
READY_FOR_REST -> REST_RUNNING <-> REST_PAUSED -> (time up / skip) -> READY
Stop: active cycle DISCARDED, back to READY.    Left / room ended: FINALIZED.
```

Timers live in PostgreSQL `timer_db` (`timers`, `cycles`, `timer_settings`, `processed_events`); the service refuses to start without it (`pnpm db:up`). Schemas in `database/schemas/` run only when the volume is first created, so after pulling a new one run `pnpm db:reset`. Session and user ids are UUIDs.

## How a timer lives
- **Opened** by Study Session's `session.participant.joined` event (RabbitMQ). A Start that beats the event opens it too.
- **Run** by the user over gRPC (through the API Gateway): pick a length (`MIN_WORK_MINUTES`..`MAX_WORK_MINUTES`, E-1), start, pause / resume, reset, stop. One active cycle at a time (unique index, E-4).
- **Completed by the server.** A background sweeper (every `TIMER_SWEEP_INTERVAL_SECONDS`) completes cycles whose time is up even when no client is watching, discards cycles paused longer than `MAX_PAUSE_MINUTES` (E-3) and retries undelivered rewards. Reads settle the same rules immediately.
- **Rewarded once.** A completed work cycle is `reward_status = PENDING`; Timer asks Study Session for the participant count (`GetParticipants`) and calls Reward `POST /api/v1/reward/award` with `user_id, session_id, cycle_id, work_duration, participant_count`. Success marks it `SENT`; failure keeps it `PENDING` for the sweeper (E-7). Reward deduplicates by `cycle_id`.
- **Finalized** by `session.participant.left` / `session.ended`: a cycle whose time had already run out is completed (so its reward is kept), anything still running is discarded, and the timer refuses further actions (E-2). Rejoining reopens it with fresh per-stay counters.

## Ports
- gRPC: `50051` (`TIMER_GRPC_PORT`): `fishertimer.studytimer.v1.StudyTimerService`, see [`proto/studytimer/v1/timer.proto`](../../proto/studytimer/v1/timer.proto). This is what the API Gateway calls.
- HTTP: `8084` (`PORT` / `TIMER_SERVICE_PORT`): the same operations over REST, plus `/health`.

## gRPC operations

| RPC | What it does |
| --- | --- |
| `StartTimer` | Starts a work cycle (`phase` WORK, default) or, after a completed work cycle, a rest period. `duration_minutes` 0 = saved default. Duplicate starts are ignored. |
| `GetTimer` | State, `remaining_seconds`, per-stay `current_cycle` / `focus_seconds`, last completed cycle, limits. With an empty `session_id`: only the user's settings. |
| `GetRoomTimers` | Every open timer in a room (others' states, idle detection in Study Session). |
| `PauseTimer` / `ResumeTimer` | Freeze / continue the active cycle. |
| `StopTimer` | Discard the active cycle, no reward (S-4). |
| `ResetTimer` | Restart the active cycle from its full length. |
| `CompleteCycle` | Complete the active cycle if its time is really up (server-checked, idempotent). |
| `SkipRest` | Skip a running or not-yet-started rest period. |
| `UpdateTimerSetting` | Save default work / rest lengths. |
| `TimerStatistics` | Sessions joined, cycles completed, focus minutes (dashboard). |

Errors: invalid input -> `InvalidArgument`; not allowed in this state or time left -> `FailedPrecondition`; timer finalized -> `PermissionDenied`.

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{"session_id":"00000000-0000-0000-0000-000000000001","user_id":"00000000-0000-0000-0000-000000000002","duration_minutes":1}' \
  localhost:50051 fishertimer.studytimer.v1.StudyTimerService/StartTimer
```

## HTTP API
`/api/v1/study-timer/{state,start,pause,resume,stop,reset,complete,skip-rest,setting}` mirror the RPCs. `GET /api/v1/study-timer/statistics?user_id=` returns `{user_id, sessions_joined, cycles_completed, total_focus_minutes, last_active, daily_focus_minutes}` (30 days, oldest first) for the Account service.

## Configuration

| Variable | Default | Meaning |
| --- | --- | --- |
| `TIMER_DATABASE_URL` | `postgres://...@localhost:5434/timer_db` | timer_db |
| `RABBITMQ_URL` | `amqp://admin:adminpassword@localhost:5672/` | broker (consumer reconnects with backoff) |
| `REWARD_SERVICE_URL` | `http://localhost:8085` | AwardReward |
| `SESSION_GRPC_TARGET` | `localhost:50052` | participant count for rewards |
| `MIN_WORK_MINUTES` / `MAX_WORK_MINUTES` | `1` / `120` | allowed focus lengths |
| `MIN_REST_MINUTES` / `MAX_REST_MINUTES` | `1` / `60` | allowed break lengths |
| `MAX_PAUSE_MINUTES` | `15` | longer pauses discard the cycle |
| `TIMER_SWEEP_INTERVAL_SECONDS` | `2` | sweeper period |

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
