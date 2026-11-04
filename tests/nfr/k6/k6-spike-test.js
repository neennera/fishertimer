import http from 'k6/http';
import { check, sleep } from 'k6';

// Spike Test: Evaluates resilience, connection pooling, and immediate recovery
// during sudden, violent surges in user traffic (from 10 to 150 VUs within 5 seconds).
export const options = {
  stages: [
    { duration: '10s', target: 10 },  // 1. Low baseline traffic (10 VUs)
    { duration: '5s',  target: 150 }, // 2. Violent spike surge (10 -> 150 VUs in 5s)
    { duration: '15s', target: 150 }, // 3. Sustained peak surge traffic
    { duration: '5s',  target: 10 },  // 4. Sharp drop back to baseline (10 VUs)
    { duration: '15s', target: 10 },  // 5. Recovery phase: verify latency returns to normal
    { duration: '5s',  target: 0 },   // 6. Ramp-down to 0
  ],
  thresholds: {
    // Under sudden flash traffic, overall p95 latency should remain below 3000ms
    http_req_duration: ['p(95)<3000'],
    // System error rate must remain under 5% during the surge
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
      'session active responds': (r) => r.status > 0,
    });
  } else if (rand < 0.90) {
    const res = http.get(`${BASE_URL}/api/leaderboard?period=all-time`);
    check(res, {
      'leaderboard status is 200': (r) => r.status === 200,
      'leaderboard responds': (r) => r.status > 0,
    });
  } else {
    const res = http.get(`${BASE_URL}/api/auth/me`, {
      responseCallback: http.expectedStatuses(200, 401),
    });
    check(res, {
      'auth status is 200 or 401': (r) => r.status === 200 || r.status === 401,
      'auth responds': (r) => r.status > 0,
    });
  }

  // Jittered think-time between 100ms and 400ms during burst traffic
  sleep(0.1 + Math.random() * 0.3);
}
