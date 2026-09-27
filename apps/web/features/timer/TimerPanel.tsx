"use client";

import { PixelAlert } from "../../components/ui/PixelAlert";
import { PixelBadge } from "../../components/ui/PixelBadge";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelPanel } from "../../components/ui/PixelPanel";
import type { TimerOwner } from "./timer.api";
import { useStudyTimer } from "./useStudyTimer";

function formatClock(totalSeconds: number): string {
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

/**
 * One participant's study timer (UC-05): a countdown with Start, Pause,
 * Resume and Reset. Only the buttons that make sense for the current state
 * are shown, so an illegal move cannot be clicked.
 */
export function TimerPanel({ owner }: { owner: TimerOwner }) {
  const { timer, remainingSeconds, pending, error, run } = useStudyTimer(owner);

  const status = timer?.status ?? "STOPPED";
  const phaseLabel = timer?.phase === "REST" ? "Rest" : "Work";
  const timeUp = status === "RUNNING" && remainingSeconds === 0;
  const clock = timer ? formatClock(remainingSeconds) : "--:--";

  return (
    <PixelPanel as="section" aria-labelledby="timer-heading" className="flex flex-col items-center gap-6 text-center">
      <div className="flex flex-wrap items-center justify-center gap-2">
        <h1 id="timer-heading" className="font-display text-2xl">
          Study Timer
        </h1>
        <PixelBadge>{phaseLabel}</PixelBadge>
        <PixelBadge>{timeUp ? "Time's up" : status}</PixelBadge>
      </div>

      <p
        role="timer"
        aria-label={`${phaseLabel} time remaining ${clock}`}
        className="font-numeric leading-none tabular-nums"
        style={{ fontSize: "calc(var(--px) * 32)" }}
      >
        {clock}
      </p>

      {timer && (
        <p className="text-bark">
          {timer.work_minutes} min work · {timer.rest_minutes} min rest · cycle {timer.current_cycle}
        </p>
      )}

      <div className="flex flex-wrap justify-center gap-4">
        {(status === "STOPPED" || timeUp) && (
          <PixelButton onClick={() => run("start")} disabled={pending || !timer}>
            {timeUp ? "Start again" : "Start"}
          </PixelButton>
        )}
        {status === "RUNNING" && !timeUp && (
          <PixelButton onClick={() => run("pause")} disabled={pending}>
            Pause
          </PixelButton>
        )}
        {status === "PAUSED" && (
          <PixelButton onClick={() => run("resume")} disabled={pending}>
            Resume
          </PixelButton>
        )}
        {status !== "STOPPED" && (
          <PixelButton variant="ghost" onClick={() => run("reset")} disabled={pending}>
            Reset
          </PixelButton>
        )}
      </div>

      {error && <PixelAlert className="w-full">{error}</PixelAlert>}
    </PixelPanel>
  );
}
