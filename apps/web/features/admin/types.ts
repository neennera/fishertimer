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
  is_banned: boolean;
  joined_at: string;
  left_at: string | null;
}

export interface AdminSessionDetail extends AdminSession {
  participants: AdminParticipant[];
}
