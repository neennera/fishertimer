// Typed API client for Fisher Timer Admin Website
// Communicates DIRECTLY with the Admin Service (Port 8087), bypassing the API Gateway

const ADMIN_SERVICE_URL =
  process.env.NEXT_PUBLIC_ADMIN_SERVICE_URL || 'http://localhost:8087';

export async function adminApiFetch<T>(
  endpoint: string,
  init?: RequestInit
): Promise<T> {
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint.slice(1) : endpoint;
  const url = `${ADMIN_SERVICE_URL}/api/v1/admin/${cleanEndpoint}`;

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
      `Admin API error ${res.status} ${res.statusText}: ${errorBody}`
    );
  }

  return res.json() as Promise<T>;
}
