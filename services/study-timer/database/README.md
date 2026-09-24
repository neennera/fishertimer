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
