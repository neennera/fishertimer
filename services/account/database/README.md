# Account Service Database

This directory contains the database configuration, connection management, and migrations for the **Account Service**.

- **Database Engine:** PostgreSQL (Supabase)
- **Default Database:** `account_db`
- **Default Port:** `5432`
- **Environment Variable:** `ACCOUNT_DATABASE_URL`

## Schemas

- [`schemas/001_create_users_table.sql`](./schemas/001_create_users_table.sql): Defines the 3NF `users` table managing authentication profile and roles (`CUSTOMER`, `ADMIN`).
