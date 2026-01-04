import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Trend } from "k6/metrics";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

const processingTimeTrend = new Trend("payment_processing_time");
const successCounter = new Counter("payment_success_total");
const failureCounter = new Counter("payment_failure_total");

export const options = {
  stages: [
    { duration: "1m", target: 20 },
    { duration: "2m", target: 20 },
    { duration: "1m", target: 200 },
    { duration: "2m", target: 200 },
    { duration: "1m", target: 20 },
    { duration: "1m", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"],
    http_req_failed: ["rate<0.01"],
    payment_processing_time: ["p(95)<10000"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8000";

export default function () {
  const reference = `ref-${Date.now()}-${__VU}-${__ITER}`;

  const payload = JSON.stringify({
    amount: randomIntBetween(1, 10000) / 100,
    currency: Math.random() > 0.5 ? "USD" : "ETB",
    reference: reference,
  });

  const params = {
    headers: { "Content-Type": "application/json" },
    tags: { name: "CreatePayment" },
  };

  const startTime = Date.now();

  const createRes = http.post(`${BASE_URL}/payments`, payload, params);

  const isCreated = check(createRes, {
    "create status is 201": (r) => r.status === 201,
  });

  if (!isCreated) {
    failureCounter.add(1, { type: "api_error" });
    return;
  }

  const paymentID = createRes.json("id");
  if (!paymentID) {
    failureCounter.add(1, { type: "missing_id" });
    return;
  }

  sleep(randomIntBetween(1, 2));

  let finalStatus = "PENDING";
  let attempts = 0;
  const maxAttempts = 20;

  while (attempts < maxAttempts) {
    const statusRes = http.get(`${BASE_URL}/payments/${paymentID}`, {
      tags: { name: "GetStatus" },
    });

    if (statusRes.status === 200) {
      finalStatus = statusRes.json("status");
      if (finalStatus === "SUCCESS" || finalStatus === "FAILED") {
        break;
      }
    }

    attempts++;
    sleep(1);
  }

  const endTime = Date.now();
  const processingTime = endTime - startTime;

  processingTimeTrend.add(processingTime);

  const isProcessed = check(finalStatus, {
    "payment eventually processed": (s) => s === "SUCCESS" || s === "FAILED",
  });

  if (isProcessed) {
    if (finalStatus === "SUCCESS") {
      successCounter.add(1);
    } else {
      failureCounter.add(1, { type: "business_failure" });
    }
  } else {
    failureCounter.add(1, { type: "processing_timeout" });
  }

  sleep(randomIntBetween(1, 5));
}
