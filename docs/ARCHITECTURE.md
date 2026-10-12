# Fisher Timer System Architecture

> **Notice to AI Agents:** This is the authoritative, living architectural documentation for the Fisher Timer system. Whenever services, endpoints, inter-service communications, database models, or infrastructure changes are introduced, this file MUST be updated in compliance with [`docs/skill-set/2-update-architecture/SKILL.md`](skill-set/2-update-architecture/SKILL.md).

---

## 1. High-Level Architecture Overview

Fisher Timer utilizes an event-aware microservices architecture on the backend coupled with a Next.js frontend client, organized within a high-performance Turborepo monorepo.

```mermaid
graph TD
    subgraph Frontend Layer
        ClientWeb["Frontend Website (Client page)<br/>apps/web (Port 3000)"]
        AdminWeb["Frontend Website (Admin page)<br/>apps/web/admin (Port 3000)"]
    end

    subgraph API Gateway Layer
        Gateway["API Gateway (Go Clean Architecture)<br/>services/api-gateway (Port 8000)"]
    end

    subgraph Backend Microservices [Go 1.25+ Clean Architecture]
        Account["Account Service (Auth & Profiles)<br/>Port: 8082"]
        Session["Study Session Service (gRPC / HTTP)<br/>Port: 8083"]
        Timer["Study Timer Service (gRPC / HTTP)<br/>gRPC: 50051 · HTTP: 8084"]
        Reward["Reward Service<br/>Port: 8085"]
        Leaderboard["Leaderboard Service<br/>Port: 8086"]
        Admin["Admin Moderation Service<br/>Port: 8087"]
    end

    subgraph Persistence Layer [Database & Cache per Service]
        AccountDB[("Account DB<br/>(PostgreSQL)")]
        SessionDB[("Study Session DB<br/>(PostgreSQL)")]
        TimerDB[("Study Timer DB<br/>(PostgreSQL)")]
        RewardDB[("Reward DB<br/>(MongoDB)")]
        LeaderboardCache[("Redis Cache<br/>(Leaderboard)")]
        AdminDB[("Admin DB<br/>(PostgreSQL)")]
    end

    subgraph Event Broker & Message Queue
        RabbitMQ["RabbitMQ 3 (Topic Exchange)<br/>Exchange: fisher.session (Port: 5672 / 15672)"]
    end

    subgraph Database Management UI
        Adminer["Adminer Web GUI<br/>Port: 8080"]
    end

    %% Client routing via API Gateway
    ClientWeb -->|REST API| Gateway
    Gateway -->|REST API| Account
    Gateway -->|gRPC :50051| Timer
    Gateway -->|REST API| Leaderboard
    Gateway -->|gRPC / REST| Session
    Gateway -->|REST API| Reward

    %% Admin routing direct to Admin Service
    AdminWeb -->|REST API| Admin

    %% Inter-service collaborations
    Admin -.->|Kick Participant / LeaveSession()| Session
    Admin -.->|End Session / EndSession()| Session
    Account -.->|Fetch History / TimerStatistics()| Timer
    Account -.->|Fetch Items / ViewRewards()| Reward
    Timer -.->|AwardReward() on CompleteCycle| Reward
    Leaderboard -.->|Fetch Rewards / ViewRewards()| Reward

    %% Asynchronous Event-Driven Messaging (RabbitMQ)
    Session -.->|Publish session.participant.left / session.ended| RabbitMQ
    RabbitMQ -.->|Consume timer.session-events queue| Timer

    %% Service Database & Cache connections
    Account --> AccountDB
    Session --> SessionDB
    Timer --> TimerDB
    Reward --> RewardDB
    Leaderboard --> LeaderboardCache
    Admin --> AdminDB

    %% DB Inspection UI
    Adminer -.-> AccountDB
    Adminer -.-> SessionDB
    Adminer -.-> TimerDB
    Adminer -.-> AdminDB
```

---

## 2. Microservice Topology & Port Registry

All backend services follow Clean / Hexagonal Architecture (Domain -> Usecase -> Adapter).

