// Timer data layer (UC-05). The only file that talks to the backend for the
// timer; components call these functions and never use fetch directly.
//
// Requests go browser -> Next.js rewrite -> API Gateway (REST) -> Study Timer
// (gRPC). The gateway acts as the signed-in user; user_id is only read when
// there is no session cookie (mock auth in development), so it is always sent.

import { apiFetch } from '../../lib/api-client';

export type TimerStatus = 'STOPPED' | 'RUNNING' | 'PAUSED';
export type TimerPhase = 'WORK' | 'REST';

/** Place in the UC-05 state machine. */
export type TimerStateName =
  | 'READY'
  | 'WORK_RUNNING'
  | 'WORK_PAUSED'
  | 'READY_FOR_REST'
  | 'REST_RUNNING'
  | 'REST_PAUSED'
  | 'FINALIZED';

export interface TimerState {
  session_id: string;
  user_id: string;
  state: TimerStateName;
  status: TimerStatus;
  phase: TimerPhase;
  /** The user's saved defaults (TimerSetting). */
  work_minutes: number;
  rest_minutes: number;
  /** Work cycles completed during this stay in the room. */
  current_cycle: number;
  focus_seconds: number;
  last_updated: string;
  /** Full length of the current (or next) phase. */
  duration_seconds: number;
  /** Time left, computed by the server when it answered. */
  remaining_seconds: number;
  cycle_id: string;
  started_at: string;
  paused_total_seconds: number;
  /** The most recent completed work cycle, for the catch reveal. */
  last_completed_cycle_id: string;
  last_completed_at: string;
  min_work_minutes: number;
  max_work_minutes: number;
  min_rest_minutes: number;
  max_rest_minutes: number;
  max_pause_minutes: number;
}

export interface TimerOwner {
  sessionId: string;
  userId: string;
}

export type TimerAction = 'start' | 'pause' | 'resume' | 'stop' | 'reset' | 'complete' | 'skip-rest';

export interface StartOptions {
  phase?: TimerPhase;
  /** 0 or omitted = the saved default for the phase. */
  minutes?: number;
}

/** Below this a completed work block catches nothing (UC-09 E-3). */
export const MIN_REWARD_MINUTES = 15;

export function getTimer({ sessionId, userId }: TimerOwner): Promise<TimerState> {
  const query = new URLSearchParams({ session_id: sessionId, user_id: userId });
  return apiFetch<TimerState>('timer', `state?${query}`);
}

export function runTimerAction(
  action: TimerAction,
  { sessionId, userId }: TimerOwner,
  options: StartOptions = {},
): Promise<TimerState> {
  return apiFetch<TimerState>('timer', action, {
    method: 'POST',
    body: JSON.stringify({
      session_id: sessionId,
      user_id: userId,
      phase: options.phase,
      duration_minutes: options.minutes ?? 0,
    }),
  });
}

/** Saves the default work / rest lengths (UpdateTimerSetting). */
export function saveTimerSettings(
  { sessionId, userId }: TimerOwner,
  workMinutes: number,
  restMinutes: number,
): Promise<TimerState> {
  return apiFetch<TimerState>('timer', 'settings', {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, user_id: userId, work_minutes: workMinutes, rest_minutes: restMinutes }),
  });
}

/** The user's settings and the allowed ranges, outside any room. */
export function getTimerSettings(userId: string): Promise<TimerState> {
  return getTimer({ sessionId: '', userId });
}
