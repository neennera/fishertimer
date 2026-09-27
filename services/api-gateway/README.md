# API Gateway Service

## Overview
Client API Gateway reverse proxy for Fisher Timer microservices. Dispatches requests from the Client frontend website to downstream microservices with CORS support.

## Port
Default port: `8000` (configurable via `API_GATEWAY_PORT` in `.env`)

## Routing
- `/api/account/*` -> Account Service (`8082`)
- `/api/auth/*` -> Account Service (`8082`)
- `/api/timer/*` -> Study Timer Service over **gRPC** (`TIMER_GRPC_TARGET`, default `localhost:50051`), see below
- `/api/leaderboard/*` -> Leaderboard Service (`8086`)
- `/api/session/*` -> Study Session Service (`8083`)
- `/api/reward/*` -> Reward Service (`8085`)

## Study Timer: REST in, gRPC out
`internal/adapter/handler/timer_handler.go` translates browser REST into
Study Timer gRPC calls. Every other route is a plain HTTP reverse proxy.

| REST | gRPC |
| --- | --- |
| `GET /api/timer/state?session_id=&user_id=` | `GetTimer` |
| `POST /api/timer/start` | `StartTimer` |
| `POST /api/timer/pause` | `PauseTimer` |
| `POST /api/timer/resume` | `ResumeTimer` |
| `POST /api/timer/reset` | `ResetTimer` |

POST bodies are JSON `{"session_id": "...", "user_id": "..."}`. Responses are
the timer state in snake_case (`status`, `phase`, `remaining_seconds`, ...).
gRPC codes become HTTP statuses: `InvalidArgument` 400, `FailedPrecondition`
409, `NotFound` 404, `Unavailable` 503, `DeadlineExceeded` 504, otherwise 502.
Server-side failures are logged and return a generic `{"error": ...}` message.

```bash
curl -X POST localhost:8000/api/timer/start -d '{"session_id":"demo","user_id":"demo"}'
```

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
