"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ProfileSections } from "../../../components/profile/ProfileSections";
import { ProfileShell, useSignedInUser } from "../../../components/profile/ProfileShell";
import { PixelPanel } from "../../../components/ui/PixelPanel";
import { getPublicProfile, ownProfileData, type ProfileData } from "../../../lib/profile";

export default function PublicProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const { user, signingOut, handleSignOut } = useSignedInUser();
  // Keyed by userId, so moving to another profile never shows the last one.
  const [fetched, setFetched] = useState<{ userId: string; data: ProfileData | null } | null>(null);

  const isOwn = user !== null && user.user_id === userId;

  useEffect(() => {
    if (!user || user.user_id === userId) {
      return;
    }
    let cancelled = false;
    void getPublicProfile(userId).then((data) => {
      if (!cancelled) {
        setFetched({ userId, data });
      }
    });
    return () => {
      cancelled = true;
    };
  }, [user, userId]);

  // undefined = still loading, null = no such user. Your own profile comes
  // from the session, the same data /account shows.
  const data: ProfileData | null | undefined = isOwn
    ? ownProfileData(user)
    : fetched?.userId === userId
      ? fetched.data
      : undefined;

  return (
    <ProfileShell title="Profile" viewer={user} signingOut={signingOut} onSignOut={handleSignOut}>
      {data === null ? (
        <PixelPanel as="section" className="text-center">
          <h2 className="font-display text-3xl leading-none">Fisher not found</h2>
          <p className="mt-3 text-sm text-bark">
            This profile doesn&rsquo;t exist, or it has moved.
          </p>
          <p className="mt-4 text-sm">
            <Link href="/account" className="pixel-link">
              Back to your account
            </Link>
          </p>
        </PixelPanel>
      ) : (
        <ProfileSections data={data ?? null} editable={false} ownPreview={isOwn} />
      )}
    </ProfileShell>
  );
}
