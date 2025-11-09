# Project: Polaris – Multi-tenant Usage & Billing Platform

Think “Stripe Metering + Internal Admin + Reporting” but scoped so one person can build it.

**What it is**

A platform other SaaS products (or internal teams) can use to:

* Track per-customer usage events (API calls, seats, storage, whatever)
* Define pricing plans + tiers
* Generate invoices based on usage
* Expose a clean GraphQL API + embeddable dashboard
* Run as a set of Go microservices on Kubernetes with gRPC between them

It’s business-y, practical, and exercises real-world architecture.

---

# Core Services (all in Go)

Each is its own service + DB schema. They talk via gRPC; public entry goes through an API Gateway with GraphQL.

1. **Identity Service**

   * Org, user, API keys.
   * Multi-tenant isolation.
   * Issues JWTs for services.
   * gRPC: `ValidateApiKey`, `GetOrg`, `GetUser`.

2. **Usage Ingestion Service**

   * Public HTTP endpoint:

     * `POST /ingest` with API key.
   * Validates key via Identity (gRPC).
   * Writes events to:

     * Postgres `usage_events` **and**
     * NATS/Kafka/RabbitMQ topic for async processing.

3. **Billing/Rating Service**

   * Subscribes to usage events.
   * For each org:

     * Applies pricing rules (e.g. first 10k events free, then $X/1k).
     * Maintains `usage_aggregates` per period.
   * Generates draft invoices at end of billing period.
   * gRPC API:

     * `GetCurrentUsage(org_id)`
     * `GetInvoices(org_id)`
     * `RecomputeInvoice(invoice_id)` for idempotent replay.

4. **Plan Management Service**

   * CRUD for:

     * Plans, prices, tiers, overages.
   * Used by Billing Service to interpret events.

5. **Notification Service** (small but real)

   * Listens for events like:

     * “invoice.generated”
     * “usage_threshold_crossed”
   * Sends email/webhook (simulate real integrations).

---

# API Gateway + GraphQL

Single **Gateway** (Go or TS):

* Northbound:

  * **GraphQL** endpoint for dashboard + external clients:

    * `me`, `organizations`, `usage(orgId)`, `invoices(orgId)`, `plans`, etc.
* Southbound:

  * Talks to services via gRPC:

    * Identity, Billing, Plans, etc.
* Also:

  * Validates JWTs.
  * Enforces per-org rate limits.

This is where you show:

* GraphQL schema design
* AuthZ across microservices
* Batching calls efficiently

---

# Kubernetes

You’re not doing toy K8s; you’re wiring something real:

* Deploy each service as its own Deployment + Service.
* Use:

  * `Ingress` for the GraphQL/API Gateway.
  * ConfigMaps/Secrets for DB creds, JWT secrets.
  * HorizontalPodAutoscaler on Usage Ingestion & Billing.
* Optional but strong:

  * Service mesh (Linkerd/Istio) later.
  * Prometheus + Grafana for metrics.

---

# Key Technical Problems (this is where the value is)

1. **Multi-tenancy**

   * Correct tenant scoping on every query.
   * Data model: `org_id` everywhere.
   * Show no cross-tenant leakage. This matters.

2. **Idempotent Usage Ingestion**

   * Handle duplicate events via `idempotency_key`.
   * Show that billing doesn’t blow up with retries.

3. **Event-driven Architecture**

   * Ingestion writes event → Billing consumes.
   * Recompute invoices by replaying aggregates.
   * You’ll touch ordering, retries, poison messages.

4. **Consistency vs Performance**

   * Current usage queries hit pre-aggregates, not raw events.
   * End-of-period invoices are deterministic from events.
   * You’ll have opinions here; document them.

5. **Observability**

   * `/metrics` on each service.
   * Tracing between gateway → services via gRPC metadata.
   * A simple dashboard: see:

     * usage timeline,
     * invoice list,
     * recent events.

---

# Stack Summary

* **Lang:** Go for all backend services.
* **RPC:** gRPC internal, JSON/GraphQL external.
* **DB:** Postgres (events, aggregates, plans, invoices).
* **Broker:** NATS or Kafka (pick one and commit).
* **Orchestration:** Kubernetes (kind or k3d locally).
* **Frontend:** React + GraphQL client (minimal but polished).
