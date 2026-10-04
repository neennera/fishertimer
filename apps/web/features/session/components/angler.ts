// How a participant is drawn on the dock, from their timer.

import type { TimerState } from '../../timer/timer.api';

export type AnglerState = 'focus' | 'paused' | 'caught' | 'rest' | 'idle';

export function anglerState(timer: TimerState | undefined): AnglerState {
  if (!timer) return 'idle';
  if (timer.status === 'PAUSED') return 'paused';
  if (timer.status !== 'RUNNING') return 'idle';
  if (timer.remaining_seconds > 0) return timer.phase === 'REST' ? 'rest' : 'focus';
  // The phase ran out and nobody has completed it yet.
  return timer.phase === 'WORK' ? 'caught' : 'idle';
}

export const ANGLER_COPY: Record<AnglerState, { chip: string; chipClass: string; mark: string }> = {
  focus: { chip: 'Focusing', chipClass: 'pixel-chip--running', mark: '' },
  paused: { chip: 'Paused', chipClass: 'pixel-chip--paused', mark: '' },
  caught: { chip: 'Block done', chipClass: 'pixel-chip--done', mark: '!' },
  rest: { chip: 'On a break', chipClass: 'pixel-chip--rest', mark: 'z' },
  idle: { chip: 'Ready', chipClass: '', mark: '' },
};

const ANGLER_COLORS = ['lake', 'rust', 'sage', 'amber', 'forest', 'lake-dp', 'oak'] as const;

/** A shirt colour per user, stable across reloads and devices. */
export function anglerColor(userId: string): string {
  const hash = [...userId].reduce((h, c) => (h * 31 + c.charCodeAt(0)) >>> 0, 7);
  return `var(--color-${ANGLER_COLORS[hash % ANGLER_COLORS.length]})`;
}
