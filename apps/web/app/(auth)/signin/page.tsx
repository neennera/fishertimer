"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelPanel } from "../../../components/ui/PixelPanel";
import { handleAuthCallback, signInWithGoogle } from "../../../lib/auth";
import { SignInError } from "./SignInError";

// No app logo yet — signpost.png / signboard.png aren't drawn. When they
// land, swap the panel header in pixel.css only (DESIGN_SYSTEM.md,
// "Replacing CSS surfaces with images"); nothing here should need to change.
function SignInPanel() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [signingIn, setSigningIn] = useState(false);

  async function handleSignIn() {
    setSigningIn(true);

    // Real mode: this navigates the whole browser away to the gateway and
    // never resolves in this page, so nothing below runs. Mock mode: it
    // resolves after a short delay — the actual new-vs-existing-user
    // decision still comes from handleAuthCallback(), reading the same
    // ?mockScenario= param /welcome's fallback reads, so
    // /signin?mockScenario=setup-default plus a click walks through the
    // whole flow in one step during testing.
    await signInWithGoogle();
    const result = await handleAuthCallback(searchParams);

    if (!result.ok) {
      setSigningIn(false);
      const nextParams = new URLSearchParams(searchParams);
      nextParams.set("error", result.error);
      router.replace(`/signin?${nextParams.toString()}`);
      return;
    }

    router.push(result.session.isFirstLogin ? "/welcome" : "/");
  }

  return (
    <PixelPanel className="pixel-panel--enter w-full max-w-sm text-center">
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
