"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { SessionHeader } from "../../components/SessionHeader";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { PixelPanel } from "../../components/ui/PixelPanel";
import { ParallaxScene } from "../../components/ui/ParallaxScene";
import { useActor } from "../../features/session/useActor";
import { getTimerSettings, type TimerState } from "../../features/timer/timer.api";
import { TimerSettingsForm } from "../../features/timer/TimerSettingsForm";
import { SIGNIN_SCENE_LAYERS } from "../../lib/scenes/signin-scene";

/** TimerSetting: the user's default focus and break lengths (UC-05). */
export default function SettingsPage() {
  const actorState = useActor();
  const actor = actorState.status === "ready" ? actorState.actor : null;
  const [timer, setTimer] = useState<TimerState | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!actor) return;
    let cancelled = false;
    getTimerSettings(actor.userId)
      .then((t) => !cancelled && setTimer(t))
      .catch(() => !cancelled && setError("Can't reach the study timer right now."));
    return () => {
      cancelled = true;
    };
  }, [actor]);

  return (
    <div className="flex min-h-screen flex-col">
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <SessionHeader />
      <main className="flex flex-1 items-start justify-center px-4 pt-8 pb-16 sm:pt-12">
        <PixelPanel as="section" aria-labelledby="settings-heading" className="w-full max-w-lg">
          <p className="font-label text-[10px] uppercase tracking-[0.12em] text-bark">Settings</p>
          <h1 id="settings-heading" className="font-display text-4xl leading-none">
            Timer defaults
          </h1>
          <p className="mt-2 text-sm text-bark">How long a focus block and a break start out as, in every room.</p>

          <div className="mt-6">
            {actorState.status === "signed_out" ? (
              <div className="pixel-placeholder text-center">
                <p>Sign in to save your timer defaults.</p>
                <Link href="/signin" className="pixel-btn mt-4">
                  Sign in
                </Link>
              </div>
            ) : error ? (
              <PixelAlert>{error}</PixelAlert>
            ) : timer && actor ? (
              <TimerSettingsForm timer={timer} userId={actor.userId} onSaved={setTimer} />
            ) : (
              <div className="flex flex-col gap-4" aria-busy="true">
                <div className="pixel-skeleton h-24" />
                <div className="pixel-skeleton h-24" />
              </div>
            )}
          </div>

          <div className="mt-6 border-t border-[var(--color-rule)] pt-4">
            <Link href="/rooms" className="font-label text-[10px] uppercase tracking-[0.12em] text-bark hover:text-ink">
              ← Back to the lake
            </Link>
          </div>
        </PixelPanel>
      </main>
    </div>
  );
}
