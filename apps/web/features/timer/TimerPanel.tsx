"use client";

import { useEffect, useState } from "react";
import { Gear } from "pixelarticons/react/Gear";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelModal } from "../../components/ui/PixelModal";
import { PixelPanel } from "../../components/ui/PixelPanel";
import { cx } from "../../lib/cx";
import { CatchReveal } from "./CatchReveal";
import { DurationPicker } from "./DurationPicker";
import { MIN_REWARD_MINUTES, type TimerAction, type TimerOwner, type StartOptions } from "./timer.api";
import { clampMinutes, formatClock, type TimerView } from "./timer-model";
import { REST_PRESETS, TimerSettingsForm, WORK_PRESETS } from "./TimerSettingsForm";
import { useStudyTimer } from "./useStudyTimer";

interface Copy {
  chip: string;
  chipClass: string;
  title: string;
  announce: string;
}

const COPY: Record<TimerView, Copy> = {
  ready: { chip: "Ready", chipClass: "", title: "Focus", announce: "Timer ready" },
  focus: { chip: "Focusing", chipClass: "pixel-chip--running", title: "Focus", announce: "Focus started" },
  "focus-paused": { chip: "Paused", chipClass: "pixel-chip--paused", title: "Focus", announce: "Focus paused" },
  finishing: { chip: "Reeling in", chipClass: "pixel-chip--done", title: "Focus", announce: "Time is up" },
  "rest-ready": { chip: "Break time", chipClass: "pixel-chip--rest", title: "Break", announce: "Focus block complete" },
  rest: { chip: "On a break", chipClass: "pixel-chip--rest", title: "Break", announce: "Break started" },
  "rest-paused": { chip: "Paused", chipClass: "pixel-chip--paused", title: "Break", announce: "Break paused" },
  closed: { chip: "Closed", chipClass: "", title: "Focus", announce: "Timer closed" },
};

const APP_TITLE = "Fisher Timer";

/**
 * One participant's study timer (UC-05): pick a focus length, run it
 * (pause / resume / stop / reset), then pick a break or skip it. Work and
 * break look different everywhere (phase track, chip, progress colour) so
 * the user always knows which one is running.
 *
 * Keyboard: Space runs the primary action, R resets, S stops.
 */
