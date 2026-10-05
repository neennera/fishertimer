'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Header } from '../../components/Header';
import { ParallaxScene } from '../../components/ui/ParallaxScene';
import { PixelPanel } from '../../components/ui/PixelPanel';
import { PixelButton } from '../../components/ui/PixelButton';
import { useSignedInUser } from '../../components/profile/ProfileShell';
import { SIGNIN_SCENE_LAYERS } from '../../lib/scenes/signin-scene';
import { getRewards } from '../../lib/profile';
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
  const { user, signingOut, handleSignOut } = useSignedInUser();
  const { data, loading, error, period, setPeriod, refresh } = useLeaderboard('monthly');
  const [userRewardScore, setUserRewardScore] = useState<number | null>(null);

  const rawRankings = data?.rankings ?? [];

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

  const rankings = rawRankings.map((entry) => {
    if (user && entry.user_id === user.user_id && entry.score === 0 && userRewardScore != null && userRewardScore > 0) {
      return { ...entry, score: userRewardScore };
    }
    return entry;
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
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col">
      {/* Fixed to the viewport by its own CSS, like the other pages. */}
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <Header
        user={{ displayName: user.display_name, avatarUrl: user.avatar_url }}
        onSignOut={handleSignOut}
        signingOut={signingOut}
      />
      <main className="mx-auto w-full max-w-2xl px-4 py-8 sm:py-10">
        <PixelPanel>
          {/* ── Page Header & Title ──────────────────────────────────── */}
          <div className="flex items-start justify-between gap-4 flex-wrap">
            <div>
              <div className="flex items-center gap-2">
                <span className="text-3xl" aria-hidden="true">
                  🏆
                </span>
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
            <PeriodTabs active={period} onChange={setPeriod} disabled={loading} />
            <PixelButton
              variant="ghost"
              onClick={refresh}
              disabled={loading}
              aria-label="Refresh rankings"
              style={{ fontSize: 'calc(var(--px) * 4)', padding: '0.25rem 0.6rem', minHeight: 0 }}
            >
              {loading ? '⏳ Updating…' : '🔄 Refresh'}
            </PixelButton>
          </div>

          {/* ── Summary Stats Strip ──────────────────────────────────── */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 mt-5">
            <div
              className="pixel-panel flex flex-col justify-center px-3 py-2.5"
              style={{ background: 'var(--color-cream-2)' }}
            >
              <span className="font-label text-[10px] uppercase text-bark">Current Leader</span>
              <span
                className="font-numeric text-lg text-ink truncate mt-0.5"
                title={topAngler?.display_name}
              >
                {topAngler ? `🥇 ${topAngler.display_name}` : '—'}
              </span>
            </div>

            <div
              className="pixel-panel flex flex-col justify-center px-3 py-2.5"
              style={{ background: 'var(--color-cream-2)' }}
            >
              <span className="font-label text-[10px] uppercase text-bark">Your Ranking</span>
              <span className="font-numeric text-lg text-ink truncate mt-0.5">
                {myEntry ? `#${myEntry.rank} of ${rankings.length} (${displayScore} pts)` : `Unranked (${displayScore} pts)`}
              </span>
            </div>
          </div>

          {/* ── Error Banner (E-3) ───────────────────────────────────── */}
          {error && (
            <div className="pixel-alert mt-4" role="alert">
              ⚠️ {error}
            </div>
          )}

          {/* ── Leaderboard Table ────────────────────────────────────── */}
          <div className="mt-4">
            {loading && !data ? (
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
              />
            )}
          </div>


          {/* ── Bottom Action Navigation ─ */}
          <div className="mt-6 pt-4 border-t border-[var(--color-rule)] flex flex-wrap items-center justify-between gap-3">
            <Link href="/">
              <PixelButton variant="ghost">⏱️ Focus Room (Timer)</PixelButton>
            </Link>
            <Link href="/account">
              <PixelButton>👤 My Account Profile</PixelButton>
            </Link>
          </div>
        </PixelPanel>
      </main>
    </div>
  );
}
