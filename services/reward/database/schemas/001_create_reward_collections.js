// =======================================================
// Reward Service Database Schema & Indexes
// Database Engine: MongoDB
// Database: reward_db
// =======================================================
// This file is run automatically by the MongoDB container
// on first startup via /docker-entrypoint-initdb.d/
// =======================================================

db = db.getSiblingDB('reward_db');

// ── 1. reward_items: catalog of unlockable item types ──────────────────────
// Used for the shop / item unlock flow. NOT the same as fish_rewards.
db.createCollection('reward_items', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['item_name', 'item_type', 'cost', 'created_at'],
      properties: {
        item_name:   { bsonType: 'string' },
        description: { bsonType: 'string' },
        item_type:   { enum: ['SKIN', 'BADGE', 'FISH_SPECIES'] },
        cost:        { bsonType: 'int', minimum: 0 },
        created_at:  { bsonType: 'date' }
      }
    }
  }
});

// ── 2. fish_rewards: earned fish per user (one doc per award event) ─────────
// This is the main collection the Leaderboard reads for ranking.
db.createCollection('fish_rewards', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['user_id', 'display_name', 'species', 'rarity', 'awarded_at'],
      properties: {
        user_id:      { bsonType: 'string', description: 'references account_db.users.user_id' },
        display_name: { bsonType: 'string' },
        species:      { bsonType: 'string' },
        rarity:       { enum: ['COMMON', 'UNCOMMON', 'RARE', 'EPIC', 'LEGENDARY'] },
        awarded_at:   { bsonType: 'date' }
      }
    }
  }
});

// Indexes on fish_rewards
db.fish_rewards.createIndex({ user_id: 1 });            // fast per-user lookup
db.fish_rewards.createIndex({ awarded_at: -1 });        // fast period-range scans
db.fish_rewards.createIndex({ user_id: 1, awarded_at: -1 }); // compound: user + period

// ── 3. user_rewards: legacy item-unlock join (kept for shop UC) ─────────────
db.createCollection('user_rewards', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['user_id', 'item_id', 'unlocked_at'],
      properties: {
        user_id:     { bsonType: 'string' },
        item_id:     { bsonType: 'objectId', description: 'refs reward_items._id' },
        unlocked_at: { bsonType: 'date' }
      }
    }
  }
});
db.user_rewards.createIndex({ user_id: 1, item_id: 1 }, { unique: true });

// ── 4. Seed data: 5 users with varied rewards for leaderboard demo ──────────
//
// Period winners (as of seed time):
//   Weekly  → user3 / LureQueen  (5 rewards this week)
//   Monthly → user5 / ReefRider  (8 rewards this month)
//   All-Time → user1 / TideAngler (12 rewards total)
//
var now = new Date();
var daysAgo = function(n) { return new Date(now.getTime() - n * 86400000); };

