import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 200,
  duration: '5m',
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'https://core-placeholder.run.app';
const TOKEN = __ENV.AUTH_TOKEN || '';

function headers() {
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${TOKEN}` };
}

export default function () {
  // List teachers
  const teachers = http.get(`${BASE_URL}/api/hr/teachers`, { headers: headers() });
  check(teachers, { 'teachers 200': (r) => r.status === 200 });

  // List subjects
  const subjects = http.get(`${BASE_URL}/api/subjects`, { headers: headers() });
  check(subjects, { 'subjects 200': (r) => r.status === 200 });

  // Dashboard stats
  const stats = http.get(`${BASE_URL}/api/dashboard/stats`, { headers: headers() });
  check(stats, { 'stats 200': (r) => r.status === 200 });

  sleep(0.5);
}
