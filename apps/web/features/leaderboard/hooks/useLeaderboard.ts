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
 * period it is null until that period arrives, never another period's rows
 * under this period's tab. While the new period loads, `previous` holds the
 * last rows of another period so the page can keep them on screen, dimmed,
 * instead of flashing a skeleton. If the new period's request fails, those
 * rows are dropped for good. A failed *refresh* of the period on screen keeps
 * its last rows, with `error` set.
 *
 * `loading` and `error` are derived from the request being asked for right now,
 * so they are right in the very render that follows a tab click.
 *
 * @param initialPeriod - the period shown first (the page passes "monthly")
 */
export function useLeaderboard(
  initialPeriod: LeaderboardPeriod = 'all-time'
): LeaderboardState {
  const [period, setPeriodState] = useState<LeaderboardPeriod>(initialPeriod);
  const [loaded, setLoaded] = useState<{ period: LeaderboardPeriod; data: LeaderboardResponse } | null>(null);
  // How the latest request settled, tagged with the request it belongs to.
  const [result, setResult] = useState<{ key: string; error: string | null } | null>(null);
  const [tick, setTick] = useState(0); // increment to trigger a manual refresh

  const key = `${period}:${tick}`;

  useEffect(() => {
    let cancelled = false;
    clientApiFetch<LeaderboardResponse>('leaderboard', `?period=${period}`)
      .then((response) => {
        if (!cancelled) {
          setLoaded({ period, data: response });
          setResult({ key, error: null });
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          // Rows of a different period are dropped for good: they must not come
          // back, dimmed, on the next Retry.
          setLoaded((prev) => (prev && prev.period !== period ? null : prev));
          setResult({ key, error: friendlyApiError(err, 'the leaderboard') });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [period, tick, key]);

  const setPeriod = useCallback((next: LeaderboardPeriod) => {
    setPeriodState(next);
  }, []);

  const refresh = useCallback(() => {
    setTick((t) => t + 1);
  }, []);

  const settled = result?.key === key;
  const data = loaded?.period === period ? loaded.data : null;
  const previous = loaded && loaded.period !== period ? loaded.data : null;

  return {
    data,
    previous,
    loading: !settled,
    error: settled ? result.error : null,
    period,
    setPeriod,
    refresh,
  };
}
