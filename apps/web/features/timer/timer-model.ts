// Client-side view of one timer. Pure functions only — no React, no fetch.
//
// The server owns the time. The client keeps a local "reading" of it — the
// last known state plus the moment it was taken — so the display can move
// every frame, and so a button press can show its result immediately
// (optimistically) before the server confirms it.

import type { StartOptions, TimerAction, TimerState, TimerStateName } from './timer.api';

/** A timer state together with when it was true, in performance.now() ms. */
export interface TimerReading {
  state: TimerState;
  /** Time left in the phase at `at`, in milliseconds. */
  remainingMs: number;
  at: number;
}

/**
 * What the panel shows. `finishing` is a running phase whose countdown hit
 * zero, waiting for the server to complete it.
 */
export type TimerView =
  | 'ready'
  | 'focus'
  | 'focus-paused'
  | 'rest-ready'
  | 'rest'
  | 'rest-paused'
  | 'finishing'
  | 'closed';

export function readingFrom(state: TimerState, at: number): TimerReading {
  return { state, remainingMs: state.remaining_seconds * 1000, at };
}

/** Milliseconds left at `now`: frozen unless the timer is running. */
export function remainingAt(reading: TimerReading, now: number): number {
  if (reading.state.status !== 'RUNNING') return reading.remainingMs;
  return Math.max(reading.remainingMs - (now - reading.at), 0);
}

const VIEW_OF_STATE: Record<TimerStateName, TimerView> = {
  READY: 'ready',
  WORK_RUNNING: 'focus',
  WORK_PAUSED: 'focus-paused',
  READY_FOR_REST: 'rest-ready',
  REST_RUNNING: 'rest',
  REST_PAUSED: 'rest-paused',
  FINALIZED: 'closed',
};

export function viewOf(reading: TimerReading, now: number): TimerView {
  const view = VIEW_OF_STATE[reading.state.state] ?? 'ready';
  if ((view === 'focus' || view === 'rest') && remainingAt(reading, now) === 0) return 'finishing';
  return view;
}

export function isActive(view: TimerView): boolean {
  return view === 'focus' || view === 'focus-paused' || view === 'rest' || view === 'rest-paused';
}

/**
 * What the server will answer for `action`, predicted locally. Mirrors the
 * rules in services/study-timer/internal/domain; the server's real answer
 * replaces it moments later. Returns null for a move the server would
 * refuse, so the UI never shows a state that will be rolled back.
 */
export function predict(
  action: TimerAction,
  reading: TimerReading,
  now: number,
  options: StartOptions = {},
): TimerReading | null {
  const view = viewOf(reading, now);
  const { state } = reading;
  const set = (name: TimerStateName, patch: Partial<TimerState>, remainingMs: number): TimerReading => ({
    state: { ...state, ...patch, state: name },
    remainingMs,
    at: now,
  });
  const readyAgain = () =>
    set('READY', { status: 'STOPPED', phase: 'WORK', duration_seconds: state.work_minutes * 60, cycle_id: '' }, state.work_minutes * 60_000);

  switch (action) {
    case 'start': {
      const phase = options.phase ?? 'WORK';
      if (phase === 'REST' && view !== 'rest-ready') return null;
      if (phase === 'WORK' && view !== 'ready' && view !== 'rest-ready') return null;
      const minutes = options.minutes || (phase === 'REST' ? state.rest_minutes : state.work_minutes);
      return set(
        phase === 'REST' ? 'REST_RUNNING' : 'WORK_RUNNING',
        { status: 'RUNNING', phase, duration_seconds: minutes * 60, paused_total_seconds: 0 },
        minutes * 60_000,
      );
    }
    case 'pause':
      if (view !== 'focus' && view !== 'rest') return null;
      return set(view === 'focus' ? 'WORK_PAUSED' : 'REST_PAUSED', { status: 'PAUSED' }, remainingAt(reading, now));
    case 'resume':
      if (view !== 'focus-paused' && view !== 'rest-paused') return null;
      return set(view === 'focus-paused' ? 'WORK_RUNNING' : 'REST_RUNNING', { status: 'RUNNING' }, reading.remainingMs);
    case 'stop':
      if (!isActive(view)) return null;
      return readyAgain();
    case 'reset':
      if (!isActive(view)) return null;
      // Back to the full length, held paused until the user resumes.
      return set(
        state.phase === 'REST' ? 'REST_PAUSED' : 'WORK_PAUSED',
        { status: 'PAUSED', paused_total_seconds: 0 },
        state.duration_seconds * 1000,
      );
    case 'skip-rest':
      if (view !== 'rest-ready' && view !== 'rest' && view !== 'rest-paused') return null;
      return readyAgain();
    case 'complete':
      return null; // only the server decides; never predicted
  }
}

/** mm:ss, rounded up like a countdown: 1499.2s left reads 25:00. */
export function formatClock(remainingMs: number): string {
  const total = Math.ceil(remainingMs / 1000);
  const minutes = Math.floor(total / 60);
  const seconds = total % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}

/** Clamps a duration choice into the server's allowed range. */
export function clampMinutes(minutes: number, min: number, max: number): number {
  return Math.min(Math.max(Math.round(minutes), min), max);
}
