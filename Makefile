.PHONY: proto generate docker-build docker-up docker-down migrate

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@which protoc > /dev/null || (echo "Error: protoc not found. Install it first." && exit 1)
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto
	@echo "Proto code generated successfully!"

# Generate GraphQL code
generate:
	@echo "Generating GraphQL code..."
	@go run github.com/99designs/gqlgen generate

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	@docker compose -f deploy/docker-compose.yml build

# Start services
docker-up:
	@echo "Starting services..."
	@docker compose -f deploy/docker-compose.yml up -d
	@echo "Waiting for postgres to be ready..."
	@sleep 5
	@echo "Running migrations..."
	@docker compose -f deploy/docker-compose.yml exec -T postgres psql -U polaris -d polaris -f /migrations/identity/001_init.sql
	@docker compose -f deploy/docker-compose.yml exec -T postgres psql -U polaris -d polaris -f /migrations/usage/001_init.sql
	@docker compose -f deploy/docker-compose.yml exec -T postgres psql -U polaris -d polaris -f /migrations/billing/001_init.sql

# Stop services
docker-down:
	@echo "Stopping services..."
	@docker compose -f deploy/docker-compose.yml down

# Run migrations manually
migrate:
	@echo "Running migrations..."
	@docker compose -f deploy/docker-compose.yml exec postgres psql -U polaris -d polaris -f /migrations/identity/001_init.sql
	@docker compose -f deploy/docker-compose.yml exec postgres psql -U polaris -d polaris -f /migrations/usage/001_init.sql
	@docker compose -f deploy/docker-compose.yml exec postgres psql -U polaris -d polaris -f /migrations/billing/001_init.sql

