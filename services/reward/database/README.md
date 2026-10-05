# Reward Service Database

This directory contains the database configuration, connection management, and migrations for the **Reward Service**.

- **Database Engine:** MongoDB
- **Default Database:** `reward_db`
- **Default Port:** `27017`
- **Environment Variable:** `REWARD_MONGODB_URI`

## Schemas

- [`schemas/001_create_reward_collections.js`](./schemas/001_create_reward_collections.js): Defines MongoDB collections with JSON Schema validation and indexes:
  - `reward_items`: Catalog of unlockable reward items (`SKIN`, `BADGE`, `FISH_SPECIES`).
  - `user_rewards`: User unlocked inventory with unique composite index `(user_id, item_id)`.
- [`schemas/002_seed_reward_items.js`](./schemas/002_seed_reward_items.js): Seeds `reward_items` with one `FISH` item per sprite in `apps/web/public/sprites/fish/`. Idempotent (upserts by `item_name`). Runs automatically on a fresh volume; on an existing one run `docker exec -i fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin < services/reward/database/schemas/002_seed_reward_items.js`.
- [`schemas/003_seed_user_rewards.js`](./schemas/003_seed_user_rewards.js): Seeds 41 user catch records across 5 demo users (`user1` to `user5`) for Leaderboard and FishTank demonstrations. Runs automatically on a fresh volume; on an existing one run `docker exec -i fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin < services/reward/database/schemas/003_seed_user_rewards.js`.
- [`schemas/004_seed_current_user_month.js`](./schemas/004_seed_current_user_month.js): Seeds this month's mock catches for the signed-in account plus 9 mock competitors (`seed-angler-*`). Deletes their existing `user_rewards` first, so it is safe to re-run. Needs `SEED_USER_ID` (and optionally `SEED_DISPLAY_NAME`): `docker exec -i -e SEED_USER_ID=<user_id> -e SEED_DISPLAY_NAME=<name> fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin < services/reward/database/schemas/004_seed_current_user_month.js`. The web app's `/dev` page runs this for you (dev only).
