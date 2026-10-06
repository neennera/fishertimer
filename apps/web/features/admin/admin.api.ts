// Admin data access. Mock-backed for now; swap the bodies for adminApiFetch
// (lib/admin-api.ts) once services/admin exposes the endpoints.

import {
  MOCK_ADMIN,
  MOCK_ADMIN_LOGS,
  MOCK_ADMIN_SESSIONS,
} from '../../lib/mocks/admin-sessions.mock';
import type {
  AdminLog,
  AdminLogAction,
  AdminSession,
  AdminSessionDetail,
} from './types';

const LATENCY_MS = 250;
const delay = () => new Promise((r) => setTimeout(r, LATENCY_MS));

/** The backend writes one admin_logs row per action; the mock does the same. */
function recordLog(
  action: AdminLogAction,
  target_id: string,
  target_label: string,
  reason: string | undefined,
) {
  MOCK_ADMIN_LOGS.unshift({
    log_id: `log-${Date.now()}`,
    admin_id: MOCK_ADMIN.id,
    admin_name: MOCK_ADMIN.name,
    action,
    target_id,
    target_label,
    reason: reason?.trim() || null,
    created_at: new Date().toISOString(),
  });
}

function closeRoom(session: AdminSessionDetail) {
  const now = new Date().toISOString();
  session.is_active = false;
  session.ended_at = now;
  session.participants.forEach((p) => {
    if (p.left_at === null) p.left_at = now;
  });
  session.participant_count = 0;
}

const snapshot = (s: AdminSessionDetail): AdminSessionDetail => ({
  ...s,
  participants: s.participants.map((p) => ({ ...p })),
});

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
 * that makes Study Session run LeaveSession() for the member. Removing the
 * last member ends the room. Logged as KICK_USER.
 * Returns the updated session, or null when session/member is not found.
 */
export async function kickParticipant(
  sessionId: string,
  userId: string,
  reason?: string,
): Promise<AdminSessionDetail | null> {
  await delay();
  const session = MOCK_ADMIN_SESSIONS.find((s) => s.session_id === sessionId);
  const member = session?.participants.find((p) => p.user_id === userId);
  if (!session || !member || !session.is_active || member.left_at !== null) {
    return null;
  }
  member.left_at = new Date().toISOString();
  session.participant_count = session.participants.filter((p) => p.left_at === null).length;
  if (session.participant_count === 0) {
    closeRoom(session);
  }
  recordLog('KICK_USER', member.user_id, member.display_name, reason);
  return snapshot(session);
}

/**
 * POST /api/v1/admin/sessions/:id/end — admin command that makes Study
 * Session run EndSession(), disconnecting everyone still in the room.
 * Logged as FORCE_CLOSE_SESSION.
 * Returns the updated session, or null when not found or already ended.
 */
export async function endSession(
  sessionId: string,
  reason?: string,
): Promise<AdminSessionDetail | null> {
  await delay();
  const session = MOCK_ADMIN_SESSIONS.find((s) => s.session_id === sessionId);
  if (!session || !session.is_active) {
    return null;
  }
  closeRoom(session);
  recordLog('FORCE_CLOSE_SESSION', session.session_id, session.title, reason);
  return snapshot(session);
}

/** GET /api/v1/admin/logs?action= — newest first. */
export async function listLogs(action?: AdminLogAction): Promise<AdminLog[]> {
  await delay();
  return MOCK_ADMIN_LOGS.filter((l) => !action || l.action === action).map((l) => ({ ...l }));
}
