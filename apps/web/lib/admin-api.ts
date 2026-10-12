// Typed API client for Fisher Timer Admin Website.
// Talks to the Admin Service, bypassing the API Gateway, via the same-origin
// /api/admin/* rewrite in next.config so the browser needs no CORS. The
// ft_session cookie rides along; the Admin Service requires role = ADMIN.

/** A non-2xx Admin Service response. `message` is its `{"error"}` text. */
export class AdminApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = 'AdminApiError';
  }
}

export async function adminApiFetch<T>(
  endpoint: string,
  init?: RequestInit
): Promise<T> {
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
  const url = `/api/admin/${cleanEndpoint}`;

  const res = await fetch(url, {
    ...init,
    credentials: 'same-origin',
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null;
    throw new AdminApiError(res.status, body?.error || `${res.status} ${res.statusText}`);
  }

  return res.json() as Promise<T>;
}
