// Mock GET /api/auth/statistics?id= in the wire shape. Missing user -> 404.

import { lastDays } from '../format';
import type { TimerStatistics } from '../profile-types';
import { MOCK_USER_IDS } from './auth.mock';

export const NEVER_ACTIVE = '0001-01-01T00:00:00Z';

// ?mockScenario=stats-unavailable -> 502.
export const MOCK_STATS_UNAVAILABLE = 'stats-unavailable';

// 30 days of minutes, oldest first; dates are filled in relative to today.
function daily(minutes: number[]): TimerStatistics['daily_focus_minutes'] {
  return lastDays(30).map((date, i) => ({ date, focus_minutes: minutes[i] ?? 0 }));
}

const ZERO_DAYS = Array<number>(30).fill(0);

function stats(
  user_id: string,
  sessions_joined: number,
  cycles_completed: number,
  total_focus_minutes: number,
  last_active: string,
  minutes: number[],
): TimerStatistics {
  return {
    user_id,
    sessions_joined,
    cycles_completed,
    total_focus_minutes,
    last_active,
    daily_focus_minutes: daily(minutes),
  };
}

export const MOCK_STATISTICS: Record<string, TimerStatistics> = {
  [MOCK_USER_IDS.signedIn]: stats(MOCK_USER_IDS.signedIn, 24, 58, 1120, '2026-09-26T19:40:00Z', [
    25, 0, 50, 75, 0, 0, 100, 25, 50, 0, 0, 120, 75, 25, 0, 50, 90, 0, 25, 0, 60, 110, 0, 50, 25, 0, 75, 100, 0, 45,
  ]),
  [MOCK_USER_IDS.mira]: stats(MOCK_USER_IDS.mira, 31, 72, 1545, '2026-09-25T08:15:00Z', [
    60, 90, 0, 120, 75, 60, 0, 0, 90, 100, 50, 0, 75, 120, 60, 0, 25, 90, 110, 0, 0, 60, 75, 120, 90, 0, 50, 100, 25, 0,
  ]),
  [MOCK_USER_IDS.tan]: stats(MOCK_USER_IDS.tan, 9, 14, 215, '2025-12-03T21:05:00Z', ZERO_DAYS),
  // Never studied: zero counts, the zero timestamp and 30 zero days.
  [MOCK_USER_IDS.newAngler]: stats(MOCK_USER_IDS.newAngler, 0, 0, 0, NEVER_ACTIVE, ZERO_DAYS),
};
