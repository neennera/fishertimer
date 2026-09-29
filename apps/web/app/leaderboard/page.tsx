'use client';

import Link from 'next/link';
import { Header } from '../../components/Header';
import { PixelPanel } from '../../components/ui/PixelPanel';
import { PixelButton } from '../../components/ui/PixelButton';
import { PeriodTabs } from '../../features/leaderboard/components/PeriodTabs';
import { LeaderboardTable } from '../../features/leaderboard/components/LeaderboardTable';
import { useLeaderboard } from '../../features/leaderboard/hooks/useLeaderboard';

/**
 * Leaderboard page — UC-08 View Leaderboard.
 *
 * Flow:
 *  NF-1  User opens Leaderboard.
 *  NF-2  System defaults to "all-time" period and loads that ranking (S-1).
 *  NF-3  Users ranked by total rewards, with display name and reward count.
 *  NF-4  User selects a different period (S-2).
 *  NF-5  System reloads ranking for the selected period.
 *  NF-6  User's own position is highlighted.
 */
const DEMO_CURRENT_USER = 'user3'; // LureQueen (Demo user)

export default function LeaderboardPage() {
  const { data, loading, error, period, setPeriod, refresh } = useLeaderboard('all-time');

  const rankings = data?.rankings ?? [];
  const isCached = data?.cached ?? false;
  const topAngler = rankings.length > 0 ? rankings[0] : null;
  const lastFetch = data?.leaderboard_last_fetch
    ? new Date(data.leaderboard_last_fetch).toLocaleTimeString()
    : null;

  return (
    <>
      <Header />
      <main className="mx-auto w-full max-w-2xl px-4 py-8 sm:py-10">
        <PixelPanel>
          {/* ── Top Navigation Bar (Clean: Home & Account only) ───────── */}
          <div className="mb-5 flex items-center justify-between pb-3 border-b border-[var(--color-rule)]">
            <Link
              href="/"
              className="inline-flex items-center gap-1.5 text-bark hover:text-ink font-label text-sm transition-colors"
            >
              ← Back to Study Room
            </Link>
            <Link
              href="/account"
              className="inline-flex items-center gap-1.5 text-bark hover:text-ink font-label text-sm transition-colors"
            >
              👤 My Account →
            </Link>
          </div>

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
                Rankings based on fish earned during focused study sessions.
              </p>
            </div>

            {/* Manual Refresh Button */}
            <PixelButton
              variant="ghost"
              onClick={refresh}
              disabled={loading}
              aria-label="Refresh rankings"
              style={{ fontSize: 'calc(var(--px) * 5)' }}
            >
              {loading ? '⏳ Updating…' : '🔄 Refresh'}
            </PixelButton>
          </div>

          {/* ── Period Tabs Selector (S-2) ────────────────────────────── */}
          <div className="mt-6">
            <PeriodTabs active={period} onChange={setPeriod} disabled={loading} />
          </div>

          {/* ── Summary Stats Strip ──────────────────────────────────── */}
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2.5 mt-5">
            <div
              className="pixel-panel flex flex-col justify-center px-3 py-2.5"
              style={{ background: 'var(--color-cream-2)' }}
            >
              <span className="font-label text-[10px] uppercase text-bark">Ranked Anglers</span>
              <span className="font-numeric text-lg text-ink mt-0.5">
                {rankings.length} Active
              </span>
            </div>

            <div
              className="pixel-panel flex flex-col justify-center px-3 py-2.5"
              style={{ background: 'var(--color-cream-2)' }}
            >
              <span className="font-label text-[10px] uppercase text-bark">Current Leader</span>
              <span className="font-numeric text-lg text-ink truncate mt-0.5" title={topAngler?.display_name}>
                {topAngler ? `🥇 ${topAngler.display_name}` : '—'}
              </span>
            </div>

            <div
              className="pixel-panel col-span-2 sm:col-span-1 flex flex-col justify-center px-3 py-2.5"
              style={{ background: 'var(--color-cream-2)' }}
            >
              <span className="font-label text-[10px] uppercase text-bark">Cache Status</span>
              <span className="font-numeric text-sm text-ink truncate mt-0.5">
                {isCached ? '📦 In-Memory Cache' : '✨ Fresh Compute'}
              </span>
            </div>
          </div>

          {/* ── Cache Timing Info ────────────────────────────────────── */}
          {lastFetch && (
            <div
              className="mt-2.5 text-right font-label text-[11px] opacity-70"
              style={{ color: 'var(--color-muted)' }}
              aria-live="polite"
            >
              Last synced at {lastFetch}
            </div>
          )}

          {/* ── Error Banner (E-3) ───────────────────────────────────── */}
          {error && (
            <div className="pixel-alert mt-4" role="alert">
              ⚠️ {error}
            </div>
          )}

          {/* ── Leaderboard Table ────────────────────────────────────── */}
          <div className="mt-4">
            {loading && !data ? (
              // Skeleton loading state
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
                currentUserId={DEMO_CURRENT_USER}
              />
            )}
          </div>

          {/* ── Footer Rules Note ────────────────────────────────────── */}
          <p
            className="mt-6 text-bark font-body"
            style={{ fontSize: '0.8rem' }}
          >
            ℹ️ Ties are resolved by earliest reward timestamp. Rankings are read-only.
          </p>

          {/* ── Bottom Action Navigation (One to Timer, One to Account) ─ */}
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
    </>
  );
}
