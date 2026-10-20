"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getSession, onSessionRecovered, type SessionUser } from "../lib/auth";
import { Header } from "./Header";
import { PixelAlert } from "./ui/PixelAlert";
import { useSignOut } from "./profile/ProfileShell";

export interface SessionHeaderProps {
  className?: string;
}

/** Header wired to the current session. Signing out goes to /signin. */
export function SessionHeader({ className }: SessionHeaderProps) {
  const router = useRouter();
  const [user, setUser] = useState<SessionUser | null>(null);
  const [signedOut, setSignedOut] = useState(false);
  const [unavailable, setUnavailable] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const { signingOut, signOutError, handleSignOut } = useSignOut(() => {
    setUser(null);
    router.replace("/signin");
  });

  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (!cancelled) {
        setUser(session.status === "signed_in" ? session.user : null);
        setSignedOut(session.status === "signed_out");
        setUnavailable(session.status === "unavailable");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [attempt]);

  // Another component's Retry got an answer: look again instead of staying on Retry.
  useEffect(() => onSessionRecovered(() => setAttempt((n) => n + 1)), []);

  return (
    <>
      <Header
        className={className}
        user={user ? { displayName: user.display_name, avatarUrl: user.avatar_url } : null}
        onSignOut={handleSignOut}
        signingOut={signingOut}
        showSignIn={signedOut}
        onRetrySession={
          unavailable
            ? () => {
                setUnavailable(false);
                setAttempt((n) => n + 1);
              }
            : undefined
        }
      />
      {signOutError && <PixelAlert className="mx-auto mt-3 w-full max-w-3xl">{signOutError}</PixelAlert>}
    </>
  );
}
