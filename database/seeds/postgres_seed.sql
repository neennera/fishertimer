-- Seed Data: postgres_seed.sql
-- Mock dataset for local development and testing

-- 1. Mock Users & Profiles
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

-- 2. Mock Study Sessions
INSERT INTO session.study_sessions (id, name, creator_id, participant_limit, status)
VALUES
  ('sess_01', 'Software Architecture Exam Cram', 'usr_01', 10, 'ACTIVE'),
  ('sess_02', 'Silent Pomodoro (Solo)', 'usr_02', 1, 'ACTIVE'),
  ('sess_03', 'Algorithms & Data Structures', 'usr_03', 6, 'ENDED')
ON CONFLICT (id) DO NOTHING;

-- 3. Mock Session Participants
INSERT INTO session.session_participants (session_id, user_id)
VALUES
  ('sess_01', 'usr_01'),
  ('sess_01', 'usr_02'),
  ('sess_02', 'usr_02')
ON CONFLICT (session_id, user_id) DO NOTHING;
