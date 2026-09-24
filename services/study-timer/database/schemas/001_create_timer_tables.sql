-- =======================================================
-- Study Timer Service Database Schema (3NF)
-- Database Engine: PostgreSQL
-- Database: timer_db
-- =======================================================

CREATE TABLE IF NOT EXISTS timer_settings (
    user_id UUID PRIMARY KEY, -- Refers logically to account_db.users(user_id)
    focus_duration INT NOT NULL DEFAULT 1500, -- Seconds (default 25 mins)
    short_break_duration INT NOT NULL DEFAULT 300, -- Seconds (default 5 mins)
    long_break_duration INT NOT NULL DEFAULT 900, -- Seconds (default 15 mins)
    cycles_before_long_break INT NOT NULL DEFAULT 4,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS timer_sessions (
    timer_session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, -- Refers logically to account_db.users(user_id)
    study_session_id UUID, -- Refers logically to session_db.study_sessions(session_id). Nullable for solo study
    status VARCHAR(20) NOT NULL CHECK (status IN ('FOCUS', 'SHORT_BREAK', 'LONG_BREAK', 'PAUSED', 'COMPLETED', 'STOPPED')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS timer_cycles (
    cycle_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timer_session_id UUID NOT NULL REFERENCES timer_sessions(timer_session_id) ON DELETE CASCADE,
    cycle_number INT NOT NULL,
    phase_type VARCHAR(20) NOT NULL CHECK (phase_type IN ('FOCUS', 'SHORT_BREAK', 'LONG_BREAK')),
    duration INT NOT NULL, -- Recorded elapsed time in seconds
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    ended_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_timer_sessions_user_id ON timer_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_timer_sessions_study_session_id ON timer_sessions(study_session_id);
CREATE INDEX IF NOT EXISTS idx_timer_cycles_timer_session_id ON timer_cycles(timer_session_id);
