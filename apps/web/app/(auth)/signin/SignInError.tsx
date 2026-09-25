"use client";

import { useSearchParams } from "next/navigation";
import { PixelAlert } from "../../../components/ui/PixelAlert";
import { readAuthError } from "../../../lib/auth";

// Reads the ?auth_error=<code> the backend's Google callback redirects back
// with, so each of 01b/01c/01d is also directly linkable, e.g.
// /signin?auth_error=access_denied. Isolated from SignInPanel so only this
// part re-renders on a searchParams change.
export function SignInError() {
  const searchParams = useSearchParams();
  const error = readAuthError(searchParams);

  if (!error) {
    return null;
  }

  return <PixelAlert className="mt-4">{error.message}</PixelAlert>;
}
