// Profile data from the account service (/api/auth/* -> /api/v1/account/*):
//   GET profile?id=     200 full user · 400 no id · 404 unknown
//   GET statistics?id=  200 TimerStatistics · 404 · 502 timer down
//   GET rewards?id=     200 RewardsSummary · 404 · 502 reward down
// Mocked unless NEXT_PUBLIC_USE_MOCKS is "false", in the same shapes.

import type { SessionUser } from './auth';
import { ClientApiError, clientApiFetch } from './client-api';
import { lastDays } from './format';
import { MOCK_STATISTICS, MOCK_STATS_UNAVAILABLE } from './mocks/account-stats.mock';
import { MOCK_SCENARIO_PARAM, mockDelay, readMockAuthState } from './mocks/auth.mock';
import { MOCK_PROFILES } from './mocks/profile.mock';
import { MOCK_REWARDS, MOCK_REWARDS_UNAVAILABLE } from './mocks/rewards.mock';
import type { RewardsSummary, TimerStatistics } from './profile-types';

const USE_MOCKS = process.env.NEXT_PUBLIC_USE_MOCKS !== 'false';

// No e-mail, no role.
export type PublicProfile = Pick<SessionUser, 'user_id' | 'display_name' | 'avatar_url' | 'created_at'>;

export type PanelData<T> =
  | { status: 'loading' }
  | { status: 'ok'; data: T }
  | { status: 'unavailable' };

export type Fetched<T> = { status: 'ok'; data: T } | { status: 'not_found' } | { status: 'unavailable' };

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function isUserId(value: string): boolean {
  return UUID.test(value);
}

export function toPublicProfile({ user_id, display_name, avatar_url, created_at }: SessionUser): PublicProfile {
  return { user_id, display_name, avatar_url, created_at };
}

// Throws like clientApiFetch when there's no value.
async function mockCall<T>(value: T | undefined, status = 404): Promise<T> {
  await mockDelay(undefined);
  if (value === undefined) {
    throw new ClientApiError('auth', status, 'mock');
  }
  return value;
}

function mockScenario(): string | null {
  return new URLSearchParams(window.location.search).get(MOCK_SCENARIO_PARAM);
}

async function fetched<T>(call: () => Promise<T>): Promise<Fetched<T>> {
  try {
    return { status: 'ok', data: await call() };
  } catch (err) {
    if (err instanceof ClientApiError && (err.status === 404 || err.status === 400)) {
      return { status: 'not_found' };
    }
    return { status: 'unavailable' };
  }
}

/** null if there's no such user; other failures throw. E-mail and role are
 * stripped here and never leave. */
export async function getPublicProfile(userId: string): Promise<PublicProfile | null> {
  // The backend answers 500 for a malformed id.
  if (!isUserId(userId)) {
    return null;
  }
  const result = await fetched(() => {
    if (USE_MOCKS) {
      const session = readMockAuthState()?.session;
      const own = session?.status === 'signed_in' && session.user.user_id === userId ? session.user : undefined;
      return mockCall(MOCK_PROFILES[userId] ?? own);
    }
    return clientApiFetch<SessionUser>('auth', `profile?id=${encodeURIComponent(userId)}`);
  });
  if (result.status === 'not_found') {
    return null;
  }
  if (result.status === 'unavailable') {
    throw new Error('profile unavailable');
  }
  return toPublicProfile(result.data);
}

export function getStatistics(userId: string): Promise<Fetched<TimerStatistics>> {
  return fetched(() =>
    USE_MOCKS
      ? mockCall(
          mockScenario() === MOCK_STATS_UNAVAILABLE ? undefined : MOCK_STATISTICS[userId],
          mockScenario() === MOCK_STATS_UNAVAILABLE ? 502 : 404,
        )
      : clientApiFetch<TimerStatistics>('auth', `statistics?id=${encodeURIComponent(userId)}`),
  );
}

export function getRewards(userId: string): Promise<Fetched<RewardsSummary>> {
  return fetched(() =>
    USE_MOCKS
      ? mockCall(
          mockScenario() === MOCK_REWARDS_UNAVAILABLE ? undefined : MOCK_REWARDS[userId],
          mockScenario() === MOCK_REWARDS_UNAVAILABLE ? 502 : 404,
        )
      : clientApiFetch<RewardsSummary>('auth', `rewards?id=${encodeURIComponent(userId)}`),
  );
}

export const NO_STATISTICS = (userId: string): TimerStatistics => ({
  user_id: userId,
  sessions_joined: 0,
  cycles_completed: 0,
  total_focus_minutes: 0,
  last_active: '0001-01-01T00:00:00Z',
  daily_focus_minutes: lastDays(30).map((date) => ({ date, focus_minutes: 0 })),
});

export const NO_REWARDS: RewardsSummary = { total_awards_earned: 0, total_score: 0, items: [] };

/** Statistics and rewards in parallel. A 404 shows as zero (study-timer
 * answers 404 for no history); other failures are "unavailable". */
export async function loadPanels(userId: string): Promise<{
  stats: PanelData<TimerStatistics>;
  rewards: PanelData<RewardsSummary>;
}> {
  const [stats, rewards] = await Promise.allSettled([getStatistics(userId), getRewards(userId)]);
  return {
    stats: toPanel(stats, NO_STATISTICS(userId)),
    rewards: toPanel(rewards, NO_REWARDS),
  };
}

function toPanel<T>(result: PromiseSettledResult<Fetched<T>>, empty: T): PanelData<T> {
  if (result.status === 'rejected' || result.value.status === 'unavailable') {
    return { status: 'unavailable' };
  }
  return { status: 'ok', data: result.value.status === 'ok' ? result.value.data : empty };
}
