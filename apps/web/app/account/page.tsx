'use client';

import Link from 'next/link';
import { Header } from '../../components/Header';
import { PixelPanel } from '../../components/ui/PixelPanel';
import { PixelButton } from '../../components/ui/PixelButton';

/**
 * Account page (UC-06).
 *
 * Provides a user profile dashboard with user stats, linked back to
 * the Study Room (timer) and Leaderboard.
 */
export default function AccountPage() {
  const demoUser = {
    displayName: 'LureQueen',
    userId: 'user3',
    email: 'lurequeen@fishertimer.local',
    role: 'Dedicated Angler',
    totalCatches: 5,
    weeklyRank: '#1 Weekly Champion',
  };

  return (
    <>
      <Header user={{ displayName: demoUser.displayName }} />
      <main className="mx-auto w-full max-w-2xl px-4 py-10">
        <PixelPanel>
          {/* Top navigation */}
          <div className="mb-6 flex items-center justify-between">
            <Link
              href="/"
              className="inline-flex items-center gap-1.5 text-bark hover:text-ink font-label text-sm transition-colors"
            >
              ← Back to Study Room
            </Link>
            <Link
              href="/leaderboard"
              className="inline-flex items-center gap-1.5 text-bark hover:text-ink font-label text-sm transition-colors"
            >
              🏆 View Leaderboard →
            </Link>
          </div>

          {/* Profile Header */}
          <div className="flex items-center gap-4 pb-6 border-b border-[var(--color-rule)] flex-wrap">
            <div
              className="w-16 h-16 flex items-center justify-center font-numeric text-2xl font-bold"
              style={{
                clipPath: 'var(--pixclip)',
                background: 'var(--color-lake)',
                color: '#fff',
                boxShadow: 'inset 0 0 0 var(--px) var(--color-lake-dp)',
              }}
            >
              LQ
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h1
                  className="font-display leading-tight"
                  style={{ fontSize: 'calc(var(--px) * 9)' }}
                >
                  {demoUser.displayName}
                </h1>
                <span className="pixel-badge" style={{ background: 'var(--color-amber)' }}>
                  ACTIVE
                </span>
              </div>
              <p className="text-bark font-label text-xs mt-1">
                ID: {demoUser.userId} · {demoUser.email}
              </p>
            </div>
          </div>

          {/* Statistics Grid */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-6">
            <div
              className="pixel-panel"
              style={{
                background: 'var(--color-cream-2)',
                padding: 'calc(var(--px) * 5)',
              }}
            >
              <div className="font-label text-xs uppercase text-bark mb-1">Weekly Standing</div>
              <div className="font-numeric text-xl text-ink flex items-center gap-2">
                <span>🥇</span> {demoUser.weeklyRank}
              </div>
              <p className="text-xs font-body text-bark mt-1">
                Ranked #1 on this week&apos;s leaderboard
              </p>
            </div>

            <div
              className="pixel-panel"
              style={{
                background: 'var(--color-cream-2)',
                padding: 'calc(var(--px) * 5)',
              }}
            >
              <div className="font-label text-xs uppercase text-bark mb-1">Total Fish Caught</div>
              <div className="font-numeric text-xl text-ink flex items-center gap-2">
                <span>🐟</span> {demoUser.totalCatches} Catches
              </div>
              <p className="text-xs font-body text-bark mt-1">
                Earned from completed focus sessions
              </p>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="mt-8 pt-5 border-t border-[var(--color-rule)] flex flex-wrap items-center justify-between gap-3">
            <Link href="/">
              <PixelButton variant="ghost">⏱️ Focus Timer</PixelButton>
            </Link>
            <Link href="/leaderboard">
              <PixelButton>🏆 Open Leaderboard</PixelButton>
            </Link>
          </div>
        </PixelPanel>
      </main>
    </>
  );
}
