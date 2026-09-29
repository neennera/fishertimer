/**
 * Leaderboard feature types — UC-08
 */

/** The three supported ranking periods (S-2). */
export type LeaderboardPeriod = 'weekly' | 'monthly' | 'all-time';

/** A single row in the ranking list (maps to backend domain.RankEntry). */
export interface RankEntry {
  user_id: string;
  display_name: string;
  rank: number;
  reward_count: number;
  period: LeaderboardPeriod;
}

/** The full response from GET /api/v1/leaderboard (maps to domain.CachedRanking). */
export interface LeaderboardResponse {
  rankings: RankEntry[];
  period: LeaderboardPeriod;
  leaderboard_last_fetch: string; // ISO 8601
  cached: boolean;
}

/** UI state shape returned by useLeaderboard hook. */
export interface LeaderboardState {
  data: LeaderboardResponse | null;
  loading: boolean;
  error: string | null;
  period: LeaderboardPeriod;
  /** Switch to a different period (triggers S-2 re-render). */
  setPeriod: (period: LeaderboardPeriod) => void;
  /** Manually refresh (busts the local cached flag check). */
  refresh: () => void;
}
