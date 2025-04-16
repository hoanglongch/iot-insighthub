/*
For load testing, we recommend tools such as k6, Locust, or Apache JMeter. Below is an example k6 test script that simulates concurrent POST requests to your /ingest endpoint. Save the file as load_test.js in a new folder (for instance, tests/load/):
Running the Test:
Install k6 (see k6 installation guide).

Run the script:

bash
Copy
k6 run tests/load/load_test.js
Review the output. Adjust thresholds and ramp-up stages based on your target SLAs.
*/
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 50 },  // Ramp-up to 50 virtual users.
    { duration: '1m', target: 50 },   // Stay at 50 users for 1 minute.
    { duration: '30s', target: 0 },   // Ramp-down to 0.
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms.
  },
};

const apiUrl = 'http://your-secure-api-endpoint/ingest';
const authToken = 'Bearer your-valid-test-token';  // Pre-generate a test token or use a fixed test token.
const payload = JSON.stringify({
  device_id: 'load-test-device',
  value: 99.9,
  time: Math.floor(Date.now() / 1000)
});

export default function () {
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': authToken
    },
  };

  let res = http.post(apiUrl, payload, params);

  check(res, {
    'status is 202': (r) => r.status === 202,
    'response time OK': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
