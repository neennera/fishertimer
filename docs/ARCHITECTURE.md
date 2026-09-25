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

    subgraph Backend Microservices [Go 1.22+ Clean Architecture]
        Account["Account Service (Auth & Profiles)<br/>Port: 8082"]
        Session["Study Session Service (gRPC / HTTP)<br/>Port: 8083"]
        Timer["Study Timer Service (gRPC / HTTP)<br/>Port: 8084"]
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

    %% Client routing via API Gateway
    ClientWeb -->|REST API| Gateway
    Gateway -->|REST API| Account
    Gateway -->|gRPC / REST| Timer
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

    %% Service Database & Cache connections
    Account --> AccountDB
    Session --> SessionDB
    Timer --> TimerDB
    Reward --> RewardDB
    Leaderboard --> LeaderboardCache
    Admin --> AdminDB
```

---

## 2. Microservice Topology & Port Registry

All backend services follow Clean / Hexagonal Architecture (Domain -> Usecase -> Adapter).

| Service Name | Port | Directory | Protocol | Primary Store | Responsibilities |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Web Client** | 3000 | `apps/web` | HTTP / WS | - | Responsive student UI, live timer render. |
| **Admin Web** | 3000 | `apps/web/admin` | HTTP / WS | - | Admin moderation portal, direct connection to Admin Service. |
| **API Gateway** | 8000 | `services/api-gateway` | HTTP / REST / gRPC Proxy | - | Client reverse proxy routing to Account, Timer, Leaderboard, Session, Reward. Verifies the account service's session JWT and forwards `X-User-Id` / `X-User-Role` / `X-Display-Name` to downstream services (passthrough only — does not itself reject unauthenticated requests). |
| **Account** | 8082 | `services/account` | HTTP / REST | Account DB (`account_db`) | Google OAuth (SignIn, SignUp, SignOut), user profiles, personal stats dashboard aggregation. |
| **Study Session** | 8083 | `services/study-session` | gRPC / HTTP / WS | Session DB (`session_db`) | Room lifecycles (create/join/leave/end), roster limits, active participant listings for admin. |
| **Study Timer** | 8084 | `services/study-timer` | gRPC / HTTP | Timer DB (`timer_db`) | Isolated user focus timers, work/break cycle execution, triggers AwardReward upon CompleteCycle. |
| **Reward** | 8085 | `services/reward` | HTTP / REST | Reward DB (`reward_db`) | Gamified fish drops, rarity table buffed by session participants, user inventory. |
| **Leaderboard** | 8086 | `services/leaderboard` | HTTP / REST | Redis Cache (In-Memory) | Read-optimized ranking of users based on total rewards fetched from Reward Service, cached in Redis. |
| **Admin** | 8087 | `services/admin` | HTTP / REST / WS | Admin DB (`admin_db`) | Real-time session monitoring, active room oversight, kicking participants and closing rooms. |

---

## 3. Service Collaboration & Communication Matrix

As defined in `docs/phase1/microservice.md`:

```
+---------------+-----------------------------+----------+------------------------------------------------+
| Source Service| Target Service & Call       | Protocol | Reason / Trigger                               |
+---------------+-----------------------------+----------+------------------------------------------------+
| Account       | StudyTimer.TimerStatistics()| gRPC     | Aggregate total sessions and focus time        |
| Account       | Reward.ViewRewards()        | REST     | Render earned fish collection on profile       |
| Study Timer   | Reward.AwardReward()        | REST     | Grant reward to user upon CompleteCycle        |
| Leaderboard   | Reward.ViewRewards()        | REST     | Fetches earned rewards for leaderboard ranking |
| Admin         | StudySession.LeaveSession() | gRPC     | Kick user from active study session room       |
| Admin         | StudySession.EndSession()   | gRPC     | Command to close/end active study session room |
+---------------+-----------------------------+----------+------------------------------------------------+
```

---

## 4. Backend Service Architecture (Hexagonal / Clean)

Every Go microservice conforms to the following layer boundaries:

```
services/<service-name>/
├── cmd/
│   └── main.go                  # Composition root: wires dependencies & starts HTTP server
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
│       ├── handler/             # HTTP / REST handlers, parameter binding, status codes
│       │   └── http_handler.go
│       └── repository/          # Database access (Supabase, Mongo, in-memory mocks)
│           └── memory_repo.go
```

**Golden Rule of Clean Architecture:**
Dependencies point **inwards**. The `domain` layer has zero dependencies on frameworks, databases, or external libraries. The `usecase` layer depends only on `domain`. The `adapter` layer implements interfaces declared by `domain`.

---

## 5. Shared Contracts & Types

Domain data structures shared across services and the web client are maintained in:
- `packages/shared-types/src/index.ts` (TypeScript interfaces for Frontend & API consumers)
- Standard Protobuf definitions for Go inter-service contracts:
  - `proto/studysession/v1/session.proto` (Study Session gRPC contracts)
  - `proto/studytimer/v1/timer.proto` (Study Timer gRPC contracts)

---

## 6. Database Architecture & 3NF Schemas

Detailed database schemas, 3NF relations, and DBML definitions are maintained in:
- [`docs/database/schema.dbml`](database/schema.dbml): Authoritative DBML specification.
- [`docs/database/README.md`](database/README.md): Detailed database documentation, port allocations, and service schema references.
- Schema DDL files:
  - Account (`account_db`): `services/account/database/schemas/001_create_users_table.sql`
  - Study Session (`session_db`): `services/study-session/database/schemas/001_create_study_sessions_tables.sql`
  - Study Timer (`timer_db`): `services/study-timer/database/schemas/001_create_timer_tables.sql`
  - Reward (`reward_db`): `services/reward/database/schemas/001_create_reward_collections.js`
  - Admin Moderation (`admin_db`): `services/admin/database/schemas/001_create_admin_logs_table.sql`
  - Leaderboard: In-memory Redis cache (`redis://localhost:6379`)

