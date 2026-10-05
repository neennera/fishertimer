-- =======================================================
-- Study Timer Service Database Schema (Phase 2 - Events & State Machine)
-- Database Engine: PostgreSQL
-- Database: timer_db
-- =======================================================

-- 1. Idempotency store for RabbitMQ events
CREATE TABLE IF NOT EXISTS processed_events (
    event_id VARCHAR(64) PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Timers per room session (OPEN / FINALIZED)
CREATE TABLE IF NOT EXISTS timers (
    timer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL, -- Refers logically to session_db.study_sessions(session_id)
    user_id UUID NOT NULL,    -- Refers logically to account_db.users(user_id)
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'FINALIZED')),
    opened_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finalized_at TIMESTAMPTZ,
    CONSTRAINT uq_timers_session_user UNIQUE (session_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_timers_session_id ON timers(session_id);
CREATE INDEX IF NOT EXISTS idx_timers_user_id ON timers(user_id);
CREATE INDEX IF NOT EXISTS idx_timers_session_status ON timers(session_id, status);

-- 3. Cycles within a timer
CREATE TABLE IF NOT EXISTS cycles (
    cycle_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timer_id UUID NOT NULL REFERENCES timers(timer_id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('WORK', 'REST')),
    status VARCHAR(20) NOT NULL DEFAULT 'RUNNING' CHECK (status IN ('RUNNING', 'PAUSED', 'COMPLETED', 'SKIPPED', 'DISCARDED')),
    duration_sec INT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paused_at TIMESTAMPTZ,
    paused_total_sec INT NOT NULL DEFAULT 0,
    ended_at TIMESTAMPTZ,
    reward_status VARCHAR(10) NOT NULL DEFAULT 'NONE' CHECK (reward_status IN ('NONE', 'PENDING', 'SENT'))
);

CREATE INDEX IF NOT EXISTS idx_cycles_timer_id ON cycles(timer_id);
-- Partial index ensuring only one active cycle per timer (UC-05 E-4)
CREATE UNIQUE INDEX IF NOT EXISTS idx_cycles_single_active ON cycles(timer_id)
    WHERE status IN ('RUNNING', 'PAUSED');
