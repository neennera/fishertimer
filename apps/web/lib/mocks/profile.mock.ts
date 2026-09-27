// MOCK ONLY: other users' public profiles, for /profile/[userId]. See
// lib/profile.ts for where the real data will come from.

import type { UserReward } from '@fishertimer/shared-types';
import type { ProfileData } from '../profile';

function catches(userId: string, counts: Record<string, number>): UserReward[] {
  return Object.entries(counts).flatMap(([itemId, count]) =>
    Array.from({ length: count }, (_, i) => ({
      id: `${userId}-${itemId}-${i}`,
      userId,
      itemId,
      unlockedAt: `2026-09-${String(10 + i).padStart(2, '0')}T12:00:00Z`,
    })),
  );
}

export const MOCK_PUBLIC_PROFILES: Record<string, Omit<ProfileData, 'fishCatalog'>> = {
  'mock-user-2': {
    profile: {
      user_id: 'mock-user-2',
      display_name: 'Mira L.',
      avatar_url: '/sprites/fish/Pufferfish.png',
      created_at: '2026-09-03T00:00:00Z',
    },
    stats: { total_sessions: 31, total_focus_minutes: 1545, rewards_earned: 12 },
    fishRewards: catches('mock-user-2', {
      'fish-angelfish': 3,
      'fish-pufferfish': 4,
      'fish-blue-tang': 2,
      'fish-bass': 1,
    }),
  },
  // Empty avatar_url: exercises the avatar fallback.
  'mock-user-3': {
    profile: {
      user_id: 'mock-user-3',
      display_name: 'Tan R.',
      avatar_url: '',
      created_at: '2026-09-12T00:00:00Z',
    },
    stats: { total_sessions: 9, total_focus_minutes: 215, rewards_earned: 3 },
    fishRewards: catches('mock-user-3', { 'fish-goldfish': 2, 'fish-catfish': 1 }),
  },
  // Zero stats and no fish: exercises the empty states.
  'mock-user-4': {
    profile: {
      user_id: 'mock-user-4',
      display_name: 'New Angler',
      avatar_url: '',
      created_at: '2026-09-26T00:00:00Z',
    },
    stats: { total_sessions: 0, total_focus_minutes: 0, rewards_earned: 0 },
    fishRewards: [],
  },
};
