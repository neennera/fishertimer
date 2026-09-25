-- =======================================================
-- Study Session Service Database Schema (3NF)
-- Database Engine: PostgreSQL
-- Database: session_db
-- =======================================================

CREATE TABLE IF NOT EXISTS study_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(150) NOT NULL,
    host_id UUID NOT NULL, -- Refers logically to account_db.users(user_id)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    max_participants INT NOT NULL DEFAULT 4,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS session_participants (
    session_id UUID NOT NULL REFERENCES study_sessions(session_id) ON DELETE CASCADE,
    user_id UUID NOT NULL, -- Refers logically to account_db.users(user_id)
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMPTZ,
    PRIMARY KEY (session_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_study_sessions_host_id ON study_sessions(host_id);
CREATE INDEX IF NOT EXISTS idx_study_sessions_is_active ON study_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_session_participants_user_id ON session_participants(user_id);
