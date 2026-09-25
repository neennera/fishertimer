"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getSession, signOut, type SessionUser } from "../lib/auth";
import { Header } from "./Header";

export interface SessionHeaderProps {
  className?: string;
}

/**
 * Header wired to the current session: shows the avatar and sign-out button
 * once GET /me reports signed_in. Signing out clears the account service's
 * ft_session / ft_signup cookies (POST /api/auth/signout), then sends the
 * browser to /signin.
 */
export function SessionHeader({ className }: SessionHeaderProps) {
  const router = useRouter();
  const [user, setUser] = useState<SessionUser | null>(null);
  const [signingOut, setSigningOut] = useState(false);

  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (!cancelled) {
        setUser(session.status === "signed_in" ? session.user : null);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleSignOut() {
    setSigningOut(true);
    try {
      await signOut();
      setUser(null);
      router.replace("/signin");
    } catch {
      // Gateway unreachable: the cookie is still set, so stay signed in and
      // let them try again.
      setSigningOut(false);
    }
  }

  return (
    <Header
      className={className}
      user={user ? { displayName: user.display_name } : null}
      onSignOut={handleSignOut}
      signingOut={signingOut}
    />
  );
}
