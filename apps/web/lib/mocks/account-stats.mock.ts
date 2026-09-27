// Mock GET /api/auth/statistics?id= in the wire shape. Missing user -> 404.

import type { TimerStatistics } from '../profile-types';
import { MOCK_USER_IDS } from './auth.mock';

export const NEVER_ACTIVE = '0001-01-01T00:00:00Z';

// ?mockScenario=stats-unavailable -> 502.
export const MOCK_STATS_UNAVAILABLE = 'stats-unavailable';

function stats(
  user_id: string,
  sessions_joined: number,
  cycles_completed: number,
  total_focus_minutes: number,
  last_active: string,
): TimerStatistics {
  return { user_id, sessions_joined, cycles_completed, total_focus_minutes, last_active };
}

export const MOCK_STATISTICS: Record<string, TimerStatistics> = {
  [MOCK_USER_IDS.signedIn]: stats(MOCK_USER_IDS.signedIn, 24, 58, 1120, '2026-09-26T19:40:00Z'),
  [MOCK_USER_IDS.mira]: stats(MOCK_USER_IDS.mira, 31, 72, 1545, '2026-09-25T08:15:00Z'),
  [MOCK_USER_IDS.tan]: stats(MOCK_USER_IDS.tan, 9, 14, 215, '2025-12-03T21:05:00Z'),
  // Never studied: zero counts and the zero timestamp.
  [MOCK_USER_IDS.newAngler]: stats(MOCK_USER_IDS.newAngler, 0, 0, 0, NEVER_ACTIVE),
};
