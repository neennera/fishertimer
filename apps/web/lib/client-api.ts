// Typed API client for Fisher Timer Client Website
// Communicates with backend microservices via the API Gateway (Port 8000)

export type ClientServiceName =
  | 'auth'
  | 'account'
  | 'timer'
  | 'session'
  | 'reward'
  | 'leaderboard';

const API_GATEWAY_URL =
  process.env.NEXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:8000';

export async function clientApiFetch<T>(
  service: ClientServiceName,
  endpoint: string,
  init?: RequestInit
): Promise<T> {
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
  const url = `${API_GATEWAY_URL}/api/${service}/${cleanEndpoint}`;

  const res = await fetch(url, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const errorBody = await res.text().catch(() => '');
    throw new Error(
      `Client API error [${service}] ${res.status} ${res.statusText}: ${errorBody}`
    );
  }

  return res.json() as Promise<T>;
}
