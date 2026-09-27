"use client";

import { useEffect } from "react";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelPanel } from "../../components/ui/PixelPanel";
import type { TimerOwner } from "./timer.api";
import { formatClock, PRIMARY_ACTION, type TimerView } from "./timer-model";
import { useStudyTimer } from "./useStudyTimer";

const COPY: Record<TimerView, { chip: string; primary: string; announce: string }> = {
  ready: { chip: "Ready", primary: "Start focus", announce: "Timer ready" },
  running: { chip: "Focusing", primary: "Pause", announce: "Focus started" },
  paused: { chip: "Paused", primary: "Resume", announce: "Timer paused" },
  done: { chip: "Done", primary: "Start again", announce: "Focus block complete" },
};

const APP_TITLE = "Fisher Timer";

/**
 * One participant's study timer (UC-05).
 *
 * Layout is fixed across states: the primary button always sits in the same
 * place and width (Start -> Pause -> Resume), and Reset is always there,
 * disabled when there is nothing to reset. So a hand that just pressed Start
 * finds Pause under the same finger.
 *
 * Keyboard: Space runs the primary action, R resets.
 */
export function TimerPanel({ owner }: { owner: TimerOwner }) {
  const { timer, view, remainingMs, progress, error, run } = useStudyTimer(owner);
  const copy = COPY[view];
  const loaded = timer !== null;
  const clock = loaded ? formatClock(remainingMs) : "--:--";
  const canReset = loaded && (view !== "ready" || remainingMs !== timer.work_minutes * 60_000);

  const TAB_LABEL: Record<TimerView, string | null> = { ready: null, running: clock, paused: "Paused", done: "Done" };
  const tabLabel = loaded ? TAB_LABEL[view] : null;
  useDocumentTitle(tabLabel ? `${tabLabel} · ${APP_TITLE}` : APP_TITLE);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (!loaded || e.repeat || e.metaKey || e.ctrlKey || e.altKey) return;
      // Leave keys alone while a control has focus: Space already presses it.
      const target = e.target as HTMLElement;
      if (target.closest("button, a, input, textarea, select, [contenteditable]")) return;

      if (e.code === "Space") {
        e.preventDefault();
        run(PRIMARY_ACTION[view]);
      } else if (e.key === "r" || e.key === "R") {
        if (canReset) run("reset");
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [loaded, view, canReset, run]);

  return (
    <PixelPanel as="section" aria-labelledby="timer-heading" className="w-full text-center">
      <div className="flex items-center justify-between gap-3">
        <h1 id="timer-heading" className="font-display text-2xl leading-none">
          Focus timer
        </h1>
        <span className={`pixel-chip pixel-chip--${view}`}>{copy.chip}</span>
      </div>

      <p role="timer" aria-label={`${clock} remaining`} data-view={view} className="pixel-clock mt-8">
        {clock}
      </p>

      <div
        className="pixel-progress mt-6"
        data-view={view}
        role="progressbar"
        aria-label="Focus progress"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(progress * 100)}
      >
        <div className="pixel-progress__fill" style={{ transform: `scaleX(${progress})` }} />
      </div>

      <p className="mt-3 text-sm text-bark">
        {view === "done"
          ? "Focus block complete. Nice work!"
          : timer
            ? `${timer.work_minutes} min focus · ${timer.rest_minutes} min break`
            : "Connecting to the timer…"}
      </p>

      <div className="mt-6 flex justify-center gap-4">
        <PixelButton className="pixel-timer__primary" onClick={() => run(PRIMARY_ACTION[view])} disabled={!loaded}>
          {copy.primary}
        </PixelButton>
        <PixelButton variant="ghost" onClick={() => run("reset")} disabled={!canReset}>
          Reset
        </PixelButton>
      </div>

      <p className="mt-5 hidden items-center justify-center gap-2 text-xs text-bark md:flex">
        <kbd className="pixel-kbd">Space</kbd> {view === "running" ? "pause" : view === "paused" ? "resume" : "start"}
        <span aria-hidden="true">·</span>
        <kbd className="pixel-kbd">R</kbd> reset
      </p>

      {/* Screen readers hear state changes, not every tick. */}
      <p className="sr-only" aria-live="polite">
        {loaded ? copy.announce : ""}
      </p>

      {error && <PixelAlert className="mt-5 text-left">{error}</PixelAlert>}
    </PixelPanel>
  );
}

/** Shows the countdown in the browser tab, so it is visible from other tabs. */
function useDocumentTitle(title: string) {
  useEffect(() => {
    document.title = title;
  }, [title]);
  useEffect(() => () => void (document.title = APP_TITLE), []);
}
