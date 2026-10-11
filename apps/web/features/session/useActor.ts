"use client";

import { useEffect, useState } from "react";
import { getSession } from "../../lib/auth";
import type { Actor } from "./session.api";

export type ActorState = { status: "loading" } | { status: "signed_out" } | { status: "ready"; actor: Actor };

/** The signed-in user, as the room API needs them. */
export function useActor(): ActorState {
  const [state, setState] = useState<ActorState>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;
    void getSession().then((session) => {
      if (cancelled) return;
      setState(
        session.status === "signed_in"
          ? { status: "ready", actor: { userId: session.user.user_id, displayName: session.user.display_name } }
          : { status: "signed_out" },
      );
    });
    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
