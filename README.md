# Fisher Timer 🎣⏱️

> A community-based study timer platform with fishing-themed gamification ("FishTank") that combines individual time management with real-time peer accountability.

---

## 👥 Group Members (Software Architecture)

| No. | Student Name | Student ID |
| :---: | :--- | :---: |
| 1 | Chanatda Konchom | `6631305321` |
| 2 | Ittichet Thongsang | `6631363721` |
| 3 | Napat Srisamut | `6631314021` |
| 4 | Naphat Serirak | `6632061321` |
| 5 | Chayut Archamongkol | `6631307621` |

---

## 🏗️ System Architecture Overview

Fisher Timer is architected as a **polyglot microservices monorepo** managed by **Turborepo** and **pnpm workspaces**, featuring a **Next.js 16** frontend client acting as an API Gateway reverse proxy and **7 Go microservices** built with **Clean / Hexagonal Architecture**.

```mermaid
graph TD
    subgraph Client & Gateway ["Frontend Layer (Port 3000)"]
        Web["Next.js 16 Client & API Gateway<br/><code>apps/web</code>"]
    end

    subgraph Backend Microservices ["Go 1.22+ Clean Architecture (services/*)"]
        Auth["Auth Service<br/>Port: 8081"]
        Account["Account Service<br/>Port: 8082"]
        Session["Study Session Service<br/>Port: 8083"]
        Timer["Study Timer Service<br/>Port: 8084"]
        Reward["Reward Service<br/>Port: 8085"]
        Leaderboard["Leaderboard Service<br/>Port: 8086"]
        Admin["Admin Moderation Service<br/>Port: 8087"]
    end

    subgraph Persistence Layer ["Polyglot Storage (docker-compose.yml / Cloud)"]
        Postgres[("Supabase / PostgreSQL 16<br/>(auth, account, session, timer, admin)")]
        Mongo[("MongoDB 7.0<br/>(fish_rewards, leaderboards)")]
    end

    %% Routing
    Web -->|/api/auth/*| Auth
    Web -->|/api/account/*| Account
    Web -->|/api/session/*| Session
    Web -->|/api/timer/*| Timer
    Web -->|/api/reward/*| Reward
    Web -->|/api/leaderboard/*| Leaderboard
    Web -->|/api/admin/*| Admin

    %% Storage connections
    Auth --> Postgres
    Account --> Postgres
    Session --> Postgres
    Timer --> Postgres
    Admin --> Postgres

    Reward --> Mongo
    Leaderboard --> Mongo
```

---

## ⚡ Architectural Decisions (ADR Summary)

- **ADR-001 (Frontend):** Next.js 16 (React 19) with App Router, server-side rendering, and built-in reverse proxy gateway.
- **ADR-002 (UI Framework):** TailwindCSS with shared component design system in `@repo/ui`.
- **ADR-003 (Backend):** Go (Golang) 1.22+ for high-concurrency timer synchronization and Hexagonal / Clean Architecture.
- **ADR-004 (Database):** Polyglot persistence: PostgreSQL (Supabase) for relational integrity & real-time presence; MongoDB for nested gamification rewards.
- **ADR-005 (Authentication):** Google OAuth 2.0 with stateless signed JWT verification across services.
- **ADR-006 (Deployment):** Docker containerization on Render with local development via `docker-compose.yml`.

---

## 🚀 Quick Start Guide

### Prerequisites
- **Node.js** `>= 22.x`
- **pnpm** `>= 9.x`
- **Go** `>= 1.22`
- **Docker & Docker Compose**

### 1. Clone & Setup Environment
```bash
git clone https://github.com/neennera/fishertimer.git
cd fishertimer

# Create local environment configuration
cp .env.example .env
```

### 2. Install Dependencies
```bash
pnpm install
```

### 3. Start Local Databases
Spins up PostgreSQL (`5432`) and MongoDB (`27017`) and automatically applies schema DDLs and seed datasets:
```bash
pnpm db:up
```

### 4. Start Development Servers
Runs Next.js web application and all 7 Go microservices concurrently:
```bash
pnpm dev
```
Visit **`http://localhost:3000`** in your browser.

---

## 🔌 Microservice Port & API Routing Registry

All backend services are proxied through Next.js reverse proxy rewrites, completely eliminating CORS issues on the client:

| Service | Internal Port | Gateway Route | Target Directory | Storage Engine |
| :--- | :---: | :--- | :--- | :--- |
| **Web Client** | `3000` | `/` | `apps/web` | - |
| **Auth** | `8081` | `/api/auth/*` | `services/auth` | PostgreSQL (`auth.users`) |
| **Account** | `8082` | `/api/account/*` | `services/account` | PostgreSQL (`account.profiles`) |
| **Study Session** | `8083` | `/api/session/*` | `services/study-session` | PostgreSQL (`session.study_sessions`) |
| **Study Timer** | `8084` | `/api/timer/*` | `services/study-timer` | PostgreSQL (`timer.timer_states`) |
| **Reward** | `8085` | `/api/reward/*` | `services/reward` | MongoDB (`fish_rewards`) |
| **Leaderboard** | `8086` | `/api/leaderboard/*` | `services/leaderboard` | MongoDB (`rank_snapshots`) |
| **Admin** | `8087` | `/api/admin/*` | `services/admin` | PostgreSQL (`admin.reports`) |

---

## 🛠️ Common Developer Commands

```bash
# Build all workspaces with Turborepo parallel caching
pnpm build

# Run unit tests across all 7 services & frontend
pnpm test

# Lint all code (ESLint + go vet)
pnpm lint

# TypeScript typechecking
pnpm check-types

# Database lifecycle commands
pnpm db:up       # Start local PostgreSQL & MongoDB
pnpm db:logs     # Follow container logs
pnpm db:down     # Stop database containers
pnpm db:reset    # Wipe volumes and re-apply schemas & seeds
```

---

## 📚 Project Documentation Index

- **Phase 1 Deliverables:**
  - [`docs/phase1/project-desc.md`](docs/phase1/project-desc.md): Problem description, target personas, detailed use cases (UC-01 to UC-03), and ADRs.
  - [`docs/phase1/microservice.md`](docs/phase1/microservice.md): Service boundaries, operations, and collaboration matrix.
  - [`docs/phase1/turborepo.md`](docs/phase1/turborepo.md): Monorepo setup, task pipelines, and reproducibility guide.
- **Living Architectural Documents:**
  - [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md): System architecture and data flow documentation.
  - [`docs/CHANGELOG.md`](docs/CHANGELOG.md): Semantic log of changes across services and packages.
  - [`docs/WAY_OF_WORK.md`](docs/WAY_OF_WORK.md): Engineering guidelines and testing standards.
  - [`docs/LESSONS_LEARNED.md`](docs/LESSONS_LEARNED.md): Knowledge base and incident log of resolved technical quirks.
- **AI Agent Skill Suite:**
  - [`docs/skill-set/`](docs/skill-set/): Specialized skills for agents maintaining this repo (explaining repo, documenting changes, architecture updates, way-of-work, database patterns, and migrations).
