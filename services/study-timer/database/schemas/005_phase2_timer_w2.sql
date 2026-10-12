-- =======================================================
-- Study Timer: Phase 2 W2 (state machine on timers / cycles)
-- Database: timer_db
-- =======================================================
-- The service now keeps every timer in `timers` and every work / rest
-- period in `cycles` (003). These indexes serve the background sweeper:
--   * running / paused cycles, to complete the ones whose time is up and
--     discard the ones paused longer than MAX_PAUSE_MINUTES (UC-05 E-3);
--   * completed work cycles whose reward Reward has not confirmed yet,
--     retried until it does (UC-05 E-7).

CREATE INDEX IF NOT EXISTS idx_cycles_active_started
    ON cycles(status, started_at) WHERE status IN ('RUNNING', 'PAUSED');

CREATE INDEX IF NOT EXISTS idx_cycles_reward_pending
    ON cycles(ended_at) WHERE reward_status = 'PENDING';

-- Lookups of the work done in the current stay (cycles since opened_at).
CREATE INDEX IF NOT EXISTS idx_cycles_timer_started ON cycles(timer_id, started_at);

COMMENT ON TABLE timer_sessions IS 'Deprecated (Phase 1). Replaced by timers + cycles; no longer written.';
COMMENT ON TABLE timer_cycles IS 'Deprecated (Phase 1). Replaced by cycles; no longer written.';
