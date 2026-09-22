export type AuthCallbackErrorCode =
  | 'consent_denied'
  | 'code_exchange_failed'
  | 'account_creation_failed';

// Copy for each error state (01b/01c/01d). Shared by auth.ts, its mock and
// /signin, which also reaches these via ?error=<code> for direct links.
// Lives in its own module, not lib/auth.ts, so lib/mocks/auth.mock.ts can
// import it without a value-level import cycle with auth.ts.
export const AUTH_CALLBACK_ERROR_MESSAGES: Record<AuthCallbackErrorCode, string> = {
  consent_denied: 'Sign-in was cancelled. No account changes were made.',
  code_exchange_failed: 'Something went wrong connecting to Google. Please try again.',
  account_creation_failed: "Couldn't create your account. Please try again.",
};
