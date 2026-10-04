"use client";

import { useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import {
  createRoom,
  getMySession,
  joinRoom,
  leaveRoom,
  SessionApiError,
  type Actor,
  type StudyRoom,
} from "./session.api";

/** What the user tried while already sitting in another room. */
type PendingAction = { kind: "join"; sessionId: string } | { kind: "create"; name: string; limit: number };

export interface RoomConflict {
  /** The room they are in now. */
  current: StudyRoom;
  pending: PendingAction;
}

export interface RoomActions {
  /** Room id being joined, for the button's busy state. */
  joiningId: string | null;
  /** Message for the last failed join, e.g. room full (UC-02 E-1/E-2). */
  joinError: string | null;
  clearJoinError: () => void;
  conflict: RoomConflict | null;
  resolving: boolean;
  join: (sessionId: string) => Promise<void>;
  /** Resolves with a message to show in the form, or null once created. */
  create: (name: string, limit: number) => Promise<string | null>;
  /** E-2/E-3 "leave that room": leave it, then retry what they wanted. */
  leaveCurrentAndContinue: () => Promise<void>;
  dismissConflict: () => void;
}

const JOIN_ERRORS: Partial<Record<SessionApiError["code"], string>> = {
  ROOM_FULL: "That dock just filled up. Here are the rooms with space.",
  ROOM_ENDED: "That room has closed. The list is fresh now.",
  ROOM_NOT_FOUND: "That room no longer exists.",
};

/**
 * Join and create, with the "you're already in a room" detour. Every
 * failure that means the list is out of date calls onStale so the page
 * refreshes it.
 */
export function useRoomActions(actor: Actor | null, onStale: () => void): RoomActions {
  const router = useRouter();
  const [joiningId, setJoiningId] = useState<string | null>(null);
  const [joinError, setJoinError] = useState<string | null>(null);
  const [conflict, setConflict] = useState<RoomConflict | null>(null);
  const [resolving, setResolving] = useState(false);

  const openConflict = useCallback(
    async (pending: PendingAction) => {
      if (!actor) return false;
      const me = await getMySession(actor).catch(() => null);
      if (!me?.in_session || !me.session) return false;
      setConflict({ current: me.session, pending });
      return true;
    },
    [actor],
  );

  const join = useCallback(
    async (sessionId: string) => {
      if (!actor) return;
      setJoiningId(sessionId);
      setJoinError(null);
      try {
        await joinRoom(actor, sessionId);
        router.push(`/rooms/${sessionId}`);
      } catch (err) {
        if (err instanceof SessionApiError && err.code === "ALREADY_IN_ROOM") {
          if (await openConflict({ kind: "join", sessionId })) return;
        }
        setJoinError(
          err instanceof SessionApiError ? (JOIN_ERRORS[err.code] ?? err.message) : "Couldn't join. Try again.",
        );
        onStale();
      } finally {
        setJoiningId(null);
      }
    },
    [actor, router, onStale, openConflict],
  );

  const create = useCallback(
    async (name: string, limit: number) => {
      if (!actor) return "Sign in to open a room.";
      try {
        const room = await createRoom(actor, name, limit);
        router.push(`/rooms/${room.id}`);
        return null;
      } catch (err) {
        if (err instanceof SessionApiError && err.code === "ALREADY_IN_ROOM") {
          if (await openConflict({ kind: "create", name, limit })) return null;
        }
        return err instanceof SessionApiError ? err.message : "Couldn't open the room. Try again.";
      }
    },
    [actor, router, openConflict],
  );

  const leaveCurrentAndContinue = useCallback(async () => {
    if (!actor || !conflict) return;
    setResolving(true);
    try {
      await leaveRoom(actor, conflict.current.id);
      const { pending } = conflict;
      setConflict(null);
      if (pending.kind === "join") await join(pending.sessionId);
      else {
        const error = await create(pending.name, pending.limit);
        if (error) setJoinError(error);
      }
    } catch {
      setJoinError("Couldn't leave your current room. Try again.");
      setConflict(null);
    } finally {
      setResolving(false);
    }
  }, [actor, conflict, join, create]);

  return {
    joiningId,
    joinError,
    clearJoinError: () => setJoinError(null),
    conflict,
    resolving,
    join,
    create,
    leaveCurrentAndContinue,
    dismissConflict: () => setConflict(null),
  };
}
