"use client";

import Link from "next/link";
import { useEffect, useState, type ReactNode } from "react";
import { getSession, onSessionRecovered, type Session } from "../lib/auth";
import { SessionUnavailableAlert } from "./profile/ProfileShell";
import { PixelPanel } from "./ui/PixelPanel";

type Gate = { status: "loading" } | Session;

/**
 * Shows `children` only to a signed-in admin. Everyone else gets a short
 * explanation instead of the page: sign in, retry, or "admins only".
 *
 * This is the page-level check. The data behind the page is protected by the
 * services themselves (the admin service refuses anyone who is not an admin).
 */
export function RequireAdmin({ children }: { children: ReactNode }) {
  const [gate, setGate] = useState<Gate>({ status: "loading" });
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (!cancelled) {
        setGate(session);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [attempt]);

  // Another component's Retry got an answer: look again.
  useEffect(() => onSessionRecovered(() => setAttempt((n) => n + 1)), []);

  if (gate.status === "signed_in" && gate.user.role === "ADMIN") {
    return <>{children}</>;
  }

  return (
    <main className="mx-auto w-full max-w-xl px-4 py-12">
      {gate.status === "loading" && (
        <PixelPanel aria-busy="true">
          <p className="text-base text-ink">Checking your account…</p>
        </PixelPanel>
      )}
      {gate.status === "unavailable" && (
        <SessionUnavailableAlert
          onRetry={() => {
            setGate({ status: "loading" });
            setAttempt((n) => n + 1);
          }}
        />
      )}
      {gate.status === "signed_out" && (
        <PixelPanel as="section" aria-labelledby="gate-heading">
          <h1 id="gate-heading" className="font-display text-2xl">
            Sign in to continue
          </h1>
          <p className="mt-2 text-base text-ink">This page is for admins. Sign in with an admin account.</p>
          <Link href="/signin" className="pixel-btn mt-4 min-h-11">
            Sign in
          </Link>
        </PixelPanel>
      )}
      {gate.status === "signed_in" && (
        <PixelPanel as="section" aria-labelledby="gate-heading">
          <h1 id="gate-heading" className="font-display text-2xl">
            Admins only
          </h1>
          <p className="mt-2 text-base text-ink">Your account doesn&apos;t have access to this page.</p>
          <Link href="/" className="pixel-btn pixel-btn--ghost mt-4 min-h-11">
            Back to home
          </Link>
        </PixelPanel>
      )}
    </main>
  );
}
