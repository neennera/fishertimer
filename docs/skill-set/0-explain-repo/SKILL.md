---
name: explain-repo
description: Explains the Fisher Timer codebase, architecture, microservice boundaries, technology stack, and directory layout for AI agents and developers.
---

# Skill 0: Explain This Repo

## Objective
Enable any AI agent to immediately comprehend the purpose, architecture, tech stack, and conventions of the Fisher Timer codebase without having to parse thousands of lines of code.

---

## 1. Project Mission & Problem Statement
Fisher Timer is a community-based study timer platform with a gamified fishing theme ("FishTank"). It solves the isolation and lack of motivation learners face when studying alone by providing:
1. **Synchronized Study Rooms:** Users join virtual rooms with peer presence while retaining control over their independent work/break timers.
2. **Gamification & Rewards:** Completing focus cycles awards fish items with varying rarity, buffed by active room presence.
3. **Dashboards & Leaderboards:** Personal statistics, streaks, and global rankings.
4. **Real-Time Moderation:** Live monitoring of rooms and moderation actions (kick/ban).

---

## 2. Technology Stack & Key Decisions (ADRs)

| Component | Technology | Rationale / ADR Reference |
| :--- | :--- | :--- |
| **Monorepo** | Turborepo + pnpm | High concurrency build cache, polyglot task orchestration. |
| **Frontend** | Next.js 16 (React 19) + TailwindCSS | Fast SSR/SEO, component modularity (ADR-001, ADR-002). |
| **Backend** | Go (Golang) 1.22+ | Extreme concurrency, low memory footprint, Clean Architecture (ADR-003). |
| **Database** | Supabase (PostgreSQL) + MongoDB | Polyglot persistence: Relational/WebSockets in Supabase, flexible JSON in MongoDB (ADR-004). |
| **Auth** | Google OAuth + JWT | Low friction, frictionless student sign-on (ADR-005). |
| **Deployment**| Render (Dockerized) | Microservice deployment with low DevOps overhead (ADR-006). |

---

## 3. Directory Layout & Mental Map

- **`apps/web`**: The user-facing web application. Feature-sliced structure (`features/timer`, `features/session`, etc.).
- **`services/*`**: 6 independent Go microservices:
  - `account` (8082): Google OAuth authentication, user profiles & statistics.
  - `study-session` (8083): Rooms & capacity limits (gRPC/HTTP).
  - `study-timer` (8084): Work/break interval execution, reward triggering (gRPC/HTTP).
  - `reward` (8085): Fish drops & rarity progression.
  - `leaderboard` (8086): Read-optimized rankings cached in Redis.
  - `admin` (8087): Moderation & live room inspection.
- **`packages/*`**: Shared libraries:
  - `shared-types`: Common TypeScript interfaces and models.
  - `ui`: Shared design system components.
  - `typescript-config`: Centralized tsconfigs.
  - `eslint-config`: Centralized linter rules.
- **`docs/`**: Project documentation, Phase 1 deliverables, system architecture, changelog, and way-of-work.

---

## 4. Key Agent Commands

```bash
# Build entire repo (all services + web)
pnpm build

# Run everything in dev mode
pnpm dev

# Run only one service
pnpm turbo dev --filter=@fishertimer/study-session-service

# Run all unit tests
pnpm test
```
