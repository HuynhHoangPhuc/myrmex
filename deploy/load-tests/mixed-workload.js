import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 100 },
    { duration: '5m', target: 500 },
    { duration: '2m', target: 500 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'https://core-placeholder.run.app';
const TOKEN = __ENV.AUTH_TOKEN || '';

function headers() {
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${TOKEN}` };
}

export default function () {
  const endpoints = [
    `${BASE_URL}/health`,
    `${BASE_URL}/api/hr/teachers`,
    `${BASE_URL}/api/subjects`,
    `${BASE_URL}/api/dashboard/stats`,
  ];
  const url = endpoints[Math.floor(Math.random() * endpoints.length)];
  const needsAuth = url.includes('/api/') && !url.includes('/auth/');
  const res = http.get(url, needsAuth ? { headers: headers() } : {});
  check(res, { 'status ok': (r) => r.status < 400 });
  sleep(0.2);
}
