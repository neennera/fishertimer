// Mock GET /api/auth/profile?id=: the full user, as the endpoint sends it.

import type { SessionUser } from '../auth';
import { MOCK_USER_IDS } from './auth.mock';

function user(user_id: string, display_name: string, avatar_url: string, created_at: string): SessionUser {
  return {
    user_id,
    email: `${display_name.toLowerCase().replace(/[^a-z]+/g, '.')}@example.com`,
    display_name,
    avatar_url,
    role: 'CUSTOMER',
    created_at,
    updated_at: created_at,
  };
}

export const MOCK_PROFILES: Record<string, SessionUser> = {
  [MOCK_USER_IDS.mira]: user(MOCK_USER_IDS.mira, 'Mira L.', '/sprites/fish/Pufferfish.png', '2026-09-03T00:00:00Z'),
  // Empty avatar_url: exercises the avatar fallback.
  [MOCK_USER_IDS.tan]: user(MOCK_USER_IDS.tan, 'Tan R.', '', '2026-09-12T00:00:00Z'),
  // Zero stats and no fish: exercises the empty states.
  [MOCK_USER_IDS.newAngler]: user(MOCK_USER_IDS.newAngler, 'New Angler', '', '2026-09-26T00:00:00Z'),
};
