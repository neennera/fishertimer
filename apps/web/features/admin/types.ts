/** Admin feature types — mirror session_db (study_sessions, session_participants). */

export interface AdminSession {
  session_id: string;
  title: string;
  host_id: string;
  host_name: string;
  is_active: boolean;
  max_participants: number;
  /** Participants currently in the room (left_at is null). */
  participant_count: number;
  created_at: string;
  ended_at: string | null;
}

export interface AdminParticipant {
  user_id: string;
  display_name: string;
  is_host: boolean;
  joined_at: string;
  left_at: string | null;
}

export interface AdminSessionDetail extends AdminSession {
  participants: AdminParticipant[];
}

/** admin_db.admin_logs — with the names the UI shows beside the raw IDs. */
export type AdminLogAction = 'KICK_USER' | 'FORCE_CLOSE_SESSION';

export interface AdminLog {
  log_id: string;
  admin_id: string;
  admin_name: string;
  action: AdminLogAction;
  /** Session ID for FORCE_CLOSE_SESSION, user ID for KICK_USER. */
  target_id: string;
  /** Session title or member name, resolved for display. */
  target_label: string;
  reason: string | null;
  created_at: string;
}
