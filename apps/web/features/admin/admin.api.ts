// Admin data access — every call hits services/admin (see its README).

import { AdminApiError, adminApiFetch } from '../../lib/admin-api';
import type {
  AdminLog,
  AdminLogAction,
  AdminParticipant,
  AdminSession,
  AdminSessionDetail,
  ModerationResult,
} from './types';

const json = (body: unknown): RequestInit => ({
  method: 'POST',
  body: JSON.stringify(body),
});

/** GET /api/v1/admin/sessions — live rooms from Study Session, newest first. */
export function listSessions(): Promise<AdminSession[]> {
  return adminApiFetch<AdminSession[]>('sessions');
}

/** GET /api/v1/admin/sessions/:id/participants — members of a live room. */
export function listParticipants(sessionId: string): Promise<AdminParticipant[]> {
  return adminApiFetch<AdminParticipant[]>(
    `sessions/${encodeURIComponent(sessionId)}/participants`,
  );
}

/**
 * GET /api/v1/admin/sessions/:id — a live room plus its members (host first).
 * Null when the room is not live (ended, or never existed).
 */
export async function getSession(id: string): Promise<AdminSessionDetail | null> {
  try {
    return await adminApiFetch<AdminSessionDetail>(`sessions/${encodeURIComponent(id)}`);
  } catch (e) {
    if (e instanceof AdminApiError && e.status === 404) return null;
    throw e;
  }
}

/**
 * POST /api/v1/admin/sessions/:id/participants/:userId/kick — Study Session
 * runs LeaveSession() for the member; removing the last member ends the room.
 * Logged as KICK_USER. Throws AdminApiError 409 when the member has already
 * left or the room has ended.
 */
export function kickParticipant(
  sessionId: string,
  userId: string,
  reason?: string,
): Promise<ModerationResult> {
  return adminApiFetch<ModerationResult>(
    `sessions/${encodeURIComponent(sessionId)}/participants/${encodeURIComponent(userId)}/kick`,
    json({ reason: reason?.trim() || undefined }),
  );
}

/**
 * POST /api/v1/admin/sessions/:id/end — Study Session runs EndSession(),
 * disconnecting everyone still in the room. Logged as FORCE_CLOSE_SESSION.
 * Throws AdminApiError 409 when the room has already ended.
 */
export function endSession(sessionId: string, reason?: string): Promise<ModerationResult> {
  return adminApiFetch<ModerationResult>(
    `sessions/${encodeURIComponent(sessionId)}/end`,
    json({ reason: reason?.trim() || undefined }),
  );
}

/** GET /api/v1/admin/logs?action= — newest first (latest 100). */
export function listLogs(action?: AdminLogAction): Promise<AdminLog[]> {
  return adminApiFetch<AdminLog[]>(action ? `logs?action=${action}` : 'logs');
}
