// =======================================================
// Reward Service Database Schema & Indexes (3NF Document Model)
// Database Engine: MongoDB
// Database: reward_db
// =======================================================

db = db.getSiblingDB('reward_db');

// 1. reward_items collection
db.createCollection('reward_items', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['item_name', 'item_type', 'cost', 'created_at'],
      properties: {
        item_name: {
          bsonType: 'string',
          description: 'must be a string and is required'
        },
        description: {
          bsonType: 'string',
          description: 'optional description text'
        },
        item_type: {
          enum: ['SKIN', 'BADGE', 'FISH_SPECIES'],
          description: 'can only be one of SKIN, BADGE, FISH_SPECIES and is required'
        },
        cost: {
          bsonType: 'int',
          minimum: 0,
          description: 'must be an integer >= 0 and is required'
        },
        created_at: {
          bsonType: 'date',
          description: 'must be a date and is required'
        }
      }
    }
  }
});

// 2. user_rewards collection
db.createCollection('user_rewards', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['user_id', 'item_id', 'unlocked_at'],
      properties: {
        user_id: {
          bsonType: 'string',
          description: 'UUID of user referring to account_db.users(user_id) and is required'
        },
        item_id: {
          bsonType: 'objectId',
          description: 'ObjectId referring to reward_items._id and is required'
        },
        unlocked_at: {
          bsonType: 'date',
          description: 'must be a date and is required'
        }
      }
    }
  }
});

// Unique composite index: (user_id, item_id)
db.user_rewards.createIndex({ user_id: 1, item_id: 1 }, { unique: true });
db.user_rewards.createIndex({ user_id: 1 });
