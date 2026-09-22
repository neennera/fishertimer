// Canned auth responses used while NEXT_PUBLIC_USE_MOCKS is on. Covers the 7
// states shown by the sign-in and first-time setup wireframes:
//   01a Sign In (Default)
//   01b Sign In (Error: Consent Denied)
//   01c Sign In (Error: Code Exchange Failed)
//   01d Sign In (Error: Account Creation Failed)
//   02a First-Time Setup (Default)
//   02b First-Time Setup (Error: Name Empty)
//   02c First-Time Setup (Error: Name Too Long)

import type { AuthCallbackErrorCode, Session } from '../auth';
import { validateDisplayName } from '../validate-display-name';

// Query param read by handleAuthCallback() in mock mode to force a given
// outcome, e.g. /auth/callback?mockScenario=signin-consent-denied
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

export const MOCK_SESSION: Session = {
  user: {
    id: 'mock-user-1',
    email: 'angler@example.com',
    displayName: 'Chayut A.',
    role: 'CUSTOMER',
  },
  isFirstLogin: false,
};

// Returned when the callback matches an existing account with no display
// name yet — 02a/02b/02c all start from this session, differing only in
// what the user has typed into the setup form.
export const MOCK_FIRST_TIME_SESSION: Session = {
  user: {
    id: 'mock-user-new',
    email: 'new-angler@example.com',
    displayName: '',
    role: 'CUSTOMER',
  },
  isFirstLogin: true,
};

// 01b, 01c, 01d
export const MOCK_AUTH_CALLBACK_ERRORS: Record<
  Extract<
    MockAuthScenario,
    | 'signin-consent-denied'
    | 'signin-code-exchange-failed'
    | 'signin-account-creation-failed'
  >,
  { code: AuthCallbackErrorCode; message: string }
> = {
  'signin-consent-denied': {
    code: 'consent_denied',
    message: 'Sign-in was cancelled.',
  },
  'signin-code-exchange-failed': {
    code: 'code_exchange_failed',
    message: "Couldn't complete sign-in. Try again.",
  },
  'signin-account-creation-failed': {
    code: 'account_creation_failed',
    message: "Couldn't create your account. Try again.",
  },
};

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

export function mockDelay<T>(value: T, ms = MOCK_SIGN_IN_DELAY_MS): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), ms));
}
