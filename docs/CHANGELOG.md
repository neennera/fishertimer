# Changelog

All notable changes to the Fisher Timer project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Agent Directive:** Whenever you modify or add features, bug fixes, or architectural changes to this repository, you MUST append your changes to this changelog following the protocol defined in [`docs/skill-set/1-document-changes/SKILL.md`](skill-set/1-document-changes/SKILL.md).

---

## [Unreleased]

### Changed
- **[architecture]**: Updated microservice specifications and system topology in `docs/phase1/microservice.md`, `docs/ARCHITECTURE.md`, and `README.md`:
  - **Study Session -> Study Timer**: Moved `AwardReward` collaborator call trigger from Study Session (`EndSession`) to Study Timer (`CompleteCycle`).
  - **Leaderboard**: Removed standalone `leaderboard_db`; replaced with Redis Cache and configured Leaderboard to read user reward records from Reward Service `ViewRewards()`.
  - **Auth -> Account**: Consolidated standalone Auth Service into Account Service (`account_db`), housing Google OAuth (`SignIn`, `SignUp`, `SignOut`) alongside user profiles and statistics dashboard.
  - **gRPC Protocol**: Adopted **gRPC** for Study Timer and Study Session microservices.
  - **Admin Collaboration**: Configured Admin service to directly send commands to Study Session (`LeaveSession()` to kick participants and `EndSession()` to close rooms).

### Removed
- **[admin]**: Removed `BanUser`, `UnbanUser`, `VerifyBanStatus`, `MonitorTimerStatus`, and moderation report operations.
- **[reward]**: Removed `ClaimReward` and `TrackProgression` operations.
- **[leaderboard]**: Removed standalone `FilterByPeriod` operation in favor of cached period filtering in `ViewLeaderboard`.
- **[study-timer]**: Removed `SwitchPhase`, `StartRest`, etc., consolidating cycle transitions into `CompleteCycle` and `SkipRest`.
- **[auth]**: Decommissioned Auth Service as an independent microservice, folding identity management into Account Service.

### Added
- **[database]**: Implemented 3NF database schemas across all microservices and authored full DBML documentation:
  - Added authoritative DBML specification in `docs/database/schema.dbml` and architecture guide in `docs/database/README.md`.
  - Created DDL migration scripts:
    - Account: `services/account/database/schemas/001_create_users_table.sql` (`users`).
    - Study Session: `services/study-session/database/schemas/001_create_study_sessions_tables.sql` (`study_sessions`, `session_participants`).
    - Study Timer: `services/study-timer/database/schemas/001_create_timer_tables.sql` (`timer_settings`, `timer_sessions`, `timer_cycles`).
    - Reward: `services/reward/database/schemas/001_create_reward_collections.js` (`reward_items`, `user_rewards`).
    - Admin Moderation: `services/admin/database/schemas/001_create_admin_logs_table.sql` (`admin_logs`).
  - Updated `docker-compose.yml` to remove `auth-db` and `leaderboard-db`, provision `redis:7-alpine`, and auto-mount DDL scripts.
  - Aligned `@fishertimer/shared-types` domain TypeScript models to reflect 3NF schemas.
- **[proto]**: Added gRPC protobuf service definitions for Study Session (`proto/studysession/v1/session.proto`) and Study Timer (`proto/studytimer/v1/timer.proto`).
- **[services]**: Updated Clean Architecture code skeletons across microservices to match updated operations and boundaries:
  - `account`: Added `SignIn`, `SignUp`, `SignOut`, `ViewProfile`, `UpdateProfile`, and `ViewStatistics` skeletons.
  - `study-session`: Implemented session lifecycle methods (`CreateSession`, `JoinSession`, `LeaveSession`, `EndSession`, `ListActiveSession`, `GetParticipants`) and decoupled reward triggers.
  - `study-timer`: Added `RewardClient` collaborator and skeleton operations (`StartTimer`, `PauseTimer`, `ResumeTimer`, `StopTimer`, `ResetTimer`, `CompleteCycle`, `SkipRest`, `UpdateTimerSetting`, `TimerStatistics`).
  - `admin`: Added `SessionClient` collaborator (`LeaveSession`, `EndSession`) and monitoring skeletons (`ViewActiveSessions`, `ViewParticipants`, `ViewSessionDetails`).
  - `leaderboard`: Replaced MongoDB configuration with `RedisURL` cache configuration and `ViewLeaderboard`/`GetRanking` endpoints.
  - `api-gateway`: Updated reverse proxy routing to map `/api/account/*` and `/api/auth/*` directly to Account Service.
