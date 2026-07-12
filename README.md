# Intunel

A queue-driven notification service built with Go and Redis Streams.

The service accepts notification requests through an HTTP API, publishes them to a Redis Stream, and processes them asynchronously using Redis Consumer Groups. Each notification is handled concurrently in its own goroutine while a semaphore limits concurrency to prevent resource exhaustion. Email delivery is currently powered by Resend.

## Architecture

```mermaid
flowchart LR
    Client[Client]
    API["Go API\nPOST /api/v1/notifications"]
    Stream["Redis Stream"]
    Group["Consumer Group"]
    Worker["Notification Worker"]
    Semaphore["Semaphore\nMax 10 Goroutines"]
    Reaper["Retry Reaper\nRuns every 60s"]
    DLQ["Dead Letter Queue\nRedis Stream"]
    Email["Resend Email API"]
    Recipient[Recipient]

    Client --> API
    API --> Stream
    Stream --> Group
    Group --> Worker
    Worker --> Semaphore
    Semaphore --> Email
    Email --> Recipient
    Worker -- "NACK / no ACK" --> Stream
    Stream -- "PEL idle 60s+" --> Reaper
    Reaper -- "max retries exceeded" --> DLQ
```

The runtime flow is:

1. A client sends a notification request to the API.
2. The API validates the request.
3. The notification is published to a Redis Stream.
4. A separate worker service continuously reads new messages using a Redis Consumer Group.
5. Each message is processed in its own goroutine.
6. A semaphore limits the number of concurrently running goroutines to 10.
7. The notification is sent through the Resend Email API.
8. Once processing succeeds, the message is acknowledged (`XACK`) to remove it from the Pending Entries List.

---

## Components

### API Service

Responsible for accepting notification requests.

Endpoints:

* `POST /api/v1/notifications`

  * validates incoming requests
  * publishes the notification to a Redis Stream
* `GET /healthz`

  * returns service health

---

### Worker Service

Runs independently from the API.

Responsibilities:

* Reads messages from Redis Streams using Consumer Groups.
* Processes notifications concurrently.
* Uses a semaphore to limit concurrency to 10 goroutines.
* Sends emails through Resend.
* Acknowledges successfully processed messages.

### Retry Reaper

Runs as a background goroutine inside the worker service.

Responsibilities:

* Wakes up every 60 seconds and scans the Pending Entries List (PEL).
* Claims messages that have been idle beyond their retry threshold.
* Applies exponential-style backoff between retries:
  * Retry 1 — after 1 minute idle
  * Retry 2 — after 5 minutes idle
  * Retry 3 — move to Dead Letter Queue
* Tracks retry count per message in a Redis Hash (`notifications:retry:<msgID>`).
* Deletes the hash key on success or after DLQ handoff.

---

### Dead Letter Queue (DLQ)

A separate Redis Stream (`notifications:stream:dead`) that holds messages which have exhausted all retry attempts.

Each DLQ entry stores:

* `data` — original notification payload
* `error` — exact error from the last failed attempt
* `failed_at` — unix timestamp of final failure
* `msg_id` — original stream message ID for traceability

---

### Redis Streams

Redis Streams act as the message broker between the API and worker.

Benefits:

* Durable message queue
* Consumer Groups for horizontal scaling
* Pending Entries List for reliability
* Message acknowledgements (`XACK`)
* Failed message recovery (`XCLAIM`/`XAUTOCLAIM` support)

---

### Email Provider

Current provider:

* Resend

The email layer is isolated behind an interface, making it easy to replace or add providers in the future.

---

## Architecture Decisions

### Asynchronous Processing

Instead of sending emails directly from the API, requests are queued in Redis Streams.

Benefits:

* Faster API responses
* Better fault tolerance
* Retry capability
* Horizontal scalability

### Consumer Groups

Consumer Groups ensure that:

* each notification is processed only once
* multiple worker instances can share the workload
* failed messages remain recoverable

### Controlled Concurrency

Although each notification is processed in a separate goroutine, a semaphore limits concurrent processing to **10** workers.

This prevents:

* excessive memory usage
* API rate-limit spikes
* overwhelming the email provider
* uncontrolled goroutine growth

### Retry Mechanism with Backoff

Failed messages are not dropped or retried immediately. Instead they remain in the Redis Pending Entries List (PEL) and a separate Reaper process claims them after an idle threshold.

Backoff windows:
* Retry 1 — 1 minute
* Retry 2 — 5 minutes
* Retry 3 — Dead Letter Queue

Retry state is stored in a Redis Hash per message with a 24 hour TTL as a safety net against orphaned keys.

### Dead Letter Queue

Messages that fail all 3 retry attempts are moved to a dedicated DLQ stream with the original payload and exact error reason preserved. This allows for future inspection, alerting, or manual replay without losing the message.

---

## Running Locally

### Using Docker

```bash
docker compose up --build
```

This starts:

* Redis
* API Service
* Notification Worker

The API will be available at:

```
http://localhost:8080
```

---

### Without Docker

Start Redis.

Run the API:

```bash
go run ./cmd/api
```

Run the worker:

```bash
go run ./cmd/worker
```

---

## Example Request

```http
POST /api/v1/notifications
Content-Type: application/json
```

```json
{
  "channel": "email",
  "to": "john@example.com",
  "title": "Welcome",
  "body": "<h1>Welcome to our platform!</h1>"
}
```

---

## Processing Pipeline

```mermaid
flowchart TD
    Request["Receive HTTP Request"]
    Validate["Validate Payload"]
    Publish["Publish to Redis Stream"]
    Read["Worker Reads Stream"]
    Semaphore["Acquire Semaphore Slot"]
    Goroutine["Start Goroutine"]
    Send["Send Email via Resend"]
    Success{"Success?"}
    Ack["XACK Message"]
    Release["Release Semaphore Slot"]
    PEL["Stay in PEL"]
    Reaper["Reaper Claims after idle threshold"]
    RetryCount{"Retry count?"}
    Retry["Retry Send"]
    DLQ["Move to DLQ"]

    Request --> Validate
    Validate --> Publish
    Publish --> Read
    Read --> Semaphore
    Semaphore --> Goroutine
    Goroutine --> Send
    Send --> Success
    Success -- "yes" --> Ack
    Ack --> Release
    Success -- "no" --> PEL
    PEL --> Reaper
    Reaper --> RetryCount
    RetryCount -- "< 3" --> Retry
    Retry --> Success
    RetryCount -- ">= 3" --> DLQ
```

---

## Technologies

* Go
* Redis Streams
* Redis Consumer Groups
* Redis Hash (retry state)
* Dead Letter Queue pattern
* Goroutines
* Channels
* Semaphore Pattern
* Resend Email API
* Docker

---

## Future Improvements

* Scheduled notifications
* Email templates
* Metrics and monitoring
* Distributed tracing
* Rate limiting
* Retry dashboards
* Provider failover

---

## Notes

* The API is intentionally lightweight and only publishes messages to Redis.
* Email sending is fully asynchronous.
* Consumer Groups enable multiple worker instances for horizontal scaling.
* Goroutine concurrency is intentionally capped at 10 using a semaphore to maintain stable resource usage.
* Additional notification channels can be added without changing the API by extending the worker layer.

