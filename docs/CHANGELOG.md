# Changelog

All notable changes to the Fisher Timer project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Agent Directive:** Whenever you modify or add features, bug fixes, or architectural changes to this repository, you MUST append your changes to this changelog following the protocol defined in [`docs/skill-set/1-document-changes/SKILL.md`](skill-set/1-document-changes/SKILL.md).

---

## [Unreleased]

### Added

- **[infra]**: Added RabbitMQ (`rabbitmq:3-management-alpine` on ports 5672 and 15672) and Adminer DB GUI (port 8080) to `docker-compose.yml`.
- **[events]**: Created shared contracts module `pkg/events` with JSON event structs (`ParticipantJoined`, `ParticipantLeft`, `SessionEnded`) and topology constants (`fisher.session`, `fisher.session.dlx`, `timer.session-events`, `timer.session-events.dlq`). Added `./pkg/events` to `go.work`.
- **[study-timer]**: Added schema migration `003_phase2_timer_and_events.sql` creating `processed_events` for idempotency, `timers` (OPEN/FINALIZED), and `cycles` (RUNNING/PAUSED/DISCARDED/COMPLETED).
- **[study-timer]**: Added temporary testing seed script `001_temp_w1_active_timer_seed.sql` for seeding an open timer and running work cycle.
- **[study-timer]**: Implemented AMQP topology auto-declarer (`internal/adapter/amqp/topology.go`) and consumer worker (`consumer.go`) with manual Ack/Nack, prefetch 10, and DLQ routing.
- **[study-timer]**: Extended gRPC interface with `GetRoomTimers` RPC in `proto/studytimer/v1/timer.proto` to fetch all active participant timers for a given room.
- **[study-timer]**: Added mock event publisher CLI test tool `scripts/publish_mock_event.go`.
- **[docs]**: Created `database-design-with-rabbitmq.md` at workspace root detailing Phase 2 3NF schema and RabbitMQ topology.
- **[reward]**: Added `003_seed_user_rewards.js` to seed 41 user catches across 5 demo users (`user1` to `user5`) for Leaderboard and FishTank demonstrations in MongoDB.
- **[reward]**: Added `GET /api/v1/reward/all-rewards` and `GET /api/v1/reward/last-update` endpoints for Leaderboard cache validation (S-1) and ranking computation (S-3).
- **[web]**: Added Leaderboard feature (`apps/web/app/leaderboard/page.tsx`, `features/leaderboard/`) with Weekly, Monthly, and All-Time period tabs, cache status indicators, and current-user highlighting.
- **[reward]**: Added `internal/domain/calculation.go` with the reward maths: one reward per full 15 work minutes, a group buff of `1 + 0.25 x (participants - 1)` (max 5 participants) that boosts every non-COMMON weight, and `DrawRewards`, which picks each reward by weight (repeats allowed). Tests in `calculation_test.go`. Not wired into `AwardReward` yet.
- **[reward]**: Added `domain.ScoreForRarity` (Common 10, Uncommon 25, Rare 50, Epic 100, Legendary 250) as the fish score system; the Mongo repository falls back to it when a catalog row has no `score_value`.
- **[reward]**: Added `database/schemas/004_seed_current_user_month.js`: seeds this month's mock catches for the signed-in account (`SEED_USER_ID`, `SEED_DISPLAY_NAME`) plus 9 `seed-angler-*` competitors; re-seeding first deletes their old `user_rewards`.
- **[web]**: Added `/dev` page and dev-only `POST /api/dev/seed` route (docker exec into the reward Mongo container) to run the 004 seed with one click. Returns 404 in production.
- **[web]**: Header shows fish sprites beside the wordmark, plus small "admin page" and "dev page" links.

### Changed

