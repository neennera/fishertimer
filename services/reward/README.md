# Reward Service

## Overview
The **Reward Service** manages gamification reward drops, fish catch inventory progression, and supplies historical reward data to the **Account Service** and **Leaderboard Service**.

It supports dual persistence:
1. **MongoDB** (`reward_db`: `reward_items`, `user_rewards`) with compound indexes for high-throughput queries.
2. **In-Memory Repository** with built-in 5-user relative seed data for instant development without Docker.

---

## Ports & Endpoints

- **Default Port**: `8085`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Service health check. |
| `GET` | `/api/v1/reward/status` | Layer status check. |
| `POST` | `/api/v1/reward/award` | Draws and stores the rewards for one completed work cycle. See below. |
| `GET` | `/api/v1/reward/rewards?user_id={id}` | Rewards for one user, joining `user_rewards` with `reward_items`. |
| `GET` | `/api/v1/reward/all-rewards` | All rewards across all users (consumed by Leaderboard for rankings). |
| `GET` | `/api/v1/reward/last-update` | `{ "reward_last_update": "<RFC3339>" }`, for Leaderboard cache validation. |

### `POST /api/v1/reward/award`

Request:

```json
{
  "user_id": "uuid",
  "session_id": "uuid",
  "cycle_id": "uuid",
  "work_minutes": 60,
  "participant_count": 3,
  "display_name": "optional, stored on each row"
}
```

- `user_id`, `cycle_id` and the cycle length in minutes (>= 0) are required; otherwise `400` with a message listing them.
- The cycle length is `work_minutes` (preferred). `work_duration` (minutes) is accepted for compatibility with Study Timer, which sends that name. Send either one; if both are sent with different values the request is rejected with `400`. Unknown fields, such as Study Timer's `reason`, are ignored. `participant_count` defaults to 1 and is clamped to 1..5. `session_id` is accepted but not stored.
- One reward per full 15 work minutes; under 15 earns nothing and stores nothing. Each reward is drawn by weight from `reward_items`, with the group buff applied to non-COMMON items (`internal/domain/calculation.go`).
- Idempotent per `cycle_id`: a retry, including a concurrent one, stores nothing and returns the rewards already stored, with `already_awarded: true`. Each row's `_id` is derived from `cycle_id` and its draw number, so Mongo's unique `_id` index rejects a duplicate award; no extra index or field is needed.

Response (`201` when rewards were stored, `200` when already awarded or nothing was earned):

```json
{
  "cycle_id": "uuid",
  "already_awarded": false,
  "rewards": [{ "user_reward_id": "…", "item_id": "…", "item_name": "Koi", "rarity": "EPIC", "...": "same fields as /rewards" }]
}
```

---

## Seed Dataset (5 Users, 41 Rewards)

Both the in-memory repository and MongoDB init scripts contain a pre-configured seed dataset with relative timestamps:

| User ID | Display Name | All-Time Total | Period Outcome |
|---------|--------------|----------------|----------------|
| `user1` | TideAngler   | 12 catches     | 🏆 **All-Time Champion** (#1) |
| `user2` | CastMaster   | 8 catches      | Regular active angler |
| `user3` | LureQueen    | 5 catches      | 📅 **Weekly Winner** (#1 on 7-day rolling) |
| `user4` | DeepDiver    | 6 catches      | Steady contender |
| `user5` | ReefRider    | 10 catches     | 🗓️ **Monthly Winner** (#1 on 30-day rolling) |

---

## Fish Score System

Each catch is worth points by rarity tier (`score_value` on `reward_items`, fallback `domain.ScoreForRarity`). The Leaderboard sums these per user.

| Tier | Score |
|------|-------|
| COMMON | 10 |
| UNCOMMON | 25 |
| RARE | 50 |
| EPIC | 100 |
| LEGENDARY | 250 |

## Seeding the signed-in user

`database/schemas/004_seed_current_user_month.js` seeds this month's catches for your real account plus 9 mock competitors (`seed-angler-*`), deleting their old `user_rewards` first. Easiest: open `/dev` in the web app (dev only) and click the button. See [`database/README.md`](./database/README.md) for the manual command.

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
