-- =======================================================
-- Study Session Service: Phase 2 (room lifecycle + event outbox)
-- Database Engine: PostgreSQL
-- Database: session_db
-- =======================================================

-- 1. Rooms: explicit lifecycle and a participant counter -----------------
-- status replaces is_active so a room's end can carry a reason, and
-- participant_count lets JoinSession check capacity with a single
-- conditional UPDATE (UC-02 E-2: two users racing for the last slot).
ALTER TABLE study_sessions
    ADD COLUMN IF NOT EXISTS status VARCHAR(10) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'ENDED')),
    ADD COLUMN IF NOT EXISTS participant_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS end_reason VARCHAR(20)
        CHECK (end_reason IN ('EMPTY', 'ADMIN_CLOSED', 'TIMEOUT_24H'));

UPDATE study_sessions SET status = 'ENDED' WHERE is_active = FALSE;
ALTER TABLE study_sessions DROP COLUMN IF EXISTS is_active;
DROP INDEX IF EXISTS idx_study_sessions_is_active;

ALTER TABLE study_sessions ALTER COLUMN max_participants SET DEFAULT 5;
ALTER TABLE study_sessions
    ADD CONSTRAINT chk_study_sessions_limit CHECK (max_participants BETWEEN 1 AND 5),
    ADD CONSTRAINT chk_study_sessions_count CHECK (participant_count BETWEEN 0 AND max_participants);

CREATE INDEX IF NOT EXISTS idx_study_sessions_status_created ON study_sessions(status, created_at);

-- 2. Participants: one row per stay -------------------------------------
-- The old (session_id, user_id) primary key made it impossible to rejoin a
-- room after leaving it. Each stay now gets its own row, and the partial
-- unique index below enforces "one active room per user" (UC-01 E-2,
-- UC-02 E-3) in the database itself.
ALTER TABLE session_participants DROP CONSTRAINT IF EXISTS session_participants_pkey;
ALTER TABLE session_participants
    ADD COLUMN IF NOT EXISTS participant_id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN IF NOT EXISTS display_name VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS leave_reason VARCHAR(20)
        CHECK (leave_reason IN ('LEFT', 'KICKED', 'DISCONNECT_TIMEOUT', 'IDLE_TIMEOUT', 'SESSION_ENDED'));
ALTER TABLE session_participants ADD PRIMARY KEY (participant_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_session_participants_one_active_room
    ON session_participants(user_id) WHERE left_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_session_participants_session_active
    ON session_participants(session_id) WHERE left_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_session_participants_last_seen
    ON session_participants(last_seen_at) WHERE left_at IS NULL;

-- 3. Transactional outbox for RabbitMQ events ---------------------------
-- Written in the same transaction as the room change; a relay goroutine
-- publishes rows with publisher confirms and stamps published_at.
CREATE TABLE IF NOT EXISTS event_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    routing_key VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMPTZ,
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT
);

CREATE INDEX IF NOT EXISTS idx_event_outbox_pending ON event_outbox(id) WHERE published_at IS NULL;
