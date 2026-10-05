'use client';

import { useEffect, useState } from 'react';
import type { RankEntry } from '../types';
import { cx } from '../../../lib/cx';
import { PixelButton } from '../../../components/ui/PixelButton';

export const LEADERBOARD_PAGE_SIZE = 5;

const PODIUM = { border: 'var(--color-amber)', bg: 'rgba(224, 138, 71, 0.12)' };
const RANK_BADGES: Record<number, { medal: string; border: string; bg?: string; label: string }> = {
  1: { medal: '🥇', ...PODIUM, label: 'Champion' },
  2: { medal: '🥈', ...PODIUM, label: 'Runner-up' },
  3: { medal: '🥉', ...PODIUM, label: 'Third Place' },
};

export interface LeaderboardTableProps {
  entries: RankEntry[];
  /** user_id of the currently signed-in user. */
  currentUserId?: string | null;
  /** Shown on the pinned badge when the user has no ranking yet. */
  currentUserName?: string | null;
}

/**
 * LeaderboardTable — read-only ranking list (UC-08), 5 rows per page.
 *
 * - The signed-in user's row gets the blue background only while it is on the
 *   page being shown.
 * - Below a divider, the user's own position is always pinned to the bottom.
 * - E-1 empty state: renders a friendly message, not an error screen.
 */
export function LeaderboardTable({ entries, currentUserId, currentUserName }: LeaderboardTableProps) {

  const [page, setPage] = useState(0);
  const pageCount = Math.max(1, Math.ceil(entries.length / LEADERBOARD_PAGE_SIZE));

  // The list can shrink (period switch / refresh); never sit past the end.
  useEffect(() => {
    setPage((p) => Math.min(p, pageCount - 1));
  }, [pageCount]);

  const current = Math.min(page, pageCount - 1);
  const visible = entries.slice(current * LEADERBOARD_PAGE_SIZE, (current + 1) * LEADERBOARD_PAGE_SIZE);
  const me = currentUserId != null ? entries.find((e) => e.user_id === currentUserId) : undefined;

  return (
    <div role="table" aria-label="Leaderboard rankings" className="flex flex-col gap-2">
      {entries.length === 0 ? (
        <div className="pixel-alert pixel-alert--warn text-center py-6" role="status">
          <div className="text-2xl mb-1">🎣</div>
          <div className="font-numeric text-base mb-1">No Catches This Period Yet</div>
          <p className="font-body text-xs opacity-80">
            Complete a study session to earn rewards and climb to the top of the leaderboard!
          </p>
        </div>
      ) : (
        <>
          {/* Header row */}
          <div
            role="row"
            className="flex items-center gap-3 px-4 py-2 border-b border-[var(--color-rule)]"
            style={{ color: 'var(--color-bark)', fontSize: 'calc(var(--px) * 3.6)' }}
            aria-hidden="true"
          >
            <span className="w-12 text-center font-label uppercase tracking-widest">Rank</span>
            <span className="flex-1 font-label uppercase tracking-widest pl-2">Users</span>
            <span className="w-28 text-right font-label uppercase tracking-widest">Score</span>
          </div>

          <ol className="flex flex-col gap-2.5" role="rowgroup">
            {visible.map((entry) => (
              <Row key={entry.user_id} entry={entry} isMe={entry.user_id === currentUserId} />
            ))}
          </ol>
        </>
      )}

      {/* Pagination: always shown, even with a single page */}
      <nav className="flex items-center justify-between gap-3 pt-1" aria-label="Leaderboard pages">
        <PixelButton
          variant="ghost"
          onClick={() => setPage(current - 1)}
          disabled={current === 0}
          aria-label="Previous page"
        >
          &lt;
        </PixelButton>
        <span className="font-label text-xs text-bark" aria-live="polite">
          Page {current + 1} / {pageCount}
        </span>
        <PixelButton
          variant="ghost"
          onClick={() => setPage(current + 1)}
          disabled={current >= pageCount - 1}
          aria-label="Next page"
        >
          &gt;
        </PixelButton>
      </nav>

      {/* Divider + the signed-in user's own position (also when unranked) */}
      {currentUserId != null && (
        <>
          <hr
            className="my-1 border-0 border-t-2 border-dashed border-[var(--color-rule)]"
            aria-hidden="true"
          />
          <ol className="flex flex-col" aria-label="Your ranking">
            {me ? (
              <Row entry={me} isMe />
            ) : (
              <Row
                isMe
                entry={{
                  user_id: currentUserId,
                  display_name: currentUserName ?? currentUserId,
                  rank: 0,
                  reward_count: 0,
                  score: 0,
                  period: 'monthly',
                }}
              />
            )}
          </ol>
        </>
      )}
    </div>
  );

}

function Row({ entry, isMe }: { entry: RankEntry; isMe: boolean }) {
  const podium = RANK_BADGES[entry.rank];
  const initials = (entry.display_name || entry.user_id)
    .split(' ')
    .map((p) => p[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  return (
    <li
      role="row"
      aria-label={`Rank ${entry.rank}: ${entry.display_name}, ${entry.score} points`}
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
          <span className="text-2xl leading-none" title={podium.label}>
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
            {entry.rank > 0 ? `#${entry.rank}` : '—'}
          </span>
        )}
      </div>

      {/* Avatar & name */}
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
            {isMe ? (entry.rank > 0 ? 'Your ranking' : 'Your ranking · Unranked') : `Fisher ID: ${entry.user_id}`}
          </span>
        </div>
      </div>

      {/* Score */}
      <div
        className="w-28 text-right flex items-center justify-end gap-1.5 shrink-0"
        role="cell"
        aria-label={`${entry.score} points from ${entry.reward_count} fish`}
        title={`${entry.reward_count} fish`}
      >
        <span className="text-xl leading-none" aria-hidden="true">
          🐟
        </span>
        <span className="font-numeric" style={{ fontSize: 'calc(var(--px) * 6.5)', lineHeight: 1 }}>
          {entry.score}
        </span>
        <span
          className="text-xs font-label opacity-75 hidden sm:inline"
          style={{ fontSize: 'calc(var(--px) * 2.8)' }}
        >
          pts
        </span>
      </div>
    </li>
  );
}
