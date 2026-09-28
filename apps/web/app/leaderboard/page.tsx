'use client';

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
 *
 * Note: currentUserId would come from the auth context once UC-06 lands.
 * For now it is hard-coded to "user3" to demonstrate the highlight on the
 * weekly leaderboard where user3 is the weekly leader.
 */
const DEMO_CURRENT_USER = 'user3'; // swap with real auth once UC-06 is merged

export default function LeaderboardPage() {
  const { data, loading, error, period, setPeriod, refresh } = useLeaderboard('all-time');

  const rankings = data?.rankings ?? [];
  const isCached = data?.cached ?? false;
  const lastFetch = data?.leaderboard_last_fetch
    ? new Date(data.leaderboard_last_fetch).toLocaleTimeString()
    : null;

  return (
    <>
      <Header />
      <main className="mx-auto w-full max-w-2xl px-4 py-10">
        <PixelPanel>
          {/* ── Page header ───────────────────────────────────────────── */}
          <div className="flex items-start justify-between gap-4 flex-wrap">
            <div>
              <h1
                className="font-display leading-tight"
                style={{ fontSize: 'calc(var(--px) * 10)' }}
              >
                🏆 Leaderboard
              </h1>
              <p className="mt-1 text-bark" style={{ fontSize: '0.85rem' }}>
                Rankings update whenever new rewards are earned.
              </p>
            </div>

            {/* Refresh button */}
            <PixelButton
              variant="ghost"
              onClick={refresh}
              disabled={loading}
              aria-label="Refresh leaderboard"
              style={{ fontSize: 'calc(var(--px) * 5)' }}
            >
              {loading ? '⏳ Loading…' : '🔄 Refresh'}
            </PixelButton>
          </div>

          {/* ── Period tabs (S-2) ─────────────────────────────────────── */}
          <div className="mt-6">
            <PeriodTabs active={period} onChange={setPeriod} disabled={loading} />
          </div>

          {/* ── Cache status badge ───────────────────────────────────── */}
          {lastFetch && (
            <p
              className="mt-3"
              style={{ fontSize: '0.75rem', color: 'var(--color-muted)' }}
              aria-live="polite"
            >
              {isCached ? '📦 Served from cache' : '✨ Fresh data'} · fetched at {lastFetch}
            </p>
          )}

          {/* ── Error state (E-3 / network errors) ───────────────────── */}
          {error && (
            <div className="pixel-alert mt-6" role="alert">
              ⚠️ {error}
            </div>
          )}

          {/* ── Rankings table ───────────────────────────────────────── */}
          <div className="mt-6">
            {loading && !data ? (
              // Skeleton: show placeholder rows while initial data loads
              <div className="flex flex-col gap-2" aria-busy="true" aria-label="Loading rankings">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div
                    key={i}
                    className="pixel-panel"
                    style={{
                      height: '3.5rem',
                      background: 'var(--color-cream-2)',
                      opacity: 1 - i * 0.1,
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

          {/* ── Footer note ──────────────────────────────────────────── */}
          <p
            className="mt-6 text-bark"
            style={{ fontSize: '0.78rem' }}
          >
            Ties are broken by the earliest reward earned at that total. Rankings are read-only.
          </p>
        </PixelPanel>
      </main>
    </>
  );
}
