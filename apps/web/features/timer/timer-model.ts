// Client-side view of one timer. Pure functions only — no React, no fetch.
//
// The server owns the time. The client keeps a local "reading" of it — the
// last known state plus the moment it was taken — so the display can move
// every frame, and so a button press can show its result immediately
// (optimistically) before the server confirms it.

import type { TimerAction, TimerState } from './timer.api';

/** A timer state together with when it was true, in performance.now() ms. */
export interface TimerReading {
  state: TimerState;
  /** Time left in the phase at `at`, in milliseconds. */
  remainingMs: number;
  at: number;
}

export type TimerView = 'ready' | 'running' | 'paused' | 'done';

export function readingFrom(state: TimerState, at: number): TimerReading {
  return { state, remainingMs: state.remaining_seconds * 1000, at };
}

/** Milliseconds left at `now`: frozen unless the timer is running. */
export function remainingAt(reading: TimerReading, now: number): number {
  if (reading.state.status !== 'RUNNING') return reading.remainingMs;
  return Math.max(reading.remainingMs - (now - reading.at), 0);
}

/** A running timer with nothing left is finished, not running. */
export function viewOf(reading: TimerReading, now: number): TimerView {
  switch (reading.state.status) {
    case 'PAUSED':
      return 'paused';
    case 'RUNNING':
      return remainingAt(reading, now) > 0 ? 'running' : 'done';
    default:
      return 'ready';
  }
}

/** The primary action for each view: the one big button. */
export const PRIMARY_ACTION: Record<TimerView, TimerAction> = {
  ready: 'start',
  running: 'pause',
  paused: 'resume',
  done: 'start',
};

/**
 * What the server will answer for `action`, predicted locally. Mirrors the
 * rules in services/study-timer/internal/domain/entity.go; the server's real
 * answer replaces it moments later. Returns null for a move the server would
 * reject, so the UI never shows a state that will be rolled back.
 */
export function predict(action: TimerAction, reading: TimerReading, now: number): TimerReading | null {
  const view = viewOf(reading, now);
  const { state } = reading;
  const workMs = state.work_minutes * 60_000;

  switch (action) {
    case 'start':
      if (view === 'running' || view === 'paused') return null;
      return {
        state: { ...state, status: 'RUNNING', phase: 'WORK', duration_seconds: workMs / 1000 },
        remainingMs: workMs,
        at: now,
      };
    case 'pause':
      if (view !== 'running') return null;
      return { state: { ...state, status: 'PAUSED' }, remainingMs: remainingAt(reading, now), at: now };
    case 'resume':
      if (view !== 'paused') return null;
      return { state: { ...state, status: 'RUNNING' }, remainingMs: reading.remainingMs, at: now };
    case 'reset':
      if (view === 'ready' && reading.remainingMs === workMs) return null;
      return {
        state: { ...state, status: 'STOPPED', phase: 'WORK', current_cycle: 0, duration_seconds: workMs / 1000 },
        remainingMs: workMs,
        at: now,
      };
  }
}

/** mm:ss, rounded up like a countdown: 1499.2s left reads 25:00. */
export function formatClock(remainingMs: number): string {
  const total = Math.ceil(remainingMs / 1000);
  const minutes = Math.floor(total / 60);
  const seconds = total % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}
