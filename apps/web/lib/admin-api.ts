// Typed API client for Fisher Timer Admin Website.
// Talks to the Admin Service, bypassing the API Gateway, via the same-origin
// /api/admin/* rewrite in next.config so the browser needs no CORS. The
// ft_session cookie rides along; the Admin Service requires role = ADMIN.

import { REQUEST_TIMEOUT_MS } from './client-api';

/**
 * A failed Admin Service call. `message` is a sentence for the admin: the
 * service's own `{"error"}` text when it sent one, otherwise one written here
 * (never a raw "500 Internal Server Error").
 *
 * `status` is the HTTP status, or 0 when there was no answer: `kind` then says
 * whether the request timed out (the action may or may not have happened) or
 * never reached the server.
 */
export class AdminApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly kind?: 'timeout' | 'network',
  ) {
    super(message);
    this.name = 'AdminApiError';
  }
}

function fallbackMessage(status: number): string {
  if (status === 401) return 'Sign in with an admin account to see this page.';
  if (status === 403) return 'This account is not an admin.';
  if (status === 404) return 'That room is no longer on the lake.';
  if (status >= 500) return 'The admin service had a problem. Try again in a moment.';
  return `The request didn't go through (${status}). Try again.`;
}

export async function adminApiFetch<T>(
  endpoint: string,
  init?: RequestInit
): Promise<T> {
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
  const url = `/api/admin/${cleanEndpoint}`;

  let res: Response;
  try {
    res = await fetch(url, {
      ...init,
      credentials: 'same-origin',
      signal: init?.signal
        ? AbortSignal.any([init.signal, AbortSignal.timeout(REQUEST_TIMEOUT_MS)])
        : AbortSignal.timeout(REQUEST_TIMEOUT_MS),
      headers: {
        'Content-Type': 'application/json',
        ...init?.headers,
      },
    });
  } catch (e) {
    if (e instanceof DOMException && (e.name === 'TimeoutError' || e.name === 'AbortError')) {
      throw new AdminApiError(0, 'The admin service took too long to answer.', 'timeout');
    }
    throw new AdminApiError(0, "Couldn't reach the admin service. Check your connection.", 'network');
  }

  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null;
    throw new AdminApiError(res.status, body?.error || fallbackMessage(res.status));
  }

  try {
    return (await res.json()) as T;
  } catch {
    throw new AdminApiError(res.status, 'The admin service sent an answer this page could not read.');
  }
}
