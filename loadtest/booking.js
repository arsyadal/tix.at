import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 100,
  duration: '10s',
};

const base = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  http.post(`${base}/events`, JSON.stringify({ id: 'concert-1', name: 'War Ticket Demo', stock: 100 }), {
    headers: { 'Content-Type': 'application/json' },
  });
}

export default function () {
  const res = http.post(`${base}/bookings`, JSON.stringify({ event_id: 'concert-1', user_id: `u-${__VU}-${__ITER}` }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'accepted or sold out': r => r.status === 202 || r.status === 409 });
}
