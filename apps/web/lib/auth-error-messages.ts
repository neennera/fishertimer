// Codes the account service puts in ?auth_error= when it redirects back from
// Google (services/account/internal/adapter/handler/http_handler.go):
//   access_denied  - Google's own ?error= passed through; what Cancel sends (01b)
//   invalid_state  - the anti-CSRF state cookie was missing or didn't match
//   login_failed   - code exchange / Google profile lookup failed (01c)
// Google can pass through other ?error= values too; readAuthError() shows
// those with the login_failed copy.
//
// account_creation_failed is the frontend's own code, not the backend's: the
// signup POST failing (500) sends /welcome back to
// /signin?auth_error=account_creation_failed, keeping 01d on the sign-in
// screen as in the wireframes.
export type AuthErrorCode =
  | 'access_denied'
  | 'invalid_state'
  | 'login_failed'
  | 'account_creation_failed';

// Copy for each error state (01b/01c/01d). Shared by auth.ts, its mock and
// /signin. Lives in its own module, not lib/auth.ts, so
// lib/mocks/auth.mock.ts can import it without a value-level import cycle
// with auth.ts.
export const AUTH_ERROR_MESSAGES: Record<AuthErrorCode, string> = {
  access_denied: 'Sign-in was cancelled. No account changes were made.',
  invalid_state: 'Your sign-in link expired. Please try again.',
  login_failed: 'Something went wrong connecting to Google. Please try again.',
  account_creation_failed: "Couldn't create your account. Please try again.",
};
