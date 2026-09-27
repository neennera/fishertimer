// MOCK ONLY: no rewards endpoint yet. Shapes follow shared-types; the real
// wire format may differ (snake_case, `_id`).
//
// SCHEMA CONFLICT: counts come from several rows per species, but
// schema.dbml's user_rewards is unique on (user_id, item_id).

import type { RewardItem, UserReward } from '@fishertimer/shared-types';

function mockFish(id: string, itemName: string, description?: string): RewardItem {
  return {
    id,
    itemName,
    description,
    itemType: 'FISH_SPECIES',
    cost: 0,
    createdAt: '2026-09-01T00:00:00Z',
  };
}

export const MOCK_FISH_CATALOG: RewardItem[] = [
  mockFish('fish-clownfish', 'Clownfish', 'Caught after a steady focus session.'),
  mockFish('fish-blue-tang', 'Blue Tang', 'Shows up in group sessions.'),
  mockFish('fish-goldfish', 'Goldfish', 'A common first catch.'),
  mockFish('fish-pufferfish', 'Pufferfish', 'Puffs up after a long streak.'),
  mockFish('fish-angelfish', 'Angelfish', 'A rare catch for deep focus.'),
  // Not caught.
  mockFish('fish-anchovy', 'Anchovy'),
  mockFish('fish-bass', 'Bass'),
  mockFish('fish-catfish', 'Catfish'),
  mockFish('fish-rainbow-trout', 'Rainbow Trout'),
];

function mockCatch(n: number, itemId: string, unlockedAt: string): UserReward {
  return { id: `reward-${n}`, userId: 'mock-user-1', itemId, unlockedAt };
}

export const MOCK_USER_REWARDS: UserReward[] = [
  mockCatch(1, 'fish-clownfish', '2026-09-02T09:15:00Z'),
  mockCatch(2, 'fish-goldfish', '2026-09-03T18:20:00Z'),
  mockCatch(3, 'fish-clownfish', '2026-09-06T14:40:00Z'),
  mockCatch(4, 'fish-blue-tang', '2026-09-08T19:05:00Z'),
  mockCatch(5, 'fish-goldfish', '2026-09-10T07:55:00Z'),
  mockCatch(6, 'fish-clownfish', '2026-09-11T08:30:00Z'),
  mockCatch(7, 'fish-pufferfish', '2026-09-13T20:10:00Z'),
  mockCatch(8, 'fish-angelfish', '2026-09-14T21:50:00Z'),
  mockCatch(9, 'fish-blue-tang', '2026-09-17T10:20:00Z'),
  mockCatch(10, 'fish-goldfish', '2026-09-18T16:00:00Z'),
  mockCatch(11, 'fish-clownfish', '2026-09-20T16:45:00Z'),
  mockCatch(12, 'fish-pufferfish', '2026-09-22T07:10:00Z'),
  mockCatch(13, 'fish-clownfish', '2026-09-24T13:35:00Z'),
  // 18 in all: over the tank's 15-fish limit.
  mockCatch(14, 'fish-goldfish', '2026-09-24T18:05:00Z'),
  mockCatch(15, 'fish-blue-tang', '2026-09-24T20:30:00Z'),
  mockCatch(16, 'fish-clownfish', '2026-09-25T08:10:00Z'),
  mockCatch(17, 'fish-goldfish', '2026-09-25T09:45:00Z'),
  mockCatch(18, 'fish-clownfish', '2026-09-25T11:20:00Z'),
];
