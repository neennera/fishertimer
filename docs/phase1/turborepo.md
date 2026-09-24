# Turborepo Monorepo Architecture & Setup Guide (Phase 1)

## 1. Overview & Architectural Context

Fisher Timer is designed with a microservice backend and a unified web frontend to support high-concurrency study sessions, real-time timer synchronization, and gamification mechanics.

To orchestrate the build, development, and testing workflows across multiple services without creating fragmented repositories, **Turborepo** was selected as the monorepo build system combined with **pnpm workspaces** and **Go multi-module workspaces (`go.work`)**.

This setup directly aligns with the decisions documented in the Architectural Decision Records (ADRs):
- **ADR-001 (Frontend Framework):** Next.js (React) application under `apps/web`.
- **ADR-002 (CSS/UI Framework):** TailwindCSS and reusable UI packages under `packages/ui`.
- **ADR-003 (Backend Framework):** Go (Golang) microservices under `services/*`.
- **ADR-006 (Deployment Infrastructure):** Unified containerization and deployment on Render with independent service builds.

---

## 2. Workspace Topology & Directory Structure

The repository is organized into three primary workspace scopes: `apps/*`, `services/*`, and `packages/*`.

```
fishertimer/
├── apps/
│   └── web/                         # Next.js 16 (React 19) web application (Port 3000)
├── services/                        # Go microservices (scaffolded via automation script)
│   ├── auth/                        # Google OAuth & Identity Service (Port 8081)
│   ├── account/                     # User Profiles & Ban Status Service (Port 8082)
│   ├── study-session/               # Study Room Lifecycle & Rosters (Port 8083)
│   ├── study-timer/                 # Independent Work/Rest Timers (Port 8084)
│   ├── reward/                      # FishTank Gamification & Progression (Port 8085)
│   ├── leaderboard/                 # High-performance Rankings Cache (Port 8086)
│   └── admin/                       # Session Monitoring & Moderation (Port 8087)
├── packages/                        # Shared libraries and configurations
│   ├── eslint-config/               # Shared ESLint rule configurations
│   ├── typescript-config/           # Shared tsconfig definitions (base, nextjs, react)
│   ├── ui/                          # Shared React UI component library
│   └── shared-types/                # Shared DTOs, domain models, and API contracts
├── docs/
│   └── phase1/
│       ├── microservice.md          # Service operations & collaborator contracts
│       ├── project-desc.md          # Requirements, use cases & ADRs
│       └── turborepo.md             # This monorepo architecture and setup documentation
├── scripts/
│   └── setup-turborepo.mjs          # Reproducible monorepo scaffolding script
├── go.work                          # Go 1.22+ workspace coordinating all Go modules
├── package.json                     # Root monorepo scripts and Turbo orchestrator
├── pnpm-workspace.yaml              # pnpm workspace definition
└── turbo.json                       # Turborepo task pipeline and caching rules
```

---

## 3. Microservices Inventory & Specifications

The microservices are scaffolded strictly based on the specifications in `microservice.md`:

| Service | Package Name | Default Port | Primary Responsibilities | Collaborating Services |
| :--- | :--- | :--- | :--- | :--- |
| **Account** | `@fishertimer/account-service` | `8082` | Google OAuth authentication, user profiles, personal stats dashboard. | Study Timer, Reward |
| **Study Session** | `@fishertimer/study-session-service` | `8083` | Room creation, join/leave lifecycle, room capacity enforcement (gRPC/HTTP). | None |
| **Study Timer** | `@fishertimer/study-timer-service` | `8084` | Independent user timers, work/break cycle calculations, complete cycle reward trigger (gRPC/HTTP). | Reward |
| **Reward** | `@fishertimer/reward-service` | `8085` | Drop rate calculations, rarity table buffed by participants, user inventory. | None |
| **Leaderboard** | `@fishertimer/leaderboard-service` | `8086` | Read-optimized rankings cached in Redis, reads from Reward Service. | Reward |
| **Admin** | `@fishertimer/admin-service` | `8087` | Live session monitoring, active room oversight, kicking participants and closing rooms. | Study Session |
| **Web** | `web` | `3000` | Next.js responsive frontend client and real-time dashboard. | All Backend Services |

