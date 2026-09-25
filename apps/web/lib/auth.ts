// Auth data layer (UC-06 Sign In / Sign Up). The only file, besides its mock
// counterpart, that knows whether auth data is real or mocked. Components
// call these functions and never touch fetch or lib/mocks/auth.mock.ts
// directly. Toggled by NEXT_PUBLIC_USE_MOCKS (see .env.example).
//
// Real contract: the account service (services/account/internal/adapter/
// handler/http_handler.go), reached through the gateway's /api/auth/* alias,
// which maps onto the service's /api/v1/account/*:
//   GET  /api/auth/google/login?next=&signup=  browser navigation to Google
//   GET  /api/auth/google/callback             Google -> backend -> redirect
//                                              to next/signup (+?auth_error=)
//   GET  /api/auth/me                          {status, user?|email?}, always 200
//   POST /api/auth/signup {display_name}       201 + user, sets ft_session
//   POST /api/auth/signout                     clears ft_session / ft_signup
// The session itself lives in HttpOnly cookies, so this file never sees a token.

import type { UserRole } from '@fishertimer/shared-types';
import { AUTH_ERROR_MESSAGES, type AuthErrorCode } from './auth-error-messages';
import { ClientApiError, clientApiFetch } from './client-api';
import {
  MOCK_SCENARIO_PARAM,
  MOCK_SIGN_IN_OUTCOMES,
  MOCK_SIGNUP_FAILS,
  MOCK_USER,
  isMockAuthScenario,
  mockDelay,
  readMockAuthState,
  writeMockAuthState,
  type MockAuthScenario,
} from './mocks/auth.mock';
import { validateDisplayName } from './validate-display-name';

export type { AuthErrorCode, UserRole };
export { AUTH_ERROR_MESSAGES };

const USE_MOCKS = process.env.NEXT_PUBLIC_USE_MOCKS !== 'false';

const GOOGLE_LOGIN_PATH = '/api/auth/google/login';

// Query param the backend tags its redirect with on a failed sign-in.
export const AUTH_ERROR_PARAM = 'auth_error';

// Where signInWithGoogle() asks the backend to send the browser back to.
// `next` also receives every ?auth_error= redirect, which is why it defaults
// to /signin (which shows the error, or forwards a signed-in user to /)
// rather than /.
export const DEFAULT_SIGN_IN_NEXT = '/signin';
export const DEFAULT_SIGN_UP_PATH = '/welcome';

