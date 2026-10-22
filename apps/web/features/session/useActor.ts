"use client";

import { useEffect, useState } from "react";
import { getSession, onSessionRecovered } from "../../lib/auth";
import type { Actor } from "./session.api";

export type ActorState =
  | { status: "loading" }
  | { status: "signed_out" }
  /** The sign-in check itself failed (offline, timeout, 5xx): not the same as signed out. */
  | { status: "unavailable"; retry: () => void }
  | { status: "ready"; actor: Actor };

/** The signed-in user, as the room API needs them. */
export function useActor(): ActorState {
  const [state, setState] = useState<ActorState>({ status: "loading" });
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (cancelled) return;
      if (session.status === "signed_in") {
        const { user_id: userId, display_name: displayName } = session.user;
        // Same person again (a Retry and the recovery notice can both land): keep
        // the same object, so hooks that depend on `actor` don't refetch.
        setState((prev) =>
          prev.status === "ready" && prev.actor.userId === userId && prev.actor.displayName === displayName
            ? prev
            : { status: "ready", actor: { userId, displayName } },
        );
      } else if (session.status === "unavailable") {
        setState({
          status: "unavailable",
          retry: () => {
            setState({ status: "loading" });
            setAttempt((n) => n + 1);
          },
        });
      } else {
        setState({ status: "signed_out" });
      }
    });
    return () => {
      cancelled = true;
    };
  }, [attempt]);

  // Another component's Retry got an answer: look again instead of staying on the banner.
  useEffect(() => onSessionRecovered(() => setAttempt((n) => n + 1)), []);

  return state;
}
