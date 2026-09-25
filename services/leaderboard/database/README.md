# Leaderboard Service Cache & Datastore

This directory documents the caching and datastore architecture for the **Leaderboard Service**.

## Architectural Decision: In-Memory Redis Cache Only

As defined in the v2 microservices architecture:
- The Leaderboard Service **does not** maintain a dedicated SQL or MongoDB database.
- It operates an in-memory **Redis Cache** (`redis://localhost:6379`) utilizing Redis Sorted Sets (`ZSET`).
- Rankings and scores are computed dynamically by querying the **Reward Service** (`ViewRewards`) and **Study Timer Service**.

- **Engine:** Redis 7
- **Default Port:** `6379`
- **Environment Variable:** `LEADERBOARD_REDIS_URL`