export function TimerPanel({ owner }: { owner: TimerOwner }) {
  const { timer, view, remainingMs, progress, error, completed, dismissCompleted, run, refresh } = useStudyTimer(owner);
  const [workChoice, setWorkChoice] = useState<number | null>(null);
  const [restChoice, setRestChoice] = useState<number | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);

  const copy = COPY[view];
  const loaded = timer !== null;
  const work = timer ? clampMinutes(workChoice ?? timer.work_minutes, timer.min_work_minutes, timer.max_work_minutes) : 0;
  const rest = timer ? clampMinutes(restChoice ?? timer.rest_minutes, timer.min_rest_minutes, timer.max_rest_minutes) : 0;

  // While choosing, the clock previews the chosen length.
  const clockMs = view === "ready" ? work * 60_000 : view === "rest-ready" ? rest * 60_000 : remainingMs;
  const clock = loaded ? formatClock(clockMs) : "--:--";
  const isRest = timer?.phase === "REST";
  const active = view === "focus" || view === "focus-paused" || view === "rest" || view === "rest-paused";
  const paused = view === "focus-paused" || view === "rest-paused";

  const primary: { label: string; action: TimerAction; options?: StartOptions } | null = (() => {
    switch (view) {
      case "ready":
        return { label: "Start focus", action: "start", options: { phase: "WORK", minutes: work } };
      case "rest-ready":
        return { label: "Start break", action: "start", options: { phase: "REST", minutes: rest } };
      case "focus":
      case "rest":
        return { label: "Pause", action: "pause" };
      case "focus-paused":
      case "rest-paused":
        return { label: "Resume", action: "resume" };
      default:
        return null;
    }
  })();

  const TAB_LABEL: Partial<Record<TimerView, string>> = {
    focus: `${clock} focus`,
    rest: `${clock} break`,
    "focus-paused": "Paused",
    "rest-paused": "Paused",
    "rest-ready": "Break time",
  };
  useDocumentTitle(loaded && TAB_LABEL[view] ? `${TAB_LABEL[view]} · ${APP_TITLE}` : APP_TITLE);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (!loaded || e.repeat || e.metaKey || e.ctrlKey || e.altKey) return;
      // Leave keys alone while a control has focus: Space already presses it.
      const target = e.target as HTMLElement;
      if (target.closest("button, a, input, textarea, select, label, [contenteditable]")) return;
      if (e.code === "Space" && primary) {
        e.preventDefault();
        run(primary.action, primary.options);
      } else if ((e.key === "r" || e.key === "R") && active) {
        run("reset");
      } else if ((e.key === "s" || e.key === "S") && (view === "focus" || view === "focus-paused")) {
        run("stop");
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [loaded, primary, active, view, run]);

  // The block being chosen or run: one after those completed in this stay.
  const block = (timer?.current_cycle ?? 0) + 1;
  const runningMinutes = Math.round((timer?.duration_seconds ?? 0) / 60);
  const caption = (() => {
    if (!timer) return "Connecting to the timer…";
    switch (view) {
      case "ready":
        return work < MIN_REWARD_MINUTES
          ? `Under ${MIN_REWARD_MINUTES} min is too short to catch a fish.`
          : `Block ${block}. Every 15 minutes of focus earns a catch.`;
      case "focus":
      case "focus-paused":
        return `Block ${block} · ${runningMinutes} min focus${runningMinutes < MIN_REWARD_MINUTES ? " · too short for a fish" : ""}`;
      case "finishing":
        return "Time! Reeling in your catch…";
      case "rest-ready":
        return "Nice block! Pick a break, or skip it and keep going.";
      case "rest":
      case "rest-paused":
        return `${runningMinutes} min break · breaks don't catch fish, they keep you fishing`;
      case "closed":
        return "This timer closed when you left the room.";
    }
  })();

  return (
    <PixelPanel
      as="section"
      aria-labelledby="timer-heading"
      className="pixel-timer w-full text-center"
      data-phase={isRest ? "REST" : "WORK"}
    >
      <div className="flex items-center justify-between gap-3">
        <h2 id="timer-heading" className="font-display text-2xl leading-none">
          {copy.title} timer
        </h2>
        <div className="flex items-center gap-2">
          <span className={cx("pixel-chip", copy.chipClass)}>{copy.chip}</span>
          <PixelButton
            variant="ghost"
            icon
            small
            aria-label="Timer settings"
            title="Default lengths"
            onClick={() => setSettingsOpen(true)}
            disabled={!loaded}
          >
            <Gear aria-hidden="true" />
          </PixelButton>
        </div>
      </div>

      <ol className="pixel-phase mt-5" aria-label="Cycle">
        <li className={cx("pixel-phase__step", !isRest && "pixel-phase__step--work")} aria-current={!isRest ? "step" : undefined}>
          Focus
        </li>
        <li className={cx("pixel-phase__step", isRest && "pixel-phase__step--rest")} aria-current={isRest ? "step" : undefined}>
          Break
        </li>
      </ol>

      <p
        role="timer"
        aria-label={`${clock} ${active ? "remaining" : "selected"}`}
        data-view={paused ? "paused" : view === "finishing" ? "done" : view}
        className="pixel-clock mt-6"
      >
        {clock}
      </p>

      <div
        className="pixel-progress mt-5"
        data-view={paused ? "paused" : view === "finishing" ? "done" : view}
        data-phase={isRest ? "REST" : "WORK"}
        role="progressbar"
        aria-label={isRest ? "Break progress" : "Focus progress"}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={active || view === "finishing" ? Math.round(progress * 100) : 0}
      >
        <div
          className="pixel-progress__fill"
          style={{ transform: `scaleX(${active || view === "finishing" ? progress : 0})` }}
        />
      </div>

      <p className="mt-3 min-h-[3.5em] text-sm text-bark">{caption}</p>

      {timer && view === "ready" && (
        <div className="mt-2 text-left">
          <DurationPicker
            label="Focus length"
            value={work}
            onChange={setWorkChoice}
            presets={WORK_PRESETS}
            min={timer.min_work_minutes}
            max={timer.max_work_minutes}
          />
        </div>
      )}
      {timer && view === "rest-ready" && (
        <div className="mt-2 text-left">
          <DurationPicker
            label="Break length"
            value={rest}
            onChange={setRestChoice}
            presets={REST_PRESETS}
            min={timer.min_rest_minutes}
            max={timer.max_rest_minutes}
          />
        </div>
      )}

      {view !== "closed" && (
        <div className="mt-6 flex flex-wrap justify-center gap-3">
          <PixelButton
            className="pixel-timer__primary"
            onClick={() => primary && run(primary.action, primary.options)}
            disabled={!loaded || !primary}
          >
            {primary?.label ?? "…"}
          </PixelButton>
          {(view === "focus" || view === "focus-paused") && (
            <>
              <PixelButton variant="ghost" onClick={() => run("reset")}>
                Reset
              </PixelButton>
              <PixelButton variant="danger" onClick={() => run("stop")}>
                Stop
              </PixelButton>
            </>
          )}
          {(view === "rest-ready" || view === "rest" || view === "rest-paused") && (
            <PixelButton variant="ghost" onClick={() => run("skip-rest")}>
              Skip break
            </PixelButton>
          )}
        </div>
      )}

      {paused && timer && (
        <p className="mt-4 text-xs text-bark">
          A pause longer than {timer.max_pause_minutes} minutes ends the {isRest ? "break" : "block without a catch"}.
        </p>
      )}

      {view === "closed" && <PixelAlert tone="warn" className="mt-5 text-left">This timer is closed.</PixelAlert>}

      <p className="mt-5 hidden items-center justify-center gap-2 text-xs text-bark md:flex">
        <kbd className="pixel-kbd">Space</kbd> {primary?.label.toLowerCase() ?? "—"}
        {active && (
          <>
            <span aria-hidden="true">·</span>
            <kbd className="pixel-kbd">R</kbd> reset
          </>
        )}
        {(view === "focus" || view === "focus-paused") && (
          <>
            <span aria-hidden="true">·</span>
            <kbd className="pixel-kbd">S</kbd> stop
          </>
        )}
      </p>

      {/* Screen readers hear state changes, not every tick. */}
      <p className="sr-only" aria-live="polite">
        {loaded ? copy.announce : ""}
      </p>

      {error && <PixelAlert className="mt-5 text-left">{error}</PixelAlert>}

      {timer && (
        <PixelModal open={settingsOpen} onClose={() => setSettingsOpen(false)} title="Timer defaults">
          <TimerSettingsForm
            timer={timer}
            userId={owner.userId}
            sessionId={owner.sessionId}
            onSaved={() => {
              setWorkChoice(null);
              setRestChoice(null);
              void refresh();
            }}
            onCancel={() => setSettingsOpen(false)}
          />
        </PixelModal>
      )}

      <CatchReveal
        block={completed}
        userId={owner.userId}
        onClose={dismissCompleted}
        breakMinutes={rest}
        onStartBreak={view === "rest-ready" ? () => run("start", { phase: "REST", minutes: rest }) : undefined}
        onSkipBreak={view === "rest-ready" ? () => run("skip-rest") : undefined}
      />
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
