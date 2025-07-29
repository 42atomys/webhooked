import { check } from "k6";
import http from "k6/http";

export const options = {
  scenarios: {
    max_rps_test: {
      executor: "ramping-arrival-rate",
      startRate: 1000,
      timeUnit: "1s",
      preAllocatedVUs: 1000,
      maxVUs: 5000,
      stages: [
        { target: 1600, duration: "10s" },
        { target: 3200, duration: "20s" },
        { target: 6400, duration: "30s" },
        { target: 12800, duration: "50s" },
        { target: 25600, duration: "1m" },
        { target: 51200, duration: "2m" },
        { target: 51200, duration: "3m" },
        { target: 0, duration: "30s" },
      ],
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.0001"],
    http_req_duration: ["p(95)<50", "p(99.9) < 100"],
  },
};

export default function () {
  const url = "http://localhost:8081/webhooks/v1alpha2/loadtesting";
  const payload = JSON.stringify({
    data: {},
    timestamp: Date.now(),
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
      "X-Hook-Secret": "test",
    },
    timeout: "10s",
  };

  const res = http.post(url, payload, params);

  check(res, {
    "status is 200": (r) => r.status >= 200 && r.status < 300,
    // NOTE: Disabled due to high response times on github actions
    // Re-enable when a custom runner are configured to run load tests
    // "response time < 100ms": (r) => r.timings.duration < 100,
  });
}
