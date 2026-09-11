-- 01-init-postgres.sql
-- Combined schema & seed execution for Docker container startup

\connect fishertimer;

-- 1. Create schemas & users table
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS account;
CREATE SCHEMA IF NOT EXISTS session;
CREATE SCHEMA IF NOT EXISTS timer;
CREATE SCHEMA IF NOT EXISTS admin;

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

CREATE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
CREATE INDEX IF NOT EXISTS idx_profiles_banned ON account.profiles(is_banned);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON session.study_sessions(status);

-- 2. Insert Seeds
INSERT INTO auth.users (id, email, display_name, avatar_url)
VALUES 
  ('usr_01', 'student1@chula.ac.th', 'FisherCaptain', 'https://api.dicebear.com/7.x/bottts/svg?seed=usr01'),
  ('usr_02', 'student2@chula.ac.th', 'QuietAngler', 'https://api.dicebear.com/7.x/bottts/svg?seed=usr02'),
  ('usr_03', 'student3@chula.ac.th', 'DeepDiver', 'https://api.dicebear.com/7.x/bottts/svg?seed=usr03'),
  ('usr_banned', 'spammer@chula.ac.th', 'BannedUser', 'https://api.dicebear.com/7.x/bottts/svg?seed=banned')
ON CONFLICT (id) DO NOTHING;

INSERT INTO account.profiles (user_id, display_name, is_banned, ban_reason, total_focus_minutes)
VALUES
  ('usr_01', 'FisherCaptain', FALSE, '', 320),
  ('usr_02', 'QuietAngler', FALSE, '', 180),
  ('usr_03', 'DeepDiver', FALSE, '', 540),
  ('usr_banned', 'BannedUser', TRUE, 'Inappropriate behavior in study room', 15)
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO session.study_sessions (id, name, creator_id, participant_limit, status)
VALUES
  ('sess_01', 'Software Architecture Exam Cram', 'usr_01', 10, 'ACTIVE'),
  ('sess_02', 'Silent Pomodoro (Solo)', 'usr_02', 1, 'ACTIVE'),
  ('sess_03', 'Algorithms & Data Structures', 'usr_03', 6, 'ENDED')
ON CONFLICT (id) DO NOTHING;

INSERT INTO session.session_participants (session_id, user_id)
VALUES
  ('sess_01', 'usr_01'),
  ('sess_01', 'usr_02'),
  ('sess_02', 'usr_02')
ON CONFLICT (session_id, user_id) DO NOTHING;
