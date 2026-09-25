// MOCK ONLY. There is no statistics endpoint on the backend yet (the account
// service's ViewStatistics is still a skeleton), so /account's stat row reads
// these fixed numbers — the 03a wireframe's values. Replace with a real call
// in lib/ once the endpoint exists; the page should not import this directly
// then.
//
// Field names and types follow main's account-service skeleton
// (domain.UserStatistics: total_sessions, total_focus_minutes,
// rewards_earned), snake_case as on the wire. Raw numbers only — the page
// formats them for display, so a real response can drop in unchanged.

export interface MockAccountStats {
  total_sessions: number;
  /** Minutes, not a display string: 1120, not "18h 40m". */
  total_focus_minutes: number;
  rewards_earned: number;
}

export const MOCK_ACCOUNT_STATS: MockAccountStats = {
  total_sessions: 24,
  total_focus_minutes: 1120,
  rewards_earned: 132,
};