| Service Name | Port | Directory | Protocol | Primary Store | Responsibilities |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Web Client** | 3000 | `apps/web` | HTTP / WS | - | Responsive student UI, live timer render. |
| **Admin Web** | 3000 | `apps/web/admin` | HTTP / WS | - | Admin moderation portal, direct connection to Admin Service. |
| **API Gateway** | 8000 | `services/api-gateway` | HTTP / REST in; HTTP proxy + gRPC client out | - | Client entry point. Reverse-proxies Account, Leaderboard and Reward over HTTP; translates `/api/timer/*` REST into Study Timer gRPC calls and `/api/session/*` REST into Study Session gRPC calls (acting as the signed-in user). Verifies the account service's session JWT and forwards `X-User-Id` / `X-User-Role` / `X-Display-Name` to downstream services (passthrough only — does not itself reject unauthenticated requests). |
| **Account** | 8082 | `services/account` | HTTP / REST | Account DB (`account_db`) | Google OAuth (SignIn, SignUp, SignOut), user profiles, personal stats dashboard aggregation. |
| **Study Session** | 50052 (gRPC), 8083 (HTTP) | `services/study-session` | gRPC / HTTP / AMQP Publisher | Session DB (`session_db`) | Room lifecycles (create/join/leave/end), atomic capacity, one room per user, presence heartbeats, 24h / disconnect / idle auto-removal. Publishes `session.*` events to RabbitMQ through a transactional outbox. |
| **Study Timer** | 50051 (gRPC), 8084 (HTTP) | `services/study-timer` | gRPC / HTTP / AMQP Consumer | Timer DB (`timer_db`) | One timer per participant per room (UC-05 state machine on `timers` / `cycles`), server-timestamp countdowns, a sweeper that completes due cycles, discards long pauses and retries AwardReward (once per `cycle_id`). Opens / finalizes timers from RabbitMQ session events (consumer reconnects on its own). |
| **Reward** | 8085 | `services/reward` | HTTP / REST | Reward DB (`reward_db`) | Gamified fish drops, rarity table buffed by session participants, user inventory. |
| **Leaderboard** | 8086 | `services/leaderboard` | HTTP / REST | Redis Cache (In-Memory) | Read-optimized ranking of users based on total rewards fetched from Reward Service, cached in Redis. |
| **Admin** | 8087 | `services/admin` | HTTP / REST / WS | Admin DB (`admin_db`) | Real-time session monitoring, active room oversight, kicking participants and closing rooms. |
| **RabbitMQ** | 5672 (AMQP), 15672 (UI) | Infrastructure (`docker-compose.yml`) | AMQP 0-9-1 | In-Memory / Mnesia Queue | Message broker for asynchronous inter-service events (`fisher.session` exchange, dead-letter DLX/DLQ). |
| **Adminer** | 8080 | Infrastructure (`docker-compose.yml`) | HTTP | - | Lightweight web-based database management GUI for local inspection of PostgreSQL databases. |

---

## 3. Service Collaboration & Communication Matrix

As defined in `docs/phase1/microservice.md`:

```
+---------------+-----------------------------+----------+------------------------------------------------+
| Source Service| Target Service & Call       | Protocol | Reason / Trigger                               |
+---------------+-----------------------------+----------+------------------------------------------------+
| Account       | StudyTimer.TimerStatistics()| gRPC     | Aggregate total sessions and focus time        |
| Account       | Reward.ViewRewards()        | REST     | Render earned fish collection on profile       |
| Study Timer   | Reward.AwardReward()        | REST     | One call per completed work cycle (cycle_id,   |
|               |                             |          | work_duration, participant_count), retried     |
| Study Timer   | StudySession.GetParticipants| gRPC     | Participant count for the reward (UC-09 S-2)   |
| Leaderboard   | Reward.ViewAllRewards()     | REST     | Fetches all earned rewards for ranking (S-3)   |
| Leaderboard   | Reward.GetLastUpdate()      | REST     | Checks reward mutation timestamp (S-1 cache)   |
| Admin         | StudySession.LeaveSession() | gRPC     | Kick user from active study session room       |
| Admin         | StudySession.EndSession()   | gRPC     | Command to close/end active study session room |
| Study Session | StudyTimer.SessionEvents    | RabbitMQ | Async events (participant.joined / .left,      |
|               |                             |          | session.ended) via transactional outbox        |
| Study Session | StudyTimer.GetRoomTimers()  | gRPC     | Idle detection (UC-05 E-8)                     |
| API Gateway   | StudySession.{Create,Join,  | gRPC     | Browser REST /api/session/* translated to gRPC |
|               |   Leave,Heartbeat,Get*}     |          | (SESSION_GRPC_TARGET, default localhost:50052) |
| API Gateway   | StudyTimer.{Start,Get,Pause,| gRPC     | Browser REST /api/timer/* translated to gRPC   |
|               |   Resume,Stop,Reset,Complete|          | (TIMER_GRPC_TARGET, default localhost:50051)   |
|               |   SkipRest,Settings,GetRoom}|          |                                                |
+---------------+-----------------------------+----------+------------------------------------------------+
```

---

## 4. Backend Service Architecture (Hexagonal / Clean)

Every Go microservice conforms to the following layer boundaries:

