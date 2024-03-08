import { Httpx } from 'https://jslib.k6.io/httpx/0.0.6/index.js';
import chai from 'https://jslib.k6.io/k6chaijs/4.3.4.3/index.js';
import redis from 'k6/experimental/redis';

chai.config.aggregateChecks = false;
chai.config.logFailures = true;

export const session = (testName) => {
  const session = new Httpx({
    baseURL: baseIntegrationURL + '/' + testName,
    headers: {
      'Content-Type': 'application/json',
      'X-Token': 'integration-test',
    }
  });
  return session;
}

export const redisClient = new redis.Client({
  socket: {
    host: __ENV.REDIS_HOST,
    port: 6379,
  },
  password: __ENV.REDIS_PASSWORD,
});

export const k6Options = {
  thresholds: {
    checks: ['rate == 1.00'],
    http_req_failed: ['rate == 0.00'],
  },
  vus: 1,
  iterations: 1
};

export const baseIntegrationURL = 'http://localhost:8080/v1alpha1/integration';
