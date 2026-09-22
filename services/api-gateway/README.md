# API Gateway Service

## Overview
Client API Gateway reverse proxy for Fisher Timer microservices. Dispatches requests from the Client frontend website to downstream microservices with CORS support.

## Port
Default port: `8080`

## Routing
- `/api/auth/*` -> Auth Service (`8081`)
- `/api/timer/*` -> Study Timer Service (`8084`)
- `/api/leaderboard/*` -> Leaderboard Service (`8086`)
- `/api/session/*` -> Study Session Service (`8083`)
- `/api/reward/*` -> Reward Service (`8085`)

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
