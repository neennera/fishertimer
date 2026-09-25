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
