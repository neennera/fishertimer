// 02-init-mongo.js
// Combined schema validator & seed execution for Docker container startup

db = db.getSiblingDB("fishertimer");

// Create fish_rewards collection with schema validator
db.createCollection("fish_rewards", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["userId", "species", "rarity", "awardedAt"],
      properties: {
        _id: { bsonType: "objectId" },
        userId: {
          bsonType: "string",
          description: "Owner user ID from auth.users"
        },
        species: {
          bsonType: "string",
          description: "Species name of the fish"
        },
        rarity: {
          enum: ["COMMON", "UNCOMMON", "RARE", "EPIC", "LEGENDARY"],
          description: "Rarity tier"
        },
        weightKg: {
          bsonType: "double",
          description: "Weight score of the fish item"
        },
        awardedAt: {
          bsonType: "date",
          description: "Timestamp when focus cycle completed"
        }
      }
    }
  }
});

db.fish_rewards.createIndex({ userId: 1 });
db.fish_rewards.createIndex({ rarity: 1 });
db.fish_rewards.createIndex({ awardedAt: -1 });

// Seed initial items
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

print("Initialized MongoDB fishertimer database with schema and seed items.");
