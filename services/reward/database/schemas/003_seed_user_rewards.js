// =======================================================
// Reward Service Seed: user_rewards for Leaderboard & FishTank demo
// Database Engine: MongoDB
// Database: reward_db
// =======================================================

db = db.getSiblingDB('reward_db');

const items = {};
db.reward_items.find().forEach((item) => {
  items[item.item_name] = item._id;
});

const defaultItemId = db.reward_items.findOne()?._id;

const now = new Date();
const daysAgo = (n) => new Date(now.getTime() - n * 86400000);

const CATCHES = [
  // ── user1: 12 rewards (All-time #1) ──
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Koi', cycle_id: 'c-u1-01', awarded_at: daysAgo(90) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Bass', cycle_id: 'c-u1-02', awarded_at: daysAgo(80) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Rainbow Trout', cycle_id: 'c-u1-03', awarded_at: daysAgo(70) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Goldfish', cycle_id: 'c-u1-04', awarded_at: daysAgo(60) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Ghostfish', cycle_id: 'c-u1-05', awarded_at: daysAgo(50) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Bass', cycle_id: 'c-u1-06', awarded_at: daysAgo(42) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Anchovy', cycle_id: 'c-u1-07', awarded_at: daysAgo(35) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Catfish', cycle_id: 'c-u1-08', awarded_at: daysAgo(28) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Blue Tang', cycle_id: 'c-u1-09', awarded_at: daysAgo(21) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Angelfish', cycle_id: 'c-u1-10', awarded_at: daysAgo(14) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Clownfish', cycle_id: 'c-u1-11', awarded_at: daysAgo(10) },
  { user_id: 'user1', display_name: 'TideAngler', item_name: 'Catfish', cycle_id: 'c-u1-12', awarded_at: daysAgo(9) },

  // ── user2: 8 rewards; 2 this week ──
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Goldfish', cycle_id: 'c-u2-01', awarded_at: daysAgo(75) },
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Octopus', cycle_id: 'c-u2-02', awarded_at: daysAgo(55) },
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Catfish', cycle_id: 'c-u2-03', awarded_at: daysAgo(40) },
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Anchovy', cycle_id: 'c-u2-04', awarded_at: daysAgo(30) },
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Rainbow Trout', cycle_id: 'c-u2-05', awarded_at: daysAgo(22) },
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Bass', cycle_id: 'c-u2-06', awarded_at: daysAgo(15) },
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Seahorse', cycle_id: 'c-u2-07', awarded_at: daysAgo(3) },
  { user_id: 'user2', display_name: 'CastMaster', item_name: 'Catfish', cycle_id: 'c-u2-08', awarded_at: daysAgo(1) },

  // ── user3: 5 rewards ALL this week → weekly #1 ──
  { user_id: 'user3', display_name: 'LureQueen', item_name: 'Angelfish', cycle_id: 'c-u3-01', awarded_at: daysAgo(6) },
  { user_id: 'user3', display_name: 'LureQueen', item_name: 'Rainbow Trout', cycle_id: 'c-u3-02', awarded_at: daysAgo(5) },
  { user_id: 'user3', display_name: 'LureQueen', item_name: 'Goldfish', cycle_id: 'c-u3-03', awarded_at: daysAgo(4) },
  { user_id: 'user3', display_name: 'LureQueen', item_name: 'Blue Tang', cycle_id: 'c-u3-04', awarded_at: daysAgo(2) },
  { user_id: 'user3', display_name: 'LureQueen', item_name: 'Ghostfish', cycle_id: 'c-u3-05', awarded_at: daysAgo(1) },

  // ── user4: 6 rewards; 3 this month ──
  { user_id: 'user4', display_name: 'DeepDiver', item_name: 'Anchovy', cycle_id: 'c-u4-01', awarded_at: daysAgo(85) },
  { user_id: 'user4', display_name: 'DeepDiver', item_name: 'Bass', cycle_id: 'c-u4-02', awarded_at: daysAgo(65) },
  { user_id: 'user4', display_name: 'DeepDiver', item_name: 'Catfish', cycle_id: 'c-u4-03', awarded_at: daysAgo(45) },
  { user_id: 'user4', display_name: 'DeepDiver', item_name: 'Goldfish', cycle_id: 'c-u4-04', awarded_at: daysAgo(20) },
  { user_id: 'user4', display_name: 'DeepDiver', item_name: 'Rainbow Trout', cycle_id: 'c-u4-05', awarded_at: daysAgo(12) },
  { user_id: 'user4', display_name: 'DeepDiver', item_name: 'Anchovy', cycle_id: 'c-u4-06', awarded_at: daysAgo(8) },

  // ── user5: 10 rewards; 8 this month → monthly #1 ──
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Bass', cycle_id: 'c-u5-01', awarded_at: daysAgo(95) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Anchovy', cycle_id: 'c-u5-02', awarded_at: daysAgo(68) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Goldfish', cycle_id: 'c-u5-03', awarded_at: daysAgo(28) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Blue Tang', cycle_id: 'c-u5-04', awarded_at: daysAgo(26) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Ghostfish', cycle_id: 'c-u5-05', awarded_at: daysAgo(24) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Rainbow Trout', cycle_id: 'c-u5-06', awarded_at: daysAgo(22) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Angelfish', cycle_id: 'c-u5-07', awarded_at: daysAgo(18) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Goldfish', cycle_id: 'c-u5-08', awarded_at: daysAgo(15) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Koi', cycle_id: 'c-u5-09', awarded_at: daysAgo(10) },
  { user_id: 'user5', display_name: 'ReefRider', item_name: 'Bass', cycle_id: 'c-u5-10', awarded_at: daysAgo(7) },
];

if (db.user_rewards.countDocuments() === 0) {
  const docs = CATCHES.map((c) => ({
    user_id: c.user_id,
    display_name: c.display_name,
    cycle_id: c.cycle_id,
    item_id: items[c.item_name] || defaultItemId,
    awarded_at: c.awarded_at,
  }));
  db.user_rewards.insertMany(docs);
  print('✅ reward_db seeded: user_rewards (' + docs.length + ' docs)');
} else {
  print('ℹ️ reward_db user_rewards already seeded (' + db.user_rewards.countDocuments() + ' docs)');
}
