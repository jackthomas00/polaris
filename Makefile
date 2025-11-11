.PHONY: proto generate docker-build docker-up docker-down migrate k8s-build k8s-load k8s-deploy k8s-undeploy k8s-migrate

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

# Build Docker images for Kubernetes (local tags)
k8s-build:
	@echo "Building Docker images for Kubernetes..."
	@docker build -f deploy/Dockerfile.gateway -t you/polaris-gateway:local .
	@docker build -f deploy/Dockerfile.identity -t you/polaris-identity:local .
	@docker build -f deploy/Dockerfile.usage -t you/polaris-usage:local .
	@docker build -f deploy/Dockerfile.billing -t you/polaris-billing:local .
	@echo "Images built successfully!"

# Load images into Kubernetes cluster (detects kind/k3d/minikube)
k8s-load:
	@echo "Loading images into Kubernetes cluster..."
	@if command -v kind > /dev/null && kind get clusters | grep -q .; then \
		echo "Detected kind cluster, loading images..."; \
		kind load docker-image you/polaris-gateway:local; \
		kind load docker-image you/polaris-identity:local; \
		kind load docker-image you/polaris-usage:local; \
		kind load docker-image you/polaris-billing:local; \
	elif command -v k3d > /dev/null && k3d cluster list | grep -q .; then \
		echo "Detected k3d cluster, loading images..."; \
		CLUSTER=$$(k3d cluster list --no-headers | head -1 | awk '{print $$1}'); \
		echo "Using cluster: $$CLUSTER"; \
		k3d image import you/polaris-gateway:local -c $$CLUSTER; \
		k3d image import you/polaris-identity:local -c $$CLUSTER; \
		k3d image import you/polaris-usage:local -c $$CLUSTER; \
		k3d image import you/polaris-billing:local -c $$CLUSTER; \
	elif command -v minikube > /dev/null && minikube status > /dev/null 2>&1; then \
		echo "Detected minikube, loading images..."; \
		minikube image load you/polaris-gateway:local; \
		minikube image load you/polaris-identity:local; \
		minikube image load you/polaris-usage:local; \
		minikube image load you/polaris-billing:local; \
	else \
		echo "No supported Kubernetes cluster detected (kind/k3d/minikube)."; \
		echo "If using a different setup, ensure images are available to your cluster nodes."; \
	fi

# Deploy to Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deploy/k8s/namespace.yaml
	@kubectl apply -f deploy/k8s/postgres.yaml
	@kubectl apply -f deploy/k8s/nats.yaml
	@kubectl apply -f deploy/k8s/identity-deploy.yaml
	@kubectl apply -f deploy/k8s/usage-deploy.yaml
	@kubectl apply -f deploy/k8s/billing-deploy.yaml
	@kubectl apply -f deploy/k8s/gateway-deploy.yaml
	@kubectl apply -f deploy/k8s/ingress.yaml
	@echo "Waiting for postgres to be ready..."
	@kubectl wait --for=condition=ready pod -l app=postgres -n polaris --timeout=120s || true
	@echo "Deployment complete! Run 'make k8s-migrate' to apply database migrations."

# Undeploy from Kubernetes
k8s-undeploy:
	@echo "Undeploying from Kubernetes..."
	@kubectl delete -f deploy/k8s/ingress.yaml --ignore-not-found=true
	@kubectl delete -f deploy/k8s/gateway-deploy.yaml --ignore-not-found=true
	@kubectl delete -f deploy/k8s/billing-deploy.yaml --ignore-not-found=true
	@kubectl delete -f deploy/k8s/usage-deploy.yaml --ignore-not-found=true
	@kubectl delete -f deploy/k8s/identity-deploy.yaml --ignore-not-found=true
	@kubectl delete -f deploy/k8s/nats.yaml --ignore-not-found=true
	@kubectl delete -f deploy/k8s/postgres.yaml --ignore-not-found=true
	@kubectl delete -f deploy/k8s/namespace.yaml --ignore-not-found=true
	@echo "Undeployment complete!"

# Run database migrations in Kubernetes
k8s-migrate:
	@echo "Running database migrations in Kubernetes..."
	@POD=$$(kubectl get pod -n polaris -l app=postgres -o jsonpath='{.items[0].metadata.name}'); \
	kubectl cp migrations/identity/001_init.sql polaris/$$POD:/tmp/identity_init.sql; \
	kubectl cp migrations/usage/001_init.sql polaris/$$POD:/tmp/usage_init.sql; \
	kubectl cp migrations/billing/001_init.sql polaris/$$POD:/tmp/billing_init.sql; \
	kubectl exec -n polaris $$POD -- psql -U polaris -d polaris -f /tmp/identity_init.sql; \
	kubectl exec -n polaris $$POD -- psql -U polaris -d polaris -f /tmp/usage_init.sql; \
	kubectl exec -n polaris $$POD -- psql -U polaris -d polaris -f /tmp/billing_init.sql
	@echo "Migrations complete!"

