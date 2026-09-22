"use client";

import { Suspense } from "react";
import { Header } from "../../components/Header";
import { ParallaxScene } from "../../components/ui/ParallaxScene";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelPanel } from "../../components/ui/PixelPanel";
import { signInWithGoogle } from "../../lib/auth";
import { SIGNIN_SCENE_LAYERS } from "../../lib/scenes/signin-scene";
import { SignInError } from "./SignInError";

// No app logo yet — signpost.png / signboard.png aren't drawn. When they
// land, swap the panel header in pixel.css only (DESIGN_SYSTEM.md,
// "Replacing CSS surfaces with images"); nothing here should need to change.
export default function SignInPage() {
  const handleSignIn = () => {
    void signInWithGoogle();
  };

  return (
    <div className="flex min-h-screen flex-col">
      {/* Fixed to the viewport by its own CSS — not sized by this layout. */}
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />

      <Header />

      <div className="flex flex-1 items-center justify-center px-4 py-12">
        <PixelPanel className="w-full max-w-sm text-center">
          <h1 className="font-display text-3xl leading-none">Fisher Timer</h1>
          <p className="mt-3 text-sm text-bark">
            Sign in to start your focus session
          </p>

          <PixelButton block className="mt-6" onClick={handleSignIn}>
            Sign in with Google
          </PixelButton>

          <Suspense fallback={null}>
            <SignInError />
          </Suspense>
        </PixelPanel>
      </div>
    </div>
  );
}
