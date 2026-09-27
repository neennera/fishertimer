"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ProfileSections } from "../../components/profile/ProfileSections";
import { ProfileShell, useSignedInUser } from "../../components/profile/ProfileShell";
import { loadPanels, toPublicProfile, type PanelData } from "../../lib/profile";
import type { RewardsSummary, TimerStatistics } from "../../lib/profile-types";

const LOADING = { status: "loading" } as const;

export default function AccountPage() {
  const router = useRouter();
  const { user, setUser, signingOut, handleSignOut } = useSignedInUser();
  const userId = user?.user_id ?? null;
  // Needs the user id from /me. Keyed by it, so a stale answer never shows.
  const [panels, setPanels] = useState<{
    userId: string;
    stats: PanelData<TimerStatistics>;
    rewards: PanelData<RewardsSummary>;
  } | null>(null);

  useEffect(() => {
    if (!userId) {
      return;
    }
    let cancelled = false;
    void loadPanels(userId).then((loaded) => {
      if (!cancelled) {
        setPanels({ userId, ...loaded });
      }
    });
    return () => {
      cancelled = true;
    };
  }, [userId]);

  const current = panels?.userId === userId ? panels : null;

  return (
    <ProfileShell title="Account" viewer={user} signingOut={signingOut} onSignOut={handleSignOut}>
      <ProfileSections
        profile={user ? toPublicProfile(user) : null}
        stats={current?.stats ?? LOADING}
        rewards={current?.rewards ?? LOADING}
        editable
        owner={
          user
            ? {
                email: user.email,
                onNameSaved: setUser,
                onSignedOut: () => router.replace("/signin"),
              }
            : undefined
        }
      />
    </ProfileShell>
  );
}
