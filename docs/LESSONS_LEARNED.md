# Lessons Learned & Agent Mistake Log

> **Notice to AI Agents:** This living log records encountered errors, environmental quirks, architectural edge cases, and actionable solutions. Before diagnosing strange build failures or runtime bugs, review this file to prevent repeating past mistakes.
>
> **Agent Protocol:** Whenever an agent discovers a non-obvious bug or failure mode and solves it, document the incident here following the template in Section 2.

---

## 1. Incident & Knowledge Archive

### Incident 001: Windows PowerShell Script Execution Policy Block
- **Date:** 2026-09-11
- **Component:** Windows CLI / pnpm / npm
- **Symptom:** Running `pnpm`, `npm`, or `turbo` in PowerShell aborted with:
  `File C:\Users\...\pnpm.ps1 cannot be loaded because running scripts is disabled on this system.`
- **Root Cause:** Windows Client PowerShell execution policy defaults to `Restricted`, which blocks `.ps1` wrapper scripts generated in `AppData\Roaming\npm`.
- **Solution:** Execute package manager commands via `cmd /c "<command>"` or invoke `powershell -NoProfile -ExecutionPolicy Bypass -Command "<command>"`.
- **Action for Future Agents:** Never rely on raw PowerShell script invocation on Windows without checking or bypassing execution policies.

---

### Incident 002: Go Build Fails under Turborepo v2 Strict Environment Filtering
- **Date:** 2026-09-11
- **Component:** Turborepo v2 `build` task & Go compiler
- **Symptom:** Running `pnpm turbo build` on Go microservices failed with:
  `build cache is required, but could not be located: GOCACHE is not defined and %LocalAppData% is not defined`
- **Root Cause:** Turborepo v2 operates in `strict` environment mode by default and strips all environment variables not explicitly configured, hiding `%LocalAppData%` and `%GOCACHE%` from the Go toolchain.
- **Solution:** Add `globalPassThroughEnv` in `turbo.json`:
  ```json
  "globalPassThroughEnv": [
    "LocalAppData",
    "LOCALAPPDATA",
    "GOPATH",
    "GOROOT",
    "GOCACHE",
    "PATH",
    "USERPROFILE",
    "SystemRoot",
    "HOME"
  ]
  ```
- **Action for Future Agents:** Whenever adding non-Node compilers (e.g. Go, Rust, Python) to a Turborepo pipeline, ensure compiler cache and OS paths are explicitly declared in `globalPassThroughEnv`.

---

### Incident 003: Go Workspaces Resolution with Subdirectory Relative Paths
- **Date:** 2026-09-11
- **Component:** Go 1.22 multi-module workspace (`go.work`)
- **Symptom:** Running `go build ./services/...` from the monorepo root failed with:
  `directory prefix services does not contain modules listed in go.work`
- **Root Cause:** In Go workspaces, modules listed in `go.work` are independent modules. Wildcard path matching from root does not automatically cross module boundaries without using the explicit module path or running inside each module.
- **Solution:** Use Turborepo task filtering (`pnpm turbo build --filter=@fishertimer/*-service`) or run commands per module directory (`cd services/<svc> && go build`).
- **Action for Future Agents:** Rely on Turborepo as the orchestrator to navigate into workspace packages rather than guessing multi-module glob patterns.

---

### Incident 004: gRPC Requires a Newer Go Than the Workspace Declared
- **Date:** 2026-09-27
- **Component:** `go.work`, `proto` module, `google.golang.org/grpc`
- **Symptom:** The workspace declared `go 1.22`, but the current `google.golang.org/grpc` cannot be used at that version; `go mod init` / `go mod tidy` also silently set the new module's `go` directive to the local toolchain (`1.27.1`), which would force every teammate onto that exact version.
- **Root Cause:** `google.golang.org/grpc` v1.84 declares `go 1.25.0` in its own `go.mod`. A dependency's minimum Go version propagates to every module that imports it, and `go.work` must be at least as new as each member module.
- **Solution:** Raised `go.work` and the modules that import gRPC (`proto`, `study-timer`, `api-gateway`) to `go 1.25.0`. Contributors need Go 1.25 or newer installed.
- **Action for Future Agents:** Before adding a Go dependency, check the `go` directive in that dependency's `go.mod`; if it is newer than `go.work`, raise the workspace deliberately (and tell the team) rather than letting `go mod tidy` do it silently.

---

