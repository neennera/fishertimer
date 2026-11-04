import http from 'k6/http';
import { check, sleep } from 'k6';

// Stress Test: Evaluates system behavior, degradation, and recovery beyond
// the normal 50 VU capacity by ramping up to 150 concurrent Virtual Users (3x normal load).
export const options = {
  stages: [
    { duration: '20s', target: 50 },  // 1. Warm-up to normal operating capacity (50 VUs)
    { duration: '20s', target: 100 }, // 2. Step 1 Stress: Push to 2x normal load (100 VUs)
    { duration: '30s', target: 150 }, // 3. Step 2 Peak Stress: Sustain at 3x capacity (150 VUs)
    { duration: '20s', target: 0 },   // 4. Recovery & Ramp-down back to 0
  ],
  thresholds: {
    // Under heavy stress (up to 150 VUs), response time should remain reasonable (< 2.5s)
    http_req_duration: ['p(95)<2500', 'p(99)<3500'],
    // System must maintain acceptable availability without catastrophic breakdown (< 5% error rate)
    http_req_failed: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost';

export default function () {
  // Realistic microservices traffic distribution:
  // - 60%: Active Study Sessions (/api/session/active -> Gateway -> Study Session -> Postgres)
  // - 30%: Leaderboard Ranking (/api/leaderboard?period=all-time -> Gateway -> Leaderboard -> Redis)
  // - 10%: Current Auth State (/api/auth/me -> Gateway -> Account -> Postgres)
  const rand = Math.random();

  if (rand < 0.60) {
    const res = http.get(`${BASE_URL}/api/session/active`);
    check(res, {
      'session active status is 200': (r) => r.status === 200,
      'session active latency < 2500ms': (r) => r.timings.duration < 2500,
    });
  } else if (rand < 0.90) {
    const res = http.get(`${BASE_URL}/api/leaderboard?period=all-time`);
    check(res, {
      'leaderboard status is 200': (r) => r.status === 200,
      'leaderboard latency < 2500ms': (r) => r.timings.duration < 2500,
    });
  } else {
    const res = http.get(`${BASE_URL}/api/auth/me`, {
      responseCallback: http.expectedStatuses(200, 401),
    });
    check(res, {
      'auth status is 200 or 401': (r) => r.status === 200 || r.status === 401,
      'auth latency < 2500ms': (r) => r.timings.duration < 2500,
    });
  }

  // Jittered think-time between 200ms and 500ms
  sleep(0.2 + Math.random() * 0.3);
}
