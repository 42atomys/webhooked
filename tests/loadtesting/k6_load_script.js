import http from 'k6/http';

export const options = {
  stages: [
    { duration: '5s', target: 10 },
    { duration: '10s', target: 200 },
    { duration: '10s', target: 1000 },
    { duration: '10s', target: 1000 },
    { duration: '10s', target: 100 },
    { duration: '10m', target: 100 },
    { duration: '10s', target: 10 },
    { duration: '5s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.0001'],
    http_req_duration: ['p(95)<50', 'p(99.9) < 100'],
  },
};

export default function () {
  const url = 'http://localhost:8080/v1alpha1/webhooks/example';
  const payload = JSON.stringify({
    data: {},
    timestamp: Date.now(),
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Hook-Secret': 'test'
    },
  };

  http.post(url, payload, params);
}
