# API Gateway Service

## Overview
Client API Gateway reverse proxy for Fisher Timer microservices. Dispatches requests from the Client frontend website to downstream microservices with CORS support.

## Port
Default port: `8000` (configurable via `API_GATEWAY_PORT` in `.env`)

## Routing
- `/api/account/*` -> Account Service (`8082`)
- `/api/auth/*` -> Account Service (`8082`)
- `/api/timer/*` -> Study Timer Service (`8084`)
- `/api/leaderboard/*` -> Leaderboard Service (`8086`)
- `/api/session/*` -> Study Session Service (`8083`)
- `/api/reward/*` -> Reward Service (`8085`)

## Identity forwarding
Every request passes through `internal/adapter/middleware.Verifier.Identity`
before routing. It reads the account service's session JWT from the
`ft_session` cookie (or an `Authorization: Bearer` header), and — if it's
valid — sets three trusted headers on the proxied request:

| Header | Source |
| :--- | :--- |
| `X-User-Id` | JWT `sub` |
| `X-User-Role` | JWT `role` |
| `X-Display-Name` | JWT `name`, URL-encoded (may contain non-ASCII characters) |

Any of these headers on the *incoming* request is always stripped first, so a
caller cannot spoof identity by setting them directly — downstream services
can trust them once the gateway is in front. A missing or invalid token is
**not** rejected here: the middleware just forwards no identity headers.
Whether a route requires a session is a decision left to each downstream
service. Uses `JWT_SECRET` / `JWT_ISSUER` from the shared root `.env` (same
values the account service signs with).

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
