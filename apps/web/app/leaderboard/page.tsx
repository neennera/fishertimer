'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import { Header } from '../../components/Header';
import { ParallaxScene } from '../../components/ui/ParallaxScene';
import { PixelPanel } from '../../components/ui/PixelPanel';
import { PixelAlert } from '../../components/ui/PixelAlert';
import { PixelButton } from '../../components/ui/PixelButton';
import { SessionUnavailableAlert, useSignedInUser } from '../../components/profile/ProfileShell';
import { SIGNIN_SCENE_LAYERS } from '../../lib/scenes/signin-scene';
import { getPublicProfile, getRewards } from '../../lib/profile';
import { PeriodTabs } from '../../features/leaderboard/components/PeriodTabs';
import { LeaderboardTable } from '../../features/leaderboard/components/LeaderboardTable';
import { useLeaderboard } from '../../features/leaderboard/hooks/useLeaderboard';

/**
 * Leaderboard page — UC-08 View Leaderboard.
 *
 * Signed-in only: with no session the user is sent to /signin. The viewer's
 * own row is highlighted (NF-6) and pinned below the table.
 */
export default function LeaderboardPage() {
  const { user, signingOut, handleSignOut, unavailable, retry } = useSignedInUser();
  const { data, previous, loading, error, period, setPeriod, refresh } = useLeaderboard('monthly');
  const [userRewardScore, setUserRewardScore] = useState<number | null>(null);
  const [userAvatars, setUserAvatars] = useState<Record<string, string>>({});

  // Switching period keeps the previous period's rows on screen, dimmed, until the
  // new ones arrive (no skeleton flash); if that request fails they are dropped.
  const board = data ?? (loading ? previous : null);
  const showingPrevious = !data && board !== null;
  // Dim the rows whenever they are being replaced (a period switch or a Refresh), so a
  // Refresh click has a visible effect and the old rows can't be operated meanwhile.
  const busy = loading && board !== null;
  const rawRankings = useMemo(() => board?.rankings ?? [], [board]);

  // Fetch user rewards score as fallback if leaderboard has score = 0 or unranked
  useEffect(() => {
    if (!user?.user_id) return;
    let cancelled = false;
    getRewards(user.user_id).then((res) => {
      if (!cancelled && res.status === 'ok') {
        setUserRewardScore(res.data.total_score);
      }
    }).catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [user?.user_id]);

  // Fetch the avatar of the users on the page being shown (the table reports
  // them as the page changes), so a long ranking costs a few requests per page
  // instead of one per user. A user is asked for once and not again, unless
  // the lookup failed, which is tried again when their page is shown next.
  // A late answer is still kept if the ranking has changed meanwhile.
  const [visibleIds, setVisibleIds] = useState<string[]>([]);
  const requestedAvatars = useRef(new Set<string>());
  const mounted = useRef(true);
  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);
  useEffect(() => {
    const toFetch = visibleIds.filter(
      (uid) => uid && uid !== user?.user_id && !requestedAvatars.current.has(uid)
    );
    toFetch.forEach((uid) => {
      requestedAvatars.current.add(uid);
      getPublicProfile(uid)
        .then((prof) => {
          if (mounted.current && prof?.avatar_url) {
            setUserAvatars((prev) => ({ ...prev, [uid]: prof.avatar_url }));
          }
        })
        .catch(() => {
          // A missing photo falls back to initials, and the next time this
          // page is shown the lookup is tried again.
          requestedAvatars.current.delete(uid);
        });
    });
  }, [visibleIds, user?.user_id]);

  const rankings = rawRankings.map((entry) => {
    const isMe = user && entry.user_id === user.user_id;
    const avatar = entry.avatar_url || (isMe ? user.avatar_url : userAvatars[entry.user_id]);
    const score = (isMe && entry.score === 0 && userRewardScore != null && userRewardScore > 0)
      ? userRewardScore
      : entry.score;

    return {
      ...entry,
      avatar_url: avatar,
      score,
    };
  });

  const topAngler = rankings.length > 0 ? rankings[0] : null;
  const myEntry = user ? rankings.find((r) => r.user_id === user.user_id) : undefined;
  const displayScore = myEntry?.score || (userRewardScore ?? 0);

  // Nothing to show until we know who is looking (signed-out -> redirect).
  if (!user) {
    return (
      <div className="flex min-h-screen flex-col">
        <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
        <Header />
        {unavailable && (
          <main className="relative mx-auto w-full max-w-2xl px-4 py-8">
            <SessionUnavailableAlert onRetry={retry} />
          </main>
        )}
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col">
      {/* Fixed to the viewport by its own CSS, like the other pages. */}
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <Header
        user={{ displayName: user.display_name, avatarUrl: user.avatar_url, isAdmin: user.role === "ADMIN" }}
        onSignOut={handleSignOut}
        signingOut={signingOut}
      />
      <main className="mx-auto w-full max-w-2xl px-4 py-8 sm:py-10">
        <PixelPanel>
          {/* ── Page Header & Title ──────────────────────────────────── */}
          <div className="flex items-start justify-between gap-4 flex-wrap">
            <div>
              <div className="flex items-center gap-2">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src="/sprites/ui/trophy.png"
                  alt=""
                  aria-hidden="true"
                  width={32}
                  height={32}
                  className="pixel-sprite"
                  style={{ width: 32, height: 32 }}
                />
                <h1
                  className="font-display leading-tight"
                  style={{ fontSize: 'calc(var(--px) * 10)' }}
                >
                  Hall of Anglers
                </h1>
              </div>
              <p className="mt-1 text-bark font-body text-sm">
                Rankings based on fish score system that earned during focused study sessions.
              </p>
            </div>
          </div>

          {/* ── Period tabs + refresh, one row (S-2) ─────────────────── */}
          <div className="mt-5 flex items-center justify-between gap-3 flex-wrap">
            <PeriodTabs active={period} onChange={setPeriod} />
            <PixelButton
              variant="ghost"
              onClick={refresh}
              aria-label="Refresh rankings"
              aria-busy={loading}
              className="min-h-11 sm:min-h-0"
              style={{ fontSize: 'calc(var(--px) * 4)', padding: '0.25rem 0.6rem' }}
            >
              Refresh
            </PixelButton>
          </div>

          {/* ── Summary Stats Strip ──────────────────────────────────── */}
          <div className={`grid grid-cols-1 sm:grid-cols-2 gap-2.5 mt-5 transition-opacity ${busy ? 'opacity-60' : ''}`}>
            <div
              className="pixel-panel flex flex-col justify-center px-3 py-2.5"
              style={{ background: 'var(--color-cream-2)' }}
            >
              <span className="font-label text-[10px] uppercase text-bark">Current Leader</span>
              <span
                className="font-numeric text-lg text-ink truncate mt-0.5"
                title={topAngler?.display_name}
              >
                {topAngler ? topAngler.display_name : '—'}
              </span>
            </div>

            <div
              className="pixel-panel flex flex-col justify-center px-3 py-2.5"
              style={{ background: 'var(--color-cream-2)' }}
            >
              <span className="font-label text-[10px] uppercase text-bark">Your Ranking</span>
              <span className="font-numeric text-lg text-ink truncate mt-0.5">
                {!board
                  ? '—'
                  : myEntry
                    ? `#${myEntry.rank} of ${rankings.length} (${displayScore} pts)`
                    : `Unranked (${displayScore} pts)`}
              </span>
            </div>
          </div>

          {/* ── Error banner: what happened, and Retry ───────────────── */}
          {error && (
            <PixelAlert className="mt-4 flex flex-wrap items-center justify-between gap-3">
              <span>
                {data ? `${error} Showing the last rankings loaded.` : error}
              </span>
              <PixelButton small className="min-h-11 sm:min-h-0" onClick={refresh} disabled={loading}>
                {loading ? 'Retrying…' : 'Retry'}
              </PixelButton>
            </PixelAlert>
          )}

          {/* ── Leaderboard Table ────────────────────────────────────── */}
          {/* With nothing loaded and a failure, show only the banner: an empty
              table would read as "no catches" when the truth is "couldn't load". */}
          {(board || !error) && (
            <div className={`mt-4 transition-opacity ${busy ? 'opacity-60' : ''}`} aria-busy={busy} inert={busy}>
              {!board ? (
                <div className="flex flex-col gap-2.5" aria-busy="true" aria-label="Loading rankings">
                  {[1, 2, 3, 4, 5].map((i) => (
                    <div
                      key={i}
                      className="pixel-panel animate-pulse"
                      style={{
                        height: '4rem',
                        background: 'var(--color-cream-2)',
                        opacity: 1 - i * 0.12,
                      }}
                      aria-hidden="true"
                    />
                  ))}
                </div>
              ) : (
                <LeaderboardTable
                  entries={rankings}
                  currentUserId={user.user_id}
                  currentUserName={user.display_name}
                  currentUserScore={displayScore}
                  currentUserAvatarUrl={user.avatar_url}
                  period={period}
                  onVisibleEntries={(shown) => {
                    // Rows about to be replaced don't need their avatars looked up.
                    if (showingPrevious) return;
                    const ids = shown.map((e) => e.user_id);
                    setVisibleIds((prev) => (prev.join(',') === ids.join(',') ? prev : ids));
                  }}
                />
              )}
            </div>
          )}

          {/* ── Bottom Action Navigation ─ */}
          <hr className="pixel-rule mt-6" aria-hidden="true" />
          <div className="pt-4 flex flex-wrap items-center justify-between gap-3">
            <Link href="/">
              <PixelButton variant="ghost">Focus Room</PixelButton>
            </Link>
            <Link href="/account">
              <PixelButton>My Account</PixelButton>
            </Link>
          </div>
        </PixelPanel>
      </main>
    </div>
  );
}
