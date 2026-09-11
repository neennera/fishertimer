-- Schema: 001_create_users_table.sql
-- Service Owner: Auth & Account Services
-- Description: Stores user accounts, Google OAuth identities, and profile moderation states.

CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS account;

CREATE TABLE IF NOT EXISTS auth.users (
    id VARCHAR(64) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    avatar_url TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS account.profiles (
    user_id VARCHAR(64) PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    display_name VARCHAR(100) NOT NULL,
    is_banned BOOLEAN DEFAULT FALSE,
    ban_reason TEXT DEFAULT '',
    banned_until TIMESTAMPTZ NULL,
    total_focus_minutes INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_profiles_banned ON account.profiles(is_banned);
