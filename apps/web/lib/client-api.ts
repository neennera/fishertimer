// Typed API client for Fisher Timer Client Website
// Communicates with backend microservices via the API Gateway, through the
// same-origin /api/<service>/* rewrites in next.config.js. Same-origin matters:
// the account service's session lives in HttpOnly cookies on this origin, so
// calling the gateway's own origin directly would never send them.

// Every name here needs a matching rewrite in next.config.js.
export type ClientServiceName =
  | 'auth'
  | 'timer'
  | 'session'
  | 'reward'
  | 'leaderboard';

// Thrown on a non-2xx response. `message` is the backend's `{"error": "..."}`
// text when it sent one, so callers can branch on `status` and still show
// something meaningful.
export class ClientApiError extends Error {
  constructor(
    readonly service: ClientServiceName,
    readonly status: number,
    message: string
  ) {
    super(message);
    this.name = 'ClientApiError';
  }
}

// A call that takes longer than this is abandoned, so a hung gateway shows an
// error instead of a spinner that never stops.
export const REQUEST_TIMEOUT_MS = 15_000;

// Thrown (as a plain TypeError / AbortError / TimeoutError, not a
// ClientApiError) when the request never gets an answer. Callers treat any
// non-ClientApiError as "couldn't reach the server".
export async function clientApiFetch<T>(
  service: ClientServiceName,
  endpoint: string,
  init?: RequestInit
): Promise<T> {
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
  // A bare query ("?period=weekly") goes on /api/<service> itself, with no
  // trailing slash for Next to redirect.
  const url = cleanEndpoint === '' || cleanEndpoint.startsWith('?')
    ? `/api/${service}${cleanEndpoint}`
    : `/api/${service}/${cleanEndpoint}`;

  const timeout = AbortSignal.timeout(REQUEST_TIMEOUT_MS);
  const res = await fetch(url, {
    credentials: 'same-origin',
    ...init,
    signal: init?.signal ? AbortSignal.any([init.signal, timeout]) : timeout,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const errorBody = await res.text().catch(() => '');
    let message = errorBody || `${res.status} ${res.statusText}`;
    try {
      const parsed = JSON.parse(errorBody) as { error?: unknown };
      if (typeof parsed.error === 'string') {
        message = parsed.error;
      }
    } catch {
      // not JSON — keep the raw body
    }
    throw new ClientApiError(service, res.status, message);
  }

  return res.json() as Promise<T>;
}
