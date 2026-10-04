# Study-session Service

## Overview
Study room lifecycle (UC-01 Create, UC-02 Join, UC-03 Leave, UC-04 Admin kick / close), participant rosters and room capacity. Every membership change is published to RabbitMQ so Study Timer can open and finalize timers.

## Ports
- HTTP `8083` (`SESSION_SERVICE_PORT`): internal REST used by the Admin service and for debugging
- gRPC `50052` (`SESSION_GRPC_PORT`): `StudySessionService` (`proto/studysession/v1/session.proto`), used by the API Gateway

## Requires
- `session_db` (PostgreSQL, `pnpm db:up`). The service refuses to start without it.
- RabbitMQ (`RABBITMQ_URL`). Optional at startup: events wait in the outbox until it is reachable.
- Study Timer gRPC (`TIMER_GRPC_TARGET`). Optional: only used to detect idle participants.

## Configuration

| Variable | Default | Meaning |
| --- | --- | --- |
| `SESSION_DATABASE_URL` | `postgres://...@localhost:5433/session_db` | session_db |
| `RABBITMQ_URL` | `amqp://admin:adminpassword@localhost:5672/` | broker |
| `TIMER_GRPC_TARGET` | `localhost:50051` | Study Timer, for idle detection |
| `SESSION_MAX_AGE_HOURS` | `24` | auto-end rooms after this (UC-03 E-6), 0 = off |
| `SESSION_DISCONNECT_TIMEOUT_SECONDS` | `60` | auto-leave without heartbeats (UC-03 E-3), 0 = off |
| `IDLE_TIMEOUT_MINUTES` | `10` | auto-leave with no running work cycle (UC-03 E-7), 0 = off |
| `SESSION_SWEEP_INTERVAL_SECONDS` | `5` | how often the rules above run |
| `SESSION_OUTBOX_INTERVAL_MS` | `500` | how often queued events are published |

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)

## Try it

```bash
pnpm db:up                       # session_db, timer_db, rabbitmq, ...
grpcurl -plaintext localhost:50052 list fishertimer.studysession.v1.StudySessionService
curl localhost:8083/api/v1/study-session/active
```

RabbitMQ UI: http://localhost:15672 (admin / adminpassword). Each join, leave and room end shows up on exchange `fisher.session` and is consumed from `timer.session-events`.
