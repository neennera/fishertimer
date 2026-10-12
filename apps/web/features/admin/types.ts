/** Admin feature types — mirror session_db (study_sessions, session_participants). */

export interface AdminSession {
  session_id: string;
  title: string;
  host_id: string;
  /** Null until Account exposes a user lookup; fall back to host_id. */
  host_name: string | null;
  is_active: boolean;
  max_participants: number;
  /** Participants currently in the room (left_at is null). */
  participant_count: number;
  created_at: string;
  ended_at: string | null;
}

export interface AdminParticipant {
  user_id: string;
  /** Null until Account exposes a user lookup; fall back to user_id. */
  display_name: string | null;
  is_host: boolean;
  joined_at: string;
  left_at: string | null;
}

export interface AdminSessionDetail extends AdminSession {
  participants: AdminParticipant[];
}

/** admin_db.admin_logs — raw IDs; admin_db stores no names. */
export type AdminLogAction = 'KICK_USER' | 'FORCE_CLOSE_SESSION';

export interface AdminLog {
  log_id: string;
  admin_id: string;
  action: AdminLogAction;
  /** Session ID for FORCE_CLOSE_SESSION, user ID for KICK_USER. */
  target_id: string;
  reason: string | null;
  created_at: string;
}

/** Response of a kick or close. */
export interface ModerationResult {
  status: 'kicked' | 'ended';
  /** True when the room is now ENDED (always for a close). */
  session_ended: boolean;
  log: AdminLog;
}
