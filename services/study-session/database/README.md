# Study Session Service Database

This directory contains the database configuration, connection management, and migrations for the **Study Session Service**.

- **Database Engine:** PostgreSQL
- **Default Database:** `session_db`
- **Default Port:** `5433`
- **Environment Variable:** `SESSION_DATABASE_URL`

## Schemas

- [`schemas/001_create_study_sessions_tables.sql`](./schemas/001_create_study_sessions_tables.sql): Defines the 3NF tables:
  - `study_sessions`: Room metadata, host ID, active state, and participant limits.
  - `session_participants`: Join table managing participant membership and joined/left timestamps.
- [`schemas/002_phase2_rooms_and_outbox.sql`](./schemas/002_phase2_rooms_and_outbox.sql): Phase 2:
  - `study_sessions`: `status` (ACTIVE / ENDED) replaces `is_active`; `participant_count` for the atomic capacity check; `end_reason` (EMPTY / ADMIN_CLOSED / TIMEOUT_24H); limit constrained to 1-5.
  - `session_participants`: one row per stay (`participant_id` PK, so a user can rejoin a room), `display_name`, `last_seen_at` (heartbeat), `leave_reason`; partial unique index `uq_session_participants_one_active_room` keeps a user in one room at a time.
  - `event_outbox`: RabbitMQ events written in the same transaction as the change, published by the outbox relay.

Init scripts only run on an empty volume: after pulling new migrations, run `pnpm db:reset`.
