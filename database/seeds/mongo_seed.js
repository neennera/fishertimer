// Seed Data: mongo_seed.js
// Mock reward collection data for local development

db = db.getSiblingDB("fishertimer");

const now = new Date();

db.fish_rewards.insertMany([
  {
    userId: "usr_01",
    species: "Golden Salmon",
    rarity: "LEGENDARY",
    weightKg: 8.45,
    awardedAt: new Date(now.getTime() - 3600 * 1000 * 2)
  },
  {
    userId: "usr_01",
    species: "Rainbow Trout",
    rarity: "RARE",
    weightKg: 3.20,
    awardedAt: new Date(now.getTime() - 3600 * 1000 * 5)
  },
  {
    userId: "usr_02",
    species: "River Perch",
    rarity: "COMMON",
    weightKg: 0.85,
    awardedAt: new Date(now.getTime() - 3600 * 1000 * 1)
  },
  {
    userId: "usr_03",
    species: "Giant Bluefin Tuna",
    rarity: "EPIC",
    weightKg: 14.80,
    awardedAt: new Date(now.getTime() - 3600 * 1000 * 12)
  }
]);

print("Seeded fish_rewards collection with initial items.");
