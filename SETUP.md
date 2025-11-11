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

## Kubernetes Setup

### Prerequisites

- Docker installed and running
- **kubectl installed** (see installation below)
- A local Kubernetes cluster (choose one):

### Installing kubectl

If you don't have kubectl installed:

```bash
# Download kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Make it executable and install to ~/.local/bin (no sudo needed)
chmod +x kubectl
mkdir -p ~/.local/bin
mv kubectl ~/.local/bin/

# Add to PATH (for current session)
export PATH="$HOME/.local/bin:$PATH"

# Add to PATH permanently (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify installation
kubectl version --client
```

**Note:** If you prefer to install system-wide (requires sudo):
```bash
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

- A local Kubernetes cluster (choose one):
  - **kind** (recommended for CI/CD): `kind create cluster`
  - **k3d** (lightweight, fast): `k3d cluster create polaris`
  - **minikube**: `minikube start`
  - **Docker Desktop**: Enable Kubernetes in Settings

### Option 1: Using kind (Kubernetes in Docker)

1. **Install kind:**
   ```bash
   # On Linux/macOS
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind
   
   # Or using package managers:
   # macOS: brew install kind
   # Ubuntu: sudo snap install kind
   ```

2. **Create a cluster:**
   ```bash
   kind create cluster --name polaris
   ```

3. **Verify cluster:**
   ```bash
   kubectl cluster-info --context kind-polaris
   ```

### Option 2: Using k3d (k3s in Docker)

1. **Install k3d:**
   ```bash
   # Using script
   curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
   
   # Or using package managers:
   # macOS: brew install k3d
   # Ubuntu: curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
   ```

2. **Create a cluster:**
   ```bash
   k3d cluster create polaris
   ```

3. **Verify cluster:**
   ```bash
   kubectl cluster-info
   ```

### Option 3: Using minikube

1. **Install minikube:**
   ```bash
   # Linux
   curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
   sudo install minikube-linux-amd64 /usr/local/bin/minikube
   
   # macOS
   brew install minikube
   ```

2. **Start minikube:**
   ```bash
   minikube start
   ```

3. **Verify cluster:**
   ```bash
   kubectl cluster-info
   ```

### Option 4: Using Docker Desktop

1. Open Docker Desktop
2. Go to Settings → Kubernetes
3. Enable Kubernetes
4. Click "Apply & Restart"
5. Wait for Kubernetes to start

### Deploying to Kubernetes

Once you have a cluster running:

1. **Build Docker images:**
   ```bash
   make k8s-build
   ```

2. **Load images into cluster** (for kind/k3d/minikube):
   ```bash
   make k8s-load
   ```
   This automatically detects your cluster type and loads the images.

3. **Deploy all services:**
   ```bash
   make k8s-deploy
   ```

4. **Run database migrations:**
   ```bash
   make k8s-migrate
   ```

5. **Verify deployment:**
   ```bash
   kubectl get pods -n polaris
   kubectl get services -n polaris
   ```

### Accessing the Services

#### Option A: Port Forwarding (Quick Testing)

```bash
# Forward API Gateway port
kubectl port-forward -n polaris service/api-gateway 8080:80

# Then access:
# - GraphQL Playground: http://localhost:8080/playground
# - GraphQL Endpoint: http://localhost:8080/query
```

#### Option B: Using Ingress (Production-like)

The project includes an Ingress configuration. To use it:

1. **Install an Ingress Controller** (if not already installed):
   ```bash
   # For kind
   kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
   
   # For k3d (Traefik is included by default)
   # No action needed
   
   # For minikube
   minikube addons enable ingress
   ```

2. **Update /etc/hosts** (or equivalent):
   ```bash
   # For kind/minikube, get the ingress IP:
   kubectl get ingress -n polaris
   
   # Add to /etc/hosts:
   # <INGRESS_IP> polaris.local
   
   # For k3d, Traefik runs on port 80, so:
   # 127.0.0.1 polaris.local
   ```

3. **Access the service:**
   ```bash
   curl http://polaris.local/query
   ```

### Useful Commands

```bash
# View all resources
kubectl get all -n polaris

# View logs
kubectl logs -n polaris deployment/api-gateway
kubectl logs -n polaris deployment/identity-svc
kubectl logs -n polaris deployment/usage-svc
kubectl logs -n polaris deployment/billing-svc

# Watch pods
kubectl get pods -n polaris -w

# Describe a pod
kubectl describe pod -n polaris <pod-name>

# Execute commands in a pod
kubectl exec -it -n polaris deployment/postgres -- psql -U polaris -d polaris

# Undeploy everything
make k8s-undeploy

# Delete the cluster (kind)
kind delete cluster --name polaris

# Delete the cluster (k3d)
k3d cluster delete polaris

# Stop minikube
minikube stop
```

### Troubleshooting

**Images not found:**
- Make sure you ran `make k8s-build` and `make k8s-load`
- For Docker Desktop, images should be available automatically
- Check with: `kubectl describe pod -n polaris <pod-name>`

**Postgres not ready:**
- Check logs: `kubectl logs -n polaris deployment/postgres`
- Verify PVC: `kubectl get pvc -n polaris`
- Check pod status: `kubectl get pods -n polaris`

**Services can't connect:**
- Verify services exist: `kubectl get svc -n polaris`
- Check DNS resolution: `kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup postgres.polaris`

**Ingress not working:**
- Verify ingress controller is running: `kubectl get pods -n ingress-nginx` (or appropriate namespace)
- Check ingress status: `kubectl describe ingress -n polaris`

