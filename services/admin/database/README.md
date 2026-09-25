# Admin Moderation Service Database

This directory contains the database configuration, connection management, and migrations for the **Admin Service**.

- **Database Engine:** PostgreSQL
- **Default Database:** `admin_db`
- **Default Port:** `5435`
- **Environment Variable:** `ADMIN_DATABASE_URL`

## Schemas

- [`schemas/001_create_admin_logs_table.sql`](./schemas/001_create_admin_logs_table.sql): Defines the 3NF `admin_logs` table for moderation audit logs (`FORCE_CLOSE_SESSION`, `KICK_USER`).