---

## 4. Pipeline & Task Configuration (`turbo.json`)

Turborepo manages task execution order, parallel execution, and computation caching across both TypeScript/Node packages and Go services.

```json
{
  "$schema": "https://turbo.build/schema.json",
  "ui": "tui",
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
  ],
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["$TURBO_DEFAULT$", ".env*"],
      "outputs": [".next/**", "!.next/cache/**", "dist/**", "bin/**"]
    },
    "lint": {
      "dependsOn": ["^lint"]
    },
    "check-types": {
      "dependsOn": ["^check-types"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "test": {
      "dependsOn": ["^build"],
      "inputs": ["$TURBO_DEFAULT$"]
    },
    "clean": {
      "cache": false
    }
  }
}
```

### Key Configurations:
1. **`globalPassThroughEnv`**: Allows Go to access native environment variables (`GOCACHE`, `LOCALAPPDATA`, `PATH`) so that `go build` and `go test` can execute seamlessly on Windows and POSIX environments.
2. **`build`**: Caches built artifacts (`bin/**` for Go binaries, `.next/**` for Next.js build output). Re-running builds on unchanged services results in instantaneous cache hits (`FULL TURBO`).
3. **`dev`**: Configured as persistent and non-cached, allowing parallel live-reloading across the frontend and backend microservices.
4. **`test`**: Automatically runs unit tests across all Go services (`go test ./...`) and frontend packages.

---

## 5. Automated Setup & Reproducibility Script

The monorepo structure and microservice skeletons were generated via an automated script (`scripts/setup-turborepo.mjs`), adhering to the requirement of scripted initialization without manually writing application logic.

### Running the Setup Script
To reproduce or re-scaffold the monorepo setup:

```bash
# Run the scaffolding script
node scripts/setup-turborepo.mjs

# Install dependencies and link all workspaces
pnpm install
```

### What the Script Automates:
1. **Initializes Workspace Scopes**: Sets up `pnpm-workspace.yaml` covering `apps/*`, `services/*`, and `packages/*`.
2. **Generates Service Packages**: Creates standard `package.json` for each service defining `dev`, `build`, `test`, `lint`, and `clean` commands.
3. **Creates Go Modules**: Initializes `go.mod` with appropriate module namespaces (`github.com/neennera/fishertimer/services/<name>`).
4. **Generates Service Entrypoints**: Creates minimal HTTP boilerplate servers (`cmd/main.go`) with `/health` endpoints and graceful shutdown handling.
5. **Configures Go Workspaces (`go.work`)**: Unifies all 7 Go services under a single workspace for multi-module IDE navigation and local dependencies.
6. **Creates `@fishertimer/shared-types`**: Exposes common TypeScript interface models (User, StudySession, TimerSession, Reward, Leaderboard) for frontend-backend contracts.

---

## 6. Common Developer Workflows

### Building the Entire Repository
```bash
pnpm build
```
Executes Next.js static and server compilation alongside Go binary builds in parallel with intelligent caching.

### Running a Specific Service
To run or build an individual service using Turborepo filters:

```bash
# Run only the Account service in dev mode
pnpm turbo dev --filter=@fishertimer/account-service

# Run only the Web frontend
pnpm turbo dev --filter=web

# Build only Go microservices
pnpm turbo build --filter=@fishertimer/*-service
```

### Running Tests
```bash
# Run all tests across all services and packages
pnpm test

# Run tests for a specific service
pnpm turbo test --filter=@fishertimer/study-timer-service
```

### Linting and Type Checking
```bash
# Lint all workspaces
pnpm lint

# Typecheck frontend and shared packages
pnpm check-types
```

---

## 7. Verification & Cache Validation

The setup has been verified and validated:
- `pnpm turbo build`: All 7 Go microservices (`bin/server`) and Next.js production build (`.next/`) successfully compile in parallel.
- **Cache Hit Verification**: A second run of `pnpm turbo build` completes in `< 200ms` with `FULL TURBO` cache hits.
- `pnpm turbo test`: Successfully traverses all 12 workspace projects and executes tests.
- Multi-module Go resolution verified via root `go.work`.
