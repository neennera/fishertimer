-- Schema: 002_create_study_sessions_table.sql
-- Service Owner: Study Session Service
-- Description: Stores active and historical study room lifecycles and participant rosters.

CREATE SCHEMA IF NOT EXISTS session;

CREATE TABLE IF NOT EXISTS session.study_sessions (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    creator_id VARCHAR(64) NOT NULL,
    participant_limit INT NOT NULL CHECK (participant_limit >= 1 AND participant_limit <= 50),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'ENDED')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS session.session_participants (
    session_id VARCHAR(64) NOT NULL REFERENCES session.study_sessions(id) ON DELETE CASCADE,
    user_id VARCHAR(64) NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMPTZ NULL,
    PRIMARY KEY (session_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_status ON session.study_sessions(status);
CREATE INDEX IF NOT EXISTS idx_participants_user ON session.session_participants(user_id);
