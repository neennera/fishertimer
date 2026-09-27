"use client";

import { ProfileSections } from "../../components/profile/ProfileSections";
import { ProfileShell, useSignedInUser } from "../../components/profile/ProfileShell";
import { ownProfileData } from "../../lib/profile";

export default function AccountPage() {
  const { user, setUser, signingOut, handleSignOut } = useSignedInUser();

  return (
    <ProfileShell title="Account" viewer={user} signingOut={signingOut} onSignOut={handleSignOut}>
      <ProfileSections
        data={user ? ownProfileData(user) : null}
        editable
        owner={user ? { email: user.email, onNameSaved: setUser } : undefined}
      />
    </ProfileShell>
  );
}