- **[web]**: Added TailwindCSS v4 and a pixel-art design system, fulfilling ADR-002:
  - `app/tokens.css`: the single source of truth — 14 colours, the `--px` art-pixel unit, the `--pixclip` one-pixel bevel, and four type roles. No other file may contain a raw hex or px value.
  - `app/pixel.css`: surface primitives (`.pixel-panel`, `.pixel-btn`, `.pixel-input`, `.pixel-alert`, `.pixel-badge`, `.pixel-tile`), all geometry derived from `--px`.
  - `app/fonts.ts`: Jersey 15, Jersey 25, Silkscreen and DotGothic16 via `next/font/google`.
- **[web]**: Added `components/ui/` primitives (`PixelButton`, `PixelPanel`, `PixelInput`, `PixelAlert`, `PixelBadge`, `StatTile`) and a shared `Header`.
- **[web]**: Added `/styleguide` route rendering every component and state from the real components — the verification surface for the design system.
- **[web]**: Added `apps/web/DESIGN_SYSTEM.md` — reference for the token rules, the `--px` art-pixel grid, the colour and type tokens, and the `pixel.css` class list.
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
- **[leaderboard]**: Added inter-service collaboration from Leaderboard Service to Reward Service:
  - Added `RewardClient.ViewRewards()` collaborator interface and usecase method `FetchUserRewards()` to fetch earned rewards for leaderboard ranking.
  - Implemented `HTTPRewardClient` adapter in `services/leaderboard/internal/adapter/client/reward_client.go`.
  - Exposed `GET /api/v1/reward/rewards` endpoint in Reward Service.
  - Updated phase 1 architecture diagram, `microservice.md`, and system architecture documentation.
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

## [0.2.0] - 2026-09-22

### Added
- **[web]**: Added a mock-backed auth data layer for UC-06 (Sign In & Sign Up):
  - `lib/auth.ts`: `signInWithGoogle()`, `handleAuthCallback()`, `completeFirstTimeSetup()` and `getSession()`, routed through `lib/client-api.ts`'s `clientApiFetch()`. Toggled by `NEXT_PUBLIC_USE_MOCKS` (see `.env.example`) — no component may import `fetch` or `lib/mocks/` directly.
  - `lib/mocks/auth.mock.ts`: canned responses covering the 7 sign-in / first-time-setup wireframe states (default, consent denied, code exchange failed, account creation failed, and the three first-time setup states).
  - `lib/validate-display-name.ts`: the display name rule (trim, 1–30 chars), pulled out of `app/styleguide/page.tsx` so both the styleguide demo and `completeFirstTimeSetup()` share one source of truth.
