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
        Gateway["API Gateway (Go Clean Architecture)<br/>services/api-gateway (Port 8080)"]
    end

    subgraph Backend Microservices [Go 1.22+ Clean Architecture]
        Auth["Auth Service<br/>Port: 8081"]
        Account["Account Service<br/>Port: 8082"]
        Session["Study Session Service<br/>Port: 8083"]
        Timer["Study Timer Service<br/>Port: 8084"]
        Reward["Reward Service<br/>Port: 8085"]
        Leaderboard["Leaderboard Service<br/>Port: 8086"]
        Admin["Admin Moderation Service<br/>Port: 8087"]
    end

    subgraph Persistence Layer [Database Per Service]
        AuthDB[("Auth DB")]
        AccountDB[("Account DB")]
        SessionDB[("Study Session DB")]
        TimerDB[("Study Timer DB")]
        RewardDB[("Reward DB")]
        LeaderboardDB[("Leaderboard DB")]
        AdminDB[("Admin DB")]
    end

    %% Client routing via API Gateway
    ClientWeb -->|REST API| Gateway
    Gateway -->|REST API| Auth
    Gateway -->|REST API| Timer
    Gateway -->|REST API| Leaderboard
    Gateway -->|REST API| Session
    Gateway -->|REST API| Reward

    %% Admin routing direct to Admin Service
    AdminWeb -->|REST API| Admin

    %% Inter-service collaborations
    Admin -.->|Kick Participant / Leave| Session
    Admin -.->|Update Ban Status| Account
    Account -.->|Fetch Stats| Timer
    Account -.->|Fetch Items| Reward
    Session -.->|AwardReward() on EndSession| Reward

    %% Service Database connections
    Auth --> AuthDB
    Account --> AccountDB
    Session --> SessionDB
    Timer --> TimerDB
    Reward --> RewardDB
    Leaderboard --> LeaderboardDB
    Admin --> AdminDB
```

---

## 2. Microservice Topology & Port Registry

All backend services follow Clean / Hexagonal Architecture (Domain -> Usecase -> Adapter).

| Service Name | Port | Directory | Protocol | Primary Store | Responsibilities |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Web Client** | 3000 | `apps/web` | HTTP / WS | - | Responsive student UI, live timer render. |
| **Admin Web** | 3000 | `apps/web/admin` | HTTP / WS | - | Admin moderation portal, direct connection to Admin Service. |
| **API Gateway** | 8080 | `services/api-gateway` | HTTP / REST | - | Client reverse proxy routing to Auth, Timer, Leaderboard, Session, Reward. |
| **Auth** | 8081 | `services/auth` | HTTP / REST | Auth DB | Google OAuth, JWT issuance & verification. |
| **Account** | 8082 | `services/account` | HTTP / REST | Account DB | Profiles, personal stats aggregation, ban records. |
| **Study Session** | 8083 | `services/study-session` | HTTP / REST / WS | Session DB | Room lifecycles, roster limits, real-time presence. |
| **Study Timer** | 8084 | `services/study-timer` | HTTP / REST | Timer DB | Isolated user focus timers, work/break cycles. |
| **Reward** | 8085 | `services/reward` | HTTP / REST | Reward DB | Gamified fish drops, inventory, milestone tracker. |
| **Leaderboard** | 8086 | `services/leaderboard` | HTTP / REST | Leaderboard DB | Fast read-cached rankings by period (weekly/all-time). |
| **Admin** | 8087 | `services/admin` | HTTP / REST / WS | Admin DB | Real-time session monitoring, reports, user bans. |

---

## 3. Service Collaboration & Communication Matrix

As defined in `docs/phase1/microservice.md`:

```
+---------------+------------------------+------------------------------------------+
| Source Service| Target Service & Call  | Reason / Trigger                         |
+---------------+------------------------+------------------------------------------+
| Account       | StudyTimer.Statistics()| Aggregate total focus time on dashboard  |
| Account       | Reward.ViewRewards()   | Render earned fish collection on profile |
| Study Session | Admin.VerifyBanStatus()| Block banned users from creating/joining |
| Study Session | Reward.AwardReward()   | Grant reward to user upon EndSession     |
| Admin         | StudySession.Leave()   | Kick banned or reported users from rooms |
| Admin         | Account.UpdateBanStatus| Write ban flag to user account record    |
+---------------+------------------------+------------------------------------------+
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
- Standard Protobuf/JSON schema definitions for Go inter-service contracts.
