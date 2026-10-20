"use client";

import { useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { Header } from "../Header";
import { PixelAlert } from "../ui/PixelAlert";
import { PixelButton } from "../ui/PixelButton";
import { getSession, onSessionRecovered, signOut, type SessionUser } from "../../lib/auth";

export const SIGN_OUT_FAILED_MESSAGE = "Couldn't sign you out. Check your connection and try again.";
export const SESSION_UNAVAILABLE_MESSAGE = "Couldn't reach the server to check your sign-in. Check your connection and try again.";

/**
 * Sign-out for a header: `error` is set when the request failed (the cookie is
 * still set, so the user is still signed in) and cleared on the next try.
 */
export function useSignOut(onSignedOut: () => void) {
  const [signingOut, setSigningOut] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSignOut() {
    setSigningOut(true);
    setError(null);
    try {
      await signOut();
      onSignedOut();
    } catch {
      setError(SIGN_OUT_FAILED_MESSAGE);
      setSigningOut(false);
    }
  }

  return { signingOut, signOutError: error, handleSignOut };
}

/**
 * The signed-in viewer, for pages that need one: redirects to /signin when
 * signed out. `user` is null while loading. When the session check itself
 * fails (offline, timeout, 5xx) the user is NOT sent to /signin: `unavailable`
 * is set instead, and `retry` asks again.
 */
export function useSignedInUser() {
  const router = useRouter();
  const [user, setUser] = useState<SessionUser | null>(null);
  const [unavailable, setUnavailable] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const { signingOut, signOutError, handleSignOut } = useSignOut(() => router.replace("/signin"));

  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (cancelled) {
        return;
      }
      if (session.status === "signed_in") {
        setUser(session.user);
      } else if (session.status === "unavailable") {
        setUnavailable(true);
      } else {
        router.replace("/signin");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [router, attempt]);

  function retry() {
    setUnavailable(false);
    setAttempt((n) => n + 1);
  }

  // Another component's Retry got an answer: look again instead of staying on the banner.
  useEffect(
    () =>
      onSessionRecovered(() => {
        setUnavailable(false);
        setAttempt((n) => n + 1);
      }),
    [],
  );

  return { user, setUser, signingOut, signOutError, handleSignOut, unavailable, retry };
}

/** The "couldn't check your sign-in" banner with its Retry button. */
export function SessionUnavailableAlert({ onRetry, className }: { onRetry: () => void; className?: string }) {
  return (
    <PixelAlert className={`flex flex-wrap items-center justify-between gap-3 ${className ?? ""}`}>
      <span>{SESSION_UNAVAILABLE_MESSAGE}</span>
      <PixelButton small className="min-h-11 sm:min-h-0" onClick={onRetry}>
        Retry
      </PixelButton>
    </PixelAlert>
  );
}

/** Wood backdrop, header and the centred column shared by profile pages. */
export function ProfileShell({
  title,
  viewer,
  signingOut,
  onSignOut,
  signOutError,
  unavailable = false,
  onRetry,
  children,
}: {
  title: string;
  viewer: SessionUser | null;
  signingOut: boolean;
  onSignOut: () => void;
  /** Set when the last sign-out attempt failed. */
  signOutError?: string | null;
  /** The sign-in check failed; shows a Retry banner. Needs `onRetry`. */
  unavailable?: boolean;
  onRetry?: () => void;
  children: ReactNode;
}) {
  return (
    <div className="pixel-wood min-h-screen">
      <Header
        user={viewer ? { displayName: viewer.display_name, avatarUrl: viewer.avatar_url } : null}
        onSignOut={onSignOut}
        signingOut={signingOut}
      />
      <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 md:py-12">
        <h1 className="sr-only">{title}</h1>
        {signOutError && <PixelAlert>{signOutError}</PixelAlert>}
        {unavailable && onRetry && <SessionUnavailableAlert onRetry={onRetry} />}
        {children}
      </main>
    </div>
  );
}
