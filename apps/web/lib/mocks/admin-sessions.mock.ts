// Mock admin data (session_db shape). Timestamps are relative to runtime.

import type { AdminLog, AdminParticipant, AdminSessionDetail } from '../../features/admin/types';

const MIN = 60_000;
const ago = (minutes: number) => new Date(Date.now() - minutes * MIN).toISOString();

function participant(
  user_id: string,
  display_name: string,
  joinedMinAgo: number,
  opts: { host?: boolean; leftMinAgo?: number } = {},
): AdminParticipant {
  return {
    user_id,
    display_name,
    is_host: opts.host ?? false,
    joined_at: ago(joinedMinAgo),
    left_at: opts.leftMinAgo === undefined ? null : ago(opts.leftMinAgo),
  };
}

function session(
  session_id: string,
  title: string,
  max_participants: number,
  createdMinAgo: number,
  endedMinAgo: number | null,
  participants: AdminParticipant[],
): AdminSessionDetail {
  const host = participants.find((p) => p.is_host)!;
  return {
    session_id,
    title,
    host_id: host.user_id,
    host_name: host.display_name,
    is_active: endedMinAgo === null,
    max_participants,
    participant_count: participants.filter((p) => p.left_at === null).length,
    created_at: ago(createdMinAgo),
    ended_at: endedMinAgo === null ? null : ago(endedMinAgo),
    participants,
  };
}

export const MOCK_ADMIN_SESSIONS: AdminSessionDetail[] = [
  session('a1f0c6e2-0001-4a10-9c11-000000000001', 'Midterm Cram — Software Architecture', 4, 50, null, [
    participant('user1', 'BobberBoss', 50, { host: true }),
    participant('user2', 'ReelDeal', 44),
    participant('user3', 'LureQueen', 31),
  ]),
  session('a1f0c6e2-0002-4a10-9c11-000000000002', 'Quiet Pond — Deep Focus', 6, 22, null, [
    participant('user4', 'NetNinja', 22, { host: true }),
    participant('user5', 'TackleTom', 20),
    participant('user6', 'SpoolSam', 15),
    participant('user7', 'HookedHana', 9),
    participant('user8', 'TroubleTrout', 6),
  ]),
  session('a1f0c6e2-0003-4a10-9c11-000000000003', 'Calculus Study Group', 4, 12, null, [
    participant('user9', 'CastAway', 12, { host: true }),
  ]),
  session('a1f0c6e2-0004-4a10-9c11-000000000004', 'Late Night Coding', 4, 300, 95, [
    participant('user2', 'ReelDeal', 300, { host: true, leftMinAgo: 95 }),
    participant('user4', 'NetNinja', 290, { leftMinAgo: 140 }),
    participant('user6', 'SpoolSam', 250, { leftMinAgo: 95 }),
  ]),
  session('a1f0c6e2-0005-4a10-9c11-000000000005', 'Thesis Writing Sprint', 5, 1500, 1380, [
    participant('user3', 'LureQueen', 1500, { host: true, leftMinAgo: 1380 }),
    participant('user1', 'BobberBoss', 1490, { leftMinAgo: 1400 }),
  ]),
];

/** The signed-in admin until the account service exposes roles. */
export const MOCK_ADMIN = { id: 'admin1', name: 'ModeratorMo' };

// Newest first, like the real query (ORDER BY created_at DESC).
export const MOCK_ADMIN_LOGS: AdminLog[] = [
  {
    log_id: 'log-3',
    admin_id: MOCK_ADMIN.id,
    admin_name: MOCK_ADMIN.name,
    action: 'FORCE_CLOSE_SESSION',
    target_id: 'a1f0c6e2-0004-4a10-9c11-000000000004',
    target_label: 'Late Night Coding',
    reason: 'Session ran past quiet hours',
    created_at: ago(95),
  },
  {
    log_id: 'log-2',
    admin_id: MOCK_ADMIN.id,
    admin_name: MOCK_ADMIN.name,
    action: 'KICK_USER',
    target_id: 'user4',
    target_label: 'NetNinja',
    reason: 'Spamming the room',
    created_at: ago(140),
  },
  {
    log_id: 'log-1',
    admin_id: MOCK_ADMIN.id,
    admin_name: MOCK_ADMIN.name,
    action: 'FORCE_CLOSE_SESSION',
    target_id: 'a1f0c6e2-0005-4a10-9c11-000000000005',
    target_label: 'Thesis Writing Sprint',
    reason: null,
    created_at: ago(1380),
  },
];
