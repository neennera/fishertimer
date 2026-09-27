"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Pencil } from "pixelarticons/react/Pencil";
import { User } from "pixelarticons/react/User";
import { countFishCaught, FishTank } from "../../components/account/FishTank";
import { Header } from "../../components/Header";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelPanel } from "../../components/ui/PixelPanel";
import { StatTile } from "../../components/ui/StatTile";
import { getSession, signOut, type SessionUser } from "../../lib/auth";
// MOCK ONLY — no statistics endpoint on the backend yet.
import { MOCK_ACCOUNT_STATS } from "../../lib/mocks/account-stats.mock";
// MOCK ONLY — no rewards endpoint on the backend yet.
import { MOCK_FISH_CATALOG, MOCK_USER_REWARDS } from "../../lib/mocks/rewards.mock";

// Placeholder: Stage 6 (UC-07 Edit Display Name, wireframes 04a–04f) builds
// this route. Until then the link 404s.
const EDIT_DISPLAY_NAME_HREF = "/account/edit";

// 1120 -> "18h 40m"; under an hour, just "40m".
function formatFocusMinutes(totalMinutes: number) {
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
}

function formatCount(count: number) {
  return count.toLocaleString("en-US");
}

function Avatar({ src }: { src: string }) {
  // Keyed by URL, so a new avatar_url gets a fresh attempt.
  const [failedSrc, setFailedSrc] = useState<string | null>(null);

  return (
    <div className="pixel-avatar">
      {src && failedSrc !== src ? (
        // A remote Google photo at a fixed box size — next/image would need
        // remotePatterns for googleusercontent.com and buys nothing here.
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={src}
          alt=""
          className="pixel-avatar__img"
          // Google's avatar CDN rejects some requests that carry a referrer.
          referrerPolicy="no-referrer"
          onError={() => setFailedSrc(src)}
        />
      ) : (
        <User aria-hidden="true" className="pixel-avatar__fallback" />
      )}
    </div>
  );
}

export default function AccountPage() {
  const router = useRouter();
  // null while GET /me is loading.
  const [user, setUser] = useState<SessionUser | null>(null);
  const [signingOut, setSigningOut] = useState(false);

  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (cancelled) {
        return;
      }
      if (session.status === "signed_in") {
        setUser(session.user);
      } else if (session.status === "needs_signup") {
        // Signed in with Google but no account yet: finish first-time setup.
        router.replace("/welcome");
      } else {
        router.replace("/signin");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [router]);

  // The Header's sign-out key is the only one on this page.
  async function handleSignOut() {
    setSigningOut(true);
    try {
      await signOut();
      router.replace("/signin");
    } catch {
      // Gateway unreachable: the cookie is still set, so stay signed in and
      // let them try again.
      setSigningOut(false);
    }
  }

  return (
    <div className="pixel-wood min-h-screen">
      <Header
        user={user ? { displayName: user.display_name } : null}
        onSignOut={handleSignOut}
        signingOut={signingOut}
      />

      <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 md:py-12">
        <h1 className="sr-only">Account</h1>

        <PixelPanel as="section" aria-label="Profile">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            {user ? (
              <Avatar src={user.avatar_url} />
            ) : (
              <div aria-hidden="true" className="pixel-avatar pixel-skeleton" />
            )}

            <div className="min-w-0 flex-1">
              {user ? (
                <div className="flex items-center gap-2">
                  <p className="min-w-0 font-display text-3xl leading-none break-words">
                    {user.display_name}
                  </p>
                  <PixelButton
                    variant="ghost"
                    icon
                    small
                    onClick={() => router.push(EDIT_DISPLAY_NAME_HREF)}
                    aria-label="Edit display name"
                    title="Edit display name"
                  >
                    <Pencil aria-hidden="true" />
                  </PixelButton>
                </div>
              ) : (
                <p className="font-display text-3xl leading-none">
                  <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text" />
                </p>
              )}
              <p className="mt-2 text-sm text-bark break-words">
                {user ? (
                  user.email
                ) : (
                  <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text" />
                )}
              </p>
            </div>
          </div>
        </PixelPanel>

        <PixelPanel as="section" aria-label="Statistics">
          {/* MOCK ONLY: these numbers come from lib/mocks/account-stats.mock.ts
              until the backend has a statistics endpoint. */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <StatTile
              value={formatCount(MOCK_ACCOUNT_STATS.total_sessions)}
              caption="Sessions joined"
            />
            <StatTile
              value={formatFocusMinutes(MOCK_ACCOUNT_STATS.total_focus_minutes)}
              caption="Total focus time"
            />
            <StatTile
              value={formatCount(MOCK_ACCOUNT_STATS.rewards_earned)}
              caption="Rewards earned"
            />
          </div>
        </PixelPanel>

        <PixelPanel as="section" aria-labelledby="sessions-by-room-type">
          <h2 id="sessions-by-room-type" className="font-display text-2xl leading-none">
            Sessions by room type
          </h2>
          {/* MOCK-ONLY placeholder: nothing backs this yet. The breakdown
              (03a shows a donut) will need two services — study-session for
              group sessions (session_participants) and study-timer for solo
              sessions (timer_sessions with no study_session_id). schema.dbml
              has no room-type column at all yet, so "room type" is not
              schema-backed until one of them adds it. */}
          <div className="pixel-placeholder mt-6 text-sm">
            Coming soon
          </div>
        </PixelPanel>

        <PixelPanel as="section" aria-labelledby="fish-tank">
          <div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
            <h2 id="fish-tank" className="font-display text-2xl leading-none">
              Fish Tank
            </h2>
            <p className="text-sm text-bark">
              {countFishCaught(MOCK_FISH_CATALOG, MOCK_USER_REWARDS)} fish caught in total
            </p>
          </div>
          <div className="mt-6">
            <FishTank catalog={MOCK_FISH_CATALOG} rewards={MOCK_USER_REWARDS} />
          </div>
        </PixelPanel>
      </main>
    </div>
  );
}
