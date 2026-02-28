import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

const checkoutAccepted = new Counter('checkout_accepted');
const checkoutRejected = new Counter('checkout_rejected');
const checkoutErrors = new Counter('checkout_errors');
const checkoutDuration = new Trend('checkout_duration');

export const options = {
  scenarios: {
    race_condition: {
      executor: 'constant-vus',
      vus: 100,
      duration: '1s',
      exec: 'checkoutStorm',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.1'],
    http_req_duration: ['p(95)<5000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  const baseUrl = __ENV.BASE_URL || BASE_URL;
  const unique = Date.now().toString(36) + Math.random().toString(36).slice(2, 8);

  http.post(
    `${baseUrl}/api/v1/auth/register`,
    JSON.stringify({
      email: `race_${unique}@test.local`,
      password: 'TestPass123!',
      name: 'Race Test User',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  const loginBody = JSON.stringify({
    email: `race_${unique}@test.local`,
    password: 'TestPass123!',
  });

  const loginRes = http.post(`${baseUrl}/api/v1/auth/login`, loginBody, {
    headers: { 'Content-Type': 'application/json' },
  });

  if (loginRes.status !== 200) {
    throw new Error(`Setup login failed: ${loginRes.status} ${loginRes.body}`);
  }

  const loginData = loginRes.json();
  const token = loginData.access_token;
  const userId = loginData.user?.id;
  if (!token || !userId) {
    throw new Error(`Setup: no token or user id in ${loginRes.body}`);
  }

  const stock = Math.floor(Math.random() * 10) + 1;
  const price = Math.floor(Math.random() * 9000000) + 100000;
  const discount = Math.floor(Math.random() * 10) + 90;

  const productRes = http.post(
    `${baseUrl}/api/v1/products/`,
    JSON.stringify({
      name: `Flash Product Race ${unique}`,
      category: 'Electronics',
      stock,
      price,
      discount,
      created_by: userId,
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
    }
  );

  if (productRes.status !== 201) {
    throw new Error(`Setup create product failed: ${productRes.status} ${productRes.body}`);
  }

  const productData = productRes.json();
  const productId = productData.id;
  if (!productId) {
    throw new Error(`Setup: no product id in ${productRes.body}`);
  }

  return { baseUrl, token, productId };
}

export function checkoutStorm(data) {
  if (!data || !data.token || !data.productId) {
    return;
  }

  const url = `${data.baseUrl}/api/v1/checkouts/`;
  const payload = JSON.stringify({
    product_id: data.productId,
    quantity: 1,
  });
  const params = {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${data.token}`,
    },
    tags: { name: 'checkout' },
  };

  const res = http.post(url, payload, params);
  const duration = res.timings.duration;
  checkoutDuration.add(duration);

  const ok = check(res, {
    'status is 202 or 400': (r) => r.status === 202 || r.status === 400,
  });

  if (res.status === 202) {
    checkoutAccepted.add(1);
  } else if (res.status === 400) {
    checkoutRejected.add(1);
  } else {
    checkoutErrors.add(1);
  }

  if (!ok) {
    console.warn(`[VU ${__VU}] unexpected status ${res.status}: ${res.body}`);
  }

  sleep(Math.random() * 0.5);
}

export function teardown(data) {
  if (data) {
    console.log(`[Teardown] Product ID: ${data.productId}`);
    console.log('Verify stock in DB: total sold should not exceed initial stock.');
  }
}
