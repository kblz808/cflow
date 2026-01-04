# How to run
i recommend using the tool [taskfile](https://taskfile.dev) to run the predefined task scripts for easier development experience

the app contains two binaries: api and worker, the api receives messages from client and publishes payment event to a message queue, the worker consumes the payment event and updates the payment status in the database.

> ___before running the application, make sure to copy the `.env.example` to `.env` and fill in the required values___

then start the application use docker-compose, this will start the database, rabbitmq, api and worker containers
```bash
docker-compose up --build
```

or using taskfile
```bash
task app
```

---
if you have go installed in your system and dont want to use docker for the api and worker, then build the binaries using

```bash
go build -o bin/api cmd/api/main.go
go build -o bin/worker cmd/worker/main.go
```

start the database and rabbitmq containers
```bash
docker-compose up db rabbitmq -d
```

then run the binaries
```bash
./bin/api
./bin/worker
```
---

by default the api server will be available on port __8000__ `http://localhost:8000`. which we can change in the `.env` `ECHO_PORT` variable

we can test the application by calling these endpoints:
- `POST` `/payments` - to create payments
    + request body: json
    ```json
    {
        "amount": 100,
        "currency": "USD",
        "reference": "arc0104-123456"
    }
    ```
- `GET` `/payments/{id}` - to get information about a payment using its id
    + response type: json
    ```json
    {
        "amount": 100,
        "currency": "USD",
        "reference": "arc0104-123456",
        "status": "SUCCESS",
        "created_at": "2026-01-04T07:27:04.123456Z"
    }
    ```

## scaling
we can scale the application using 2 mechanisms.

1. by increasing the number of workers in the `.env` file which will increase the number of goroutines processing the payment events.
2. by using `--scale worker=2` with docker-compose. which will create 2 worker containers. these containers will share the same database and message queue and run in parallel.

___note___:
if we used both combinations, lets say `--scale worker=2` and `MQ_WORKER_COUNT=16` then docker will create 2 worker containers and each container will have 16 workers (goroutines), so total of 32 workers.

## load testing
to test the application under a simulated environment of heavy transaction and user activities, ive included a k6 load testing script in the `tests/` directory.

the script will test the application payment creation and status retrieval.
example:
```js
stages: [
{ duration: "1m", target: 20 },  // normal: up to 20 users
{ duration: "2m", target: 20 }, 
{ duration: "1m", target: 200 }, // surge: spike to 200 users
{ duration: "2m", target: 200 },
{ duration: "1m", target: 20 },  // normal: ramp down to 20 users
{ duration: "1m", target: 0 },   // normal: ramp down to 0
],
```

we can run the load test using docker
```bash
docker run --rm -i --network="host" grafana/k6 run - <tests/load_test.js
```

or using taskfile
```bash
task test:load
```

my test result using 1 worker process and `MQ_WORKER_COUNT=16` (16 workers):
+ the 95th percentile (p95) for request duration was only 10.73ms
+ there were zero HTTP failures (0.00% error rate) across 18,472 requests.
+ the system handled the surge to 200 concurrent users without any increase in latency or error rates.
+ the test has a mandatory 1 or 2 seconds wait after create payment request before getting the payment status to allow the worker to process the payment event and update the payment status in the database. the result shows a (p95) of 2.01 seconds