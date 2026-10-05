'use client';

import { useState } from 'react';
import { Header } from '../../components/Header';
import { ParallaxScene } from '../../components/ui/ParallaxScene';
import { PixelButton } from '../../components/ui/PixelButton';
import { PixelPanel } from '../../components/ui/PixelPanel';
import { useSignedInUser } from '../../components/profile/ProfileShell';
import { SIGNIN_SCENE_LAYERS } from '../../lib/scenes/signin-scene';

type Result = { ok: true; output: string } | { ok: false; error: string; output?: string };

/** Developer tools. Signed-in only; the seed API refuses to run in production. */
export default function DevPage() {
  const { user, signingOut, handleSignOut } = useSignedInUser();
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<Result | null>(null);

  async function runSeed() {
    if (!user) {
      return;
    }
    setRunning(true);
    setResult(null);
    try {
      const res = await fetch('/api/dev/seed', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: user.user_id, display_name: user.display_name }),
      });
      const body = (await res.json().catch(() => ({}))) as {
        output?: string;
        error?: string;
      };
      setResult(
        res.ok
          ? { ok: true, output: body.output ?? '' }
          : { ok: false, error: body.error ?? `Request failed (${res.status})`, output: body.output },
      );
    } catch {
      setResult({ ok: false, error: 'Could not reach the dev server.' });
    } finally {
      setRunning(false);
    }
  }

  return (
    <div className="flex min-h-screen flex-col">
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <Header
        user={user ? { displayName: user.display_name, avatarUrl: user.avatar_url } : null}
        onSignOut={handleSignOut}
        signingOut={signingOut}
      />
      <main className="mx-auto w-full max-w-2xl px-4 py-8 sm:py-10">
        <PixelPanel>
          <h1 className="font-display leading-tight" style={{ fontSize: 'calc(var(--px) * 10)' }}>
            🛠️ Dev Tools
          </h1>
          <p className="mt-1 text-bark font-body text-sm">
            Seeds this month&apos;s mock fish for your account, plus mock competitors for the
            leaderboard. Existing seeded rewards are deleted first.
          </p>

          <div className="mt-5 flex flex-wrap items-center gap-3">
            <PixelButton onClick={runSeed} disabled={!user || running}>
              {running ? '⏳ Seeding…' : '🌱 Seed my fish (this month)'}
            </PixelButton>
            {user && (
              <span className="font-label text-xs text-bark">
                as {user.display_name} ({user.user_id})
              </span>
            )}
          </div>

          {result && (
            <div
              className={result.ok ? 'pixel-alert mt-4' : 'pixel-alert pixel-alert--warn mt-4'}
              role="status"
            >
              <div className="font-numeric">{result.ok ? '✅ Seeded' : `⚠️ ${result.error}`}</div>
              {result.output && (
                <pre className="mt-2 whitespace-pre-wrap break-words font-body text-xs">
                  {result.output}
                </pre>
              )}
            </div>
          )}
        </PixelPanel>
      </main>
    </div>
  );
}
