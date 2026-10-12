-- =======================================================
-- Study Timer: one timer_sessions row per participant per room
-- Database: timer_db
-- =======================================================
-- Until now the repository used the room's session id as
-- timer_session_id (the primary key), so a room could hold only one row:
-- the second participant to press Start overwrote the first participant's
-- timer, and completed cycles were counted per room instead of per person.
--
-- The room id now lives in study_session_id (the column was already meant
-- for it), timer_session_id is a generated id again, and a unique index
-- keeps exactly one live row per (room, participant).

-- Rows written before this migration stored the room id in the primary key.
UPDATE timer_sessions
SET study_session_id = timer_session_id
WHERE study_session_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_timer_sessions_room_user
    ON timer_sessions(study_session_id, user_id);
