-- =======================================================
-- Study Timer Service: Temporary Seed for Phase 2 W1 Testing
-- Database: timer_db
-- =======================================================
-- Description:
-- Seeds an active study room timer with OPEN status and a RUNNING work cycle.
-- This allows testing RabbitMQ event consumer (LeaveSession, EndSession)
-- before JoinSession is implemented in W2.

-- 1. Seed user timer in room
INSERT INTO timers (timer_id, session_id, user_id, status, opened_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'OPEN',
    CURRENT_TIMESTAMP
)
ON CONFLICT (session_id, user_id) DO UPDATE
SET status = 'OPEN', finalized_at = NULL;

-- 2. Seed active WORK cycle currently RUNNING
INSERT INTO cycles (cycle_id, timer_id, type, status, duration_sec, started_at, reward_status)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'WORK',
    'RUNNING',
    1500,
    CURRENT_TIMESTAMP,
    'NONE'
)
ON CONFLICT (cycle_id) DO UPDATE
SET status = 'RUNNING', ended_at = NULL;
