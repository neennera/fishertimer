"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  getTimer,
  runTimerAction,
  type StartOptions,
  type TimerAction,
  type TimerOwner,
  type TimerState,
} from "./timer.api";
import { predict, readingFrom, remainingAt, viewOf, type TimerReading, type TimerView } from "./timer-model";

/** A work block that just completed (for the catch reveal). */
export interface CompletedBlock {
  cycleId: string;
  completedAt: string;
  minutes: number;
}

export interface StudyTimer {
  /** null until the first answer from the server. */
  timer: TimerState | null;
  view: TimerView;
  /** Milliseconds left right now; updated every animation frame while running. */
  remainingMs: number;
  /** Share of the phase already done, 0 to 1. */
  progress: number;
  error: string | null;
  /** Set when a work block completes while the panel is open; clear with dismissCompleted. */
  completed: CompletedBlock | null;
  dismissCompleted: () => void;
  /** Applies the action on screen at once, then confirms it with the server. */
  run: (action: TimerAction, options?: StartOptions) => void;
  /** Re-reads the server, e.g. after the settings changed. */
  refresh: () => Promise<void>;
}

/** While running, the server state is re-read this often (other tabs, sweeper). */
const RESYNC_MS = 15_000;

/**
 * Keeps one participant's timer in sync with the Study Timer service.
 *
 * The server owns the time (progress comes from server timestamps); this
 * hook only makes it feel instant and smooth:
 * - A press is predicted locally and shown immediately, then replaced by the
 *   server's answer. Presses are sent one at a time, in order, and only the
 *   answer to the last one is shown, so quick double presses never flicker.
 * - While running, the display advances every animation frame.
 * - When the countdown reaches zero it asks the server to complete the
 *   cycle (the server also completes it on its own if the tab is closed).
 * - It re-reads the server on load, on tab focus and every 15s.
 */
export function useStudyTimer(owner: TimerOwner): StudyTimer {
  const { sessionId, userId } = owner;
  const [reading, setReading] = useState<TimerReading | null>(null);
  const [now, setNow] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [completed, setCompleted] = useState<CompletedBlock | null>(null);

  const readingRef = useRef<TimerReading | null>(null);
  const queue = useRef<Promise<void>>(Promise.resolve());
  const inFlight = useRef(0);
  // The last completed block we know of; undefined until the first answer,
  // so a block finished before the page opened is not announced.
  const knownCompleted = useRef<string | undefined>(undefined);

  const show = useCallback((next: TimerReading) => {
    readingRef.current = next;
    setReading(next);
    setNow(next.at);
  }, []);

  const accept = useCallback(
    (state: TimerState) => {
      const previous = knownCompleted.current;
      knownCompleted.current = state.last_completed_cycle_id;
      if (previous !== undefined && state.last_completed_cycle_id && state.last_completed_cycle_id !== previous) {
        const minutes = readingRef.current?.state.duration_seconds ?? 0;
        setCompleted({
          cycleId: state.last_completed_cycle_id,
          completedAt: state.last_completed_at,
          minutes: Math.round(minutes / 60),
        });
      }
      show(readingFrom(state, performance.now()));
      setError(null);
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

  const send = useCallback(
    (action: TimerAction, options?: StartOptions, failureMessage?: string) => {
      inFlight.current += 1;
      queue.current = queue.current.then(async () => {
        try {
          const state = await runTimerAction(action, { sessionId, userId }, options);
          if (inFlight.current === 1) accept(state);
        } catch {
          // Undo the prediction by showing what the server really has.
          await load();
          if (failureMessage) setError(failureMessage);
        } finally {
          inFlight.current -= 1;
        }
      });
    },
    [accept, load, sessionId, userId],
  );

  const run = useCallback(
    (action: TimerAction, options: StartOptions = {}) => {
      const current = readingRef.current;
      if (!current) return;
      const predicted = predict(action, current, performance.now(), options);
      if (!predicted) return; // the server would refuse it anyway
      show(predicted);
      send(action, options, `Couldn't ${action.replace("-", " ")} — showing the timer's latest state.`);
    },
    [send, show],
  );

  // Load once, again whenever the tab comes back into view, and every 15s.
  useEffect(() => {
    void refresh();
    const onVisible = () => {
      if (document.visibilityState === "visible") void refresh();
    };
    document.addEventListener("visibilitychange", onVisible);
    const id = window.setInterval(() => {
      if (document.visibilityState === "visible") void refresh();
    }, RESYNC_MS);
    return () => {
      document.removeEventListener("visibilitychange", onVisible);
      window.clearInterval(id);
    };
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

  // At zero, ask the server to complete the cycle, then keep re-reading
  // until it has. If the server still sees time left (clocks a little
  // apart), its answer puts a second back on the clock and this runs again.
  useEffect(() => {
    if (view !== "finishing") return;
    send("complete");
    const id = window.setInterval(() => void refresh(), 2000);
    return () => window.clearInterval(id);
  }, [view, send, refresh]);

  const durationMs = (reading?.state.duration_seconds ?? 0) * 1000;
  const progress = durationMs > 0 ? 1 - remainingMs / durationMs : 0;

  return {
    timer: reading?.state ?? null,
    view,
    remainingMs,
    progress,
    error,
    completed,
    dismissCompleted: () => setCompleted(null),
    run,
    refresh,
  };
}
