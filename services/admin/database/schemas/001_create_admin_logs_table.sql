-- =======================================================
-- Admin Service Database Schema (3NF)
-- Database Engine: PostgreSQL
-- Database: admin_db
-- =======================================================

CREATE TABLE IF NOT EXISTS admin_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL, -- Refers logically to account_db.users(user_id)
    action VARCHAR(50) NOT NULL CHECK (action IN ('FORCE_CLOSE_SESSION', 'KICK_USER')),
    target_id UUID NOT NULL, -- ID of target session or user
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_logs_admin_id ON admin_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_logs_action ON admin_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_logs_created_at ON admin_logs(created_at);
