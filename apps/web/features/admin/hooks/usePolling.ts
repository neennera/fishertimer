import { useCallback, useEffect, useRef, useState } from 'react';
import { AdminApiError } from '../../../lib/admin-api';

/** How often admin monitoring re-reads Study Session (UC-04 step 5). */
export const ADMIN_POLL_MS = 5000;

interface PollState<T> {
  data: T | undefined;
  /** Set only while there is no data to show (first load failed). */
  error: unknown;
  /** True when the last refresh failed but older data is still shown (UC-04 E-2). */
  stale: boolean;
  /** Reloads now (also the Retry button), and resumes polling after a sign-in error. */
  refresh: () => Promise<void>;
}

/**
 * Loads `fetcher` now and every `intervalMs` until the page unmounts.
 * A failed refresh keeps the last snapshot and marks it stale; the next
 * success replaces it. Change `key` to restart with a fresh load.
 *
 * Polling is polite: it skips a tick while the previous request is still in
 * flight (a slow server is not hammered), pauses while the tab is hidden and
 * refreshes as soon as it is shown again, and stops after a 401 / 403 (signing
 * in or a Retry restarts it). Only the newest answer is ever applied, so a slow
 * old response cannot overwrite a newer one.
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
  /** Bumped per request and per restart; only the latest request may apply. */
  const latest = useRef(0);
  /** Bumped per restart (new key); requests of an older generation don't count as in flight. */
  const generation = useRef(0);
  const inFlight = useRef(0);
  const authBlocked = useRef(false);

  const refresh = useCallback(async () => {
    authBlocked.current = false;
    const mine = ++latest.current;
    const gen = generation.current;
    inFlight.current += 1;
    try {
      const next = await fetcherRef.current();
      if (!alive.current || mine !== latest.current) return;
      hasData.current = true;
      setData(next);
      setError(null);
      setStale(false);
    } catch (e) {
      if (!alive.current || mine !== latest.current) return;
      if (e instanceof AdminApiError && (e.status === 401 || e.status === 403)) {
        // Signed out (or not an admin): what is on screen can no longer be
        // refreshed or acted on, so show the sign-in banner instead of an
        // "out of date" note that would never resolve. A Retry restarts polling.
        authBlocked.current = true;
        hasData.current = false;
        setData(undefined);
        setStale(false);
        setError(e);
        return;
      }
      if (hasData.current) setStale(true);
      else setError(e);
    } finally {
      // Only a request of the current generation is counted: a restart reset the
      // counter, so an old request settling must not drive it below the truth.
      if (gen === generation.current) inFlight.current -= 1;
    }
  }, []);

  useEffect(() => {
    alive.current = true;
    latest.current += 1; // drop any answer still in flight for the previous key
    generation.current += 1;
    inFlight.current = 0;
    hasData.current = false;
    authBlocked.current = false;
    setData(undefined);
    setError(null);
    setStale(false);
    refresh();

    const tick = () => {
      if (document.hidden || authBlocked.current || inFlight.current > 0) return;
      void refresh();
    };
    const timer = setInterval(tick, intervalMs);
    const onVisible = () => {
      if (!document.hidden && !authBlocked.current && inFlight.current === 0) void refresh();
    };
    document.addEventListener('visibilitychange', onVisible);
    return () => {
      alive.current = false;
      clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisible);
    };
  }, [key, intervalMs, refresh]);

  return { data, error, stale, refresh };
}
