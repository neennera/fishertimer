// Timer data layer (UC-05). The only file that talks to the backend for the
// timer; components call these functions and never use fetch directly.
//
// Requests go browser -> Next.js rewrite -> API Gateway (REST) -> Study Timer
// (gRPC). The gateway answers in the shape below.

import { apiFetch } from '../../lib/api-client';

export type TimerStatus = 'STOPPED' | 'RUNNING' | 'PAUSED';
export type TimerPhase = 'WORK' | 'REST';

export interface TimerState {
  session_id: string;
  user_id: string;
  status: TimerStatus;
  phase: TimerPhase;
  work_minutes: number;
  rest_minutes: number;
  current_cycle: number;
  last_updated: string;
  /** Full length of the current phase. */
  duration_seconds: number;
  /** Time left, computed by the server when it answered. */
  remaining_seconds: number;
}

export interface TimerOwner {
  sessionId: string;
  userId: string;
}

export type TimerAction = 'start' | 'pause' | 'resume' | 'reset';

export function getTimer({ sessionId, userId }: TimerOwner): Promise<TimerState> {
  const query = new URLSearchParams({ session_id: sessionId, user_id: userId });
  return apiFetch<TimerState>('timer', `state?${query}`);
}

export function runTimerAction(action: TimerAction, { sessionId, userId }: TimerOwner): Promise<TimerState> {
  return apiFetch<TimerState>('timer', action, {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, user_id: userId }),
  });
}
