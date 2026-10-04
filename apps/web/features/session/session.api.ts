// Study room data layer (UC-01 Create, UC-02 Join, UC-03 Leave). The only
// file that talks to the backend for rooms; components call these functions
// and never use fetch directly.
//
// Requests go browser -> Next.js rewrite -> API Gateway (REST) -> Study
// Session (gRPC). The gateway acts as the signed-in user; `user_id` /
// `display_name` are only read when there is no session cookie (mock auth in
// development), so they are always sent.

import type { TimerState } from '../timer/timer.api';

export type SessionStatus = 'ACTIVE' | 'ENDED';

export interface StudyRoom {
  id: string;
  name: string;
  creator_id: string;
  participant_limit: number;
  participant_count: number;
  status: SessionStatus;
  created_at: string;
  ended_at?: string;
  end_reason?: 'EMPTY' | 'ADMIN_CLOSED' | 'TIMEOUT_24H';
}

export interface RoomParticipant {
  user_id: string;
  display_name: string;
  joined_at: string;
  last_seen_at: string;
}

export interface RoomSnapshot {
  session: StudyRoom;
  participants: RoomParticipant[];
}

export interface MySession {
  in_session: boolean;
  session?: StudyRoom;
  participant?: RoomParticipant;
}

export interface LeaveResult {
  left: boolean;
  session_ended: boolean;
  session_id: string;
  joined_at: string;
  left_at: string;
}

/** Why a participant is no longer in the room (from the heartbeat). */
export type LeaveReason = 'LEFT' | 'KICKED' | 'DISCONNECT_TIMEOUT' | 'IDLE_TIMEOUT' | 'SESSION_ENDED';

export interface Heartbeat {
  active: boolean;
  reason: LeaveReason | '';
}

/** Stable codes from the gateway, so pages branch without parsing text. */
export type SessionErrorCode =
  | 'INVALID'
  | 'SIGNED_OUT'
  | 'ROOM_NOT_FOUND'
  | 'ROOM_ENDED'
  | 'ROOM_FULL'
  | 'ALREADY_IN_ROOM'
  | 'UNAVAILABLE';

export class SessionApiError extends Error {
  constructor(
    readonly code: SessionErrorCode,
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'SessionApiError';
  }
}

/** Room rules, mirrored from services/study-session/internal/domain. */
export const ROOM_RULES = {
  minLimit: 1,
  maxLimit: 5,
  maxNameLength: 60,
} as const;

export interface Actor {
  userId: string;
  displayName: string;
}

async function call<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`/api/session/${path}`, {
      credentials: 'same-origin',
      ...init,
      headers: { 'Content-Type': 'application/json', ...init?.headers },
    });
  } catch {
    throw new SessionApiError('UNAVAILABLE', 0, "Can't reach the study rooms right now.");
  }
  const body = (await res.json().catch(() => ({}))) as { error?: string; code?: SessionErrorCode };
  if (!res.ok) {
    throw new SessionApiError(
      body.code ?? 'UNAVAILABLE',
      res.status,
      body.error ?? `Study rooms answered ${res.status}.`,
    );
  }
  return body as T;
}

function post<T>(path: string, actor: Actor, body: Record<string, unknown> = {}): Promise<T> {
  return call<T>(path, {
    method: 'POST',
    body: JSON.stringify({ ...body, user_id: actor.userId, display_name: actor.displayName }),
  });
}

function query(actor: Actor, extra: Record<string, string> = {}): string {
  return new URLSearchParams({ ...extra, user_id: actor.userId }).toString();
}

export async function listActiveRooms(): Promise<StudyRoom[]> {
  const { sessions } = await call<{ sessions: StudyRoom[] }>('active');
  return sessions;
}

export function getMySession(actor: Actor): Promise<MySession> {
  return call<MySession>(`me?${query(actor)}`);
}

export function getRoom(actor: Actor, sessionId: string): Promise<RoomSnapshot> {
  return call<RoomSnapshot>(`room?${query(actor, { session_id: sessionId })}`);
}

export function createRoom(actor: Actor, name: string, participantLimit: number): Promise<StudyRoom> {
  return post<StudyRoom>('create', actor, { name, participant_limit: participantLimit });
}

export function joinRoom(actor: Actor, sessionId: string): Promise<StudyRoom> {
  return post<StudyRoom>('join', actor, { session_id: sessionId });
}

export function leaveRoom(actor: Actor, sessionId: string): Promise<LeaveResult> {
  return post<LeaveResult>('leave', actor, { session_id: sessionId });
}

export function sendHeartbeat(actor: Actor, sessionId: string): Promise<Heartbeat> {
  return post<Heartbeat>('heartbeat', actor, { session_id: sessionId });
}

/** Everyone's timer in the room, for the dock (GET /api/timer/room). */
export async function getRoomTimers(sessionId: string): Promise<TimerState[]> {
  const res = await fetch(`/api/timer/room?${new URLSearchParams({ session_id: sessionId })}`, {
    credentials: 'same-origin',
  });
  if (!res.ok) throw new Error(`timer room ${res.status}`);
  const { timers } = (await res.json()) as { timers: TimerState[] };
  return timers;
}

/** Validates the create form the same way the server does (UC-01 E-1). */
export function validateRoom(name: string, limit: number): { name?: string; limit?: string } {
  const errors: { name?: string; limit?: string } = {};
  const trimmed = name.trim();
  if (!trimmed) errors.name = 'Give your room a name.';
  else if ([...trimmed].length > ROOM_RULES.maxNameLength)
    errors.name = `Keep it under ${ROOM_RULES.maxNameLength} characters.`;
  if (!Number.isInteger(limit) || limit < ROOM_RULES.minLimit || limit > ROOM_RULES.maxLimit)
    errors.limit = `Pick ${ROOM_RULES.minLimit} to ${ROOM_RULES.maxLimit} spots.`;
  return errors;
}
