# STUDY-SESSION Service Architecture

## Architecture Pattern: Hexagonal / Clean Architecture

This service is structured according to Ports & Adapters (Hexagonal) principles to ensure zero coupling between business logic and infrastructure.

```
study-session/
├── cmd/
│   └── main.go                    # Composition root (Dependency Injection, workers)
├── config/
│   └── config.go                  # Configuration and environment variables
├── database/schemas/
│   ├── 001_create_study_sessions_tables.sql
│   └── 002_phase2_rooms_and_outbox.sql   # status, counters, one-room index, event_outbox
├── internal/
│   ├── domain/                    # Core Enterprise Logic (No external deps)
│   │   ├── entity.go              # StudySession, Participant, OutboxEvent, errors, room rules
│   │   └── repository.go          # Ports: Repository, EventPublisher, TimerReader
│   ├── usecase/                   # Application Use Cases
│   │   ├── service.go             # Create / Join / Leave / End / reads / Heartbeat
│   │   ├── outbox_relay.go        # Publishes queued events to RabbitMQ, in order
│   │   ├── sweeper.go             # 24h auto-end, disconnect timeout, idle timeout
│   │   ├── service_test.go        # Room rules against the in-memory repository
│   │   └── workers_test.go        # Relay and sweeper
│   └── adapter/                   # Technical Adapters
│       ├── handler/
│       │   ├── grpc_handler.go    # Inbound: StudySessionService (Gateway, Admin)
│       │   └── http_handler.go    # Inbound: internal REST (Admin kick / close)
│       ├── repository/
│       │   ├── postgres_repo.go   # Outbound: session_db, one transaction per command
│       │   └── memory_repo.go     # Outbound: same rules in memory (tests)
│       ├── amqp/
│       │   ├── publisher.go       # Outbound: RabbitMQ with publisher confirms, lazy reconnect
│       │   └── topology.go        # Declares exchange + Timer queue + DLX/DLQ (idempotent)
│       └── client/
│           └── timer_client.go    # Outbound: Study Timer GetRoomTimers (idle detection)
```

## Room rules (where they are enforced)

| Rule | Where |
| --- | --- |
| Name 1-60 chars, limit 1-5 (UC-01 E-1) | `usecase.CreateSession` |
| Last seat taken by exactly one of two racing users (UC-02 E-2) | one conditional `UPDATE ... WHERE participant_count < max_participants` |
| One active room per user (UC-01 E-2, UC-02 E-3) | partial unique index `session_participants(user_id) WHERE left_at IS NULL` |
| Re-joining your own room is a no-op (UC-02 E-5) | `JoinSession` checks the open participation first |
| Leaving twice is a no-op (UC-04 E-4) | `LeaveSession` updates only `left_at IS NULL` rows |
| Last one out ends the room (UC-03 S-2) | same transaction as the leave, with the room row locked (`FOR UPDATE`) |

## Events (RabbitMQ)

Contract: `pkg/events`. Exchange `fisher.session` (topic), consumed by Study Timer from `timer.session-events`.

| Change | Events |
| --- | --- |
| CreateSession / JoinSession | `session.participant.joined` |
| LeaveSession (reason LEFT, KICKED, DISCONNECT_TIMEOUT, IDLE_TIMEOUT) | `session.participant.left`, plus `session.ended` (EMPTY) if it emptied the room |
| EndSession (ADMIN_CLOSED, TIMEOUT_24H) | `session.ended` only: Timer finalizes every timer in the room |

**Transactional outbox.** A command writes its events into `event_outbox` in the same transaction as the room change. `OutboxRelay` publishes pending rows oldest first, waits for the broker's confirm, then stamps `published_at`. So:
- an event is never lost while RabbitMQ is down (rows wait, `attempts` / `last_error` record the retries);
- an event is never sent for a change that rolled back;
- delivery is at-least-once (a crash between confirm and stamp re-sends), which the Timer consumer absorbs by deduplicating on `event_id`.

The relay stops at the first failure so events keep their order. It assumes one Study Session instance; running several would need `FOR UPDATE SKIP LOCKED` when claiming rows.

## Presence and automatic removal

The room page sends `Heartbeat` every 10s. `Sweeper` (every 5s) removes participants through the normal `LeaveSession`, so the same events go out:
- no heartbeat for 60s: `DISCONNECT_TIMEOUT` (UC-03 E-3);
- no running work cycle for 10 minutes, read from Study Timer (`usecase.IsIdle`, UC-05 E-8): `IDLE_TIMEOUT`. Skipped while Study Timer is unreachable, so an outage never kicks anyone;
- rooms older than 24 hours are ended with `TIMEOUT_24H` (UC-03 E-6).

## Rules for Extending This Service
1. **Never import `adapter` from `domain` or `usecase`**.
2. Define repository interfaces in `internal/domain/repository.go`.
3. Implement database adapters in `internal/adapter/repository/` (e.g. Supabase, MongoDB, Redis).
4. Any new membership change must enqueue its event in the same transaction.
5. Run unit tests with `pnpm test`.
