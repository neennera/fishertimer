"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { TimerState } from "../timer/timer.api";
import {
  getRoom,
  getRoomTimers,
  sendHeartbeat,
  SessionApiError,
  type Actor,
  type LeaveReason,
  type RoomParticipant,
  type StudyRoom,
} from "./session.api";

/** Roster refresh: the NFR asks for room changes within 2 seconds. */
const ROOM_REFRESH_MS = 2000;
/** Everyone's timer state on the dock. */
const TIMERS_REFRESH_MS = 3000;
/** Presence ping; the server auto-leaves after ~60s without one (UC-03 E-3). */
const HEARTBEAT_MS = 10_000;

export type RoomPresence =
  | { kind: "loading" }
  | { kind: "member" }
  /** Viewing a room the user is not in (e.g. a shared link). */
  | { kind: "visitor" }
  /** Was in the room and is not any more, for a reason they did not choose. */
  | { kind: "removed"; reason: LeaveReason | "" }
  | { kind: "not_found" };

export interface LiveRoom {
  room: StudyRoom | null;
  participants: RoomParticipant[];
  /** Timer per user id, for the dock. */
  timers: Record<string, TimerState>;
  presence: RoomPresence;
  /** Last refresh failed; the view keeps the previous snapshot. */
  stale: boolean;
  refresh: () => Promise<void>;
  /** Call before leaving on purpose, so it is not reported as a removal. */
  markLeaving: () => void;
}

/**
 * One room, kept live by polling: the roster every 2s, timers every 3s, and
 * a heartbeat every 10s while the user is a member. Detects being kicked,
 * timed out or the room closing, and reports why.
 */
export function useRoom(actor: Actor | null, sessionId: string): LiveRoom {
  const [room, setRoom] = useState<StudyRoom | null>(null);
  const [participants, setParticipants] = useState<RoomParticipant[]>([]);
  const [timers, setTimers] = useState<Record<string, TimerState>>({});
  const [presence, setPresence] = useState<RoomPresence>({ kind: "loading" });
  const [stale, setStale] = useState(false);

  const wasMember = useRef(false);
  const leaving = useRef(false);
  const removed = useRef(false);

  const markLeaving = useCallback(() => {
    leaving.current = true;
  }, []);

  const noticeRemoval = useCallback(
    async (fallback: LeaveReason | "") => {
      if (!actor || leaving.current || removed.current) return;
      removed.current = true;
      let reason: LeaveReason | "" = fallback;
      try {
        const beat = await sendHeartbeat(actor, sessionId);
        if (beat.reason) reason = beat.reason;
      } catch {
        // keep the fallback
      }
      setPresence({ kind: "removed", reason });
    },
    [actor, sessionId],
  );

  const refresh = useCallback(async () => {
    if (!actor || removed.current) return;
    try {
      const snapshot = await getRoom(actor, sessionId);
      setRoom(snapshot.session);
      setParticipants(snapshot.participants);
      setStale(false);

      const isMember = snapshot.participants.some((p) => p.user_id === actor.userId);
      if (isMember) {
        wasMember.current = true;
        setPresence({ kind: "member" });
      } else if (wasMember.current) {
        void noticeRemoval(snapshot.session.status === "ENDED" ? "SESSION_ENDED" : "");
      } else {
        setPresence({ kind: "visitor" });
      }
    } catch (err) {
      if (err instanceof SessionApiError && (err.code === "ROOM_NOT_FOUND" || err.code === "INVALID")) {
        setPresence({ kind: "not_found" });
      } else {
        setStale(true);
      }
    }
  }, [actor, sessionId, noticeRemoval]);

  const refreshTimers = useCallback(async () => {
    try {
      const list = await getRoomTimers(sessionId);
      setTimers(Object.fromEntries(list.map((t) => [t.user_id, t])));
    } catch {
      // The dock just keeps the last known states.
    }
  }, [sessionId]);

  const member = presence.kind === "member";

  useEffect(() => {
    void refresh();
    const id = window.setInterval(() => void refresh(), ROOM_REFRESH_MS);
    return () => window.clearInterval(id);
  }, [refresh]);

  useEffect(() => {
    void refreshTimers();
    const id = window.setInterval(() => void refreshTimers(), TIMERS_REFRESH_MS);
    return () => window.clearInterval(id);
  }, [refreshTimers]);

  // Heartbeat only while a member. An inactive answer means the server has
  // already removed us (kick, idle, timeout, room closed).
  useEffect(() => {
    if (!actor || !member) return;
    const beat = async () => {
      try {
        const res = await sendHeartbeat(actor, sessionId);
        if (!res.active) void noticeRemoval(res.reason);
      } catch {
        // A missed beat is fine; the server allows several.
      }
    };
    void beat();
    const id = window.setInterval(beat, HEARTBEAT_MS);
    return () => window.clearInterval(id);
  }, [actor, member, sessionId, noticeRemoval]);

  return { room, participants, timers, presence, stale, refresh, markLeaving };
}
