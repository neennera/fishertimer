# Changelog

All notable changes to the Fisher Timer project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Agent Directive:** Whenever you modify or add features, bug fixes, or architectural changes to this repository, you MUST append your changes to this changelog following the protocol defined in [`docs/skill-set/1-document-changes/SKILL.md`](skill-set/1-document-changes/SKILL.md).

---

## [Unreleased]

### Added
- **[web]**: Added TailwindCSS v4 and a pixel-art design system, fulfilling ADR-002:
  - `app/tokens.css`: the single source of truth — 14 colours, the `--px` art-pixel unit, the `--pixclip` one-pixel bevel, and four type roles. No other file may contain a raw hex or px value.
  - `app/pixel.css`: surface primitives (`.pixel-panel`, `.pixel-btn`, `.pixel-input`, `.pixel-alert`, `.pixel-badge`, `.pixel-tile`), all geometry derived from `--px`.
  - `app/fonts.ts`: Jersey 15, Jersey 25, Silkscreen and Nunito via `next/font/google`.
- **[web]**: Added `components/ui/` primitives (`PixelButton`, `PixelPanel`, `PixelInput`, `PixelAlert`, `PixelBadge`, `StatTile`) and a shared `Header`.
- **[web]**: Added `/styleguide` route rendering every component and state from the real components — the verification surface for the design system.
- **[web]**: Added `apps/web/DESIGN_SYSTEM.md` — reference for the token rules, the `--px` art-pixel grid, the colour and type tokens, and the `pixel.css` class list.
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

### Changed
- **[web]**: Replaced the Turborepo starter `layout.tsx`, `page.tsx` and `globals.css`; the app is no longer titled "Create Next App".

### Fixed
- **[web]**: Fixed `pnpm lint` failure in `apps/web/next.config.js` — added a Node globals override in `apps/web/eslint.config.js` so `process` is recognized, and declared the 7 gateway service URL env vars (`AUTH_SERVICE_URL`, `ACCOUNT_SERVICE_URL`, `SESSION_SERVICE_URL`, `TIMER_SERVICE_URL`, `REWARD_SERVICE_URL`, `LEADERBOARD_SERVICE_URL`, `ADMIN_SERVICE_URL`) in `turbo.json`'s `build` task so Turborepo hashes them correctly and `turbo/no-undeclared-env-vars` stops flagging them.
- **[shared-types]**: Fixed `pnpm check-types` failure caused by `packages/shared-types/tsconfig.json` extending the nonexistent `@fishertimer/typescript-config/base.json`; corrected to `@repo/typescript-config/base.json`.

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
