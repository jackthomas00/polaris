Good. Let’s turn **Polaris** into a scoped, buildable v1 that’s technically deep but finishable.

This version teaches you Go, gRPC, GraphQL, Postgres, and K8s — but without drowning you.

---

## 🚀 v1 Goal

A **multi-tenant usage and billing MVP** with:

* Event ingestion
* Periodic aggregation
* Invoice generation
* gRPC microservice boundaries
* GraphQL gateway
* Kubernetes deployability

No external payments, no Kafka cluster — local Postgres + NATS (or even Redis Streams) is enough.

---

## 🧩 Architecture (3 microservices + gateway)

### 1. **Identity Service**  (`identity-svc`)

Handles orgs, users, API keys, and authentication.

**DB tables:**

* `organizations(id, name, created_at)`
* `api_keys(id, org_id, key, created_at)`

**gRPC API:**

```proto
service Identity {
  rpc ValidateApiKey(ValidateApiKeyRequest) returns (ValidateApiKeyResponse);
  rpc GetOrganization(GetOrganizationRequest) returns (GetOrganizationResponse);
}
```

Used by gateway + ingestion service for auth.

---

### 2. **Usage Service** (`usage-svc`)

Receives usage events and maintains aggregates.

**DB tables:**

* `usage_events(id, org_id, metric, quantity, occurred_at, idempotency_key)`
* `usage_aggregates(org_id, metric, period_start, period_end, total)`

**Endpoints:**

* **HTTP** `POST /ingest`

  * Header: `X-API-Key`
  * Body: `{ "metric": "api_calls", "quantity": 5, "timestamp": "..." }`
* **gRPC internal:**

  ```proto
  service Usage {
    rpc RecordUsage(RecordUsageRequest) returns (RecordUsageResponse);
    rpc GetUsageSummary(GetUsageSummaryRequest) returns (GetUsageSummaryResponse);
  }
  ```
* Emits internal events like `UsageRecorded` via NATS (simple JSON pub/sub).

**Aggregation Job (cron inside service):**

* Every 5 minutes:

  * Sum usage per org per metric per day.
  * Upsert into `usage_aggregates`.
  * Publish `"aggregate.updated"` event.

---

### 3. **Billing Service** (`billing-svc`)

Turns aggregates into invoices.

**DB tables:**

* `plans(id, org_id, name, metric, unit_price, free_quota)`
* `invoices(id, org_id, period_start, period_end, total_amount, status)`

**gRPC API:**

```proto
service Billing {
  rpc GenerateInvoice(GenerateInvoiceRequest) returns (Invoice);
  rpc ListInvoices(ListInvoicesRequest) returns (ListInvoicesResponse);
}
```

**Logic:**

* Subscribes to `"aggregate.updated"`.
* For each org & metric:

  * total = max(0, usage - free_quota) × unit_price
* Writes invoice draft rows.

---

### 4. **API Gateway** (`api-gateway`)

Single public entrypoint (GraphQL).

**GraphQL schema (simplified):**

```graphql
type Query {
  me: Organization!
  usage(metric: String!): [UsageAggregate!]!
  invoices: [Invoice!]!
}

type Mutation {
  recordUsage(metric: String!, quantity: Int!): Boolean!
}

type Organization {
  id: ID!
  name: String!
}

type UsageAggregate {
  metric: String!
  total: Float!
  periodStart: String!
  periodEnd: String!
}

type Invoice {
  id: ID!
  totalAmount: Float!
  status: String!
  periodStart: String!
  periodEnd: String!
}
```

**Gateway responsibilities:**

* Accepts JWT or API key.
* Validates via `identity-svc` (gRPC).
* Delegates:

  * `recordUsage` → `usage-svc`
  * `usage(...)` → `usage-svc`
  * `invoices` → `billing-svc`

Use **gRPC + grpc-gateway** or native **gRPC-web** bridging.

---

## 🛠️ Tech Stack Summary

| Layer          | Choice                              | Why                                   |
| -------------- | ----------------------------------- | ------------------------------------- |
| Language       | Go                                  | Great for gRPC + concurrency          |
| Transport      | gRPC internal, GraphQL external     | Demonstrates real API gateway pattern |
| DB             | PostgreSQL                          | Strong relational base                |
| Message broker | NATS (or Redis Streams)             | Simple async pub/sub                  |
| Orchestration  | Kubernetes (k3d, kind, or minikube) | For real deployment experience        |
| Config         | Env vars + ConfigMap + Secrets      | Best practices                        |
| Observability  | Prometheus metrics + logs           | Keep it real but small                |
| CI             | Docker Compose locally              | Dev convenience                       |

---

## ⚙️ Data Flow Example

**Client → Billing Flow**

1. Client posts usage via API Gateway
   → Gateway verifies API key (gRPC Identity)
   → Calls `RecordUsage` on `usage-svc`

2. `usage-svc` saves event, publishes NATS event.

3. Billing service consumes NATS event, recomputes invoice total.

4. GraphQL dashboard queries show current usage + invoices.

---

## 🧱 Implementation Order

**Phase 1 — Core Foundations**

1. Scaffold repos (`cmd/identity`, `cmd/usage`, `cmd/billing`, `cmd/gateway`).
2. Dockerfiles + Compose file.
3. gRPC setup between services.
4. Basic Postgres migrations.
5. Identity + API key validation working end-to-end.

**Phase 2 — Usage Recording**

1. `/ingest` endpoint → DB insert.
2. Gateway → `usage.RecordUsage` integration.
3. CLI test: record and fetch usage.

**Phase 3 — Billing**

1. Add `plans` + simple invoice logic.
2. Manual `GenerateInvoice` RPC call.
3. GraphQL `invoices` query.

**Phase 4 — Kubernetes Deployment**

1. Convert Compose → manifests:

   * Deployments, Services, Ingress.
   * Postgres via StatefulSet.
2. Test local k3d/minikube deployment.
3. Add metrics.

**Phase 5 — (Optional polish)**

* React dashboard to view usage & invoices.
* JWT-based org login.
* Automated NATS-triggered invoice recompute.
* CI pipeline.

---

## 🧭 Learning Focus Breakdown

| Concept                      | Where You’ll Learn It            |
| ---------------------------- | -------------------------------- |
| Go concurrency, interfaces   | Billing + Usage services         |
| gRPC IDL, stubs, errors      | All inter-service calls          |
| GraphQL schema & resolvers   | API Gateway                      |
| Database design + migrations | Postgres per service             |
| Pub/Sub eventing             | NATS integration                 |
| Multi-tenancy                | Identity auth & `org_id` scoping |
| Kubernetes basics            | Deployments, Ingress, Services   |
| Observability                | Metrics, healthchecks, logs      |

---

## 💡 Extension Ideas (future)

* Add **seat-based pricing** or **storage-based metrics**.
* Build a **React-based admin UI**.
* Replace NATS with Kafka for durability.
* Add **webhooks** for invoice created events.
* Integrate Stripe for actual payments.
