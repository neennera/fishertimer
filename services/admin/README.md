# Admin Service

## Overview
Real-time room monitoring (UC-04) and moderation (UC-10) for administrators.
Live room state is read from the Study Session service; `admin_db` stores only
the moderation audit log (`admin_logs`).

## Port
Default port: `8087`

## Authentication
The admin website calls this service directly (Next.js rewrite
`/api/admin/*`), not through the API Gateway, so the service verifies the
account service's session JWT itself (`ft_session` cookie or
`Authorization: Bearer`), using `JWT_SECRET` / `JWT_ISSUER` from the root
`.env`. Missing or expired token: `401`. Role other than `ADMIN`: `403`.
`/health` and `/api/v1/admin/status` are public.

## API
| Method | Path | Operation |
| :--- | :--- | :--- |
| GET | `/api/v1/admin/sessions` | ViewActiveSessions: live rooms, newest first |
| GET | `/api/v1/admin/sessions/{id}` | ViewSessionDetails: room + participants (host first) |
| GET | `/api/v1/admin/sessions/{id}/participants` | ViewParticipants |
| POST | `/api/v1/admin/sessions/{id}/participants/{userId}/kick` | Study Session `LeaveSession()` (`KICKED`), logs `KICK_USER` |
| POST | `/api/v1/admin/sessions/{id}/end` | Study Session `EndSession()` (`ADMIN_CLOSED`), logs `FORCE_CLOSE_SESSION` |
| GET | `/api/v1/admin/logs?action=` | Audit log, newest first (max 100) |

Kick / end take an optional body `{"reason": "..."}` (max 500 characters, counted
as characters, not bytes) and return `{"status", "session_ended", "log"}`. The room's
host can't be kicked (`400`): close the room instead.

Errors are `{"error": "..."}` with a sentence meant to be shown to the admin: `400` invalid input, `404` room not live,
`409` the room has already ended or the participant has already left (nothing
changed or logged), `502` Study Session unreachable (retry).

The audit row and the Study Session command share one transaction: the row
commits only if the command succeeded, and no command is sent while
`admin_db` is unreachable. The service does not start without `admin_db`
(`pnpm db:up`).

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