```
services/<service-name>/
├── cmd/
│   └── main.go                  # Composition root: wires dependencies & starts HTTP/gRPC servers
├── config/
│   └── config.go                # Strongly typed environment variable loader
├── internal/
│   ├── domain/                  # CORE: Entities, Value Objects, Domain errors, Repository interfaces
│   │   ├── entity.go
│   │   └── repository.go
│   ├── usecase/                 # APPLICATION: Business logic orchestration (independent of frameworks)
│   │   ├── service.go
│   │   └── service_test.go
│   └── adapter/                 # INFRASTRUCTURE: Technical implementations (driving & driven)
│       ├── handler/             # Driving adapters: HTTP / REST and gRPC handlers over the same usecase
│       │   ├── http_handler.go
│       │   └── grpc_handler.go  # (services that expose gRPC, e.g. study-timer)
│       ├── amqp/                # Message broker adapters (RabbitMQ topology, consumers, producers)
│       │   ├── topology.go
│       │   └── consumer.go
│       └── repository/          # Database access (PostgreSQL, Mongo, in-memory mocks)
│           ├── memory_repo.go
│           └── postgres_repo.go
```

**Golden Rule of Clean Architecture:**
Dependencies point **inwards**. The `domain` layer has zero dependencies on frameworks, databases, or external libraries. The `usecase` layer depends only on `domain`. The `adapter` layer implements interfaces declared by `domain`.

---

## 5. Shared Contracts & Types

Domain data structures shared across services and the web client are maintained in:
- `packages/shared-types/src/index.ts` (TypeScript interfaces for Frontend & API consumers)
- `pkg/events/` (Go shared events module: `github.com/neennera/fishertimer/pkg/events` in `go.work`)
  - Shared AMQP message schemas: `ParticipantJoined`, `ParticipantLeft`, `SessionEnded`
  - RabbitMQ exchange constants (`fisher.session`), dead-letter exchange (`fisher.session.dlx`), queues (`timer.session-events`), routing keys (`session.participant.joined`, `session.participant.left`, `session.ended`), and bindings (`session.#`).
- Standard Protobuf definitions for Go inter-service contracts:
  - `proto/studysession/v1/session.proto` (Study Session gRPC contracts)
  - `proto/studytimer/v1/timer.proto` (Study Timer gRPC contracts: `StartTimer`, `GetTimer`, `PauseTimer`, `ResumeTimer`, `ResetTimer`, `GetRoomTimers`, `GetTimerStatistics`)
  - `proto/` is its own Go module (`github.com/neennera/fishertimer/proto`) in `go.work`. Generated `*.pb.go` files are committed; regenerate with `pnpm proto:gen` after editing a `.proto`. Consumers require it with `replace ... => ../../proto` so they also build with `GOWORK=off`.

---

## 6. Database Architecture & 3NF Schemas

Detailed database schemas, 3NF relations, and DBML definitions are maintained in:
- [`docs/database/schema.dbml`](database/schema.dbml): Authoritative DBML specification.
- [`docs/database/README.md`](database/README.md): Detailed database documentation, port allocations, and service schema references.
- Schema DDL files:
  - Account (`account_db`, Port 5432): `services/account/database/schemas/001_create_users_table.sql`
  - Study Session (`session_db`, Port 5433):
    - `services/study-session/database/schemas/001_create_study_sessions_tables.sql` (Phase 1: `study_sessions`, `session_participants`)
    - `services/study-session/database/schemas/002_phase2_rooms_and_outbox.sql` (Phase 2: room `status` / `participant_count` / `end_reason`, one row per stay with `leave_reason` and `last_seen_at`, one-active-room unique index, `event_outbox`)
  - Study Timer (`timer_db`, Port 5434):
    - `services/study-timer/database/schemas/001_create_timer_tables.sql` (Phase 1 legacy: `timer_settings`, `timer_sessions`, `timer_cycles`)
    - `services/study-timer/database/schemas/003_phase2_timer_and_events.sql` (Phase 2 3NF: `timers`, `cycles`, `processed_events`)
    - `services/study-timer/database/schemas/004_timer_sessions_per_participant.sql` (one `timer_sessions` row per participant per room, keyed by `(study_session_id, user_id)`)
    - `services/study-timer/database/schemas/005_phase2_timer_w2.sql` (sweeper indexes; the service now runs on `timers` / `cycles`, Phase 1 tables deprecated)
    - `services/study-timer/database/seeds/001_temp_w1_active_timer_seed.sql` (Phase 2 mock seed: temporary test fixture for isolated local testing, waiting for Study Session Service Role A integration in W2)
  - Reward (`reward_db`, Port 27017):
    - `services/reward/database/schemas/001_create_reward_collections.js` (3NF collections: `reward_items`, `user_rewards`)
    - `services/reward/database/schemas/002_seed_reward_items.js` (Seeds 15 fish sprite catalog items)
    - `services/reward/database/schemas/003_seed_user_rewards.js` (Seeds 41 user catches for 5 demo users)
  - Admin Moderation (`admin_db`, Port 5435): `services/admin/database/schemas/001_create_admin_logs_table.sql`
  - Leaderboard: In-memory Redis cache (`redis://localhost:6379`)
  - Message Queue: RabbitMQ 3 (`amqp://localhost:5672`, Management UI `http://localhost:15672`)
  - Database Management UI: Adminer (`http://localhost:8080`)

