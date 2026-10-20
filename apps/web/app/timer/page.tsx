"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { SessionHeader } from "../../components/SessionHeader";
import { ParallaxScene } from "../../components/ui/ParallaxScene";
import { getMySession } from "../../features/session/session.api";
import { useActor } from "../../features/session/useActor";
import { SIGNIN_SCENE_LAYERS } from "../../lib/scenes/signin-scene";

/**
 * The timer only runs inside a room (UC-05: the user is an active
 * participant), so this sends the user to the room they are in, or to the
 * lake to pick or open one. The lake also handles being signed out.
 */
export default function TimerPage() {
  const router = useRouter();
  const actorState = useActor();

  useEffect(() => {
    if (actorState.status === "loading") return;
    if (actorState.status === "signed_out") {
      router.replace("/rooms");
      return;
    }
    let cancelled = false;
    getMySession(actorState.actor)
      .then((me) => {
        if (!cancelled) router.replace(me.in_session && me.session ? `/rooms/${me.session.id}` : "/rooms");
      })
      .catch(() => {
        if (!cancelled) router.replace("/rooms");
      });
    return () => {
      cancelled = true;
    };
  }, [actorState, router]);

  return (
    <div className="flex min-h-screen flex-col">
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <SessionHeader />
      <main className="flex flex-1 items-start justify-center px-4 pt-12 pb-16">
        <p className="pixel-panel w-full max-w-md text-center text-sm text-bark" aria-live="polite">
          Finding your dock…
        </p>
      </main>
    </div>
  );
}
