// Canned auth responses used while NEXT_PUBLIC_USE_MOCKS is on, shaped exactly
// like the account service's real responses (snake_case users, GET /me's
// status union). Covers the 7 states shown by the sign-in and first-time
// setup wireframes:
//   01a Sign In (Default)
//   01b Sign In (Error: Consent Denied)
//   01c Sign In (Error: Code Exchange Failed)
//   01d Sign In (Error: Account Creation Failed)
//   02a First-Time Setup (Default)
//   02b First-Time Setup (Error: Name Empty)
//   02c First-Time Setup (Error: Name Too Long)

import type { AuthErrorCode } from '../auth-error-messages';
import type { Session, SessionUser } from '../auth';
import { validateDisplayName } from '../validate-display-name';

// Query param read by the mock signInWithGoogle() (on /signin) and the mock
// getSession() (on /welcome) to force a given outcome, e.g.
// /signin?mockScenario=signin-consent-denied, then click "Sign in with Google".
export const MOCK_SCENARIO_PARAM = 'mockScenario';

export type MockAuthScenario =
  | 'signin-default'
  | 'signin-consent-denied'
  | 'signin-code-exchange-failed'
  | 'signin-account-creation-failed'
  | 'setup-default'
  | 'setup-name-empty'
  | 'setup-name-too-long';

export const MOCK_SIGN_IN_DELAY_MS = 400;

export const MOCK_USER: SessionUser = {
  user_id: 'mock-user-1',
  email: 'angler@example.com',
  display_name: 'Chayut A.',
  avatar_url: '',
  role: 'CUSTOMER',
  created_at: '2026-09-01T00:00:00Z',
  updated_at: '2026-09-01T00:00:00Z',
};

// The e-mail GET /me reports for a verified Google identity with no account
// yet. The real needs_signup response carries only the e-mail — no Google
// display name — so 02a/02b/02c all start from an empty name field.
export const MOCK_SIGNUP_EMAIL = 'new-angler@example.com';

// What the mock "Google round trip" does for each scenario: which of the two
// paths passed to signInWithGoogle() the backend would redirect to, with
// which ?auth_error=, leaving which GET /me state behind.
export interface MockSignInOutcome {
  redirectTo: 'next' | 'signup';
  authError?: AuthErrorCode;
  session: Session;
}

const NEEDS_SIGNUP: MockSignInOutcome = {
  redirectTo: 'signup',
  session: { status: 'needs_signup', email: MOCK_SIGNUP_EMAIL },
};

export const MOCK_SIGN_IN_OUTCOMES: Record<MockAuthScenario, MockSignInOutcome> = {
  // 01a
  'signin-default': {
    redirectTo: 'next',
    session: { status: 'signed_in', user: MOCK_USER },
  },
  // 01b, 01c — errors come back to `next`, like the real GoogleCallback
  'signin-consent-denied': {
    redirectTo: 'next',
    authError: 'access_denied',
    session: { status: 'signed_out' },
  },
  'signin-code-exchange-failed': {
    redirectTo: 'next',
    authError: 'login_failed',
    session: { status: 'signed_out' },
  },
  // 01d — sign-in succeeds; the signup POST is what fails (MOCK_SIGNUP_FAILS)
  'signin-account-creation-failed': NEEDS_SIGNUP,
  // 02a, 02b, 02c
  'setup-default': NEEDS_SIGNUP,
  'setup-name-empty': NEEDS_SIGNUP,
  'setup-name-too-long': NEEDS_SIGNUP,
};

// Scenarios whose POST /signup answers 500 "could not create account".
export const MOCK_SIGNUP_FAILS: ReadonlySet<MockAuthScenario> = new Set([
  'signin-account-creation-failed',
]);

// 02b, 02c — the exact copy the first-time setup form shows for each
// scenario. These come from validateDisplayName() at typing time, not from
// a network response; kept here so the wireframe states have one source of
// truth to test against.
export const MOCK_SETUP_FORM_ERRORS: Record<
  Extract<MockAuthScenario, 'setup-name-empty' | 'setup-name-too-long'>,
  string
> = {
  'setup-name-empty': validateDisplayName('') as string,
  'setup-name-too-long': validateDisplayName('a'.repeat(31)) as string,
};

export function isMockAuthScenario(value: string | null): value is MockAuthScenario {
  return value !== null && value in MOCK_SIGN_IN_OUTCOMES;
}

// Stand-in for the ft_session / ft_signup cookies: the mock sign-in is a real
// full-page navigation, so the state it leaves behind has to survive a reload.
// sessionStorage can be unavailable (private mode, blocked storage) — every
// access is guarded and falls back to "no state".
const MOCK_STATE_KEY = 'ft_mock_auth';

export interface MockAuthState {
  scenario: MockAuthScenario;
  session: Session;
}

export function readMockAuthState(): MockAuthState | null {
  try {
    const raw = window.sessionStorage.getItem(MOCK_STATE_KEY);
    return raw ? (JSON.parse(raw) as MockAuthState) : null;
  } catch {
    return null;
  }
}

export function writeMockAuthState(state: MockAuthState | null): void {
  try {
    if (state) {
      window.sessionStorage.setItem(MOCK_STATE_KEY, JSON.stringify(state));
    } else {
      window.sessionStorage.removeItem(MOCK_STATE_KEY);
    }
  } catch {
    // storage unavailable — mock state just won't persist
  }
}

export function mockDelay<T>(value: T, ms = MOCK_SIGN_IN_DELAY_MS): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), ms));
}
