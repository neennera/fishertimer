// Session summary shown on leave (UC-03 step 5): completed cycles, total
// focus time and the rewards earned during this stay.
//
// No service keeps per-room statistics, so the summary is put together here:
// - cycles and focus time come from the user's own timer for this room, read
//   just before leaving (Study Timer counts the work cycles completed during
//   this stay: current_cycle and focus_seconds);
// - rewards are the user's catches from Reward with awarded_at inside the
//   stay (joined_at .. left_at).

import type { TimerState } from '../timer/timer.api';
import type { RewardRarity } from '../../lib/profile-types';

export interface CaughtFish {
  name: string;
  rarity: RewardRarity;
  asset_url: string;
  score_value: number;
  count: number;
}

export interface SessionSummary {
  roomName: string;
  /** How long the user was in the room. */
  stayMinutes: number;
  completedCycles: number;
  focusMinutes: number;
  /** null when Reward could not be reached. */
  fish: CaughtFish[] | null;
  /** True when the user left during a work cycle, which earns nothing (UC-03 S-1). */
  forfeitedCycle: boolean;
}

/** One entry of GET /api/reward/rewards?user_id= */
interface UnlockedReward {
  item_name: string;
  rarity: string;
  asset_url: string;
  score_value: number;
  awarded_at: string;
}

const RARITY_ORDER: RewardRarity[] = ['LEGENDARY', 'EPIC', 'RARE', 'UNCOMMON', 'COMMON'];

/** True while a work cycle is running or paused: leaving would forfeit it. */
export function workInProgress(timer: TimerState | null): boolean {
  if (!timer) return false;
  return timer.state === 'WORK_PAUSED' || (timer.state === 'WORK_RUNNING' && timer.remaining_seconds > 0);
}

/** The user's catches between `from` and `to`, grouped by fish. */
export async function fishCaughtBetween(userId: string, from: Date, to: Date): Promise<CaughtFish[] | null> {
  try {
    const res = await fetch(`/api/reward/rewards?${new URLSearchParams({ user_id: userId })}`, {
      credentials: 'same-origin',
    });
    if (!res.ok) return null;
    const rewards = ((await res.json()) as UnlockedReward[] | null) ?? [];
    const byName = new Map<string, CaughtFish>();
    for (const r of rewards) {
      const at = new Date(r.awarded_at).getTime();
      if (at < from.getTime() || at > to.getTime()) continue;
      const fish = byName.get(r.item_name);
      if (fish) fish.count += 1;
      else
        byName.set(r.item_name, {
          name: r.item_name,
          rarity: (r.rarity?.toUpperCase() as RewardRarity) ?? 'COMMON',
          asset_url: r.asset_url,
          score_value: r.score_value,
          count: 1,
        });
    }
    return [...byName.values()].sort(
      (a, b) => RARITY_ORDER.indexOf(a.rarity) - RARITY_ORDER.indexOf(b.rarity) || b.count - a.count,
    );
  } catch {
    return null;
  }
}

export async function buildSummary(args: {
  userId: string;
  roomName: string;
  joinedAt: string;
  leftAt: string;
  timer: TimerState | null;
}): Promise<SessionSummary> {
  const joined = new Date(args.joinedAt);
  const left = args.leftAt ? new Date(args.leftAt) : new Date();
  return {
    roomName: args.roomName,
    stayMinutes: Math.max(0, Math.round((left.getTime() - joined.getTime()) / 60_000)),
    completedCycles: args.timer?.current_cycle ?? 0,
    focusMinutes: Math.round((args.timer?.focus_seconds ?? 0) / 60),
    // A little slack after left_at: a cycle that completed just before
    // leaving may be awarded a moment later.
    fish: await fishCaughtBetween(args.userId, joined, new Date(left.getTime() + 30_000)),
    forfeitedCycle: workInProgress(args.timer),
  };
}