### Incident 005: Shared Proto Module Must Also Build Outside the Workspace
- **Date:** 2026-09-27
- **Component:** `proto` module consumed by `study-timer` and `api-gateway`
- **Symptom:** Importing `github.com/neennera/fishertimer/proto/...` works under `go.work`, but `GOWORK=off go build` (as a deploy target such as Render builds a single service) cannot find the module, because it is not published anywhere.
- **Root Cause:** `go.work` only resolves local modules during workspace builds; a standalone module build resolves imports from its own `go.mod` and the module proxy.
- **Solution:** Each consumer's `go.mod` requires `github.com/neennera/fishertimer/proto v0.0.0` with `replace github.com/neennera/fishertimer/proto => ../../proto`, then `GOWORK=off go mod tidy`.
- **Action for Future Agents:** When a service imports a local workspace module, add the `require` + `replace` pair and verify with `GOWORK=off go build ./...`.

---

### Incident 006: Leaderboard Empty After Database Reset Due to Unseeded `user_rewards`
- **Date:** 2026-09-30
- **Component:** `services/reward`, `services/leaderboard`, MongoDB, Redis Cache
- **Symptom:** After running `pnpm db:reset`, querying `GET /api/v1/leaderboard` returned `"rankings": []` with `"cached": true`.
- **Root Cause:** `002_seed_reward_items.js` seeded only the catalog collection (`reward_items`), leaving `user_rewards` with 0 documents. The Reward service returned an empty reward list to Leaderboard, which cached the empty ranking into Redis.
- **Solution:** Added `services/reward/database/schemas/003_seed_user_rewards.js` containing 41 demo user catches across 5 users (`user1` to `user5`). When cache is stale after seeding, run `docker exec -i fishertimer-redis redis-cli FLUSHALL`.
- **Action for Future Agents:** Whenever normalized 3NF schemas separate catalog items from user interaction records, ensure both collections have corresponding initialization seed scripts in `/docker-entrypoint-initdb.d/`.

---

### Incident 007: Docker Port 5432 Conflict with Host PostgreSQL & Selective Container Starting
- **Date:** 2026-09-30
- **Component:** `docker-compose.yml`, `account-db`
- **Symptom:** Running `docker compose up -d` failed with `listen tcp 0.0.0.0:5432: bind: address already in use`.
- **Root Cause:** A host PostgreSQL daemon (installed via Homebrew or native macOS service) was already actively bound to local port 5432, preventing `account-db` from binding to 5432.
- **Solution:** For targeted service development (e.g. Phase 2 Role B Study Timer), run only non-conflicting containers: `docker compose up -d timer-db rabbitmq adminer` (where `timer-db` is on 5434, RabbitMQ is on 5672/15672, Adminer is on 8080).
- **Action for Future Agents:** If host port conflicts occur during Docker startup, instruct developers or run selective container subsets rather than failing all containers, or remap host binding ports.

---

### Incident 008: At-Least-Once Delivery Idempotency with `processed_events` in Database Transactions
- **Date:** 2026-09-30
- **Component:** `services/study-timer/internal/adapter/amqp/consumer.go`, `timer_db.processed_events`
- **Symptom:** RabbitMQ consumers re-deliver messages upon reconnection or unacknowledged failures, potentially causing duplicate cycle terminations or state corruption.
- **Root Cause:** AMQP guarantees at-least-once message delivery, meaning network blips or pod restarts can cause the same message (`event_id`) to be received multiple times.
- **Solution:** Implemented `processed_events` table (`event_id`, `event_type`, `processed_at`) checked and recorded within the same atomic SQL transaction (`tx`) as timer finalization. If the `event_id` is already present, the transaction rolls back gracefully and the AMQP consumer acks the message without reapplying side effects.
- **Action for Future Agents:** Always combine message deduplication keys in the same database transaction as business state mutations when implementing AMQP consumers.

---

## 2. Template for Recording New Lessons Learned

When documenting a new learning, append to Section 1 using this markdown structure:

```markdown
### Incident <Number>: <Short Descriptive Title>
- **Date:** YYYY-MM-DD
- **Component:** <Affected service, package, or tool>
- **Symptom:** <Exact error message or undesirable behavior>
- **Root Cause:** <Technical reason why the failure occurred>
- **Solution:** <Code or configuration change that fixed the problem>
- **Action for Future Agents:** <Specific rule or heuristic to avoid this in future turns>
```
