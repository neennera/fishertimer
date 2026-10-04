# STUDY-TIMER Service Architecture

## Architecture Pattern: Hexagonal / Clean Architecture

This service is structured according to Ports & Adapters (Hexagonal) principles to ensure zero coupling between business logic and infrastructure.

```
study-timer/
├── cmd/
│   └── main.go                    # Composition root: DI, gRPC/HTTP servers, consumer, sweeper
├── config/
│   └── config.go                  # Environment, including duration limits
├── database/schemas/              # 001-002 Phase 1 (deprecated tables), 003 timers/cycles,
│                                  # 004 per-participant fix, 005 sweeper indexes
├── internal/
│   ├── domain/                    # Core Enterprise Logic (No external deps)
│   │   ├── entity.go              # Timer, Cycle, Limits, State machine, View
│   │   ├── entity_test.go         # Rule tests with explicit timestamps
│   │   ├── collaborator.go        # Ports: RewardClient (AwardRequest), SessionClient
│   │   └── repository.go          # Port: timers, cycles, settings, events
│   ├── usecase/
│   │   ├── service.go             # Actions, lazy settling, reward delivery, Sweep
│   │   └── service_test.go        # State machine, rewards, retries, events
│   └── adapter/
│       ├── handler/
│       │   ├── grpc_handler.go    # gRPC (port 50051), used by the API Gateway and Study Session
│       │   ├── grpc_handler_test.go
│       │   └── http_handler.go    # HTTP / REST (port 8084), statistics for Account
│       ├── amqp/
│       │   ├── topology.go        # Exchange, queue, DLX / DLQ (idempotent)
│       │   └── consumer.go        # joined / left / ended, manual ack, reconnect with backoff
│       ├── client/
│       │   ├── reward_client.go   # Reward AwardReward (HTTP)
│       │   └── session_client.go  # Study Session GetParticipants (gRPC)
│       └── repository/
│           ├── postgres_repo.go   # timers, cycles, timer_settings, processed_events
│           └── memory_repo.go     # Same rules in memory (tests)
```

## Key decisions
- **Time comes from the server.** A cycle stores `started_at`, `paused_at`, `paused_total_sec` and `duration_sec`; remaining time is computed, never counted down in storage.
- **Cycles are rows.** Every work and rest period is a `cycles` row with its own `cycle_id`, so a reward is tied to exactly one cycle (Reward deduplicates on it) and undelivered rewards can be found and retried (`reward_status`).
- **Optimistic, conditional updates.** A cycle is updated only if it is still in the status it was read in (`UpdateCycle(..., from...)`), so two tabs or the sweeper racing a click cannot both win; the loser just returns the current state.
- **The server completes cycles.** Clients may call `CompleteCycle` at zero, but the sweeper and every read complete due cycles anyway, so a closed laptop still earns its reward.
- **Events are idempotent.** Each event inserts its `event_id` into `processed_events` in the same transaction as its effect.

## Rules for Extending This Service
1. **Never import `adapter` from `domain` or `usecase`**.
2. Define repository interfaces in `internal/domain/repository.go`.
3. Implement database adapters in `internal/adapter/repository/` (e.g. Supabase, MongoDB, Redis).
4. Run unit tests with `pnpm test`.
