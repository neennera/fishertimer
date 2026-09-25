// Auth data layer (UC-06 Sign In / Sign Up). The only file, besides its mock
// counterpart, that knows whether auth data is real or mocked. Components
// call these functions and never touch fetch or lib/mocks/auth.mock.ts
// directly. Toggled by NEXT_PUBLIC_USE_MOCKS (see .env.example).

import type { User } from '@fishertimer/shared-types';
import { AUTH_CALLBACK_ERROR_MESSAGES, type AuthCallbackErrorCode } from './auth-error-messages';
import { clientApiFetch } from './client-api';
import {
  MOCK_AUTH_CALLBACK_ERRORS,
  MOCK_FIRST_TIME_SESSION,
  MOCK_SCENARIO_PARAM,
  MOCK_SESSION,
  mockDelay,
  type MockAuthScenario,
} from './mocks/auth.mock';
import { validateDisplayName } from './validate-display-name';

export type { AuthCallbackErrorCode };
export { AUTH_CALLBACK_ERROR_MESSAGES };

const USE_MOCKS = process.env.NEXT_PUBLIC_USE_MOCKS !== 'false';

const API_GATEWAY_URL =
  process.env.NEXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:8080';

export type UserRole = 'CUSTOMER' | 'ADMIN';

export interface SessionUser {
  id: string;
  email: string;
  displayName: string;
  role: UserRole;
}

export interface Session {
  user: SessionUser;
  isFirstLogin: boolean;
}

export type AuthCallbackResult =
  | { ok: true; session: Session }
  | { ok: false; error: AuthCallbackErrorCode; message: string };

export type CompleteFirstTimeSetupResult =
  | { ok: true; session: Session }
  | { ok: false; error: string };

// Raw JSON shape the account service actually sends (snake_case). Covers the
// fields common to both UserAccount and Profile in
// services/account/internal/domain/entity.go — the two response shapes the
// 'auth' gateway alias can proxy to for callback/setup/session.
interface AccountUserWire {
  user_id: string;
  email: string;
  display_name: string;
}

// Maps the account service's real response onto our existing Session shape.
function mapAccountUserToSession(wire: AccountUserWire): Session {
  // `role` lines up with shared-types User['role'], but entity.go's
  // UserAccount/Profile structs don't have a role field yet — unresolved,
  // defaulting to CUSTOMER rather than inventing a real value.
  const role: User['role'] = 'CUSTOMER';
  return {
    user: {
      id: wire.user_id,
      email: wire.email,
      displayName: wire.display_name,
      role,
    },
    // `isFirstLogin` doesn't exist anywhere in the real schema either —
    // unresolved, same as `role` above.
    isFirstLogin: false,
  };
}

// In-memory cache of the current session, populated by handleAuthCallback(),
// completeFirstTimeSetup() and getSession() itself.
let currentSession: Session | null = null;

export function signInWithGoogle(): Promise<Session | void> {
  if (!USE_MOCKS) {
    window.location.href = `${API_GATEWAY_URL}/api/auth/google`;
    return Promise.resolve();
  }

  return mockDelay(MOCK_SESSION).then((session) => {
    currentSession = session;
    return session;
  });
}

export async function handleAuthCallback(
  params: URLSearchParams
): Promise<AuthCallbackResult> {
  if (!USE_MOCKS) {
    try {
      const query = params.toString();
      const wire = await clientApiFetch<AccountUserWire>(
        'auth',
        `callback${query ? `?${query}` : ''}`
      );
      const session = mapAccountUserToSession(wire);
      currentSession = session;
      return { ok: true, session };
    } catch (err) {
      return {
        ok: false,
        error: 'code_exchange_failed',
        message:
          err instanceof Error
            ? err.message
            : AUTH_CALLBACK_ERROR_MESSAGES.code_exchange_failed,
      };
    }
  }

  const scenario = params.get(MOCK_SCENARIO_PARAM) as MockAuthScenario | null;

  if (scenario && scenario in MOCK_AUTH_CALLBACK_ERRORS) {
    const { code, message } =
      MOCK_AUTH_CALLBACK_ERRORS[scenario as keyof typeof MOCK_AUTH_CALLBACK_ERRORS];
    return mockDelay({ ok: false, error: code, message });
  }

  const session = scenario?.startsWith('setup-') ? MOCK_FIRST_TIME_SESSION : MOCK_SESSION;
  return mockDelay({ ok: true as const, session }).then((result) => {
    currentSession = result.session;
    return result;
  });
}

export async function completeFirstTimeSetup(
  displayName: string
): Promise<CompleteFirstTimeSetupResult> {
  const error = validateDisplayName(displayName);
  if (error) {
    return { ok: false, error };
  }

  const trimmed = displayName.trim();

  if (!USE_MOCKS) {
    const wire = await clientApiFetch<AccountUserWire>('auth', 'setup', {
      method: 'POST',
      body: JSON.stringify({ displayName: trimmed }),
    });
    const session = mapAccountUserToSession(wire);
    currentSession = session;
    return { ok: true, session };
  }

  const base = currentSession ?? MOCK_FIRST_TIME_SESSION;
  const session: Session = {
    user: { ...base.user, displayName: trimmed },
    isFirstLogin: false,
  };
  return mockDelay({ ok: true as const, session }).then((result) => {
    currentSession = result.session;
    return result;
  });
}

export async function getSession(): Promise<Session | null> {
  if (currentSession) {
    return currentSession;
  }

  if (USE_MOCKS) {
    return null;
  }

  try {
    const wire = await clientApiFetch<AccountUserWire>('auth', 'session');
    const session = mapAccountUserToSession(wire);
    currentSession = session;
    return session;
  } catch {
    return null;
  }
}
