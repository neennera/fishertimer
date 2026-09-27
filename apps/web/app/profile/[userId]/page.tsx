"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ProfileSections } from "../../../components/profile/ProfileSections";
import { ProfileShell, useSignedInUser } from "../../../components/profile/ProfileShell";
import { PixelPanel } from "../../../components/ui/PixelPanel";
import {
  getPublicProfile,
  isUserId,
  loadPanels,
  type PanelData,
  type PublicProfile,
} from "../../../lib/profile";
import type { RewardsSummary, TimerStatistics } from "../../../lib/profile-types";

const LOADING = { status: "loading" } as const;

interface Loaded {
  userId: string;
  /** null = no such user; "error" = the profile itself couldn't load. */
  profile: PublicProfile | null | "error";
  stats: PanelData<TimerStatistics>;
  rewards: PanelData<RewardsSummary>;
}

function Notice({ title, children }: { title: string; children: string }) {
  return (
    <PixelPanel as="section" className="text-center">
      <h2 className="font-display text-3xl leading-none">{title}</h2>
      <p className="mt-3 text-sm text-bark">{children}</p>
      <p className="mt-4 text-sm">
        <Link href="/account" className="pixel-link">
          Back to your account
        </Link>
      </p>
    </PixelPanel>
  );
}

export default function PublicProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const { user, signingOut, handleSignOut } = useSignedInUser();
  const [loaded, setLoaded] = useState<Loaded | null>(null);

  // All three in parallel, alongside /me.
  useEffect(() => {
    let cancelled = false;
    void Promise.allSettled([
      getPublicProfile(userId),
      isUserId(userId) ? loadPanels(userId) : Promise.resolve(null),
    ]).then(([profile, panels]) => {
      if (cancelled) {
        return;
      }
      const both = panels.status === "fulfilled" ? panels.value : null;
      setLoaded({
        userId,
        profile: profile.status === "fulfilled" ? profile.value : "error",
        stats: both?.stats ?? { status: "unavailable" },
        rewards: both?.rewards ?? { status: "unavailable" },
      });
    });
    return () => {
      cancelled = true;
    };
  }, [userId]);

  const current = loaded?.userId === userId ? loaded : null;
  const isOwn = user !== null && user.user_id === userId;

  return (
    <ProfileShell title="Profile" viewer={user} signingOut={signingOut} onSignOut={handleSignOut}>
      {current?.profile === null ? (
        <Notice title="Fisher not found">This profile doesn&rsquo;t exist, or it has moved.</Notice>
      ) : current?.profile === "error" ? (
        <Notice title="Couldn&rsquo;t load profile">Please try again in a moment.</Notice>
      ) : (
        <ProfileSections
          profile={current?.profile ?? null}
          stats={current?.stats ?? LOADING}
          rewards={current?.rewards ?? LOADING}
          editable={false}
          ownPreview={isOwn}
        />
      )}
    </ProfileShell>
  );
}
