// =======================================================
// Reward Service Seed: this month's mock fish for the signed-in user
// Database Engine: MongoDB
// Database: reward_db
// =======================================================
//
// Gives the REAL signed-in account (from the auth session) a month of mock
// catches, plus a handful of mock competitors so the leaderboard has several
// pages. Scores come from the rarity tier (reward_items.score_value), so the
// ranking is the same one the Leaderboard service computes.
//
// Re-seeding is safe: every user_rewards row for the target user and for the
// mock competitors (user_id "seed-angler-*") is deleted first.
//
// Usage (the user id is the account's user_id, e.g. from GET /api/auth/me):
//
//   docker exec -i -e SEED_USER_ID=<user_id> -e SEED_DISPLAY_NAME="<name>" \
//     fishertimer-reward-db mongosh -u mongoadmin -p mongopassword \
//     --authenticationDatabase admin < services/reward/database/schemas/004_seed_current_user_month.js
//
// Needs 002_seed_reward_items.js to have run (it picks fish by rarity).

db = db.getSiblingDB('reward_db');

const USER_ID = process.env.SEED_USER_ID;
const DISPLAY_NAME = process.env.SEED_DISPLAY_NAME || 'You';

if (!USER_ID) {
  throw new Error('Set SEED_USER_ID to the signed-in account user_id (see header comment).');
}

// item ids by rarity tier
const byRarity = {};
db.reward_items.find({ category: 'FISH' }).forEach((item) => {
  (byRarity[item.rarity] = byRarity[item.rarity] || []).push(item._id);
});
if (Object.keys(byRarity).length === 0) {
  throw new Error('reward_items is empty - run 002_seed_reward_items.js first.');
}

const now = new Date();
const monthStart = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 1));
const spanMs = Math.max(now.getTime() - monthStart.getTime(), 3600000);

// Deterministic spread of timestamps across this month so that re-seeding
// gives stable orderings: the i-th of n catches lands at (i+1)/(n+1) of the span.
const at = (i, n) => new Date(monthStart.getTime() + Math.floor((spanMs * (i + 1)) / (n + 1)));

// A catch list is an array of rarity tiers.
const C = 'COMMON', U = 'UNCOMMON', R = 'RARE', E = 'EPIC', L = 'LEGENDARY';

const PLAYERS = [
  // Mock competitors. Scores: C=10 U=25 R=50 E=100 L=250
  { user_id: 'seed-angler-1', display_name: 'TideAngler', catches: [L, E, E, R, R, U, C, C] }, // 595
  { user_id: 'seed-angler-2', display_name: 'ReefRider', catches: [L, E, R, R, U, U, C] }, // 510
  { user_id: 'seed-angler-3', display_name: 'CastMaster', catches: [E, E, R, R, U, C, C, C] }, // 355
  { user_id: 'seed-angler-4', display_name: 'DeepDiver', catches: [E, R, R, U, U, C, C] }, // 270
  { user_id: 'seed-angler-5', display_name: 'KoiKeeper', catches: [E, R, U, U, C, C] }, // 220
  { user_id: 'seed-angler-6', display_name: 'NetWeaver', catches: [R, R, U, C, C, C] }, // 155
  { user_id: 'seed-angler-7', display_name: 'BaitBoss', catches: [R, U, C, C, C] }, // 105
  { user_id: 'seed-angler-8', display_name: 'PondPilot', catches: [U, C, C, C] }, // 55
  { user_id: 'seed-angler-9', display_name: 'BobberBuddy', catches: [C, C, C] }, // 30
  // The signed-in user lands mid-table so the pinned "your ranking" row and
  // pagination are both visible.
  { user_id: USER_ID, display_name: DISPLAY_NAME, catches: [E, R, R, U, C, C, C] }, // 255
];

// Wipe old reward sessions for everyone we are about to (re)seed.
const ids = PLAYERS.map((p) => p.user_id);
const removed = db.user_rewards.deleteMany({
  $or: [{ user_id: { $in: ids } }, { user_id: /^seed-angler-/ }],
}).deletedCount;

const docs = [];
PLAYERS.forEach((p) => {
  p.catches.forEach((rarity, i) => {
    const pool = byRarity[rarity] || byRarity[C];
    docs.push({
      user_id: p.user_id,
      display_name: p.display_name,
      cycle_id: `seed-${p.user_id}-${i + 1}`,
      item_id: pool[i % pool.length],
      awarded_at: at(i, p.catches.length),
    });
  });
});

db.user_rewards.insertMany(docs);
print(
  `reward_db user_rewards re-seeded for ${USER_ID} (${DISPLAY_NAME}): removed ${removed}, inserted ${docs.length} catches across ${PLAYERS.length} anglers for ${monthStart.toISOString().slice(0, 7)}`,
);
