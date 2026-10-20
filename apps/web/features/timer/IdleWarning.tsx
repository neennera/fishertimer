"use client";

import { useEffect, useState } from "react";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { formatClock } from "./timer-model";
import { IDLE_TIMEOUT_MINUTES, type TimerState } from "./timer.api";

/** How long before the removal the warning shows. */
const WARN_MS = 3 * 60_000;

/**
 * When an idle user will be moved out of the room (UC-05 E-8), in epoch ms,
 * or null while a cycle is running or paused (not idle). Mirrors
 * services/study-session usecase.IsIdle: the idle clock starts when the
 * user joined or when their last cycle ended, whichever is later.
 */
export function idleDeadline(timer: TimerState | null, joinedAt: string | undefined): number | null {
  if (!timer || !joinedAt || timer.status !== "STOPPED") return null;
  const since = [joinedAt, timer.last_updated].map((iso) => Date.parse(iso)).filter(Number.isFinite);
  if (since.length === 0) return null;
  return Math.max(...since) + IDLE_TIMEOUT_MINUTES * 60_000;
}

/**
 * Counts down the last few minutes before an idle user is moved off the
 * dock, so the removal is never a surprise. Hidden until then.
 */
export function IdleWarning({ deadline }: { deadline: number }) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(id);
  }, []);

  const left = deadline - now;
  if (left > WARN_MS) return null;

  return (
    <PixelAlert tone="warn" className="mt-4 text-left">
      {left > 0 ? (
        <>
          No focus block is running. Start one within{" "}
          {/* The clock ticks every second; screen readers get it once. */}
          <strong className="font-numeric" aria-hidden="true">
            {formatClock(left)}
          </strong>
          <span className="sr-only">a few minutes</span> or you&apos;ll be moved off the dock.
        </>
      ) : (
        `Moving you off the dock: no focus block ran for ${IDLE_TIMEOUT_MINUTES} minutes.`
      )}
    </PixelAlert>
  );
}