// domain.UserAccount, field for field. `role` is passed through as sent:
// Google sign-in only creates CUSTOMERs, but ADMIN rows are pre-provisioned.
export interface SessionUser {
  user_id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

// GET /me's response, exactly.
export type Session =
  | { status: 'signed_in'; user: SessionUser }
  | { status: 'needs_signup'; email: string }
  | { status: 'signed_out' };

export type SessionStatus = Session['status'];

export interface AuthError {
  // The raw ?auth_error= value — may be a Google code outside AuthErrorCode.
  code: string;
  message: string;
}

export type CompleteFirstTimeSetupResult =
  | { ok: true; user: SessionUser }
  | {
      ok: false;
      // invalid: name rejected (400) · expired: ft_signup ticket gone (401)
      // account_creation_failed: anything else from the server (01d)
      code: 'invalid' | 'expired' | 'account_creation_failed';
      error: string;
    };

const SIGN_UP_EXPIRED_MESSAGE = 'Your sign-up has expired. Please sign in with Google again.';

// Real mode only: GET /me's last answer, updated by completeFirstTimeSetup()
// and signOut(). Signing in is a full-page navigation, so it never goes stale
// across that.
let cachedSession: Session | null = null;

function currentMockScenario(): MockAuthScenario | null {
  const value = new URLSearchParams(window.location.search).get(MOCK_SCENARIO_PARAM);
  return isMockAuthScenario(value) ? value : null;
}

// Sends the browser to Google via the backend. Never resolves into anything
// useful — the page is navigated away; the backend redirects back to `next`
// (signed in, or ?auth_error=) or `signup` (new e-mail) on its own.
export async function signInWithGoogle({
  next = DEFAULT_SIGN_IN_NEXT,
  signup = DEFAULT_SIGN_UP_PATH,
}: { next?: string; signup?: string } = {}): Promise<void> {
  if (!USE_MOCKS) {
    const query = new URLSearchParams({ next, signup });
    window.location.assign(`${GOOGLE_LOGIN_PATH}?${query.toString()}`);
    return;
  }

  // Mock: play out the backend's redirect for ?mockScenario= (default 01a).
  const scenario = currentMockScenario() ?? 'signin-default';
  const outcome = MOCK_SIGN_IN_OUTCOMES[scenario];
  await mockDelay(undefined);

  writeMockAuthState(
    outcome.session.status === 'signed_out' ? null : { scenario, session: outcome.session }
  );
  const path = outcome.redirectTo === 'next' ? next : signup;
  window.location.assign(
    outcome.authError ? `${path}?${AUTH_ERROR_PARAM}=${outcome.authError}` : path
  );
}

// Reads the ?auth_error= the backend's callback redirect left on the URL.
// Unknown codes (other Google ?error= values) get the login_failed copy.
export function readAuthError(params: URLSearchParams): AuthError | null {
  const code = params.get(AUTH_ERROR_PARAM);
  if (!code) {
    return null;
  }
  const message =
    code in AUTH_ERROR_MESSAGES
      ? AUTH_ERROR_MESSAGES[code as AuthErrorCode]
      : AUTH_ERROR_MESSAGES.login_failed;
  return { code, message };
}

export async function getSession(): Promise<Session> {
  if (USE_MOCKS) {
    const state = readMockAuthState();
    if (state) {
      return mockDelay(state.session);
    }
    // No mock "cookie" yet: /welcome?mockScenario=setup-* is still directly
    // reachable, as if the user had just come back from Google.
    const scenario = currentMockScenario();
    const session: Session =
      scenario && MOCK_SIGN_IN_OUTCOMES[scenario].redirectTo === 'signup'
        ? MOCK_SIGN_IN_OUTCOMES[scenario].session
        : { status: 'signed_out' };
    return mockDelay(session);
  }

  if (cachedSession) {
    return cachedSession;
  }

  try {
    cachedSession = await clientApiFetch<Session>('auth', 'me');
    return cachedSession;
  } catch {
    // /me always answers 200, so this is the gateway being unreachable.
    return { status: 'signed_out' };
  }
}

// Creates the account for the pending Google identity (needs_signup) and signs
// the user in. Network failures throw; server answers come back as results.
export async function completeFirstTimeSetup(
  displayName: string
): Promise<CompleteFirstTimeSetupResult> {
  const invalid = validateDisplayName(displayName);
  if (invalid) {
    return { ok: false, code: 'invalid', error: invalid };
  }

  const trimmed = displayName.trim();

  if (!USE_MOCKS) {
    try {
      const user = await clientApiFetch<SessionUser>('auth', 'signup', {
        method: 'POST',
        body: JSON.stringify({ display_name: trimmed }),
      });
      cachedSession = { status: 'signed_in', user };
      return { ok: true, user };
    } catch (err) {
      if (!(err instanceof ClientApiError)) {
        throw err;
      }
      if (err.status === 400) {
        return { ok: false, code: 'invalid', error: err.message };
      }
      if (err.status === 401) {
        return { ok: false, code: 'expired', error: SIGN_UP_EXPIRED_MESSAGE };
      }
      return {
        ok: false,
        code: 'account_creation_failed',
        error: AUTH_ERROR_MESSAGES.account_creation_failed,
      };
    }
  }

  const state = readMockAuthState();
  const scenario = state?.scenario ?? currentMockScenario();
  const pending = state?.session ?? (await getSession());

  if (pending.status !== 'needs_signup' || !scenario) {
    return mockDelay<CompleteFirstTimeSetupResult>({ ok: false, code: 'expired', error: SIGN_UP_EXPIRED_MESSAGE });
  }

  if (MOCK_SIGNUP_FAILS.has(scenario)) {
    return mockDelay<CompleteFirstTimeSetupResult>({
      ok: false,
      code: 'account_creation_failed',
      error: AUTH_ERROR_MESSAGES.account_creation_failed,
    });
  }

  const user: SessionUser = {
    ...MOCK_USER,
    user_id: 'mock-user-new',
    email: pending.email,
    display_name: trimmed,
  };
  writeMockAuthState({ scenario: 'signin-default', session: { status: 'signed_in', user } });
  return mockDelay<CompleteFirstTimeSetupResult>({ ok: true, user });
}

export async function signOut(): Promise<void> {
  if (USE_MOCKS) {
    writeMockAuthState(null);
    await mockDelay(undefined);
    return;
  }

  await clientApiFetch<{ signed_out: boolean }>('auth', 'signout', { method: 'POST' });
  cachedSession = { status: 'signed_out' };
}
