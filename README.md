# Terminus

This project is a deliberately small but serious backend system designed to model how real production services behave under stress, change, and failure.

At first glance, it looks like a simple store. Under the hood, it is an event-driven, multi-tenant backend that explores correctness, resilience, and operational discipline rather than surface-level product features.

The backend is written in Go and implements an asynchronous checkout flow with idempotency, retries, background processing, and feature flags that genuinely alter runtime behavior. A lightweight Next.js frontend exists only to make the system easier to demo, inspect, and reason about—it is not the focus.

As the system evolved, it grew into something closer to a “Shopify-lite” backend: capable of serving multiple stores, handling real scaling concerns, and supporting analytics, fault detection, and load testing. The goal is not completeness, but realism.

---

> Embrace Complexity

## Why this project exists

 This project intentionally goes further than a CRUD application.

The project exists to practice and demonstrate the kinds of problems that show up in real backend systems:

Requests are retried.

Work is processed asynchronously.

Events arrive more than once—or not at all.

Requirements change after code is already deployed.

Systems must scale, degrade gracefully, and remain observable.

Instead of optimizing for features or polish, the system optimizes for correctness under failure. It asks questions like:

What happens if the same checkout is processed twice?

How do you guarantee idempotency across async boundaries?

How do you roll out behavior changes safely without redeploying?

How do you detect and reason about faults before users report them?

How does a multi-tenant system fail, and how do you see it happening?
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




