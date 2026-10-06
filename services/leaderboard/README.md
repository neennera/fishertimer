# Leaderboard Service

## Overview
The **Leaderboard Service** provides read-only ranking of users based on fish rewards earned during focused study sessions (Use Case **UC-08: View Leaderboard**).

It is a stateless, cache-optimized service designed with Hexagonal / Clean Architecture. It queries the **Reward Service** for data and maintains a high-performance in-memory cache backed by Redis (with automatic fallback to an in-memory repository for local development without Docker).

---

## Ports & Endpoints

- **Default Port**: `8086`
- **Internal Routes**:
  - `GET /health` — Service health check.
  - `GET /api/v1/leaderboard/status` — Architecture status.
  - `GET /api/v1/leaderboard?period={weekly|monthly|all-time}` — Full ranking response with cache metadata (`CachedRanking`).
  - `GET /api/v1/leaderboard/rankings?period={weekly|monthly|all-time}` — Lightweight array of rankings (`[]RankEntry`).

---

## UC-08 Caching & Invalidation Architecture

The service implements the **S-1 Cache Protocol**:

```
Client Request
      │
      ▼
Check Reward Service: GET /api/v1/reward/last-update  ──► reward_last_update
      │
      ▼
Compare with stored cache: leaderboard_last_fetch[period]
      │
   Equal? (leaderboard_last_fetch >= reward_last_update)
  ┌───┴───┐
  │ YES   │ NO (Cache Miss / Stale)
  ▼       ▼
Return  Query Reward Service: GET /api/v1/reward/all-rewards
Cache   Filter by period window (Rolling 7 days / Rolling 30 days / All-Time)
(S-1)   Aggregate per user & Sort descending by count
        Tie-break: earliest reward timestamp (E-4)
        Persist in Redis Cache (with 10-minute TTL safety net)
        Set leaderboard_last_fetch = reward_last_update
        Return fresh ranking
```

---

## Ranking

Users are ranked by total fish score in the period: the sum of each catch's `score_value` (set by rarity tier in the Reward service: 10 / 25 / 50 / 100 / 250). Ties go to the user whose earliest catch in the period is older, then by user id. Each `RankEntry` carries `score` and `reward_count`.

---

## Configuration & Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` or `LEADERBOARD_SERVICE_PORT` | HTTP server listening port | `8086` |
| `LEADERBOARD_REDIS_URL` or `REDIS_URL` | Redis server connection URI | `redis://localhost:6379/0` |
| `REWARD_SERVICE_URL` | Base URL of the Reward Microservice | `http://localhost:8085` |
| `ENV` | Environment identifier | `development` |

*Note: If Redis is unavailable, the service logs a notice and seamlessly falls back to an in-memory cache.*

---

## Development & Testing

```powershell
# Run the service locally
go run cmd/main.go

# Run unit tests (tests S-1 cache hit/miss, weekly/all-time filtering, empty state)
go test ./... -v

# Static analysis
go vet ./...
```
