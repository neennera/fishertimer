"use client";

import { useCallback, useEffect, useState } from "react";
import { getMySession, listActiveRooms, type Actor, type MySession, type StudyRoom } from "./session.api";

/** How often the room list refreshes, so counts stay current (UC-02 S-2). */
const REFRESH_MS = 4000;

export interface RoomList {
  rooms: StudyRoom[] | null;
  /** The room the user is already in, if any. */
  mine: MySession | null;
  error: string | null;
  refresh: () => Promise<void>;
}

/**
 * The public room list, refreshed every few seconds and whenever the tab
 * comes back into view. A failed refresh keeps the last list and shows an
 * error, rather than blanking the page.
 */
export function useRoomList(actor: Actor | null): RoomList {
  const [rooms, setRooms] = useState<StudyRoom[] | null>(null);
  const [mine, setMine] = useState<MySession | null>(null);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const [list, me] = await Promise.all([listActiveRooms(), actor ? getMySession(actor) : Promise.resolve(null)]);
      setRooms(list);
      setMine(me);
      setError(null);
    } catch {
      setError("Can't reach the study rooms. Retrying…");
    }
  }, [actor]);

  useEffect(() => {
    void refresh();
    const timer = window.setInterval(() => {
      if (document.visibilityState === "visible") void refresh();
    }, REFRESH_MS);
    const onVisible = () => {
      if (document.visibilityState === "visible") void refresh();
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, [refresh]);

  return { rooms, mine, error, refresh };
}
