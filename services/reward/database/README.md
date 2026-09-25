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
