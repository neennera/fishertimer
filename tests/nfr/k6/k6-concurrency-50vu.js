import http from 'k6/http';
import { check, sleep } from 'k6';

// NFR-PF-03: The system shall maintain acceptable response time when supporting
// at least 50 concurrent users.
export const options = {
  stages: [
    { duration: '15s', target: 20 }, // Ramp-up to 20 users
    { duration: '30s', target: 50 }, // Ramp-up to 50 concurrent users
    { duration: '30s', target: 50 }, // Sustained load at 50 concurrent users
    { duration: '15s', target: 0 },  // Ramp-down to 0 users
  ],
  thresholds: {
    // Under 50 concurrent users, response time must remain acceptable (< 2s)
    http_req_duration: ['p(95)<2000', 'p(99)<3000'],
    http_req_failed: ['rate<0.02'], // Error rate under 2%
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost';

export default function () {
  // Simulate active user traffic hitting active sessions endpoint (/api/session/active)
  const res = http.get(`${BASE_URL}/api/session/active`);
  check(res, {
    'session active query returns 200': (r) => r.status === 200,
    'latency is within acceptable threshold (<2s)': (r) => r.timings.duration < 2000,
  });

  // Short random pause between 200ms and 600ms to simulate real-world user activity
  sleep(0.3 + Math.random() * 0.3);
}
