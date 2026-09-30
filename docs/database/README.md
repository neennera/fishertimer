# Fisher Timer - Database Architecture (v2 3NF Specification)

This directory documents the database architecture and 3NF schema definitions for the Fisher Timer microservices platform.

## Architecture Pattern: Database-per-Service

In adherence to microservice autonomy and high cohesion, each service manages its own isolated datastore. Cross-service references use logical UUIDs without physical foreign keys across database instances.

| Service | Database / Datastore | Engine | Port | Primary Entities / Collections |
| :--- | :--- | :--- | :--- | :--- |
| **Account** | `account_db` | PostgreSQL (Supabase) | `5432` | `users` |
| **Study Session** | `session_db` | PostgreSQL | `5433` | `study_sessions`, `session_participants` |
| **Study Timer** | `timer_db` | PostgreSQL | `5434` | `timer_settings`, `timers`, `cycles`, `processed_events`<br>*(Legacy: `timer_sessions`, `timer_cycles`)* |
| **Reward** | `reward_db` | MongoDB | `27017` | `reward_items`, `user_rewards` |
| **Admin Moderation**| `admin_db` | PostgreSQL | `5435` | `admin_logs` |
| **Leaderboard** | Redis Cache | Redis | `6379` | In-memory sorted sets (no dedicated SQL/Mongo DB) |
| **Message Broker** | RabbitMQ | RabbitMQ (Management) | `5672`<br>`15672` (UI) | Exchange: `fisher.session`, DLX: `fisher.session.dlx`<br>Queue: `timer.session-events`, DLQ: `timer.session-events.dlq` |
| **Database GUI** | Adminer | Docker Web GUI | `8080` | Web inspection for all SQL databases |

---

## DBML Schema Reference

The canonical DBML representation is available in [`schema.dbml`](./schema.dbml).

### Summary of Tables and Relationships

#### 1. Account Service (`account_db`)
- `users`: Managed user accounts with Google OAuth 2.0 integration.
  - `user_id` (UUID, PK)
  - `email` (VARCHAR(255), UNIQUE, NOT NULL)
  - `display_name` (VARCHAR(100), NOT NULL)
  - `avatar_url` (VARCHAR(255))
  - `role` (VARCHAR(20), NOT NULL, DEFAULT `'CUSTOMER'`)
  - `created_at`, `updated_at` (TIMESTAMPTZ)

#### 2. Study Session Service (`session_db`)
- `study_sessions`: Collaborative study room definitions.
  - `session_id` (UUID, PK)
  - `title` (VARCHAR(150), NOT NULL)
  - `host_id` (UUID, NOT NULL, logical ref to `account_db.users.user_id`)
  - `is_active` (BOOLEAN, DEFAULT TRUE)
  - `max_participants` (INT, DEFAULT 4)
  - `created_at`, `ended_at` (TIMESTAMPTZ)
- `session_participants`: Join table tracking participant room roster.
  - `session_id` (UUID, FK -> `study_sessions.session_id`)
  - `user_id` (UUID, logical ref to `account_db.users.user_id`)
  - `joined_at`, `left_at` (TIMESTAMPTZ)
  - Composite PK: `(session_id, user_id)`

#### 3. Study Timer Service (`timer_db` - Phase 2 Specification)
- `timer_settings`: Per-user pomodoro configuration.
  - `user_id` (UUID, PK, logical ref to `account_db.users.user_id`)
  - `focus_duration` / `work_duration_seconds` (INT, default 1500 seconds / 25 mins)
  - `short_break_duration` / `rest_duration_seconds` (INT, default 300 seconds / 5 mins)
  - `updated_at` (TIMESTAMPTZ)
- `timers`: Session timer status per user per study room (Phase 2).
  - `timer_id` (UUID, PK)
  - `session_id` (UUID, logical ref to `session_db.study_sessions.session_id`)
  - `user_id` (UUID, logical ref to `account_db.users.user_id`)
  - `status` (`OPEN` | `FINALIZED`)
  - `opened_at` (TIMESTAMPTZ, NOT NULL)
  - `finalized_at` (TIMESTAMPTZ, nullable)
  - Unique Constraint: `(session_id, user_id)`
- `cycles`: Detailed cycle tracking within a timer (Phase 2).
  - `cycle_id` (UUID, PK)
  - `timer_id` (UUID, FK -> `timers.timer_id` ON DELETE CASCADE)
  - `type` (`WORK` | `REST`)
  - `status` (`RUNNING` | `PAUSED` | `COMPLETED` | `SKIPPED` | `DISCARDED`)
  - `duration_sec` (INT, target interval seconds)
  - `started_at` (TIMESTAMPTZ, NOT NULL)
  - `paused_at` (TIMESTAMPTZ, nullable)
  - `paused_total_sec` (INT, total paused duration)
  - `ended_at` (TIMESTAMPTZ, nullable)
  - `reward_status` (`NONE` | `PENDING` | `SENT`)
  - Partial Unique Index: `(timer_id) WHERE status IN ('RUNNING', 'PAUSED')`
- `processed_events`: RabbitMQ at-least-once message deduplication store.
  - `event_id` (VARCHAR(64), PK)
  - `event_type` (VARCHAR(50), NOT NULL)
  - `processed_at` (TIMESTAMPTZ, NOT NULL)
- *(Legacy Phase 1 tables maintained for backward compatibility: `timer_sessions`, `timer_cycles`)*
- *Seed Note (`001_temp_w1_active_timer_seed.sql`):* Temporary mock fixture for Week 1 isolated testing. Seeds an `OPEN` timer and `RUNNING` cycle solely to verify AMQP consumer finalization locally. In production, timers/cycles are created dynamically via API/gRPC; this mock fixture will be superseded once integrated with Study Session Service (Role A) in Week 2.

#### 4. Reward Service (`reward_db`)
- `reward_items`: Catalog of unlockable reward items.
  - `_id` (ObjectId, PK)
  - `item_name` (VARCHAR(100), NOT NULL)
  - `description` (TEXT)
  - `item_type` (`SKIN` | `BADGE` | `FISH_SPECIES`)
  - `cost` (INT, DEFAULT 0)
  - `created_at` (TIMESTAMPTZ)
- `user_rewards`: Unlocked user reward inventory.
  - `_id` (ObjectId, PK)
  - `user_id` (UUID, logical ref to `account_db.users.user_id`)
  - `item_id` (ObjectId, FK -> `reward_items._id`)
  - `unlocked_at` (TIMESTAMPTZ)
  - Unique index: `(user_id, item_id)`

#### 5. Admin Service (`admin_db`)
- `admin_logs`: Audit trail for moderation actions.
  - `log_id` (UUID, PK)
  - `admin_id` (UUID, logical ref to `account_db.users.user_id`)
  - `action` (`FORCE_CLOSE_SESSION` | `KICK_USER`)
  - `target_id` (UUID, target session or user)
  - `reason` (TEXT)
  - `created_at` (TIMESTAMPTZ)

#### 6. Leaderboard Service
- Utilizes Redis sorted sets (`ZSET`) cached in memory on port `6379`.
- Dynamic scores are queried from Reward Service and Study Timer Service.
