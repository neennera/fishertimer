import Link from 'next/link';
import { PixelAlert } from '../../../components/ui/PixelAlert';
import { AdminApiError } from '../../../lib/admin-api';

/** Readable text for an admin API failure. */
export function adminErrorMessage(e: unknown, fallback: string): string {
  if (e instanceof AdminApiError) {
    if (e.status === 401) return 'Sign in with an admin account to see this page.';
    if (e.status === 403) return 'This account is not an admin.';
    if (e.status === 502) return 'Study session service is unavailable. Try again in a moment.';
    return e.message;
  }
  return e instanceof Error ? e.message : fallback;
}

/** Error banner; offers Sign in when the session is missing or expired. */
export function AdminError({ error, fallback }: { error: unknown; fallback: string }) {
  const signIn = error instanceof AdminApiError && error.status === 401;
  return (
    <PixelAlert className="mt-5 flex flex-wrap items-center justify-between gap-3" role="alert">
      <span>{adminErrorMessage(error, fallback)}</span>
      {signIn && (
        <Link href="/signin" className="pixel-btn pixel-btn--sm">
          Sign in
        </Link>
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
