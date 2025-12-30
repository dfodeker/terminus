# Terminus

This project is a small but serious backend system built to practice and demonstrate how production services behave under retries, failures, and changing requirements.

It looks like a simple store on the surface, but the focus is on backend correctness and system design rather than UI or product features.

The backend is written in Go and models an event-driven checkout flow with idempotency, async processing, and feature flags that actually change behavior. A lightweight Next.js frontend is included to make the system easy to demo and reason about.

---

> Embrace Complexity

## Why this project exists

Most side projects stop at CRUD. This one intentionally goes a step further and focuses on problems that show up in real systems:

- What happens when requests are retried?
- How do you safely process work asynchronously?
- How do you avoid double-charging or overselling?
- How do you roll out behavior changes without redeploying?

The goal is not to build a full ecommerce product, but to model the kinds of patterns and tradeoffs used in larger production backends.

---

## High-level architecture

- **API service (Go)**  
  Handles HTTP requests, validates input, writes domain state, and records events.

- **Worker service (Go)**  
  Processes domain events asynchronously using an outbox table and retry logic.

- **Postgres**  
  Primary datastore for orders, inventory, payments, feature flags, and events.

- **Next.js frontend**  
  Simple UI for browsing products, checking out, viewing order status, and toggling feature flags.

Everything runs locally via Docker Compose.

---

## Core concepts implemented

### Event-driven checkout

Checkout does not directly “do everything.” Instead, it records a `PaymentAuthorized` event and lets a worker process the rest of the workflow. This keeps the API responsive and makes retries safer.

### Outbox pattern

Events are written to the database in the same transaction as state changes. A worker polls the outbox and processes events exactly once, even across restarts.

### Idempotency

Checkout requests require an idempotency key. Retrying the same request will never create duplicate payments or inventory changes.

### Feature flags that affect backend behavior

Feature flags are not just UI toggles. For example:

- `checkout.async_enabled`
  - When enabled, checkout returns immediately and finishes in the background.
  - When disabled, checkout processes synchronously.

This makes it easy to test different execution paths and rollout strategies.

### Inventory safety

Inventory updates are done atomically to prevent overselling, even under concurrent checkouts.

---

## What’s intentionally not included (yet)

- Authentication / user accounts
- Real payment providers
- External message brokers (Kafka, NATS, etc.)
- Complex UI or styling

Those are deliberately left out to keep the focus on backend behavior and correctness.

---

## Project structure
/backend
  /cmd/api        # HTTP API
    /cmd/api/main.go
  /cmd/worker
    /cmd/worker/main.go
  
    go.mod
    go.sum
  
/frontend         # Next.js app
docker-compose.yml

Running the project locally

Prerequisites:

Docker

Docker Compose

Start everything:

docker-compose up --build


This will start:

Postgres

Go API service

Go worker service

Next.js frontend

How to try it

Open the frontend and browse products

Create an order and checkout

Toggle checkout.async_enabled in the flags page

Observe how checkout behavior changes

Retry the same checkout request and verify it’s safe

Watch order status transition as events are processed

Known limitations

Single-node setup

Simplified domain model

Minimal error handling for external integrations

No authentication or authorization

These are tradeoffs made to keep the project focused and understandable.

Future work

Failure compensation (refunds, cancellations)

Webhook delivery with retries and dead-lettering

Observability (tracing, metrics, dashboards)

External message broker integration

Load testing and concurrency benchmarks

Final note

This project is meant to be read as much as it is meant to be run.
Code clarity, explicit tradeoffs, and correctness matter more here than feature count.




