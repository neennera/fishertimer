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
1. **Isolated PostgreSQL Databases (3NF)**:
   - `account-db` (Port `5432`, DB `account_db`)
   - `session-db` (Port `5433`, DB `session_db`)
   - `timer-db` (Port `5434`, DB `timer_db`)
   - `admin-db` (Port `5435`, DB `admin_db`)
2. **Isolated MongoDB Database**:
   - `reward-db` (Port `27017`, DB `reward_db`)
3. **In-Memory Cache**:
   - `redis` (Port `6379`, Redis Cache for Leaderboard)

---

## 3. Adding Migrations and Schemas

Because each service manages its own database:
1. Place service-specific DDL scripts and migrations inside `services/<service-name>/database/schemas/`.
2. Follow idempotent SQL conventions (e.g., `CREATE TABLE IF NOT EXISTS`).
3. Update [`docs/ARCHITECTURE.md`](../../ARCHITECTURE.md), [`docs/database/schema.dbml`](../../database/schema.dbml), and [`docs/CHANGELOG.md`](../../CHANGELOG.md).

---

## 4. Connecting Directly via CLI / GUI

- **PostgreSQL CLI (`psql`):**
  ```bash
  docker exec -it fishertimer-account-db psql -U postgres -d account_db
  docker exec -it fishertimer-session-db psql -U postgres -d session_db
  docker exec -it fishertimer-timer-db psql -U postgres -d timer_db
  docker exec -it fishertimer-admin-db psql -U postgres -d admin_db
  ```
- **MongoDB CLI (`mongosh`):**
  ```bash
  docker exec -it fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin
  ```
- **Redis CLI (`redis-cli`):**
  ```bash
  docker exec -it fishertimer-redis redis-cli ping
  ```
- **GUI Clients (DBeaver, TablePlus, Compass):**
  - Postgres: `localhost:5432` (`account_db`), `localhost:5433` (`session_db`), `localhost:5434` (`timer_db`), `localhost:5435` (`admin_db`), user: `postgres`, password: `postgrespassword`
  - MongoDB: `mongodb://mongoadmin:mongopassword@localhost:27017/?authSource=admin`
  - Redis: `localhost:6379`

