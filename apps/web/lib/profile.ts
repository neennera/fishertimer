// Profile data for /account and /profile/[userId].
//
// MOCK ONLY: the account service only has /me, so there is no way yet to
// fetch another user. Real data will come from several services, fetched in
// parallel: account (profile), study-session / study-timer (stats) and reward
// (fish).

import type { RewardItem, UserReward } from '@fishertimer/shared-types';
import type { SessionUser } from './auth';
import { MOCK_ACCOUNT_STATS, type MockAccountStats } from './mocks/account-stats.mock';
import { mockDelay } from './mocks/auth.mock';
import { MOCK_PUBLIC_PROFILES } from './mocks/profile.mock';
import { MOCK_FISH_CATALOG, MOCK_USER_REWARDS } from './mocks/rewards.mock';

// The user wire type without anything private: no e-mail (or role).
export type PublicProfile = Pick<SessionUser, 'user_id' | 'display_name' | 'avatar_url' | 'created_at'>;

export interface ProfileData {
  profile: PublicProfile;
  stats: MockAccountStats;
  fishCatalog: RewardItem[];
  fishRewards: UserReward[];
}

export function toPublicProfile({ user_id, display_name, avatar_url, created_at }: SessionUser): PublicProfile {
  return { user_id, display_name, avatar_url, created_at };
}

// The signed-in user's own data, for /account and their own public view.
export function ownProfileData(user: SessionUser): ProfileData {
  return {
    profile: toPublicProfile(user),
    stats: MOCK_ACCOUNT_STATS,
    fishCatalog: MOCK_FISH_CATALOG,
    fishRewards: MOCK_USER_REWARDS,
  };
}

/** null = no such user. */
export async function getPublicProfile(userId: string): Promise<ProfileData | null> {
  const entry = MOCK_PUBLIC_PROFILES[userId];
  return mockDelay(entry ? { ...entry, fishCatalog: MOCK_FISH_CATALOG } : null);
}
