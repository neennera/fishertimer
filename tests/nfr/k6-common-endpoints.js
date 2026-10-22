import http from 'k6/http';
import { check, sleep } from 'k6';

// NFR-PF-01: The system shall respond to common user requests within 2 seconds
// under normal operating conditions.
export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    // 95% of all requests must finish within 2000ms (2 seconds SLA)
    http_req_duration: ['p(95)<2000', 'p(99)<2500'],
    // Less than 1% failed requests under normal load
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost';

export default function () {
  // 1. Common user request: List Active Study Sessions (/api/session/active)
  const roomsRes = http.get(`${BASE_URL}/api/session/active`);
  check(roomsRes, {
    'rooms status is 200': (r) => r.status === 200,
    'rooms response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  // 2. Common user request: View Leaderboard
  const lbRes = http.get(`${BASE_URL}/api/leaderboard?period=all-time`);
  check(lbRes, {
    'leaderboard status is 200': (r) => r.status === 200,
    'leaderboard response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  // 3. Common user request: Check Session / Current User
  // Note: 200 (signed in) and 401 (guest/unauthenticated) are both valid status codes.
  const authRes = http.get(`${BASE_URL}/api/auth/me`, {
    responseCallback: http.expectedStatuses(200, 401),
  });
  check(authRes, {
    'auth response status is valid (200 or 401)': (r) => r.status === 200 || r.status === 401,
    'auth response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  sleep(1);
}
