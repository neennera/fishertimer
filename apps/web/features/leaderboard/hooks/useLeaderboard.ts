'use client';

import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../../../lib/api-client';
import type { LeaderboardPeriod, LeaderboardResponse, LeaderboardState } from '../types';

/**
 * useLeaderboard — UC-08 hook
 *
 * Fetches the leaderboard from GET /api/leaderboard?period=<period>.
 * The request is routed via the Next.js reverse-proxy rewrite (next.config.js)
 * → API Gateway → Leaderboard Service.
 *
 * Automatically re-fetches whenever `period` changes (S-2 filter).
 *
 * @param initialPeriod - defaults to "all-time" per UC-08 Normal Flow step 2
 */
export function useLeaderboard(
  initialPeriod: LeaderboardPeriod = 'all-time'
): LeaderboardState {
  const [period, setPeriodState] = useState<LeaderboardPeriod>(initialPeriod);
  const [data, setData] = useState<LeaderboardResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [tick, setTick] = useState(0); // increment to trigger a manual refresh

  const fetchLeaderboard = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await apiFetch<LeaderboardResponse>(
        'leaderboard',
        `?period=${period}`
      );
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load leaderboard');
    } finally {
      setLoading(false);
    }
  }, [period, tick]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    fetchLeaderboard();
  }, [fetchLeaderboard]);

  const setPeriod = useCallback((next: LeaderboardPeriod) => {
    setPeriodState(next);
  }, []);

  const refresh = useCallback(() => {
    setTick((t) => t + 1);
  }, []);

  return { data, loading, error, period, setPeriod, refresh };
}
