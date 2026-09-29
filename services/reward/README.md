# Reward Service

## Overview
The **Reward Service** manages gamification reward drops, fish catch inventory progression, and supplies historical reward data to the **Leaderboard Service**.

It supports dual persistence:
1. **MongoDB** (`reward_db.fish_rewards`) with compound indexes for high-throughput queries.
2. **In-Memory Repository** with built-in 5-user relative seed data for instant development without Docker.

---

## Ports & Endpoints

- **Default Port**: `8085`
- **Internal Routes**:
  - `GET /health` — Service health check.
  - `GET /api/v1/reward/status` — Layer status check.
  - `POST /api/v1/reward/award` — Awards a fish reward to a user (`{ user_id, reason }`).
  - `GET /api/v1/reward/rewards?user_id={id}` — Returns fish rewards for a specific user.
  - `GET /api/v1/reward/all-rewards` — Returns all rewards across all users (consumed by Leaderboard).
  - `GET /api/v1/reward/last-update` — Returns `{ "reward_last_update": "<RFC3339>" }` timestamp for cache validation.

---

## Seed Dataset (5 Users, 41 Rewards)

Both the in-memory repository and MongoDB init script (`001_create_reward_collections.js`) contain a pre-configured seed dataset with relative timestamps:

| User ID | Display Name | All-Time Total | Period Outcome |
|---------|--------------|----------------|----------------|
| `user1` | TideAngler   | 12 catches     | 🏆 **All-Time Champion** (#1) |
| `user2` | CastMaster   | 8 catches      | Regular active angler |
| `user3` | LureQueen    | 5 catches      | 📅 **Weekly Winner** (#1 on 7-day rolling) |
| `user4` | DeepDiver    | 6 catches      | Steady contender |
| `user5` | ReefRider    | 10 catches     | 🗓️ **Monthly Winner** (#1 on 30-day rolling) |

---

## Configuration & Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` or `REWARD_SERVICE_PORT` | HTTP server listening port | `8085` |
| `REWARD_MONGODB_URI` or `MONGODB_URI` | MongoDB connection URI | `mongodb://mongoadmin:mongopassword@localhost:27017/reward_db?authSource=admin` |
| `ENV` | Environment identifier | `development` |

*Note: If MongoDB is unavailable, the service automatically logs a notice and operates with in-memory seeded storage.*

---

## Development & Testing

```powershell
# Run service locally
go run cmd/main.go

# Run unit tests
go test ./... -v

# Static analysis
go vet ./...
```
