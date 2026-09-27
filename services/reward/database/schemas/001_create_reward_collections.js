// =======================================================
// Reward Service Database Schema & Indexes (3NF Document Model)
// Database Engine: MongoDB
// Database: reward_db
// =======================================================

db = db.getSiblingDB('reward_db');

// 1. reward_items collection: the catalog of things a user can catch/unlock.
// _id is the natural Mongo primary key (the DBML's item_id).
db.createCollection('reward_items', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['item_name', 'category', 'rarity', 'base_weight', 'score_value', 'asset_url'],
      properties: {
        item_name: {
          bsonType: 'string',
          description: 'must be a string and is required'
        },
        category: {
          enum: ['FISH', 'DECORATION', 'ROD'],
          description: 'can only be one of FISH, DECORATION, ROD and is required'
        },
        rarity: {
          enum: ['COMMON', 'UNCOMMON', 'RARE', 'EPIC', 'LEGENDARY'],
          description: 'can only be one of COMMON, UNCOMMON, RARE, EPIC, LEGENDARY and is required'
        },
        base_weight: {
          bsonType: ['double', 'int'],
          minimum: 0,
          description: 'relative drop-rate weight and is required'
        },
        score_value: {
          bsonType: 'int',
          minimum: 0,
          description: 'points this item contributes to a leaderboard score and is required'
        },
        asset_url: {
          bsonType: 'string',
          description: 'URL of the item\'s sprite/icon asset and is required'
        }
      }
    }
  }
});

// 2. user_rewards collection: one document per catch/drop. No uniqueness on
// (user_id, item_id) - a user can catch the same item more than once.
db.createCollection('user_rewards', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['user_id', 'cycle_id', 'item_id', 'awarded_at'],
      properties: {
        user_id: {
          bsonType: 'string',
          description: 'UUID of user referring to account_db.users(user_id) and is required'
        },
        cycle_id: {
          bsonType: 'string',
          description: 'UUID referring to timer_db.timer_cycles(cycle_id) (WORK cycle only) and is required'
        },
        item_id: {
          bsonType: 'objectId',
          description: 'ObjectId referring to reward_items._id and is required'
        },
        awarded_at: {
          bsonType: 'date',
          description: 'reward drop timestamp and is required; also used by Leaderboard Cache to detect updates'
        }
      }
    }
  }
});

db.user_rewards.createIndex({ user_id: 1, item_id: 1 });
db.user_rewards.createIndex({ awarded_at: 1 });
