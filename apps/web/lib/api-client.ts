// Generic typed API client for Fisher Timer microservices
// Routes requests through Next.js API Gateway rewrites (next.config.js) to eliminate CORS

export type ServiceName =
  | 'auth'
  | 'account'
  | 'session'
  | 'timer'
  | 'reward'
  | 'leaderboard'
  | 'admin';

export async function apiFetch<T>(
  service: ServiceName,
  endpoint: string,
  init?: RequestInit
): Promise<T> {
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
  // Route through Next.js reverse proxy gateway (avoids CORS issues on the client)
  const url = `/api/${service}/${cleanEndpoint}`;

  const res = await fetch(url, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const errorBody = await res.text().catch(() => '');
    throw new Error(`API error [${service}] ${res.status} ${res.statusText}: ${errorBody}`);
  }

  return res.json() as Promise<T>;
}
