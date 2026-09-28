'use client';

import type { RankEntry } from '../types';
import { cx } from '../../../lib/cx';

const RANK_MEDALS: Record<number, string> = {
  1: '🥇',
  2: '🥈',
  3: '🥉',
};

const RARITY_COLOURS: Record<string, string> = {};

export interface LeaderboardTableProps {
  entries: RankEntry[];
  /** user_id of the currently signed-in user — their row is highlighted (UC-08 NF-6). */
  currentUserId?: string | null;
}

/**
 * LeaderboardTable — read-only ranking list (UC-08).
 *
 * Rules:
 * - Rankings are read-only (no mutation actions).
 * - Current user's row is highlighted so they can locate their position (NF-6 / E-2).
 * - E-1 empty state: renders a friendly message, not an error screen.
 */
export function LeaderboardTable({ entries, currentUserId }: LeaderboardTableProps) {
  if (entries.length === 0) {
    return (
      <div className="pixel-alert pixel-alert--warn" role="status">
        🎣 No catches recorded for this period yet — be the first!
      </div>
    );
  }

  return (
    <div role="table" aria-label="Leaderboard rankings">
      {/* Header row */}
      <div
        role="row"
        className="flex items-center gap-3 px-3 py-2 mb-1"
        style={{ color: 'var(--color-bark)', fontSize: 'calc(var(--px) * 3.5)' }}
        aria-hidden="true"
      >
        <span className="w-10 text-center font-label uppercase tracking-widest">#</span>
        <span className="flex-1 font-label uppercase tracking-widest">Angler</span>
        <span className="w-20 text-right font-label uppercase tracking-widest">Rewards</span>
      </div>

      {/* Data rows */}
      <ol className="flex flex-col gap-2" role="rowgroup">
        {entries.map((entry) => {
          const isMe = currentUserId != null && entry.user_id === currentUserId;
          const medal = RANK_MEDALS[entry.rank] ?? null;

          return (
            <li
              key={entry.user_id}
              role="row"
              aria-label={`Rank ${entry.rank}: ${entry.display_name}, ${entry.reward_count} rewards`}
              className={cx(
                'pixel-panel flex items-center gap-3',
                isMe && 'leaderboard-row--me',
              )}
              style={
                isMe
                  ? {
                      background: 'var(--color-lake)',
                      color: '#fff',
                      boxShadow: `inset 0 0 0 var(--px) var(--color-lake-dp), inset 0 calc(var(--px) * -1) 0 var(--color-lake-dp)`,
                    }
                  : undefined
              }
            >
              {/* Rank number / medal */}
              <span
                className="w-10 text-center font-numeric"
                style={{ fontSize: 'calc(var(--px) * 6)', lineHeight: 1 }}
                role="cell"
                aria-label={`Rank ${entry.rank}`}
              >
                {medal ?? entry.rank}
              </span>

              {/* Display name */}
              <span className="flex-1 font-body" role="cell">
                {entry.display_name || entry.user_id}
                {isMe && (
                  <span
                    className="ml-2 pixel-badge"
                    style={{ background: 'var(--color-lake-dp)', verticalAlign: 'middle' }}
                    aria-label="You"
                  >
                    YOU
                  </span>
                )}
              </span>

              {/* Reward count */}
              <span
                className="w-20 text-right font-numeric"
                style={{ fontSize: 'calc(var(--px) * 7)', lineHeight: 1 }}
                role="cell"
                aria-label={`${entry.reward_count} rewards`}
              >
                🐟&nbsp;{entry.reward_count}
              </span>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
