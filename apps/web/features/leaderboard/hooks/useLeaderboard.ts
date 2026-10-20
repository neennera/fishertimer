'use client';

import { useCallback, useEffect, useState } from 'react';
import { clientApiFetch, friendlyApiError } from '../../../lib/client-api';
import type { LeaderboardPeriod, LeaderboardResponse, LeaderboardState } from '../types';

/**
 * useLeaderboard — UC-08 hook
 *
 * Fetches the leaderboard from GET /api/leaderboard?period=<period>.
 * The request is routed via the Next.js reverse-proxy rewrite (next.config.js)
 * → API Gateway → Leaderboard Service, and gives up after
 * REQUEST_TIMEOUT_MS (lib/client-api.ts).
 *
 * Automatically re-fetches whenever `period` changes (S-2 filter).
 *
 * `data` is only ever the answer for the current `period`: after switching
 * period, or when that period's request fails, it is null — never another
 * period's rows under this period's tab. A failed *refresh* of the period on
 * screen keeps its last rows, with `error` set.
 *
 * @param initialPeriod - the period shown first (the page passes "monthly")
 */
export function useLeaderboard(
  initialPeriod: LeaderboardPeriod = 'all-time'
): LeaderboardState {
  const [period, setPeriodState] = useState<LeaderboardPeriod>(initialPeriod);
  const [loaded, setLoaded] = useState<{ period: LeaderboardPeriod; data: LeaderboardResponse } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tick, setTick] = useState(0); // increment to trigger a manual refresh

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    clientApiFetch<LeaderboardResponse>('leaderboard', `?period=${period}`)
      .then((result) => {
        if (!cancelled) {
          setLoaded({ period, data: result });
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(friendlyApiError(err, 'the leaderboard'));
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [period, tick]);

  const setPeriod = useCallback((next: LeaderboardPeriod) => {
    setPeriodState(next);
  }, []);

  const refresh = useCallback(() => {
    setTick((t) => t + 1);
  }, []);

  const data = loaded?.period === period ? loaded.data : null;

  return { data, loading, error, period, setPeriod, refresh };
}
