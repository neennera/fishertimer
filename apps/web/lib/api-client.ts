// Generic typed API client for Fisher Timer microservices

const SERVICE_PORTS = {
  auth: 8081,
  account: 8082,
  studySession: 8083,
  studyTimer: 8084,
  reward: 8085,
  leaderboard: 8086,
  admin: 8087,
} as const;

export async function fetchFromService<T>(
  service: keyof typeof SERVICE_PORTS,
  endpoint: string,
  init?: RequestInit
): Promise<T> {
  const port = SERVICE_PORTS[service];
  const url = `http://localhost:${port}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;
  const res = await fetch(url, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!res.ok) {
    throw new Error(`Service ${service} responded with ${res.status}: ${res.statusText}`);
  }

  return res.json() as Promise<T>;
}
