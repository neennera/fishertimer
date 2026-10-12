import { useCallback, useEffect, useRef, useState } from 'react';

/** How often admin monitoring re-reads Study Session (UC-04 step 5). */
export const ADMIN_POLL_MS = 5000;

interface PollState<T> {
  data: T | undefined;
  /** Set only while there is no data to show (first load failed). */
  error: unknown;
  /** True when the last refresh failed but older data is still shown (UC-04 E-2). */
  stale: boolean;
  refresh: () => Promise<void>;
}

/**
 * Loads `fetcher` now and every `intervalMs` until the page unmounts.
 * A failed refresh keeps the last snapshot and marks it stale; the next
 * success replaces it. Change `key` to restart with a fresh load.
 */
export function usePolling<T>(
  fetcher: () => Promise<T>,
  key: string,
  intervalMs = ADMIN_POLL_MS,
): PollState<T> {
  const [data, setData] = useState<T | undefined>(undefined);
  const [error, setError] = useState<unknown>(null);
  const [stale, setStale] = useState(false);
  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;
  const hasData = useRef(false);
  const alive = useRef(true);

  const refresh = useCallback(async () => {
    try {
      const next = await fetcherRef.current();
      if (!alive.current) return;
      hasData.current = true;
      setData(next);
      setError(null);
      setStale(false);
    } catch (e) {
      if (!alive.current) return;
      if (hasData.current) setStale(true);
      else setError(e);
    }
  }, []);

  useEffect(() => {
    alive.current = true;
    hasData.current = false;
    setData(undefined);
    setError(null);
    setStale(false);
    refresh();
    const timer = setInterval(refresh, intervalMs);
    return () => {
      alive.current = false;
      clearInterval(timer);
    };
  }, [key, intervalMs, refresh]);

  return { data, error, stale, refresh };
}
