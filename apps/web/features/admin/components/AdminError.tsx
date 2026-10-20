import Link from 'next/link';
import { PixelAlert } from '../../../components/ui/PixelAlert';
import { PixelButton } from '../../../components/ui/PixelButton';
import { AdminApiError } from '../../../lib/admin-api';

/** Readable text for an admin API failure. */
export function adminErrorMessage(e: unknown, fallback: string): string {
  if (e instanceof AdminApiError) {
    // The service's own 401 / 403 texts are terse ("sign in required"); say what to do.
    if (e.status === 401) return 'Sign in with an admin account to see this page.';
    if (e.status === 403) return 'This account is not an admin.';
    if (e.status === 502) return 'Study session service is unavailable. Try again in a moment.';
    return e.message;
  }
  return e instanceof Error ? e.message : fallback;
}

/**
 * A moderation call that got no answer (timed out, or the connection dropped):
 * the kick or close may or may not have happened, so the page refreshes instead
 * of offering a blind retry.
 */
export const isUnknownOutcome = (e: unknown) => e instanceof AdminApiError && e.kind !== undefined;

export const UNKNOWN_OUTCOME_MESSAGE =
  "The admin service didn't answer, so it isn't clear whether that went through. Check the list before trying again.";

/**
 * Error banner. Offers Sign in when the session is missing or expired, and
 * Retry for anything that may pass (not for a 403: a different account is the fix).
 */
export function AdminError({
  error,
  fallback,
  onRetry,
}: {
  error: unknown;
  fallback: string;
  onRetry?: () => void;
}) {
  const status = error instanceof AdminApiError ? error.status : undefined;
  const signIn = status === 401;
  const retry = onRetry && status !== 401 && status !== 403;
  return (
    <PixelAlert className="mt-5 flex flex-wrap items-center justify-between gap-3">
      <span>{adminErrorMessage(error, fallback)}</span>
      {signIn && (
        <Link href="/signin" className="pixel-btn pixel-btn--sm min-h-11 sm:min-h-0">
          Sign in
        </Link>
      )}
      {retry && (
        <PixelButton small className="min-h-11 sm:min-h-0" onClick={onRetry}>
          Retry
        </PixelButton>
      )}
    </PixelAlert>
  );
}

/** Shown when a refresh failed and the data on screen may be out of date (UC-04 E-2). */
export function StaleNotice({ stale }: { stale: boolean }) {
  if (!stale) return null;
  return (
    <span className="font-label text-[10px] uppercase text-rust" role="status">
      Out of date · retrying
    </span>
  );
}
