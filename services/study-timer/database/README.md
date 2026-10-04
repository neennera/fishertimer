# Study Timer Service Database

This directory contains the database configuration, connection management, and migrations for the **Study Timer Service**.

- **Database Engine:** PostgreSQL
- **Default Database:** `timer_db`
- **Default Port:** `5434`
- **Environment Variable:** `TIMER_DATABASE_URL`

## Schemas

- [`schemas/001_create_timer_tables.sql`](./schemas/001_create_timer_tables.sql): Defines the 3NF tables:
  - `timer_settings`: Per-user pomodoro durations (focus, short break, long break, cycle count).
  - `timer_sessions`: Active and historical timer sessions (supports solo or room-bound study).
  - `timer_cycles`: Completed interval cycles with phase type, duration, and completion status.
- `schemas/002_add_timer_progress.sql`: Phase 1 progress columns on `timer_sessions`.
- `schemas/003_phase2_timer_and_events.sql`: Phase 2 tables, used by the service today:
  - `timers`: one per participant per room (`OPEN` / `FINALIZED`), opened by `session.participant.joined`.
  - `cycles`: every work / rest period with `started_at`, `paused_at`, `paused_total_sec`, `ended_at`, `status` (RUNNING / PAUSED / COMPLETED / SKIPPED / DISCARDED) and `reward_status` (NONE / PENDING / SENT). A partial unique index keeps one active cycle per timer.
  - `processed_events`: RabbitMQ event ids, for idempotent consumers.
- `schemas/004_timer_sessions_per_participant.sql`: per-participant key for the Phase 1 `timer_sessions` table.
- `schemas/005_phase2_timer_w2.sql`: indexes for the sweeper (active cycles, pending rewards, per-stay lookups); marks `timer_sessions` / `timer_cycles` deprecated (no longer written).

`timer_settings` holds each user's default lengths (`focus_duration`, `short_break_duration`, in seconds).

Init scripts only run on an empty volume: after pulling new migrations, run `pnpm db:reset`.
