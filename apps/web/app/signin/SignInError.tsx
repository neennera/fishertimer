"use client";

import { useSearchParams } from "next/navigation";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { AUTH_CALLBACK_ERROR_MESSAGES, type AuthCallbackErrorCode } from "../../lib/auth";

function isAuthCallbackErrorCode(value: string | null): value is AuthCallbackErrorCode {
  return value !== null && value in AUTH_CALLBACK_ERROR_MESSAGES;
}

// Reads ?error=<code> so each of 01b/01c/01d is directly linkable, e.g.
// /signin?error=consent_denied. Isolated from SignInPage so only this part
// needs the Suspense boundary useSearchParams() requires.
export function SignInError() {
  const searchParams = useSearchParams();
  const code = searchParams.get("error");

  if (!isAuthCallbackErrorCode(code)) {
    return null;
  }

  return <PixelAlert className="mt-4">{AUTH_CALLBACK_ERROR_MESSAGES[code]}</PixelAlert>;
}