db.fish_rewards.insertMany([
  // ── user1 / TideAngler: 12 rewards over ~90 days (all-time #1) ──
  { user_id:'user1', display_name:'TideAngler', species:'Legendary Koi',      rarity:'LEGENDARY', awarded_at: daysAgo(90) },
  { user_id:'user1', display_name:'TideAngler', species:'Silver Bass',        rarity:'UNCOMMON',  awarded_at: daysAgo(80) },
  { user_id:'user1', display_name:'TideAngler', species:'Rainbow Trout',      rarity:'RARE',      awarded_at: daysAgo(70) },
  { user_id:'user1', display_name:'TideAngler', species:'Golden Carp',        rarity:'RARE',      awarded_at: daysAgo(60) },
  { user_id:'user1', display_name:'TideAngler', species:'Phantom Eel',        rarity:'EPIC',      awarded_at: daysAgo(50) },
  { user_id:'user1', display_name:'TideAngler', species:'Silver Bass',        rarity:'UNCOMMON',  awarded_at: daysAgo(42) },
  { user_id:'user1', display_name:'TideAngler', species:'Sunfish',            rarity:'COMMON',    awarded_at: daysAgo(35) },
  { user_id:'user1', display_name:'TideAngler', species:'River Perch',        rarity:'COMMON',    awarded_at: daysAgo(28) },
  { user_id:'user1', display_name:'TideAngler', species:'Blue Tuna',          rarity:'EPIC',      awarded_at: daysAgo(21) },
  { user_id:'user1', display_name:'TideAngler', species:'Glowing Jellyfish',  rarity:'RARE',      awarded_at: daysAgo(14) },
  { user_id:'user1', display_name:'TideAngler', species:'Sunfish',            rarity:'COMMON',    awarded_at: daysAgo(10) },
  { user_id:'user1', display_name:'TideAngler', species:'River Perch',        rarity:'COMMON',    awarded_at: daysAgo(9)  },

  // ── user2 / CastMaster: 8 rewards; 2 this week ──
  { user_id:'user2', display_name:'CastMaster', species:'Golden Salmon',      rarity:'RARE',      awarded_at: daysAgo(75) },
  { user_id:'user2', display_name:'CastMaster', species:'Phantom Eel',        rarity:'EPIC',      awarded_at: daysAgo(55) },
  { user_id:'user2', display_name:'CastMaster', species:'River Perch',        rarity:'COMMON',    awarded_at: daysAgo(40) },
  { user_id:'user2', display_name:'CastMaster', species:'Sunfish',            rarity:'COMMON',    awarded_at: daysAgo(30) },
  { user_id:'user2', display_name:'CastMaster', species:'Rainbow Trout',      rarity:'RARE',      awarded_at: daysAgo(22) },
  { user_id:'user2', display_name:'CastMaster', species:'Silver Bass',        rarity:'UNCOMMON',  awarded_at: daysAgo(15) },
  { user_id:'user2', display_name:'CastMaster', species:'Blue Tuna',          rarity:'EPIC',      awarded_at: daysAgo(3)  }, // this week
  { user_id:'user2', display_name:'CastMaster', species:'River Perch',        rarity:'COMMON',    awarded_at: daysAgo(1)  }, // this week

  // ── user3 / LureQueen: 5 rewards ALL this week → weekly #1 ──
  { user_id:'user3', display_name:'LureQueen',  species:'Glowing Jellyfish',  rarity:'RARE',      awarded_at: daysAgo(6)  },
  { user_id:'user3', display_name:'LureQueen',  species:'Rainbow Trout',      rarity:'RARE',      awarded_at: daysAgo(5)  },
  { user_id:'user3', display_name:'LureQueen',  species:'Golden Carp',        rarity:'RARE',      awarded_at: daysAgo(4)  },
  { user_id:'user3', display_name:'LureQueen',  species:'Blue Tuna',          rarity:'EPIC',      awarded_at: daysAgo(2)  },
  { user_id:'user3', display_name:'LureQueen',  species:'Phantom Eel',        rarity:'EPIC',      awarded_at: daysAgo(1)  },

  // ── user4 / DeepDiver: 6 rewards; 3 this month ──
  { user_id:'user4', display_name:'DeepDiver',  species:'Sunfish',            rarity:'COMMON',    awarded_at: daysAgo(85) },
  { user_id:'user4', display_name:'DeepDiver',  species:'Silver Bass',        rarity:'UNCOMMON',  awarded_at: daysAgo(65) },
  { user_id:'user4', display_name:'DeepDiver',  species:'River Perch',        rarity:'COMMON',    awarded_at: daysAgo(45) },
  { user_id:'user4', display_name:'DeepDiver',  species:'Golden Salmon',      rarity:'RARE',      awarded_at: daysAgo(20) }, // this month
  { user_id:'user4', display_name:'DeepDiver',  species:'Rainbow Trout',      rarity:'RARE',      awarded_at: daysAgo(12) }, // this month
  { user_id:'user4', display_name:'DeepDiver',  species:'Sunfish',            rarity:'COMMON',    awarded_at: daysAgo(8)  }, // this month

  // ── user5 / ReefRider: 10 rewards; 8 this month → monthly #1 ──
  { user_id:'user5', display_name:'ReefRider',  species:'Silver Bass',        rarity:'UNCOMMON',  awarded_at: daysAgo(95) },
  { user_id:'user5', display_name:'ReefRider',  species:'Sunfish',            rarity:'COMMON',    awarded_at: daysAgo(68) },
  { user_id:'user5', display_name:'ReefRider',  species:'Golden Carp',        rarity:'RARE',      awarded_at: daysAgo(28) }, // this month
  { user_id:'user5', display_name:'ReefRider',  species:'Blue Tuna',          rarity:'EPIC',      awarded_at: daysAgo(26) }, // this month
  { user_id:'user5', display_name:'ReefRider',  species:'Phantom Eel',        rarity:'EPIC',      awarded_at: daysAgo(24) }, // this month
  { user_id:'user5', display_name:'ReefRider',  species:'Rainbow Trout',      rarity:'RARE',      awarded_at: daysAgo(22) }, // this month
  { user_id:'user5', display_name:'ReefRider',  species:'Glowing Jellyfish',  rarity:'RARE',      awarded_at: daysAgo(18) }, // this month
  { user_id:'user5', display_name:'ReefRider',  species:'Golden Salmon',      rarity:'RARE',      awarded_at: daysAgo(15) }, // this month
  { user_id:'user5', display_name:'ReefRider',  species:'Legendary Koi',      rarity:'LEGENDARY', awarded_at: daysAgo(10) }, // this month
  { user_id:'user5', display_name:'ReefRider',  species:'Silver Bass',        rarity:'UNCOMMON',  awarded_at: daysAgo(7)  }, // this month
]);

print('✅ reward_db seeded: fish_rewards (' + db.fish_rewards.countDocuments() + ' docs)');
