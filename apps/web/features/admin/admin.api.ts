// Admin data access. Mock-backed for now; swap the bodies for adminApiFetch
// (lib/admin-api.ts) once services/admin exposes the endpoints.

import { MOCK_ADMIN_SESSIONS } from '../../lib/mocks/admin-sessions.mock';
import type { AdminSession, AdminSessionDetail } from './types';

const LATENCY_MS = 250;
const delay = () => new Promise((r) => setTimeout(r, LATENCY_MS));

/** GET /api/v1/admin/sessions?active=true — only rooms still open. */
export async function listSessions(): Promise<AdminSession[]> {
  await delay();
  return MOCK_ADMIN_SESSIONS.filter((s) => s.is_active).map(({ participants: _p, ...s }) => s);
}

/** GET /api/v1/admin/sessions/:id — null when not found. */
export async function getSession(id: string): Promise<AdminSessionDetail | null> {
  await delay();
  return MOCK_ADMIN_SESSIONS.find((s) => s.session_id === id) ?? null;
}

/**
 * POST /api/v1/admin/sessions/:id/participants/:userId/kick — admin command
 * that makes Study Session run LeaveSession() for the member.
 * Returns the updated session, or null when session/member is not found.
 */
export async function kickParticipant(
  sessionId: string,
  userId: string,
): Promise<AdminSessionDetail | null> {
  await delay();
  const session = MOCK_ADMIN_SESSIONS.find((s) => s.session_id === sessionId);
  const member = session?.participants.find((p) => p.user_id === userId);
  if (!session || !member || member.left_at !== null) {
    return null;
  }
  member.left_at = new Date().toISOString();
  session.participant_count = session.participants.filter((p) => p.left_at === null).length;
  return { ...session, participants: session.participants.map((p) => ({ ...p })) };
}

/**
 * POST /api/v1/admin/sessions/:id/end — admin command that makes Study
 * Session run EndSession(), disconnecting everyone still in the room.
 * Returns the updated session, or null when not found or already ended.
 */
export async function endSession(sessionId: string): Promise<AdminSessionDetail | null> {
  await delay();
  const session = MOCK_ADMIN_SESSIONS.find((s) => s.session_id === sessionId);
  if (!session || !session.is_active) {
    return null;
  }
  const now = new Date().toISOString();
  session.is_active = false;
  session.ended_at = now;
  session.participants.forEach((p) => {
    if (p.left_at === null) p.left_at = now;
  });
  session.participant_count = 0;
  return { ...session, participants: session.participants.map((p) => ({ ...p })) };
}
