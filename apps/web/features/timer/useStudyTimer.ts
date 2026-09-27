"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  getTimer,
  runTimerAction,
  type TimerAction,
  type TimerOwner,
  type TimerState,
} from "./timer.api";

const TICK_MS = 250;

export interface StudyTimer {
  /** null until the first answer from the server. */
  timer: TimerState | null;
  /** Seconds left right now, counted down locally between server answers. */
  remainingSeconds: number;
  /** True while an action is waiting for the server. */
  pending: boolean;
  error: string | null;
  run: (action: TimerAction) => Promise<void>;
}

/**
 * Keeps one participant's timer in sync with the Study Timer service.
 *
 * The server owns the time (FR: progress comes from server timestamps). Each
 * answer carries `remaining_seconds`; between answers the hook only counts
 * down locally for display, and re-reads the server when the tab becomes
 * visible again or the countdown reaches zero, so drift never accumulates.
 */
export function useStudyTimer(owner: TimerOwner): StudyTimer {
  const { sessionId, userId } = owner;
  const [timer, setTimer] = useState<TimerState | null>(null);
  const [receivedAt, setReceivedAt] = useState(0);
  const [now, setNow] = useState(0);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const resyncedAtZero = useRef(false);

  const accept = useCallback((state: TimerState) => {
    const at = Date.now();
    setTimer(state);
    setReceivedAt(at);
    setNow(at);
    setError(null);
    // A finished phase still reads RUNNING with 0 left; only a state with
    // time on the clock arms the next zero re-sync, or it would loop.
    if (state.remaining_seconds > 0) resyncedAtZero.current = false;
  }, []);

  const refresh = useCallback(async () => {
    try {
      accept(await getTimer({ sessionId, userId }));
    } catch {
      setError("Could not reach the study timer. Check that the services are running, then try again.");
    }
  }, [accept, sessionId, userId]);

  const run = useCallback(
    async (action: TimerAction) => {
      setPending(true);
      try {
        accept(await runTimerAction(action, { sessionId, userId }));
      } catch {
        setError(`Could not ${action} the timer. Showing its latest state.`);
        await refresh();
      } finally {
        setPending(false);
      }
    },
    [accept, refresh, sessionId, userId],
  );

  // Load the timer once, and again whenever the tab comes back into view.
  useEffect(() => {
    void refresh();
    const onVisible = () => {
      if (document.visibilityState === "visible") void refresh();
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => document.removeEventListener("visibilitychange", onVisible);
  }, [refresh]);

  const running = timer?.status === "RUNNING";

  // Tick the display only while the timer runs.
  useEffect(() => {
    if (!running) return;
    const id = window.setInterval(() => setNow(Date.now()), TICK_MS);
    return () => window.clearInterval(id);
  }, [running]);

  const elapsed = running ? Math.floor((now - receivedAt) / 1000) : 0;
  const remainingSeconds = timer ? Math.max(timer.remaining_seconds - elapsed, 0) : 0;

  // When the countdown reaches zero, confirm the final state with the server.
  useEffect(() => {
    if (running && remainingSeconds === 0 && !resyncedAtZero.current) {
      resyncedAtZero.current = true;
      void refresh();
    }
  }, [running, remainingSeconds, refresh]);

  return { timer, remainingSeconds, pending, error, run };
}
