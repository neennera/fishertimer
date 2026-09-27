"use client";

import { useState } from "react";
import Link from "next/link";
import { DisplayName } from "../account/DisplayName";
import { countFishCaught, FishTank } from "../account/FishTank";
import { PixelPanel } from "../ui/PixelPanel";
import { StatTile } from "../ui/StatTile";
import { FocusHistoryChart } from "./FocusHistoryChart";
import type { SessionUser } from "../../lib/auth";
import { cx } from "../../lib/cx";
import { formatCount, formatFocusMinutes } from "../../lib/format";
import { initials } from "../../lib/initials";
import type { PanelData, PublicProfile } from "../../lib/profile";
import type { RewardsSummary, TimerStatistics } from "../../lib/profile-types";

// "26 Sept" this year, "Dec 2025" before; the zero timestamp means never.
function formatLastActive(iso: string) {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime()) || date.getUTCFullYear() <= 1) {
    return "Never";
  }
  const thisYear = date.getFullYear() === new Date().getFullYear();
  return date.toLocaleDateString(
    "en-GB",
    thisYear ? { day: "numeric", month: "short" } : { month: "short", year: "numeric" },
  );
}

function Avatar({ src, name }: { src: string; name: string }) {
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
        <span aria-hidden="true" className="pixel-avatar__initials">
          {initials(name)}
        </span>
      )}
    </div>
  );
}

/** What only the owner gets: the e-mail and the name editor. */
export interface OwnerControls {
  email: string;
  onNameSaved: (user: SessionUser) => void;
  onSignedOut: () => void;
}

// Fixed width, so the two swap in the exact same box.
const SMALL_LINK_BUTTON = "pixel-btn pixel-btn--ghost pixel-btn--sm w-36";

export function ProfilePanel({
  profile,
  editable,
  owner,
  ownPreview = false,
}: {
  /** null while loading. */
  profile: PublicProfile | null;
  editable: boolean;
  /** Required once loaded when editable. */
  owner?: OwnerControls;
  /** The owner viewing their own public profile: same layout as /account. */
  ownPreview?: boolean;
}) {
  return (
    <PixelPanel as="section" aria-label="Profile" className="pixel-panel--profile">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        {profile ? (
          <Avatar src={profile.avatar_url} name={profile.display_name} />
        ) : (
          <div aria-hidden="true" className="pixel-avatar pixel-skeleton" />
        )}

        <div className={cx("min-w-0 flex-1", ownPreview && "pixel-profile-preview")}>
          {!profile ? (
            <>
              <div className="pixel-name-row" aria-hidden="true">
                <span className="pixel-skeleton pixel-skeleton--text h-3/4" />
              </div>
              {(editable || ownPreview) && (
                <>
                  <p className="pixel-inline-error" aria-hidden="true" />
                  {/* Same line as the loaded one, button-height included. */}
                  <div
                    className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1 text-sm"
                    aria-hidden="true"
                  >
                    <p className="min-w-0 flex-1">
                      <span className="pixel-skeleton pixel-skeleton--text" />
                    </p>
                    <span className={`${SMALL_LINK_BUTTON} invisible`}>View public profile</span>
                  </div>
                </>
              )}
            </>
          ) : editable && owner ? (
            <DisplayName
              editable
              name={profile.display_name}
              email={owner.email}
              onSaved={owner.onNameSaved}
              onSignedOut={owner.onSignedOut}
              aside={
                <Link
                  href={`/profile/${encodeURIComponent(profile.user_id)}`}
                  className={SMALL_LINK_BUTTON}
                >
                  View public profile
                </Link>
              }
            />
          ) : ownPreview ? (
            <DisplayName
              name={profile.display_name}
              aside={
                <Link href="/account" className={SMALL_LINK_BUTTON}>
                  Back to account
                </Link>
              }
            />
          ) : (
            <DisplayName name={profile.display_name} />
          )}
        </div>
      </div>
    </PixelPanel>
  );
}

// A span, so it can sit inside the fish tank's message <p>.
function Unavailable({ what }: { what: string }) {
  return (
    <span className="pixel-placeholder min-h-0 text-sm" role="status">
      Couldn&rsquo;t load {what} right now.
    </span>
  );
}

