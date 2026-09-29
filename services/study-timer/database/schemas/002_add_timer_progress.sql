-- =======================================================
-- Study Timer: persist the live progress of a timer
-- Database: timer_db
-- =======================================================
-- Remaining time is derived from server timestamps (see domain.TimerState),
-- so the row must keep when the timer last started or resumed, how much of
-- the current phase had already elapsed before that, and which phase it is
-- in. Without these a reload could not tell how much time is left.

ALTER TABLE timer_sessions
    ADD COLUMN IF NOT EXISTS phase VARCHAR(4) NOT NULL DEFAULT 'WORK' CHECK (phase IN ('WORK', 'REST')),
    ADD COLUMN IF NOT EXISTS running_since TIMESTAMPTZ, -- NULL unless the timer is running
    ADD COLUMN IF NOT EXISTS elapsed_ms BIGINT NOT NULL DEFAULT 0; -- Milliseconds spent in the phase before running_since
