"use client";

import { useState } from "react";
import Link from "next/link";
import { User } from "pixelarticons/react/User";
import { DisplayName } from "../account/DisplayName";
import { countFishCaught, FishTank } from "../account/FishTank";
import { PixelPanel } from "../ui/PixelPanel";
import { StatTile } from "../ui/StatTile";
import type { SessionUser } from "../../lib/auth";
import type { MockAccountStats } from "../../lib/mocks/account-stats.mock";
import type { ProfileData, PublicProfile } from "../../lib/profile";

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

/** What only the owner gets: the e-mail and the name editor. */
export interface OwnerControls {
  email: string;
  onNameSaved: (user: SessionUser) => void;
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
          <Avatar src={profile.avatar_url} />
        ) : (
          <div aria-hidden="true" className="pixel-avatar pixel-skeleton" />
        )}

        <div className="min-w-0 flex-1">
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

export function StatsPanel({ stats }: { stats: MockAccountStats | null }) {
  const tiles: [string, string | null][] = [
    ["Sessions joined", stats && formatCount(stats.total_sessions)],
    ["Total focus time", stats && formatFocusMinutes(stats.total_focus_minutes)],
    ["Rewards earned", stats && formatCount(stats.rewards_earned)],
  ];

  return (
    <PixelPanel as="section" aria-label="Statistics" className="pixel-panel--stats">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
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

export function RoomTypePanel() {
  return (
    <PixelPanel as="section" aria-labelledby="sessions-by-room-type" className="pixel-panel--rooms">
      <h2 id="sessions-by-room-type" className="font-display text-2xl leading-none">
        Sessions by room type
      </h2>
      {/* MOCK-ONLY placeholder: needs study-session (group) and study-timer
          (solo) data, and schema.dbml has no room-type column yet. */}
      <div className="pixel-placeholder mt-6 text-sm">Coming soon</div>
    </PixelPanel>
  );
}

export function FishTankPanel({ data, editable }: { data: ProfileData | null; editable: boolean }) {
  return (
    <PixelPanel as="section" aria-labelledby="fish-tank" className="pixel-panel--tank">
      <div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
        <h2 id="fish-tank" className="font-display text-2xl leading-none">
          Fish Tank
        </h2>
        <p className="text-sm text-bark">
          {data ? (
            `${countFishCaught(data.fishCatalog, data.fishRewards)} fish caught in total`
          ) : (
            <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text w-32" />
          )}
        </p>
      </div>
      <div className="mt-6">
        {/* While loading: the empty tank, with a placeholder where the
            message or cards will be. */}
        <FishTank
          catalog={data?.fishCatalog ?? []}
          rewards={data?.fishRewards ?? []}
          emptyMessage={
            !data ? (
              <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text w-1/2" />
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
 * The public view never receives an e-mail: `owner` is only passed on
 * /account.
 */
export function ProfileSections({
  data,
  editable,
  owner,
  ownPreview,
}: {
  /** null while loading. */
  data: ProfileData | null;
  editable: boolean;
  owner?: OwnerControls;
  ownPreview?: boolean;
}) {
  return (
    <>
      <ProfilePanel
        profile={data?.profile ?? null}
        editable={editable}
        owner={owner}
        ownPreview={ownPreview}
      />
      <StatsPanel stats={data?.stats ?? null} />
      <RoomTypePanel />
      <FishTankPanel data={data} editable={editable} />
    </>
  );
}
