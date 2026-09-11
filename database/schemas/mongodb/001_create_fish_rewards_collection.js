// Schema & Validator: 001_create_fish_rewards_collection.js
// Service Owner: Reward Service
// Description: MongoDB collection definition with JSON schema validation for gamification items.

db = db.getSiblingDB("fishertimer");

db.createCollection("fish_rewards", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["userId", "species", "rarity", "awardedAt"],
      properties: {
        _id: { bsonType: "objectId" },
        userId: {
          bsonType: "string",
          description: "Owner user ID from auth.users (string reference)"
        },
        species: {
          bsonType: "string",
          description: "Species name of the caught fish (e.g. Golden Salmon, Blue Fin Tuna)"
        },
        rarity: {
          enum: ["COMMON", "UNCOMMON", "RARE", "EPIC", "LEGENDARY"],
          description: "Rarity tier buffed by community presence"
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

// Indexes for fast lookup
db.fish_rewards.createIndex({ userId: 1 });
db.fish_rewards.createIndex({ rarity: 1 });
db.fish_rewards.createIndex({ awardedAt: -1 });

print("Created collection: fish_rewards with schema validator and indexes.");
