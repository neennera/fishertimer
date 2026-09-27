# Study-timer Service

## Overview
Independent user timer execution, focus cycles, and rest intervals

## Port
Default port: `8084`

## HTTP API

| Method | Route | Purpose |
| :--- | :--- | :--- |
| `GET` | `/api/v1/study-timer/statistics?user_id=<id>` | Returns `{user_id, sessions_joined, cycles_completed, total_focus_minutes, last_active, daily_focus_minutes}` for a user, where `daily_focus_minutes` is a 30-entry `[{date, focus_minutes}]` array covering today and the previous 29 days (oldest first). Called by the Account service's `GET /api/v1/account/statistics?id=<user_id>` collaborator. |

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