- **[leaderboard]**: Rankings are now the sum of each catch's tier score (ties: earliest catch, then user id). `RankEntry` has a new `score` field; `FishReward` carries `score_value`.
- **[web]**: `/leaderboard` requires a signed-in user (redirects to `/signin`; the demo `LureQueen` user is gone), uses the scene background, defaults to Monthly, and drops the cache panel and "Last synced". Summary tiles are Current Leader and Your Ranking. The table pages 5 rows at a time with an always-visible `<` / `>` pager, highlights ranks 1-3 in one colour, and pins the viewer's own row (or an "Unranked" row) under a divider. Period tabs are compact and share a row with Refresh. The top Back/Account links are removed.
- **[web]**: Removed all emoji from `/leaderboard`. Ranks 1-3 use new gold, silver and bronze medal sprites and the title uses a trophy sprite (`public/sprites/ui/`); scores use the Goldfish sprite; the empty state uses the Ghostfish sprite; buttons and the error banner are text only.
- **[web]**: New `.pixel-rule` divider in `pixel.css` (one art pixel of `--color-cream-2`, with a `--dashed` variant on the pixel grid). The `/leaderboard` table header line, the line above the viewer's own row, and the line above the bottom buttons use it instead of 1-2px `--color-rule` borders.
- **[web]**: Leaderboard row colours (top-3 rows, the viewer's own row, avatars, rank numbers and the YOU / LEADER badges) moved from inline styles into `.pixel-rank-row` in `pixel.css`, using colour tokens instead of hard-coded `#fff` and `rgba(...)` values.
- **[web]**: Home page (`/`) uses the scene background and clickable feature cards (Study Timer, FishTank Rewards, Leaderboard; Study Sessions is "Coming soon"). Styleguide and Admin Portal links removed.
- **[reward]**: Unified `domain.UnlockedReward` entity to support both the 3NF database schema (items/user drops) for the Account service and leaderboard fields (`ID`, `DisplayName`, `Species`, `AwardedAt`) with `type FishReward = UnlockedReward` for backwards compatibility.
- **[reward]**: Implemented MongoDB repository lookup joining `user_rewards` with `reward_items` supporting all-reward listing (`userID == ""`) and `GetLastUpdate` sorted by `awarded_at: -1`.
- **[reward]**: Configured `cmd/main.go` to connect to MongoDB when available and gracefully fall back to the 5-user in-memory seeded store for isolated local development without Docker.
- **[web]**: Merged Home page (`apps/web/app/page.tsx`) navigation links to include both the Study Timer (`/timer`) and Leaderboard (`/leaderboard`).
- **[web]**: Kept full production implementation of `/account` from `main` using `ProfileShell`, `ProfileSections`, and live session management.
- **[reward]**: `POST /api/v1/reward/award` now really awards rewards (replaces the fake "Golden Salmon" stub). New request `{ user_id, session_id, cycle_id, work_minutes, participant_count, display_name? }`, response `{ cycle_id, already_awarded, rewards }`. Draws one reward per 15 work minutes from `reward_items` and stores one `user_rewards` row per draw. Safe to retry: the same `cycle_id` returns the stored rewards instead of awarding twice (rows get an `_id` derived from the cycle, so no schema change is needed). **Breaking:** study-timer's current `{ user_id, reason }` payload now gets a 400 until it sends the new fields.
- **[reward]**: Changed drop weights to COMMON 30, UNCOMMON 25, RARE 12, EPIC 6, LEGENDARY 3 (was 50 / 25 / 10 / 4 / 1) in `002_seed_reward_items.js` and `memory_repo.go`. Re-run 002 on an existing volume to apply. Updated `database/README.md` to match.

### Fixed

- **[leaderboard]**: Fixed empty ranking issue where `user_rewards` was unseeded after `pnpm db:reset`, causing Leaderboard to cache an empty response into Redis.
- **[study-timer]**: The Postgres repository now persists the live timer progress (`phase`, `running_since`, `elapsed_ms` on `timer_sessions`, added by `database/schemas/002_add_timer_progress.sql`) so remaining time survives reloads with the timestamp-based timer; phase is read from its own column instead of being inferred from the last completed cycle. `TimerStatistics` over gRPC maps `total_sessions` from `SessionsJoined`. Existing local databases need `pnpm db:reset` to pick up the new columns.
- **[web]**: `/timer`'s fixed demo owner uses placeholder UUIDs instead of `demo`, since `timer_db` keys timers by UUID.
- **[web]**: `/account` and `/profile/[userId]` show a Focus history chart (7D / 14D / 30D, from `statistics.daily_focus_minutes`) in place of the "Sessions by room type" placeholder. New `.pixel-chart` classes and `--chart-*` tokens.
- **[web]**: With no profile picture, the profile avatar shows the user's initials, matching the header (was a generic icon).
- **[web]**: The header avatar is a key like the sign-out button (same size and press, cream), and the page always reserves the scrollbar's space so the header doesn't shift between scrolling and non-scrolling pages.
- **[web]**: The header shows a small "Sign in" button (to `/signin`) once the session is known to be signed out; the bar's height is unchanged.
- **[web]**: `/account` and `/profile/[userId]` use the real account service (`profile`, `statistics`, `rewards`, `PATCH update-profile`). Each panel loads and fails on its own; stats add "Last active"; the fish tank shows only fish.
- **[web]**: Follow the account service's new `GET /me` (`prem/account`, `bed051e`): it returns the user or 401, with no pending sign-up state. `/welcome` now always shows the form (no e-mail) and sends expired sign-ups back to `/signin`.
- **[web]**: The header stays pinned to the top of the page while scrolling, and its avatar shows the user's picture (initials if there is none or it fails to load) and links to `/account`.
- **[web]**: The `/account` display-name editor now edits the name in place: the name itself becomes an underlined input in the same font and position, with small check / close icon buttons in the pencil's spot, so nothing shifts: the error line between the name and the e-mail is always reserved. The underline colour, the check / close keys, the error and the saving dim fade in over about 120ms (off under reduced motion). This intentionally departs from wireframe 04b's large input and Cancel / Save buttons; all six 04a–04f states, the validator and the mock `updateDisplayName()` are unchanged. New `.pixel-name-row`, `.pixel-inline-field`, `.pixel-inline-input`, `.pixel-inline-error` classes and a `.pixel-btn--busy` modifier.
- **[architecture]**: Updated microservice specifications and system topology in `docs/phase1/microservice.md`, `docs/ARCHITECTURE.md`, and `README.md`:
  - **Study Session -> Study Timer**: Moved `AwardReward` collaborator call trigger from Study Session (`EndSession`) to Study Timer (`CompleteCycle`).
  - **Leaderboard**: Removed standalone `leaderboard_db`; replaced with Redis Cache and configured Leaderboard to read user reward records from Reward Service `ViewRewards()`.
  - **Auth -> Account**: Consolidated standalone Auth Service into Account Service (`account_db`), housing Google OAuth (`SignIn`, `SignUp`, `SignOut`) alongside user profiles and statistics dashboard.
  - **gRPC Protocol**: Adopted **gRPC** for Study Timer and Study Session microservices.
  - **Admin Collaboration**: Configured Admin service to directly send commands to Study Session (`LeaveSession()` to kick participants and `EndSession()` to close rooms).
- **[go.work]**: Raised the workspace Go version from 1.22 to 1.25.0, the minimum required by `google.golang.org/grpc` v1.84.
- **[proto]**: Revised the Study Timer contract (`proto/studytimer/v1/timer.proto`): added a `GetTimer` RPC so the CRUD set is complete (Start / Get / Pause-Resume / Reset), replaced the free-text `status` and `phase` fields with the `TimerStatus` and `TimerPhase` enums, and added `duration_seconds` and `remaining_seconds` to `TimerStateResponse` so clients can render a countdown.

### Fixed

- **[study-timer]**: `CompleteCycle` awarded a reward and counted a cycle after rest phases too; only a completed work phase now does (UC-05 S-2). The in-memory repository returned shared pointers, letting concurrent requests race on one timer; it now stores copies and reports a missing timer as `ErrNotFound`.

### Fixed

- **[reward]**: `database/schemas/001_create_reward_collections.js` had a unique index on `user_rewards (user_id, item_id)`, and `MongoRepository.Award` upserted with `$setOnInsert` against it - so catching the same fish species a second time silently no-opped instead of being recorded. Dropped the unique constraint (kept the plain compound index for query performance) and switched `Award` to a plain `InsertOne`, so `user_rewards` is a log of every catch rather than a one-time "has unlocked" flag; `GetUserInventory`/`GET /api/v1/reward/rewards` now lists repeat catches as separate entries. `BADGE`/`SKIN` items being granted only once, if desired, is left as an application-layer rule for whoever implements `AwardReward`'s item-selection logic, not a database constraint.

### Removed

- **[admin]**: Removed `BanUser`, `UnbanUser`, `VerifyBanStatus`, `MonitorTimerStatus`, and moderation report operations.
- **[reward]**: Removed `ClaimReward` and `TrackProgression` operations.
- **[leaderboard]**: Removed standalone `FilterByPeriod` operation in favor of cached period filtering in `ViewLeaderboard`.
- **[study-timer]**: Removed `SwitchPhase`, `StartRest`, etc., consolidating cycle transitions into `CompleteCycle` and `SkipRest`.
- **[auth]**: Decommissioned Auth Service as an independent microservice, folding identity management into Account Service.

### Changed

- **[docs]**: Synced architecture docs with the gRPC timer: `docs/ARCHITECTURE.md` (Gateway -> Timer edge is gRPC :50051, timer ports 50051/8084, gateway role, new Gateway -> StudyTimer row in the communication matrix, `grpc_handler.go` in the layer template, `proto` module and `pnpm proto:gen`), root `README.md` and skill 0 (Go 1.25+), and the study-timer / api-gateway READMEs (RPC table, REST-to-gRPC mapping, grpcurl and curl examples).
- **[api-gateway]**: `/api/timer/*` now reaches Study Timer over **gRPC** instead of reverse-proxying HTTP, matching the microservice design (Gateway -> gRPC -> Study Timer). `TimerHandler` translates `GET /api/timer/state` and `POST /api/timer/{start,pause,resume,reset}` (JSON `{session_id, user_id}`) into `GetTimer` / `StartTimer` / `PauseTimer` / `ResumeTimer` / `ResetTimer`, returns the same snake_case JSON shape as the Timer's HTTP API with short enum names (`RUNNING`, `WORK`), and maps gRPC codes to HTTP (400 / 409 / 404 / 503 / 504 / 502). Server-side failures are logged and return a generic message so internal addresses never reach the browser. Config `TimerServiceURL` (`TIMER_SERVICE_URL`) is replaced by `TimerGRPCTarget` (`TIMER_GRPC_TARGET`, default `localhost:50051`); `TIMER_GRPC_PORT` and `TIMER_GRPC_TARGET` are listed in `turbo.json` `globalEnv`.

### Added

- **[reward]**: Seed `reward_items` with the 15 fish that have web sprites (`services/reward/database/schemas/002_seed_reward_items.js`, idempotent); new Koi, Octopus, Seahorse, Globefish and Ghostfish sprites, and the Dungeness crab, added to the web sprite map.
- **[web]**: Redesigned `/timer` for smoother use. Presses apply instantly (optimistic, reconciled with the server; queued so double presses never flicker); the countdown and a new progress bar advance every animation frame; controls keep a fixed layout across states; clear `Ready / Focusing / Paused / Done` states; Space / R shortcuts; countdown in the tab title; state changes announced to screen readers. Adds a CSS-drawn lakeside scene (`.pixel-scene`) and timer surfaces (`.pixel-chip`, `.pixel-clock`, `.pixel-progress`, `.pixel-kbd`) to `pixel.css`, built only from tokens and `--px`.
- **[web]**: `/timer` page with a working study timer (UC-05 demo, not yet linked to a signed-in user or room — it uses a fixed demo owner with placeholder UUIDs). `features/timer/` holds `timer.api.ts` (the only place that calls the backend, via `apiFetch` -> gateway REST -> Study Timer gRPC), `useStudyTimer` (counts down locally from the server's `remaining_seconds` and re-syncs on tab focus and when the countdown hits zero, so the server stays the source of truth) and `TimerPanel` (countdown plus Start / Pause / Resume / Reset, showing only the buttons valid for the current state). Linked from the home page.
- **[study-timer]**: gRPC server for `StudyTimerService` on `TIMER_GRPC_PORT` (default `50051`), running alongside the HTTP server. `GRPCHandler` is a second driving adapter over the same usecase and maps domain errors to gRPC codes (`InvalidArgument`, `FailedPrecondition`, `NotFound`, `Internal`). Server reflection is enabled outside production so `grpcurl` can list and call RPCs without the `.proto` file. The module requires `proto` through a `replace` to `../../proto`, so it also builds with `GOWORK=off`.
- **[study-timer]**: Timer now keeps real time (UC-05). Remaining time is derived from server timestamps (`RunningSince` + accumulated `Elapsed`), so pause freezes it, resume continues it, and a reload never alters it. State-transition rules live on `domain.TimerState` and are unit tested; `ErrInvalidState` rejects illegal moves (e.g. pausing a stopped timer), and a duplicate Start is ignored (UC-05 E-4). Added `GetTimer` usecase and `GET /api/v1/study-timer/state`; HTTP responses now include `duration_seconds` and `remaining_seconds` and return 400 / 409 instead of a blanket 500.
- **[proto]**: Made `proto/` its own Go module (`github.com/neennera/fishertimer/proto`, registered in `go.work`) holding the generated gRPC code for Study Timer and Study Session, so every service imports one shared contract. Regenerate with `pnpm proto:gen` (`scripts/gen-proto.mjs`); the generated `*.pb.go` files are committed so only contract authors need `protoc`.
- **[account]**: Implemented Google OAuth 2.0 sign-in (UC-06 SignIn / SignUp / SignOut) in the Account Service:
  - `GET /api/v1/account/google/login` (redirect to Google with an anti-CSRF `state` cookie), `GET /api/v1/account/google/callback` (code exchange, account match-or-create, session cookie), `GET /api/v1/account/me`, `POST /api/v1/account/signout`.
  - Accounts are matched by e-mail per UC-06; new accounts are created with `role = CUSTOMER`, and an existing `ADMIN` row keeps its role (E-4).
  - Driven ports `domain.OAuthProvider` and `domain.TokenService` with adapters `internal/adapter/oauth` (Google) and `internal/adapter/token` (7-day HS256 JWT carrying `user_id`, `role` and `display_name`).
  - PostgreSQL adapter for `account_db.users`; the service fails fast at startup when `account_db` is unreachable instead of running on a volatile store.
  - `profile` and `statistics` endpoints now require a valid session instead of a `user_id` query parameter.
  - Added `PATCH /api/v1/account/update-profile` — `{display_name}` body, requires a session (`ft_session` cookie or `Authorization: Bearer`); renames the user, persists it via a new `Repository.UpdateProfile(userID, displayName)` method, and re-issues `ft_session` (`200` + updated user) since the JWT carries `display_name`. Errors: `400` (missing/too-long name), `401` (no/expired session).
  - Added `GET /api/v1/account/profile?id=<user_id>` — looks up any user by id via `Usecase.GetProfile`, no session required (`200` + user JSON, `400` missing `id`, `404` not found).
  - Added `GET /api/v1/account/statistics?id=<user_id>` — pulls a user's Pomodoro timer stats (sessions joined, cycles completed, total focus time) from the Study Timer service by id. Confirms the account exists via `Repository.GetUserByID` before calling out, then delegates to a new `domain.TimerClient` port (`internal/adapter/client.HTTPTimerClient`, calling `GET /api/v1/study-timer/statistics?user_id=<id>` on the URL from the new `TIMER_SERVICE_URL` config var, default `http://localhost:8084`). Errors: `400` (missing `id`), `404` (no such account), `502` (Study Timer unreachable/errored).
- **[account]**: Added `GET /api/v1/account/rewards?id=<user_id>` — pulls everything a user has unlocked from the Reward service by id. Confirms the account exists via `Repository.GetUserByID` before calling out, then delegates to a new `domain.RewardClient` port (`internal/adapter/client.HTTPRewardClient`, calling `GET /api/v1/reward/rewards?user_id=<id>` on the URL from the new `REWARD_SERVICE_URL` config var, default `http://localhost:8085`). Errors: `400` (missing `id`), `404` (no such account), `502` (Reward service unreachable/errored). Responds with a shaped summary rather than the raw per-catch list: `usecase.summarizeRewards` groups the Reward service's flat catches by `item_id` into `{total_awards_earned, total_score, items: [{name, rarity, asset_url, type, score_value, count}, ...]}`, where `total_awards_earned`/`total_score` count every catch (including repeats - `total_score` sums `score_value` since the catalog has no `cost` field) and each `items` entry's `count` is how many times that distinct item has been caught. This grouping lives in account (not reward) because Leaderboard's `ViewRewards()` collaborator still needs reward's flat per-catch list.
- **[study-timer]**: Extended `domain.TimerHistory` (returned by `GET /api/v1/study-timer/statistics?user_id=<id>`) with a `cycles_completed` field alongside the renamed `sessions_joined` and `total_focus_minutes`, so the Account service's new statistics collaborator has all three figures it needs.
- **[study-timer]**: Added a `daily_focus_minutes` field to `domain.TimerHistory` (`GET /api/v1/study-timer/statistics?user_id=<id>`) — a 30-entry `[{date, focus_minutes}]` array covering today and the previous 29 days (oldest first).
- **[study-timer]**: Implemented `internal/adapter/repository.PostgresRepository` against `timer_db` (`timer_settings`, `timer_sessions`, `timer_cycles`) and wired it into `cmd/main.go` in place of the in-memory stub, which is now a unit-test double only. `GetHistory` computes real `sessions_joined` (`COUNT` of `timer_sessions`), `cycles_completed` (`COUNT` of completed `timer_cycles`) and `total_focus_minutes` (`SUM` of `FOCUS`-phase cycle durations) instead of returning a hardcoded `{12, 40, 300}` for every user. `SaveTimer` reconstructs `TimerState.Phase` from the schema's `status` enum while running, or from the last completed cycle's phase (its opposite - `CompleteCycle` always flips) once paused/stopped, and records a `timer_cycles` row whenever `CurrentCycle` advances. The service now fails fast at startup if `timer_db` is unreachable, matching the account service's pattern. Added `lib/pq` as a direct dependency.
- **[reward]**: Implemented `internal/adapter/repository.MongoRepository` against `reward_db` (`reward_items`, `user_rewards`) and wired it into `cmd/main.go` in place of the in-memory stub, which is now a unit-test double only. `ListByUser`/`GetUserInventory` runs a real `$lookup` aggregation joining `user_rewards` with `reward_items` instead of returning a hardcoded single fish for every user. Renamed `domain.FishReward` (fake `Species`/`Rarity` fields) to `domain.UnlockedReward`, matching the actual seeded schema (`item_id`, `item_name`, `item_type`, `cost`, `unlocked_at`). The service now fails fast at startup if `reward_db` is unreachable. `AwardReward`/`POST /api/v1/reward/award` remains an unimplemented stub - deciding how a `reason` (e.g. `CompleteCycle`) maps to a specific catalog item is out of scope for this change. Added `go.mongodb.org/mongo-driver/v2` as a direct dependency.
- **[reward]**: Redesigned `reward_items`/`user_rewards` in `database/schemas/001_create_reward_collections.js` (and synced `docs/database/schema.dbml`, which had drifted from what was actually implemented) to match the intended design: `reward_items` drops `description`/`cost`/`created_at` and `item_type` in favor of `category` (`FISH`/`DECORATION`/`ROD`), and gains `rarity` (`COMMON`/`UNCOMMON`/`RARE`/`EPIC`/`LEGENDARY`), `base_weight` (relative drop-rate weight) and `score_value` (leaderboard points), all required. `user_rewards` renames `unlocked_at` to `awarded_at`, gains a required `cycle_id` (the WORK cycle in `timer_db.timer_cycles` that earned the drop - not yet wired to a real value anywhere, since `AwardReward` is still a stub), and its own `user_reward_id` (the document's own `_id`, now exposed in the API). Indexes are now a plain `user_id` index plus one on `awarded_at`, matching the DBML (still no uniqueness - repeat catches are still allowed). `domain.UnlockedReward`, `MongoRepository`, the seed data, and the 3 catalog items were all updated to match; the seeded `user_rewards` rows reference the two real FOCUS cycles already seeded in `services/study-timer/database/schemas/002_seed_test_timer.sql` for realism.
- **[api-gateway]**: Added identity-forwarding middleware (`internal/adapter/middleware.Verifier`) wrapping the whole proxy mux in `cmd/main.go`:
  - Verifies the account service's session JWT (`ft_session` cookie or `Authorization: Bearer` header) using a standard-library-only copy of its HS256 check (shared `JWT_SECRET`/`JWT_ISSUER`).
  - On a valid token, sets `X-User-Id`, `X-User-Role` and `X-Display-Name` (URL-encoded) on the proxied request; always strips any client-supplied copies of these headers first so downstream services can trust them.
  - Does not itself reject unauthenticated requests — passthrough only; enforcing that a route requires a session is left to each downstream service.
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
- **[web]**: Added `/account` (UC-07, wireframe 03a). Redirects to `/signin` when signed out and to `/welcome` when sign-up is pending; sign-out is the header's button.
  - Profile panel: avatar (falls back to an icon), display name and e-mail. New `.pixel-avatar` class and `PixelButton` `small` prop.
  - Edit display name in place (UC-07, wireframes 04a–04f): the pencil swaps the name for an input with Cancel / Save, using the shared `validateDisplayName()`. Saving updates the name and header without a reload. New `updateDisplayName()` in `lib/auth.ts` is **mock-only** (no backend endpoint yet; with mocks off it always fails). `/account?mockScenario=edit-name-save-failed` shows the save-failed state. `PixelButton` now accepts a `ref`.
  - Stat tiles, **mock-only** (`lib/mocks/account-stats.mock.ts`), using `main`'s account-service field names (`total_sessions`, `total_focus_minutes`, `rewards_earned`).
  - "Sessions by room type": placeholder only. It needs study-session and study-timer data, and `schema.dbml` has no room-type column yet.
  - Fish Tank, **mock-only** (`lib/mocks/rewards.mock.ts`): up to 15 caught fish swim in an aquarium (`components/account/FishTank.tsx`, `useFishSwim.ts`), with a card per species showing its count and the total in the panel header. Motion is off under `prefers-reduced-motion`. Sprites are mapped in `lib/fish-sprites.ts`; the background layers are in `lib/scenes/fishtank.ts`.
  - **Schema conflict:** per-species counts need several `user_rewards` rows for one item, but `schema.dbml` makes `(user_id, item_id)` unique.
  - Wood-plank page background (`public/sprites/scene/wood.png`) with soft lamp lighting; it stays still while the page scrolls.
- **[web]**: Fish tank bubbles (`components/account/useBubbles.ts`, sprites `big_bubble.png` / `small_bubble.png` drawn by the team): every 10–30s a short stream of translucent bubbles rises from the sand and fades after 3 seconds or at the surface, never past it. Easter egg: tapping the tank quickly many times can release bubbles where you tap. Off under reduced motion.
- **[web]**: Added a public profile view, `/profile/[userId]`: avatar, display name, stats, sessions by room type and fish tank, without the e-mail or the name editor. It shows a "Fisher not found" panel for unknown ids; signed-out viewers go to `/signin`. On `/account` a small "View public profile" button sits at the end of the e-mail line; your own public view has a "Back to account" button in the exact same spot, with the name where it is on `/account`. `PixelButton`'s `small` modifier now also works for text buttons. Both pages render the same shared components (`components/profile/`, with an `editable` prop). Loading skeletons use the real panel structure, and panels have locked minimum heights (`--panel-min-*` tokens), so loading and loaded pages are the same size. Data is **mock-only** (`lib/profile.ts`, `lib/mocks/profile.mock.ts`); the public type is the user wire type without the e-mail, and real data will come from the account, study-session / study-timer and reward services.
- **[web]**: Tiles inside panels are now filled with a new `--color-tile` token (warm tan).
- **[web]**: Added `apps/web/CREDITS.md` for third-party art.
- **[web]**: Added `@fishertimer/shared-types` as an `apps/web` dependency and `'account'` to `client-api.ts`'s `ClientServiceName`, for future profile/statistics calls (`'auth'` stays as the existing gateway alias used by `lib/auth.ts`).

---

### Changed

- **[web]**: Replaced the Turborepo starter `layout.tsx`, `page.tsx` and `globals.css`; the app is no longer titled "Create Next App".
- **[web]**: Rewrote the sign-in / sign-up code (`lib/auth.ts`, `lib/client-api.ts`, `lib/mocks/auth.mock.ts`) to match the real account service API:
  - Sign-in now redirects the browser to `/api/auth/google/login`; the backend sends it back to `/welcome` (new account) or `/signin` (existing account, or `?auth_error=`). `handleAuthCallback()` is replaced by `readAuthError()`.
  - `getSession()` uses `GET /api/auth/me` (`signed_in` / `needs_signup` / `signed_out`); user fields are now the backend's snake_case names, with `role` passed through. Sign-up posts to `/api/auth/signup`; added `signOut()`.
  - API calls go through the same-origin Next rewrites so the backend's login cookies are sent. Mock scenarios 01a–01d / 02a–02c still work via `?mockScenario=`.
- **[web]**: Added a sign-out button to the header (top right): a square red `PixelButton` with the `Logout` icon from the new `pixelarticons` dependency. It shows only when signed in, calls `signOut()` (`POST /api/auth/signout`), then goes to `/signin`. New `components/SessionHeader.tsx` loads the session for the presentational `Header`; used on `/`. `PixelButton` gains `variant="danger"` and `icon` (`.pixel-btn--danger`, `.pixel-btn--icon`), with new `--color-rust-dk` / `--color-rust-dp` tokens.
- **[web]**: Disabled buttons are now gray (new `--color-stone`, `--color-stone-dk`, `--color-stone-ink` tokens) instead of beige hex values hard-coded in `pixel.css`, and no longer show a variant's hover colour.

### Fixed

- **[web]**: `globals.css` used `overflow-x: hidden` on `html`/`body`, which made `body` its own scroll container and broke sticky positioning; it now uses `overflow-x: clip`.
- **[web]**: Link classes (`.pixel-link`, Tailwind's `underline`) had no effect: an unlayered `a` reset in `globals.css` overrode them. It now sits in `@layer base`.
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