export function StatsPanel({
  stats,
  rewards,
}: {
  stats: PanelData<TimerStatistics>;
  rewards: PanelData<RewardsSummary>;
}) {
  if (stats.status === "unavailable") {
    return (
      <PixelPanel as="section" aria-label="Statistics" className="pixel-panel--stats">
        <Unavailable what="statistics" />
      </PixelPanel>
    );
  }
  const s = stats.status === "ok" ? stats.data : null;
  // Counts every reward type, so it can exceed the fish total.
  const earned =
    rewards.status === "ok"
      ? formatCount(rewards.data.total_awards_earned)
      : rewards.status === "unavailable"
        ? "—"
        : null;
  const tiles: [string, string | null][] = [
    ["Sessions joined", s && formatCount(s.sessions_joined)],
    ["Total focus time", s && formatFocusMinutes(s.total_focus_minutes)],
    ["Rewards earned", s && earned],
    ["Last active", s && formatLastActive(s.last_active)],
  ];

  return (
    <PixelPanel as="section" aria-label="Statistics" className="pixel-panel--stats">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4">
        {tiles.map(([caption, value]) =>
          value === null ? (
            // Same structure as StatTile, so it is the same size; the value
            // line stays empty until the number arrives.
            <div key={caption} className="pixel-tile" aria-hidden="true">
              <div className="pixel-tile__value">{"\u00a0"}</div>
              <div className="pixel-tile__caption">{caption}</div>
            </div>
          ) : (
            <StatTile key={caption} value={value} caption={caption} />
          ),
        )}
      </div>
    </PixelPanel>
  );
}

export function FocusHistoryPanel({ stats }: { stats: PanelData<TimerStatistics> }) {
  return (
    <PixelPanel as="section" aria-labelledby="focus-history">
      {stats.status === "unavailable" ? (
        <>
          <h2 id="focus-history" className="font-display text-2xl leading-none">
            Focus history
          </h2>
          <div className="mt-6">
            <Unavailable what="focus history" />
          </div>
        </>
      ) : (
        <FocusHistoryChart
          titleId="focus-history"
          days={stats.status === "ok" ? stats.data.daily_focus_minutes : null}
        />
      )}
    </PixelPanel>
  );
}

export function FishTankPanel({
  rewards,
  editable,
}: {
  rewards: PanelData<RewardsSummary>;
  editable: boolean;
}) {
  const items = rewards.status === "ok" ? rewards.data.items : [];
  return (
    <PixelPanel as="section" aria-labelledby="fish-tank" className="pixel-panel--tank">
      <div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
        <h2 id="fish-tank" className="font-display text-2xl leading-none">
          Fish Tank
        </h2>
        <p className="text-sm text-bark">
          {rewards.status === "ok" ? (
            `${countFishCaught(items)} fish caught in total`
          ) : rewards.status === "loading" ? (
            <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text w-32" />
          ) : null}
        </p>
      </div>
      <div className="mt-6">
        <FishTank
          items={items}
          emptyMessage={
            rewards.status === "loading" ? (
              <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text w-1/2" />
            ) : rewards.status === "unavailable" ? (
              <Unavailable what="fish" />
            ) : editable ? (
              "No fish yet. Finish a focus session to catch one."
            ) : (
              "No fish caught yet."
            )
          }
        />
      </div>
    </PixelPanel>
  );
}

/**
 * The profile as /account (editable) and /profile/[userId] (public) show it.
 * Only public fields reach it: `owner` (with the e-mail) is only passed on
 * /account.
 */
export function ProfileSections({
  profile,
  stats,
  rewards,
  editable,
  owner,
  ownPreview,
}: {
  /** null while loading. */
  profile: PublicProfile | null;
  stats: PanelData<TimerStatistics>;
  rewards: PanelData<RewardsSummary>;
  editable: boolean;
  owner?: OwnerControls;
  ownPreview?: boolean;
}) {
  return (
    <>
      <ProfilePanel profile={profile} editable={editable} owner={owner} ownPreview={ownPreview} />
      <StatsPanel stats={stats} rewards={rewards} />
      <FocusHistoryPanel stats={stats} />
      <FishTankPanel rewards={rewards} editable={editable} />
    </>
  );
}
