// =======================================================
// Reward Service Seed: reward_items catalog
// Database Engine: MongoDB
// Database: reward_db
// =======================================================
//
// One FISH item per sprite in apps/web/public/sprites/fish/. asset_url is the
// sprite's path in the web app (apps/web/lib/fish-sprites.ts matches it by
// file name). Upserts by item_name, so re-running it updates rather than
// duplicates. Runs after 001 on a fresh volume (docker-entrypoint-initdb.d);
// on an existing one, run it by hand with mongosh.

db = db.getSiblingDB('reward_db');

// Per rarity: relative drop weight and leaderboard points. These are the base
// weights and reward values of Table 2 in the project description (UC-09);
// internal/domain/calculation_test.go checks the maths against the same numbers.
const TIERS = {
  COMMON: { base_weight: 30, score_value: NumberInt(10) },
  UNCOMMON: { base_weight: 25, score_value: NumberInt(25) },
  RARE: { base_weight: 12, score_value: NumberInt(50) },
  EPIC: { base_weight: 6, score_value: NumberInt(100) },
  LEGENDARY: { base_weight: 3, score_value: NumberInt(250) },
};

const FISH = [
  { item_name: 'Anchovy', rarity: 'COMMON', sprite: 'Anchovy.png' },
  { item_name: 'Goldfish', rarity: 'COMMON', sprite: 'Goldfish.png' },
  { item_name: 'Bass', rarity: 'COMMON', sprite: 'Bass.png' },
  { item_name: 'Catfish', rarity: 'COMMON', sprite: 'Catfish.png' },
  { item_name: 'Clownfish', rarity: 'COMMON', sprite: 'Clownfish.png' },
  { item_name: 'Blue Tang', rarity: 'UNCOMMON', sprite: 'Surgeonfish.png' },
  { item_name: 'Angelfish', rarity: 'UNCOMMON', sprite: 'Angelfish.png' },
  { item_name: 'Rainbow Trout', rarity: 'RARE', sprite: 'Rainbow Trout.png' },
  { item_name: 'Pufferfish', rarity: 'RARE', sprite: 'Pufferfish.png' },
  { item_name: 'Dungeness Crab', rarity: 'RARE', sprite: 'Crab - Dungeness.png' },
  { item_name: 'Koi', rarity: 'EPIC', sprite: 'Koi.png' },
  { item_name: 'Octopus', rarity: 'EPIC', sprite: 'Octopus.png' },
  { item_name: 'Seahorse', rarity: 'EPIC', sprite: 'Seahorse.png' },
  { item_name: 'Globefish', rarity: 'LEGENDARY', sprite: 'Globefish.png' },
  { item_name: 'Ghostfish', rarity: 'LEGENDARY', sprite: 'Ghostfish.png' },
];

FISH.forEach(({ item_name, rarity, sprite }) => {
  db.reward_items.updateOne(
    { item_name },
    {
      $set: {
        item_name,
        category: 'FISH',
        rarity,
        ...TIERS[rarity],
        asset_url: `/sprites/fish/${sprite}`,
      },
    },
    { upsert: true },
  );
});