- **[web]**: Added `components/ui/ParallaxScene.tsx` — a generic, full-viewport layered background scene (`ParallaxLayer[]` props, not signin-specific), combining two rendering modes since one CSS technique can't serve both art styles without either stretching or gapping:
  - `"cover"` layers (`public/sprites/scene/skies/`: `Sky_sky.png` as the base, `sky_clouds.png` composited over it through its own transparency) are smooth painterly/gradient art with no pixel grid to misalign, so `.pixel-parallax__layer--cover` is plain `background-size: cover` — always preserves aspect ratio, only ever crops the overflow, and by definition fully covers any viewport shape (wide, narrow, short, tall). No stretch, boundless.
  - `"tile"` layers (`public/sprites/scene/parallax-lake/`: mountains, forest-far/mid/near, valley-fill, foreground, water) are pixel-art — scaling those by any fraction (a percentage, or `cover`/`contain`) blurs their pixel grid, which reads as "stretched", so `.pixel-parallax__layer--tile` scales by the sprite's native size times `--px` instead (`--scene-tile-w`/`--scene-tile-h` tokens, the same `calc(var(--px) * n)` idiom as every other sprite scale) and tiles horizontally with an animated `background-position-x`, one tile-width per loop. This only ever covers a bounded band anchored to the bottom, not the whole viewport — a `"cover"` sky layer behind it fills whatever space opens up above that band on a tall viewport, so the two compose without a gap.
  - Both variants share `background-position: bottom`, so they crop/anchor consistently with each other. `background-color: var(--color-sky-fill)` (new token, sampled from `Sky_sky.png`'s last opaque pixel before it fades to transparent) is the last-resort fallback below all of it.
  - A `"cover"` layer can optionally drift (`driftSeconds`, e.g. the clouds' slow 90s-per-leg sway): since `cover` already fixes the image's scale, animating `background-position-x` as a percentage only pans within the art rather than resizing it, so it's safe in a way a percentage-driven size never was for `"tile"`. `sky_clouds.png`'s edges don't line up as a seamless tile, so `.pixel-parallax__layer--drift` eases back and forth (`animation-direction: alternate`) instead of looping in one direction, which would jump-cut at the seam.
  - `.pixel-parallax` itself is `position: fixed; inset: 0; width: 100vw; height: 100dvh` — a container sized to its content or a wrapping box never reaches the top of a taller window — with `z-index: -1`, so it paints behind the page's normal-flow content without any component needing its own `z-index`.
  - `lib/scenes/sky.ts` (the 2 `"cover"` layers) and `lib/scenes/parallax-lake.ts` (the 7 `"tile"` ground layers, no longer including that set's own now-unused `sky.png`/`clouds.png`) combine into `lib/scenes/signin-scene.ts`'s `SIGNIN_SCENE_LAYERS`, shared by `/signin` and the styleguide's "Scene" demo (boxed there via `contain: layout`, since the viewport-fixed scene would otherwise take over the whole preview page).
- **[web]**: Added `/signin` (UC-06) and `/welcome` (UC-06 S-1) under `app/(auth)/`, a route group sharing one `app/(auth)/layout.tsx` (`SceneShell`: background + header + centred panel) so the background doesn't remount when navigating between them.
  - `/signin`: Google button wired to `signInWithGoogle()` → `handleAuthCallback()`, routing to `/welcome` or `/` based on the result. Failures set `?error=<code>`, read by `SignInError.tsx`. `?mockScenario=` drives the whole flow in mock mode.
  - `/welcome`: `PixelInput` wired to `completeFirstTimeSetup()`; both validation states (empty, too long) come from the shared `validateDisplayName()` (`lib/validate-display-name.ts`), matching the wireframe copy exactly.
- **[web]**: Added the hanging sign (signpost, signboard, bird) to `SceneShell`, above the panel on every auth screen (`public/sprites/ui/{signpost,signboard}.png`, `public/sprites/items/sign-bird.png`). Sized via `calc(<native px> * var(--px))` in `app/pixel.css`: the post sits behind the panel, the board sways and carries "Fisher Timer" as real text, the bird perches on the panel's top-right corner and idles between frames.
- **[web]**: `/welcome`'s panel now renders immediately regardless of the session load; only the name field and email line show a `.pixel-skeleton` placeholder while `session` is loading, sized to match the real elements so nothing shifts when they swap in.
- **[web]**: `/welcome`'s display-name error now uses `.pixel-field-error` (`app/pixel.css`) instead of a boxed `PixelAlert` — red, pixel-label styled, and always reserves its line so the panel doesn't shift when it appears.
- **[web]**: Added a required Privacy Notice consent checkbox to `/welcome`, gating `Continue` until checked. New `PixelCheckbox` and `PixelModal` components (`components/ui/`), with matching `.pixel-checkbox`, `.pixel-link` and `.pixel-modal` classes in `app/pixel.css`.
