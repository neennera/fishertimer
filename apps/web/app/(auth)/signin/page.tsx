"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelPanel } from "../../../components/ui/PixelPanel";
import { getSession, signInWithGoogle } from "../../../lib/auth";
import { SignInError } from "./SignInError";

function SignInPanel() {
  const router = useRouter();
  const [signingIn, setSigningIn] = useState(false);

  // The backend sends every Google round trip back here (signInWithGoogle()'s
  // default `next`) — a successful sign-in lands signed in, so move on.
  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (!cancelled && session.status === "signed_in") {
        router.replace("/");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [router]);

  async function handleSignIn() {
    setSigningIn(true);
    // A full-page navigation to Google via the backend (mock mode plays out
    // the same redirect for ?mockScenario=), so the button just stays in its
    // busy state until the page unloads.
    await signInWithGoogle();
  }

  return (
    <PixelPanel className="w-full max-w-sm text-center">
      <h1 className="font-display text-3xl leading-none">Fisher Timer</h1>
      <p className="mt-3 text-sm text-bark">
        Sign in to start your focus session
      </p>

      <PixelButton block className="mt-6" onClick={handleSignIn} disabled={signingIn}>
        {signingIn ? "Signing in…" : "Sign in with Google"}
      </PixelButton>

      <SignInError />
    </PixelPanel>
  );
}

export default function SignInPage() {
  return (
    <Suspense fallback={null}>
      <SignInPanel />
    </Suspense>
  );
}
