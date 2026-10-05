# Reward Service Database

This directory contains the database configuration, connection management, and migrations for the **Reward Service**.

- **Database Engine:** MongoDB
- **Default Database:** `reward_db`
- **Default Port:** `27017`
- **Environment Variable:** `REWARD_MONGODB_URI`

## Schemas

- [`schemas/001_create_reward_collections.js`](./schemas/001_create_reward_collections.js): Defines MongoDB collections with JSON Schema validation and indexes:
  - `reward_items`: Catalog of reward items (`category`: `FISH`, `DECORATION`, `ROD`; `rarity`: `COMMON` to `LEGENDARY`), each with a `base_weight` (drop weight) and `score_value` (leaderboard points).
  - `user_rewards`: One document per catch. Indexed on `(user_id, item_id)` and `awarded_at`, neither unique: a user can catch the same item more than once.
- [`schemas/002_seed_reward_items.js`](./schemas/002_seed_reward_items.js): Seeds `reward_items` with one `FISH` item per sprite in `apps/web/public/sprites/fish/`. Base weights (30 / 25 / 12 / 6 / 3) and score values (10 / 25 / 50 / 100 / 250) follow Table 2 of the project description (UC-09). Idempotent (upserts by `item_name`), so re-running it updates the weights of an existing catalogue. Runs automatically on a fresh volume; on an existing one run `docker exec -i fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin < services/reward/database/schemas/002_seed_reward_items.js`.
- [`schemas/003_seed_user_rewards.js`](./schemas/003_seed_user_rewards.js): Seeds 41 user catch records across 5 demo users (`user1` to `user5`) for Leaderboard and FishTank demonstrations. Runs automatically on a fresh volume; on an existing one run `docker exec -i fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin < services/reward/database/schemas/003_seed_user_rewards.js`.
