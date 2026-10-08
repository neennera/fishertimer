'use client';

import { useEffect, useState } from 'react';
import type { RankEntry } from '../types';
import { cx } from '../../../lib/cx';
import { PixelButton } from '../../../components/ui/PixelButton';

export const LEADERBOARD_PAGE_SIZE = 5;

// Top-3 medal sprites (public/sprites/ui, 16x16; the rank number is drawn on the medal).
// Their row colours are .pixel-rank-row--podium in pixel.css.
const RANK_BADGES: Record<number, { medal: string; label: string }> = {
  1: { medal: '/sprites/ui/medal-gold.png', label: 'Champion' },
  2: { medal: '/sprites/ui/medal-silver.png', label: 'Runner-up' },
  3: { medal: '/sprites/ui/medal-bronze.png', label: 'Third Place' },
};

// Score icon and empty-state art, from public/sprites/fish (16x16, drawn at whole multiples).
const SCORE_FISH = '/sprites/fish/Goldfish.png';
const EMPTY_FISH = '/sprites/fish/Ghostfish.png';

export interface LeaderboardTableProps {
  entries: RankEntry[];
  /** user_id of the currently signed-in user. */
  currentUserId?: string | null;
  /** Shown on the pinned badge when the user has no ranking yet. */
  currentUserName?: string | null;
  /** Fallback score for the user if unranked or fetched from reward service. */
  currentUserScore?: number | null;
  /** Ranking period; switching periods resets to page 1 (0-indexed 0). */
  period?: string;
  /** Current user's avatar url if available. */
  currentUserAvatarUrl?: string | null;
}

/**
 * LeaderboardTable — read-only ranking list (UC-08), 5 rows per page.
 *
 * - The signed-in user's row gets the blue background only while it is on the
 *   page being shown.
 * - Below a divider, the user's own position is always pinned to the bottom.
 * - E-1 empty state: renders a friendly message, not an error screen.
 */
export function LeaderboardTable({
  entries,
  currentUserId,
  currentUserName,
  currentUserScore,
  period,
  currentUserAvatarUrl,
}: LeaderboardTableProps) {

  const [page, setPage] = useState(0);
  const pageCount = Math.max(1, Math.ceil(entries.length / LEADERBOARD_PAGE_SIZE));

  // When switching period (weekly, monthly, all-time), always reset to page 1
  useEffect(() => {
    setPage(0);
  }, [period]);

  // The list can shrink (period switch / refresh); never sit past the end.
  useEffect(() => {
    setPage((p) => Math.min(p, pageCount - 1));
  }, [pageCount]);

  const current = Math.min(page, pageCount - 1);
  const visible = entries.slice(current * LEADERBOARD_PAGE_SIZE, (current + 1) * LEADERBOARD_PAGE_SIZE);
  const meIndex = currentUserId != null ? entries.findIndex((e) => e.user_id === currentUserId) : -1;
  const me = meIndex !== -1 ? entries[meIndex] : undefined;

  const handleGoToMyPage = () => {
    if (meIndex !== -1) {
      const targetPage = Math.floor(meIndex / LEADERBOARD_PAGE_SIZE);
      setPage(targetPage);
    }
  };

  return (
    <div role="table" aria-label="Leaderboard rankings" className="flex flex-col gap-2">
      {entries.length === 0 ? (
        <div className="pixel-alert pixel-alert--warn text-center py-6" role="status">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={EMPTY_FISH}
            alt=""
            width={32}
            height={32}
            className="pixel-sprite mx-auto mb-2 opacity-60"
            style={{ width: 32, height: 32 }}
          />
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
            className="flex items-center gap-3 px-4 pt-2"
            style={{ color: 'var(--color-bark)', fontSize: 'calc(var(--px) * 3.6)' }}
            aria-hidden="true"
          >
            <span className="w-12 text-center font-label uppercase tracking-widest">Rank</span>
            <span className="flex-1 font-label uppercase tracking-widest pl-2">Users</span>
            <span className="w-28 text-right font-label uppercase tracking-widest">Score</span>
          </div>
          <hr className="pixel-rule" aria-hidden="true" />

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
          <hr className="pixel-rule pixel-rule--dashed my-1" aria-hidden="true" />
          <ol className="flex flex-col" aria-label="Your ranking">
            {me ? (
              <Row
                entry={me}
                isMe
                onClick={handleGoToMyPage}
                clickable={page !== Math.floor(meIndex / LEADERBOARD_PAGE_SIZE)}
              />
            ) : (
              <Row
                isMe
                entry={{
                  user_id: currentUserId,
                  display_name: currentUserName ?? currentUserId,
                  avatar_url: currentUserAvatarUrl ?? undefined,
                  rank: 0,
                  reward_count: 0,
                  score: currentUserScore ?? 0,
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

function Row({
  entry,
  isMe,
  onClick,
  clickable = false,
}: {
  entry: RankEntry;
  isMe: boolean;
  onClick?: () => void;
  clickable?: boolean;
}) {
  const [failedAvatar, setFailedAvatar] = useState(false);
  const podium = RANK_BADGES[entry.rank];
  const initials = (entry.display_name || entry.user_id)
    .split(' ')
    .map((p) => p[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  const showAvatar = Boolean(entry.avatar_url && !failedAvatar);

  return (
    <li
      role="row"
      aria-label={`Rank ${entry.rank}: ${entry.display_name}, ${entry.score} points`}
      onClick={clickable ? onClick : undefined}
      className={cx(
        'pixel-panel flex items-center gap-3 px-4 py-3 transition-transform duration-75',
        isMe ? 'pixel-rank-row--me' : podium && 'pixel-rank-row--podium',
        clickable && 'cursor-pointer hover:opacity-90 active:scale-[0.99]',
      )}
      title={clickable ? 'Click to jump to your page in the table' : undefined}
    >
      {/* Rank number / medal */}
      <div
        className="w-12 flex flex-col items-center justify-center font-numeric shrink-0"
        role="cell"
        aria-label={`Rank ${entry.rank}`}
      >
        {podium ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={podium.medal}
            alt=""
            title={podium.label}
            width={32}
            height={32}
            className="pixel-sprite"
            style={{ width: 32, height: 32 }}
          />
        ) : (
          <span
            className="pixel-rank-row__rank font-numeric"
            style={{ fontSize: 'calc(var(--px) * 6)', lineHeight: 1 }}
          >
            {entry.rank > 0 ? `#${entry.rank}` : '—'}
          </span>
        )}
      </div>

      {/* Avatar & name */}
      <div className="flex-1 flex items-center gap-3 min-w-0" role="cell">
        <span
          className="pixel-rank-row__avatar w-9 h-9 shrink-0 flex items-center justify-center font-label font-bold text-xs overflow-hidden"
          aria-hidden="true"
        >
          {showAvatar ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={entry.avatar_url!}
              alt=""
              className="w-full h-full object-cover"
              referrerPolicy="no-referrer"
              onError={() => setFailedAvatar(true)}
            />
          ) : (
            initials
          )}
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
                className="pixel-badge pixel-rank-row__badge pixel-rank-row__badge--you"
                aria-label="Your Position"
              >
                YOU
              </span>
            )}
            {entry.rank === 1 && !isMe && (
              <span className="pixel-badge pixel-rank-row__badge pixel-rank-row__badge--leader">
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
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={SCORE_FISH}
          alt=""
          aria-hidden="true"
          width={16}
          height={16}
          className="pixel-sprite"
          style={{ width: 16, height: 16 }}
        />
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
