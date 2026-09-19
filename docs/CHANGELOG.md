# Changelog

All notable changes to the Fisher Timer project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Agent Directive:** Whenever you modify or add features, bug fixes, or architectural changes to this repository, you MUST append your changes to this changelog following the protocol defined in [`docs/skill-set/1-document-changes/SKILL.md`](skill-set/1-document-changes/SKILL.md).

---

## [Unreleased]

### Added
- **[architecture]**: Restructured system architecture according to phase 1 design diagram:
  - Saved and embedded architecture diagram in `docs/phase1/diagram.png` and `docs/phase1/microservice.md`.
  - Separated Client and Admin websites in `apps/web`:
    - Client portal at `/` (Student Portal) communicating via the dedicated Client API Gateway.
    - Admin portal at `/admin` (Admin Moderation Portal) communicating directly with `services/admin` (Port 8087).
    - Created dedicated typed clients: `apps/web/lib/client-api.ts` and `apps/web/lib/admin-api.ts`.
  - Implemented Client API Gateway in `services/api-gateway` (Port 8080) in Go Clean Architecture, routing `/api/auth/*`, `/api/timer/*`, `/api/leaderboard/*`, `/api/session/*`, and `/api/reward/*` with CORS support.
  - Added inter-service collaboration: Study Session Service triggers `Reward Service AwardReward()` upon `EndSession`.
  - Restructured data persistence to Database-per-Service architecture:
    - Added `database/` directories with service-level configurations in each microservice.
    - Updated `docker-compose.yml` to provision isolated databases per service (`auth-db`, `account-db`, `session-db`, `timer-db`, `admin-db`, `reward-db`, `leaderboard-db`).
    - Updated service configuration loaders and `.env.example`/`.env` with service-specific database variables.
- Standard Hexagonal / Clean Architecture templates across all 7 Go microservices (`auth`, `account`, `study-session`, `study-timer`, `reward`, `leaderboard`, `admin`).
- Feature-driven modular architecture template for Next.js web application (`apps/web`).
- AI Agent Skill Suite in `docs/skill-set/`:
  - `0-explain-repo`: System overview and onboarding guide for agents.
  - `1-document-changes`: Strict protocol for logging updates in `CHANGELOG.md`.
  - `2-update-architecture`: Automatic architecture document synchronization guidelines.
  - `3-way-of-work`: Testing practices, bug tracking, and agent self-improvement guide.
- Shared governance documents: `ARCHITECTURE.md`, `WAY_OF_WORK.md`, and `LESSONS_LEARNED.md`.
- **[database]**: Added `docker-compose.yml` defining local PostgreSQL 16 (`5432`) and MongoDB 7.0 (`27017`) instances.
- **[database]**: Added sample schema DDLs and validators under `database/schemas/`:
  - PostgreSQL Table 1: `auth.users` & `account.profiles`
  - PostgreSQL Table 2: `session.study_sessions` & `session_participants`
  - MongoDB Collection 1: `fish_rewards` with JSON Schema validator and indexes
- **[database]**: Added mock seed data in `database/seeds/` and automatic Docker container initialization in `database/init/`.
- **[skill-set]**: Added Skill 4 (`4-database-and-repository`) and Skill 5 (`5-migration-and-seeding`).
- **[package.json]**: Added database lifecycle scripts: `db:up`, `db:down`, `db:logs`, `db:reset`.
- **[docs]**: Created root `README.md` with system architecture diagrams, team roster, quickstart steps, and document index.
- **[config]**: Created root `.env.example` defining central environment variables, database strings, and OAuth/JWT secrets.
- **[web]**: Configured Next.js API Gateway reverse proxy rewrites in `apps/web/next.config.js` to proxy `/api/*` to backend microservices, eliminating CORS.
- **[web]**: Updated `apps/web/lib/api-client.ts` to fetch through the unified relative gateway route.

---

## [0.1.0] - 2026-09-11

### Added
- Initialized Turborepo monorepo orchestrated with `pnpm` workspaces.
- Created microservices skeletons for 7 services with default port allocations:
  - `auth` (8081)
  - `account` (8082)
  - `study-session` (8083)
  - `study-timer` (8084)
  - `reward` (8085)
  - `leaderboard` (8086)
  - `admin` (8087)
- Created Next.js 16 (React 19) frontend client in `apps/web` (Port 3000).
- Created shared packages:
  - `@fishertimer/shared-types`: Common domain interfaces and DTOs.
  - `@repo/ui`: Shared UI library.
  - `@repo/typescript-config`: Shared TypeScript compiler configurations.
  - `@repo/eslint-config`: Shared ESLint configurations.
- Multi-module Go workspace `go.work` linking all 7 Go services.
- Reproducible scaffolding script in `scripts/setup-turborepo.mjs`.
- Phase 1 documentation in `docs/phase1/turborepo.md`, `microservice.md`, and `project-desc.md`.
