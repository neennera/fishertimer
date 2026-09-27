"use client";

import { useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { Header } from "../Header";
import { getSession, signOut, type SessionUser } from "../../lib/auth";

/**
 * The signed-in viewer, for pages that need one: redirects to /signin when
 * signed out (or /welcome mid-sign-up). `user` is null while loading.
 */
export function useSignedInUser() {
  const router = useRouter();
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
        router.replace("/welcome");
      } else {
        router.replace("/signin");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [router]);

  async function handleSignOut() {
    setSigningOut(true);
    try {
      await signOut();
      router.replace("/signin");
    } catch {
      // Gateway unreachable: the cookie is still set, so stay signed in.
      setSigningOut(false);
    }
  }

  return { user, setUser, signingOut, handleSignOut };
}

/** Wood backdrop, header and the centred column shared by profile pages. */
export function ProfileShell({
  title,
  viewer,
  signingOut,
  onSignOut,
  children,
}: {
  title: string;
  viewer: SessionUser | null;
  signingOut: boolean;
  onSignOut: () => void;
  children: ReactNode;
}) {
  return (
    <div className="pixel-wood min-h-screen">
      <Header
        user={viewer ? { displayName: viewer.display_name } : null}
        onSignOut={onSignOut}
        signingOut={signingOut}
      />
      <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 md:py-12">
        <h1 className="sr-only">{title}</h1>
        {children}
      </main>
    </div>
  );
}
