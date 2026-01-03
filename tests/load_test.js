import http from "k6/http";
import { check, sleep } from "k6";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

export const options = {
  stages: [
    { duration: "2m", target: 100 },
    { duration: "4m", target: 100 },
    { duration: "2m", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"],
    http_req_failed: ["rate<0.01"],
  },
};

const BASE_URL = "http://localhost:8000";

export default function () {
  // Create payment
  const payload = {
    amount: randomIntBetween(10, 1000),
    currency: "USD",
    reference: `ref-${__VU}-${__ITER}`,
  };

  const createRes = http.post(`${BASE_URL}/payments`, JSON.stringify(payload), {
    headers: { "Content-Type": "application/json" },
  });

  check(createRes, {
    "create status 202": (r) => r.status === 201,
  });

  if (createRes.status !== 201) {
    return;
  }

  const paymentID = createRes.json("id");

  let finalStatus;
  let attempts = 0;
  while (attempts < 20) {
    const statusRes = http.get(`${BASE_URL}/payments/${paymentID}`);
    finalStatus = statusRes.json("status");
    if (finalStatus === "SUCCESS" || finalStatus === "FAILED") break;
    sleep(1);
    attempts++;
  }

  check(finalStatus, {
    "eventually processed": (s) => s === "SUCCESS" || s === "FAILED",
  });

  sleep(1);
}
