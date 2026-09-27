# Reward Service

## Overview
Gamification reward drop calculation and inventory progression

## Port
Default port: `8085`

## HTTP API

| Method | Route | Purpose |
| :--- | :--- | :--- |
| `GET` | `/api/v1/reward/rewards?user_id=<id>` | Lists every catch/drop for a user, joining `user_rewards` with `reward_items`: `[{user_reward_id, item_id, user_id, cycle_id, item_name, category, rarity, base_weight, score_value, asset_url, awarded_at}, ...]` (empty array if none). `user_rewards` has no uniqueness constraint on `(user_id, item_id)` - a user can catch the same item more than once, and each catch is its own row. This is a flat per-catch log (also read by Leaderboard's `ViewRewards()` collaborator); the Account service's `GET /api/v1/account/rewards?id=<user_id>` is the one that groups it into totals + per-item counts.|


## Environment

| Variable | Purpose |
| :--- | :--- |
| `REWARD_MONGODB_URI` | `reward_db` connection string (default `mongodb://mongoadmin:mongopassword@localhost:27017/reward_db?authSource=admin`). **Required** - the service exits if the database is unreachable. |

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
