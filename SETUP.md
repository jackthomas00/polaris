# Setup Instructions

## Current Status

All the skeleton code has been created:
- ✅ Identity service (migration, store, service, main)
- ✅ Usage service (migration, store, service, main) 
- ✅ Billing service (migration, store, service, main)
- ✅ API Gateway (GraphQL schema, resolvers, middleware)
- ✅ Docker Compose configuration
- ✅ Dockerfiles for all services

## Next Steps

### 1. Generate Protobuf Code

You need to generate the Go code from the `.proto` files:

```bash
# Install protoc (if not already installed)
# Ubuntu/Debian:
sudo apt-get install protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
# Note: The go_package option in the proto files will place generated files in subdirectories
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/*.proto

# After generation, move files to the correct subdirectories (if needed)
# The go_package option should handle this automatically, but if files end up in proto/,
# move them to proto/identityv1/, proto/usagev1/, proto/billingv1/ as needed
```

This will create:
- `proto/usagev1/usage.pb.go`
- `proto/usagev1/usage_grpc.pb.go`
- `proto/identityv1/identity.pb.go`
- `proto/identityv1/identity_grpc.pb.go`
- `proto/billingv1/billing.pb.go`
- `proto/billingv1/billing_grpc.pb.go`

### 2. Generate GraphQL Code

After generating proto code, generate GraphQL code:

```bash
# Install gqlgen globally (recommended)
go install github.com/99designs/gqlgen@latest

# Then run generate
gqlgen generate

# Or use go run (may require additional dependency setup)
# go run github.com/99designs/gqlgen generate
```

This will create:
- `internal/gateway/graphql/generated/generated.go`
- `internal/gateway/graphql/models_gen.go`
- `internal/gateway/resolver_gen.go`

### 3. Update go.mod

After generating code, run:

```bash
go mod tidy
```

### 4. Start Services

```bash
# Build and start all services
docker-compose -f deploy/docker-compose.yml up --build

# Or use Makefile
make docker-up
```

### 5. Test the API

```bash
# Test with curl
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-12345" \
  -d '{
    "query": "query { me { id name } }"
  }'
```

Or visit http://localhost:8080/playground in your browser.

## Known Issues

1. **Proto code not generated**: The proto files exist but Go code needs to be generated
2. **GraphQL code not generated**: The schema exists but generated code is needed
3. **Billing service queries usage_events**: Currently queries directly from same DB. In production, should query via usage service gRPC.

## Hardcoded Test Data

- **Organization ID**: `org-1`
- **Organization Name**: `Test Organization`
- **API Key**: `test-api-key-12345`
- **Plan**: Default plan with $0.01 per unit after 1000 free quota

## Database Setup

All services use the same PostgreSQL database in Docker Compose. Migrations are run automatically on startup.

For local development, you can run migrations manually:
```bash
psql -U polaris -d polaris -f migrations/identity/001_init.sql
psql -U polaris -d polaris -f migrations/usage/001_init.sql
psql -U polaris -d polaris -f migrations/billing/001_init.sql
```

