import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const proofGenerationTime = new Trend('proof_generation_time');
const verificationTime = new Trend('verification_time');

// Configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 10 },    // Stay at 10 users
    { duration: '30s', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 50 },    // Stay at 50 users
    { duration: '30s', target: 100 },  // Ramp up to 100 users
    { duration: '1m', target: 100 },   // Stay at 100 users
    { duration: '30s', target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    'http_req_duration': ['p(95)<5000'], // 95% of requests should be below 5s
    'errors': ['rate<0.1'],               // Error rate should be below 10%
    'http_req_failed': ['rate<0.05'],     // Failed requests should be below 5%
  },
};

// Environment variables
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'test_zapiki_key_1230ab3c044056686e2552fb5a2648cd';

export default function () {
  // Test 1: Health check
  testHealthCheck();

  // Test 2: List systems
  testListSystems();

  // Test 3: Generate commitment proof (fast, synchronous)
  testCommitmentProof();

  // Test 4: List templates
  testListTemplates();

  // Test 5: Verify proof
  testVerifyProof();

  sleep(1);
}

function testHealthCheck() {
  const res = http.get(`${BASE_URL}/health`);

  check(res, {
    'health check status is 200': (r) => r.status === 200,
    'health check returns healthy': (r) => {
      const body = JSON.parse(r.body);
      return body.status === 'healthy';
    },
  });

  errorRate.add(res.status !== 200);
}

function testListSystems() {
  const params = {
    headers: {
      'X-API-Key': API_KEY,
    },
  };

  const res = http.get(`${BASE_URL}/api/v1/systems`, params);

  check(res, {
    'list systems status is 200': (r) => r.status === 200,
    'list systems returns 4 systems': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.systems && body.systems.length === 4;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(res.status !== 200);
}

function testCommitmentProof() {
  const payload = JSON.stringify({
    proof_system: 'commitment',
    data: {
      type: 'string',
      value: `Test message ${Date.now()}`,
    },
  });

  const params = {
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
  };

  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/proofs`, payload, params);
  const duration = Date.now() - start;

  const success = check(res, {
    'commitment proof status is 200': (r) => r.status === 200,
    'commitment proof is completed': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'completed';
      } catch (e) {
        return false;
      }
    },
    'commitment proof < 500ms': (r) => duration < 500,
  });

  if (success) {
    proofGenerationTime.add(duration);
  }

  errorRate.add(res.status !== 200);

  return res;
}

function testListTemplates() {
  const params = {
    headers: {
      'X-API-Key': API_KEY,
    },
  };

  const res = http.get(`${BASE_URL}/api/v1/templates`, params);

  check(res, {
    'list templates status is 200': (r) => r.status === 200,
    'list templates returns templates': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.templates && body.templates.length > 0;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(res.status !== 200);
}

function testVerifyProof() {
  // First generate a proof
  const proofRes = testCommitmentProof();

  if (proofRes.status !== 200) {
    return;
  }

  let proofData;
  try {
    proofData = JSON.parse(proofRes.body);
  } catch (e) {
    return;
  }

  // Now verify it
  const payload = JSON.stringify({
    proof_system: 'commitment',
    proof: proofData.proof,
    verification_key: proofData.verification_key,
  });

  const params = {
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
  };

  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/verify`, payload, params);
  const duration = Date.now() - start;

  const success = check(res, {
    'verify proof status is 200': (r) => r.status === 200,
    'verify proof returns valid': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.valid === true;
      } catch (e) {
        return false;
      }
    },
    'verification < 100ms': (r) => duration < 100,
  });

  if (success) {
    verificationTime.add(duration);
  }

  errorRate.add(res.status !== 200);
}

// Summary function to display results
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'load-test-results.json': JSON.stringify(data),
  };
}

function textSummary(data, options) {
  // Simple text summary
  const indent = options.indent || '';
  const colors = options.enableColors;

  let summary = '\n' + indent + '='.repeat(60) + '\n';
  summary += indent + '  Zapiki Load Test Summary\n';
  summary += indent + '='.repeat(60) + '\n\n';

  summary += indent + `VUs: ${data.metrics.vus.values.value}\n`;
  summary += indent + `Duration: ${(data.state.testRunDurationMs / 1000).toFixed(2)}s\n\n`;

  summary += indent + 'HTTP Metrics:\n';
  summary += indent + `  Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += indent + `  Failed: ${data.metrics.http_req_failed.values.rate.toFixed(4) * 100}%\n`;
  summary += indent + `  Duration (avg): ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += indent + `  Duration (p95): ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += indent + `  Duration (p99): ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n\n`;

  if (data.metrics.errors) {
    summary += indent + `Error Rate: ${(data.metrics.errors.values.rate * 100).toFixed(2)}%\n`;
  }

  summary += indent + '='.repeat(60) + '\n';

  return summary;
}
