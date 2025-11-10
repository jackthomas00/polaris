# Polaris - Multi-tenant Usage & Billing Platform

A multi-tenant usage and billing platform built with Go, gRPC, GraphQL, and PostgreSQL.

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Protocol Buffers compiler (`protoc`)
- gRPC Go plugins

### Setup

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Install protoc and plugins:**
   ```bash
   # Install protoc (varies by OS)
   # On Ubuntu/Debian:
   sudo apt-get install protobuf-compiler
   
   # Install Go plugins
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

3. **Generate protobuf code:**
   ```bash
   make proto
   # Or manually:
   protoc --go_out=. --go_opt=paths=source_relative \
     --go-grpc_out=. --go-grpc_opt=paths=source_relative \
     proto/*.proto
   ```

4. **Generate GraphQL code:**
   ```bash
   make generate
   # Or manually:
   go run github.com/99designs/gqlgen generate
   ```

5. **Start services with Docker Compose:**
   ```bash
   make docker-up
   # Or manually:
   docker-compose -f deploy/docker-compose.yml up -d
   ```

6. **Run migrations:**
   ```bash
   make migrate
   ```

### Testing the API

The API Gateway runs on port 8080. You can access:
- GraphQL Playground: http://localhost:8080/playground
- GraphQL Endpoint: http://localhost:8080/query

**Test API Key:** `test-api-key-12345` (hardcoded in migration)

**Example GraphQL Query:**
```graphql
query {
  me {
    id
    name
  }
  usage(metric: "api_calls") {
    metric
    total
    periodStart
    periodEnd
  }
  invoices {
    id
    totalAmount
    status
  }
}
```

**Example GraphQL Mutation:**
```graphql
mutation {
  recordUsage(metric: "api_calls", quantity: 10)
}
```

**Using curl:**
```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-12345" \
  -d '{
    "query": "query { me { id name } }"
  }'
```

## Architecture

- **identity-svc** (port 50051): Organization and API key management
- **usage-svc** (port 50052): Usage event ingestion and aggregation
- **billing-svc** (port 50053): Invoice generation based on usage
- **api-gateway** (port 8080): GraphQL API gateway with authentication

## Development

### Running locally (without Docker)

1. Start PostgreSQL:
   ```bash
   docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=polaris -e POSTGRES_USER=polaris -e POSTGRES_DB=polaris postgres:15-alpine
   ```

2. Run migrations:
   ```bash
   psql -U polaris -d polaris -f migrations/identity/001_init.sql
   psql -U polaris -d polaris -f migrations/usage/001_init.sql
   psql -U polaris -d polaris -f migrations/billing/001_init.sql
   ```

3. Start services (in separate terminals):
   ```bash
   # Identity service
   IDENTITY_DB_DSN="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go run cmd/identity-svc/main.go
   
   # Usage service
   USAGE_DB_DSN="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go run cmd/usage-svc/main.go
   
   # Billing service
   BILLING_DB_DSN="postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable" go run cmd/billing-svc/main.go
   
   # API Gateway
   IDENTITY_SVC_ADDR="localhost:50051" USAGE_SVC_ADDR="localhost:50052" BILLING_SVC_ADDR="localhost:50053" go run cmd/api-gateway/main.go
   ```

## Project Structure

```
.
├── cmd/              # Service entry points
│   ├── identity-svc/
│   ├── usage-svc/
│   ├── billing-svc/
│   └── api-gateway/
├── internal/         # Internal packages
│   ├── identity/
│   ├── usage/
│   ├── billing/
│   └── gateway/
├── proto/            # Protocol buffer definitions
├── migrations/       # Database migrations
├── deploy/           # Docker and K8s configs
└── pkg/              # Shared packages
```

