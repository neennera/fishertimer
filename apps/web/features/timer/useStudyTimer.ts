"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { getTimer, runTimerAction, type TimerAction, type TimerOwner, type TimerState } from "./timer.api";
import { predict, readingFrom, remainingAt, viewOf, type TimerReading, type TimerView } from "./timer-model";

export interface StudyTimer {
  /** null until the first answer from the server. */
  timer: TimerState | null;
  view: TimerView;
  /** Milliseconds left right now; updated every animation frame while running. */
  remainingMs: number;
  /** Share of the phase already done, 0 to 1. */
  progress: number;
  error: string | null;
  /** Applies the action on screen at once, then confirms it with the server. */
  run: (action: TimerAction) => void;
}

/**
 * Keeps one participant's timer in sync with the Study Timer service.
 *
 * The server owns the time (FR: progress comes from server timestamps); this
 * hook only makes it feel instant and smooth:
 * - A press is predicted locally and shown immediately, then replaced by the
 *   server's answer. Presses are sent one at a time, in order, and only the
 *   answer to the last one is shown, so quick double presses never flicker.
 * - While running, the display advances every animation frame.
 * - It re-reads the server on load, when the tab becomes visible again, and
 *   once when the countdown reaches zero, so drift never builds up.
 */
export function useStudyTimer(owner: TimerOwner): StudyTimer {
  const { sessionId, userId } = owner;
  const [reading, setReading] = useState<TimerReading | null>(null);
  const [now, setNow] = useState(0);
  const [error, setError] = useState<string | null>(null);

  // The latest reading, for callbacks that must not re-subscribe every frame.
  const readingRef = useRef<TimerReading | null>(null);
  const queue = useRef<Promise<void>>(Promise.resolve());
  const inFlight = useRef(0);
  const resyncedAtZero = useRef(false);

  const show = useCallback((next: TimerReading) => {
    readingRef.current = next;
    setReading(next);
    setNow(next.at);
  }, []);

  const accept = useCallback(
    (state: TimerState) => {
      show(readingFrom(state, performance.now()));
      setError(null);
      // A finished phase still reads RUNNING with 0 left; only a state with
      // time on the clock arms the next zero re-sync, or it would loop.
      if (state.remaining_seconds > 0) resyncedAtZero.current = false;
    },
    [show],
  );

  const load = useCallback(async () => {
    try {
      accept(await getTimer({ sessionId, userId }));
    } catch {
      setError("Can't reach the study timer. Check that the services are running.");
    }
  }, [accept, sessionId, userId]);

  // Background re-sync. Skipped while a press is in flight: its answer is
  // newer than anything a read could return.
  const refresh = useCallback(async () => {
    if (inFlight.current === 0) await load();
  }, [load]);

  const run = useCallback(
    (action: TimerAction) => {
      const current = readingRef.current;
      if (!current) return;
      const predicted = predict(action, current, performance.now());
      if (!predicted) return; // the server would refuse it anyway
      show(predicted);

      inFlight.current += 1;
      queue.current = queue.current.then(async () => {
        try {
          const state = await runTimerAction(action, { sessionId, userId });
          if (inFlight.current === 1) accept(state);
        } catch {
          // Undo the prediction by showing what the server really has.
          await load();
          setError(`Couldn't ${action} the timer — showing its latest state.`);
        } finally {
          inFlight.current -= 1;
        }
      });
    },
    [accept, load, sessionId, userId, show],
  );

  // Load once, and again whenever the tab comes back into view.
  useEffect(() => {
    void refresh();
    const onVisible = () => {
      if (document.visibilityState === "visible") void refresh();
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => document.removeEventListener("visibilitychange", onVisible);
  }, [refresh]);

  const running = reading?.state.status === "RUNNING";

  // Advance the display every frame while running.
  useEffect(() => {
    if (!running) return;
    let frame = requestAnimationFrame(function tick(t) {
      setNow(t);
      frame = requestAnimationFrame(tick);
    });
    return () => cancelAnimationFrame(frame);
  }, [running]);

  const remainingMs = reading ? remainingAt(reading, now) : 0;
  const view: TimerView = reading ? viewOf(reading, now) : "ready";

  // When the countdown reaches zero, confirm the final state with the server.
  useEffect(() => {
    if (view === "done" && !resyncedAtZero.current) {
      resyncedAtZero.current = true;
      void refresh();
    }
  }, [view, refresh]);

  const durationMs = (reading?.state.duration_seconds ?? 0) * 1000;
  const progress = durationMs > 0 ? 1 - remainingMs / durationMs : 0;

  return { timer: reading?.state ?? null, view, remainingMs, progress, error, run };
}
