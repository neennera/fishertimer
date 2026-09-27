// Mock GET /api/auth/rewards?id= in the wire shape. Missing user -> 404.

import type { RewardSummaryItem, RewardsSummary } from '../profile-types';
import { MOCK_USER_IDS } from './auth.mock';

// ?mockScenario=rewards-unavailable -> 502.
export const MOCK_REWARDS_UNAVAILABLE = 'rewards-unavailable';

function item(
  name: string,
  count: number,
  asset_url: string,
  type: RewardSummaryItem['type'] = 'FISH',
  rarity: RewardSummaryItem['rarity'] = 'COMMON',
  score_value = 10,
): RewardSummaryItem {
  return { name, rarity, asset_url, type, score_value, count };
}

function summary(items: RewardSummaryItem[]): RewardsSummary {
  return {
    total_awards_earned: items.reduce((sum, i) => sum + i.count, 0),
    total_score: items.reduce((sum, i) => sum + i.count * i.score_value, 0),
    items,
  };
}

export const MOCK_REWARDS: Record<string, RewardsSummary> = {
  [MOCK_USER_IDS.signedIn]: summary([
    item('Clownfish', 7, '/sprites/fish/Clownfish.png'),
    // No asset_url: matched by name.
    item('Blue Tang', 3, '', 'FISH', 'UNCOMMON', 25),
    item('Goldfish', 5, '/sprites/fish/Goldfish.png'),
    item('Pufferfish', 2, '/sprites/fish/Pufferfish.png', 'FISH', 'RARE', 50),
    item('Angelfish', 1, '/sprites/fish/Angelfish.png', 'FISH', 'UNCOMMON', 25),
    // Unknown sprite: placeholder.
    item('Golden Betta', 1, 'https://assets.example/fish/golden-betta.png', 'FISH', 'LEGENDARY', 250),
    // Not fish: in rewards earned, not in the tank.
    item('Night Owl Badge', 2, '/sprites/items/night-owl.png', 'DECORATION', 'RARE', 50),
    item('Bamboo Rod', 1, '/sprites/items/bamboo-rod.png', 'ROD', 'UNCOMMON', 25),
  ]),
  [MOCK_USER_IDS.mira]: summary([
    item('Angelfish', 3, '/sprites/fish/Angelfish.png', 'FISH', 'UNCOMMON', 25),
    item('Pufferfish', 4, '/sprites/fish/Pufferfish.png', 'FISH', 'RARE', 50),
    item('Blue Tang', 2, '/sprites/fish/Surgeonfish.png', 'FISH', 'UNCOMMON', 25),
    item('Bass', 1, '/sprites/fish/Bass.png'),
  ]),
  [MOCK_USER_IDS.tan]: summary([
    item('Goldfish', 2, '/sprites/fish/Goldfish.png'),
    item('Catfish', 1, '/sprites/fish/Catfish.png'),
  ]),
  // No fish at all: the empty tank.
  [MOCK_USER_IDS.newAngler]: summary([]),
};
