'use client';

import type { RankEntry } from '../types';
import { cx } from '../../../lib/cx';

const RANK_BADGES: Record<number, { medal: string; border: string; bg?: string; label: string }> = {
  1: { medal: '🥇', border: 'var(--color-amber)', bg: 'rgba(224, 138, 71, 0.12)', label: 'Champion' },
  2: { medal: '🥈', border: '#b0bec5', bg: 'rgba(176, 190, 197, 0.12)', label: 'Runner-up' },
  3: { medal: '🥉', border: '#cd7f32', bg: 'rgba(205, 127, 50, 0.12)', label: 'Third Place' },
};

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
      <div className="pixel-alert pixel-alert--warn text-center py-6" role="status">
        <div className="text-2xl mb-1">🎣</div>
        <div className="font-numeric text-base mb-1">No Catches This Period Yet</div>
        <p className="font-body text-xs opacity-80">
          Complete a study session to earn rewards and climb to the top of the leaderboard!
        </p>
      </div>
    );
  }

  return (
    <div role="table" aria-label="Leaderboard rankings" className="flex flex-col gap-2">
      {/* Header row */}
      <div
        role="row"
        className="flex items-center gap-3 px-4 py-2 border-b border-[var(--color-rule)]"
        style={{ color: 'var(--color-bark)', fontSize: 'calc(var(--px) * 3.6)' }}
        aria-hidden="true"
      >
        <span className="w-12 text-center font-label uppercase tracking-widest">Rank</span>
        <span className="flex-1 font-label uppercase tracking-widest pl-2">Angler</span>
        <span className="w-28 text-right font-label uppercase tracking-widest">Catches</span>
      </div>

      {/* Data rows */}
      <ol className="flex flex-col gap-2.5" role="rowgroup">
        {entries.map((entry) => {
          const isMe = currentUserId != null && entry.user_id === currentUserId;
          const podium = RANK_BADGES[entry.rank];
          const initials = (entry.display_name || entry.user_id)
            .split(' ')
            .map((p) => p[0])
            .join('')
            .slice(0, 2)
            .toUpperCase();

          return (
            <li
              key={entry.user_id}
              role="row"
              aria-label={`Rank ${entry.rank}: ${entry.display_name}, ${entry.reward_count} rewards`}
              className={cx(
                'pixel-panel flex items-center gap-3 px-4 py-3 transition-transform duration-75',
                isMe && 'leaderboard-row--me',
              )}
              style={
                isMe
                  ? {
                      background: 'var(--color-lake)',
                      color: '#ffffff',
                      boxShadow: `inset 0 0 0 var(--px) var(--color-lake-dp), inset 0 calc(var(--px) * -2) 0 var(--color-lake-dp)`,
                    }
                  : podium
                    ? {
                        background: podium.bg,
                        boxShadow: `inset 0 0 0 var(--px) ${podium.border}, inset 0 calc(var(--px) * -1) 0 rgba(107, 75, 53, 0.15)`,
                      }
                    : undefined
              }
            >
              {/* Rank number / medal */}
              <div
                className="w-12 flex flex-col items-center justify-center font-numeric shrink-0"
                role="cell"
                aria-label={`Rank ${entry.rank}`}
              >
                {podium ? (
                  <span
                    className="text-2xl leading-none"
                    title={podium.label}
                  >
                    {podium.medal}
                  </span>
                ) : (
                  <span
                    className="font-numeric"
                    style={{
                      fontSize: 'calc(var(--px) * 6)',
                      lineHeight: 1,
                      color: isMe ? '#fff' : 'var(--color-muted)',
                    }}
                  >
                    #{entry.rank}
                  </span>
                )}
              </div>

              {/* Angler Avatar & Name */}
              <div className="flex-1 flex items-center gap-3 min-w-0" role="cell">
                <span
                  className="w-9 h-9 shrink-0 flex items-center justify-center font-label font-bold text-xs"
                  style={{
                    clipPath: 'var(--pixclip)',
                    background: isMe
                      ? 'var(--color-lake-dp)'
                      : podium
                        ? podium.border
                        : 'var(--color-cream-2)',
                    color: isMe || podium ? '#ffffff' : 'var(--color-ink)',
                  }}
                  aria-hidden="true"
                >
                  {initials}
                </span>

                <div className="flex flex-col min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span
                      className="font-numeric text-base tracking-wide truncate"
                      style={{ fontSize: 'calc(var(--px) * 5.2)' }}
                    >
                      {entry.display_name || entry.user_id}
                    </span>
                    {isMe && (
                      <span
                        className="pixel-badge"
                        style={{
                          background: 'var(--color-lake-dp)',
                          color: '#fff',
                          fontSize: 'calc(var(--px) * 2.8)',
                          padding: 'calc(var(--px) * 1) calc(var(--px) * 2.5)',
                        }}
                        aria-label="Your Position"
                      >
                        YOU
                      </span>
                    )}
                    {entry.rank === 1 && !isMe && (
                      <span
                        className="pixel-badge"
                        style={{
                          background: 'var(--color-amber)',
                          color: '#fff',
                          fontSize: 'calc(var(--px) * 2.5)',
                          padding: 'calc(var(--px) * 0.8) calc(var(--px) * 2)',
                        }}
                      >
                        LEADER
                      </span>
                    )}
                  </div>
                  <span
                    className="text-xs truncate font-label opacity-75"
                    style={{ fontSize: 'calc(var(--px) * 2.8)' }}
                  >
                    {isMe ? 'Your registered profile' : `Fisher ID: ${entry.user_id}`}
                  </span>
                </div>
              </div>

              {/* Reward count */}
              <div
                className="w-28 text-right flex items-center justify-end gap-1.5 shrink-0"
                role="cell"
                aria-label={`${entry.reward_count} rewards`}
              >
                <span className="text-xl leading-none" aria-hidden="true">
                  🐟
                </span>
                <span
                  className="font-numeric"
                  style={{ fontSize: 'calc(var(--px) * 6.5)', lineHeight: 1 }}
                >
                  {entry.reward_count}
                </span>
                <span
                  className="text-xs font-label opacity-75 hidden sm:inline"
                  style={{ fontSize: 'calc(var(--px) * 2.8)' }}
                >
                  fish
                </span>
              </div>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
