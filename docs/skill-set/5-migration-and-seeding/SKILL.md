---
name: migration-and-seeding
description: Step-by-step procedures for managing database migrations, running local seeds, and managing Docker database lifecycles in Fisher Timer.
---

# Skill 5: Database Migration & Seeding Protocol

## Objective
Enable AI agents and engineers to manage local databases, apply schema migrations, seed test datasets, and ensure reproducible database states across all environments.

---

## 1. Quick Start Commands

From the monorepo root, run:

```bash
# Spin up local PostgreSQL and MongoDB containers
pnpm db:up

# View database container logs
pnpm db:logs

# Stop database containers
pnpm db:down

# Completely reset databases and re-run all initialization schemas & seeds
pnpm db:reset
```

---

## 2. Automated Initialization Lifecycle (`docker-compose.yml`)

When you run `pnpm db:up` for the first time:
1. **PostgreSQL** mounts [`database/init/01-init-postgres.sql`](../../../database/init/01-init-postgres.sql) to `/docker-entrypoint-initdb.d/`:
   - Creates schemas: `auth`, `account`, `session`, `timer`, `admin`.
   - Executes DDL creating tables (`auth.users`, `account.profiles`, `session.study_sessions`, etc.).
   - Inserts seed data (`usr_01`, `sess_01`, etc.).
2. **MongoDB** mounts [`database/init/02-init-mongo.js`](../../../database/init/02-init-mongo.js) to `/docker-entrypoint-initdb.d/`:
   - Creates the `fishertimer` database.
   - Enforces JSON Schema validator for `fish_rewards`.
   - Creates performance indexes.
   - Inserts mock fish rewards.

---

## 3. How to Add a New PostgreSQL Migration

When adding or altering tables:
1. Create a sequentially numbered file in `database/schemas/postgres/`:
   ```bash
   database/schemas/postgres/003_create_timer_sessions_table.sql
   ```
2. Write idempotent SQL DDL (use `CREATE TABLE IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`, etc.).
3. If this change affects active containers, append the SQL statement to [`database/init/01-init-postgres.sql`](../../../database/init/01-init-postgres.sql).
4. Run `pnpm db:reset` locally to verify the new migration applies cleanly.
5. Update [`docs/ARCHITECTURE.md`](../../ARCHITECTURE.md) and [`docs/CHANGELOG.md`](../../CHANGELOG.md).

---

## 4. How to Add a New MongoDB Migration / Validator

MongoDB is schemaless, but we enforce document consistency using schema validators:
1. Create a script in `database/schemas/mongodb/`:
   ```bash
   database/schemas/mongodb/002_create_leaderboard_snapshots_collection.js
   ```
2. Define the validator using `db.createCollection("<name>", { validator: { ... } })` and create relevant compound indexes.
3. Append the initialization logic to [`database/init/02-init-mongo.js`](../../../database/init/02-init-mongo.js).
4. Run `pnpm db:reset` to verify.

---

## 5. Connecting Directly via CLI / GUI

- **PostgreSQL CLI (`psql`):**
  ```bash
  docker exec -it fishertimer-postgres psql -U postgres -d fishertimer
  ```
- **MongoDB CLI (`mongosh`):**
  ```bash
  docker exec -it fishertimer-mongodb mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin
  ```
- **GUI Clients (DBeaver, TablePlus, Compass):**
  - Postgres: `localhost:5432`, user: `postgres`, password: `postgrespassword`, db: `fishertimer`
  - MongoDB: `mongodb://mongoadmin:mongopassword@localhost:27017/?authSource=admin`
