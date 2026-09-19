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
1. **Isolated PostgreSQL Databases**:
   - `auth-db` (Port `5431`, DB `auth_db`)
   - `account-db` (Port `5432`, DB `account_db`)
   - `session-db` (Port `5433`, DB `session_db`)
   - `timer-db` (Port `5434`, DB `timer_db`)
   - `admin-db` (Port `5435`, DB `admin_db`)
2. **Isolated MongoDB Databases**:
   - `reward-db` (Port `27017`, DB `reward_db`)
   - `leaderboard-db` (Port `27018`, DB `leaderboard_db`)

---

## 3. Adding Migrations and Schemas

Because each service manages its own database:
1. Place service-specific DDL scripts and migrations inside `services/<service-name>/database/`.
2. Follow idempotent SQL conventions (e.g., `CREATE TABLE IF NOT EXISTS`).
3. Update [`docs/ARCHITECTURE.md`](../../ARCHITECTURE.md) and [`docs/CHANGELOG.md`](../../CHANGELOG.md).

---

## 4. Connecting Directly via CLI / GUI

- **PostgreSQL CLI (`psql`):**
  ```bash
  docker exec -it fishertimer-auth-db psql -U postgres -d auth_db
  docker exec -it fishertimer-session-db psql -U postgres -d session_db
  ```
- **MongoDB CLI (`mongosh`):**
  ```bash
  docker exec -it fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin
  docker exec -it fishertimer-leaderboard-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin
  ```
- **GUI Clients (DBeaver, TablePlus, Compass):**
  - Postgres: `localhost:5432`, user: `postgres`, password: `postgrespassword`, db: `fishertimer`
  - MongoDB: `mongodb://mongoadmin:mongopassword@localhost:27017/?authSource=admin`
