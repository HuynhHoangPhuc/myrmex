# Deployment Guide

## Table of Contents

1. [Local Development Setup](#local-development-setup)
2. [Environment Variables](#environment-variables)
3. [Database Migrations](#database-migrations)
4. [Building Services](#building-services)
5. [Running Services](#running-services)
6. [Docker Compose Deployment](#docker-compose-deployment)
7. [Troubleshooting](#troubleshooting)
8. [Production Deployment](#production-deployment-future)

---

## Local Development Setup

### Prerequisites

- **Go**: 1.26+ (download from [golang.org](https://golang.org/dl))
- **Node.js**: 18+ (download from [nodejs.org](https://nodejs.org))
- **Docker & Docker Compose**: 3.8+ (download from [docker.com](https://www.docker.com))
- **PostgreSQL Client** (optional, for manual DB inspection): `psql` or pgAdmin
- **Protoc Tools** (if modifying .proto files):
  ```bash
  go install github.com/bufbuild/buf/cmd/buf@latest
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

### Step 1: Clone Repository

```bash
git clone https://github.com/yourusername/myrmex.git
cd myrmex
```

### Step 2: Start Infrastructure

```bash
# Start PostgreSQL, NATS, Redis via Docker Compose
make up

# Verify services are running
docker ps
# Should show: postgres, nats, redis containers

# Wait for PostgreSQL to be ready (health check)
docker exec myrmex-postgres-1 pg_isready -U myrmex
# Output: accepting connections
```

### Step 3: Configure Environment

Create `.env` file in project root (Git-ignored by default):

```bash
# Database
DATABASE_URL="postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable"

# NATS
NATS_URL="nats://localhost:4222"

# Core Service
CORE_JWT_SECRET="your-secret-key-min-32-chars-long!!"
CORE_HTTP_PORT=8000
CORE_GRPC_PORT=50051
CORE_LLM_PROVIDER="claude" # or "openai"
CORE_LLM_MODEL="claude-haiku-4-5-20251001"

# LLM API Key (add one)
CLAUDE_API_KEY="sk-ant-..." # If using Claude
# OPENAI_API_KEY="sk-..." # If using OpenAI

# HR Service
HR_GRPC_PORT=50052

# Subject Service
SUBJECT_GRPC_PORT=50053

# Timetable Service
TIMETABLE_GRPC_PORT=50054

# Frontend
VITE_API_URL="http://localhost:8000"
```

Load environment variables:

```bash
# Export to shell
export $(cat .env | xargs)

# Or use direnv (optional)
echo "export \$(cat .env | xargs)" > .envrc
direnv allow
```

### Step 4: Initialize Database

Run migrations for all services:

```bash
# Set DATABASE_URL
export DATABASE_URL="postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable"

# Run all migrations
make migrate

# Verify schema created
psql $DATABASE_URL -c "\dt"
# Should list tables: users, departments, teachers, subjects, etc.

# Check schema structure
psql $DATABASE_URL -c "\dn"
# Should list schemas: public, core, hr, subject, timetable
```

### Step 5: Generate Protobuf Code

```bash
# Generate Go code from .proto files
make proto

# Verify gen/go/ directory populated
ls -la gen/go/core/v1/
# Should show: *.pb.go files
```

### Step 6: Build All Services

```bash
# Compile all Go services
make build

# Or build individually
cd services/core && go build ./cmd/server && cd ../..
cd services/module-hr && go build ./cmd/server && cd ../..
cd services/module-subject && go build ./cmd/server && cd ../..
cd services/module-timetable && go build ./cmd/server && cd ../..
```

### Step 7: Run Services

Open 4 separate terminal tabs/windows and start each service:

**Terminal 1: Core Service**
```bash
cd services/core
go run ./cmd/server
# Output: Server listening on :8000 (HTTP) and :50051 (gRPC)
```

**Terminal 2: HR Service**
```bash
cd services/module-hr
go run ./cmd/server
# Output: Server listening on :50052 (gRPC)
```

**Terminal 3: Subject Service**
```bash
cd services/module-subject
go run ./cmd/server
# Output: Server listening on :50053 (gRPC)
```

**Terminal 4: Timetable Service**
```bash
cd services/module-timetable
go run ./cmd/server
# Output: Server listening on :50054 (gRPC)
```

### Step 8: Start Frontend

Open another terminal:

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
# Output: Local: http://localhost:3000
```

### Step 9: Verify Setup

1. **HTTP Gateway**: `curl http://localhost:8000/api/health`
   - Should return `200 OK`

2. **Frontend**: Open http://localhost:3000
   - Should show login page

3. **Register & Login**:
   - Click "Register" → Create account
   - Login with email/password
   - Should see dashboard

---

## Environment Variables

### Core Service (`services/core/config/local.yaml`)

```yaml
server:
  http_port: 8000          # HTTP gateway port
  grpc_port: 50051         # gRPC server port
  request_timeout: 30s     # Timeout for gRPC calls

auth:
  jwt_secret: "your-secret-key-min-32-chars!!"
  access_token_ttl: 15m    # Access token lifetime
  refresh_token_ttl: 7d    # Refresh token lifetime

llm:
  provider: "claude"                                    # "claude" or "openai"
  model: "claude-haiku-4-5-20251001"                  # Model name
  api_key: "${CLAUDE_API_KEY}"                         # From env var
  timeout: 30s                                         # Request timeout

database:
  url: "postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable"
  max_connections: 25

nats:
  url: "nats://localhost:4222"
  connection_timeout: 5s

logging:
  level: "info"            # "debug", "info", "warn", "error"
  format: "json"           # "json" or "console"
```

### Module Services (HR, Subject, Timetable)

Similar structure:

```yaml
server:
  grpc_port: 50052         # Different for each module

database:
  url: "postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable"
  max_connections: 20

nats:
  url: "nats://localhost:4222"

logging:
  level: "info"
```

### Frontend (`frontend/.env.local`)

```env
VITE_API_URL=http://localhost:8000
VITE_CHAT_WS_URL=ws://localhost:8000/ws/chat
```

---

## Database Migrations

### Running Migrations

```bash
# Set DATABASE_URL
export DATABASE_URL="postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable"

# Run migrations for all services
make migrate

# Or manually for specific service
cd services/core
go tool goose -dir migrations postgres "$DATABASE_URL" up
cd ../..
```

### Creating Migrations

```bash
# Create new migration for module-hr
cd services/module-hr
go tool goose create add_teacher_specializations sql
# Creates: migrations/NNN_add_teacher_specializations.sql

# Edit migration file
cat migrations/NNN_add_teacher_specializations.sql
# Add up/down SQL

# Test migration
go tool goose -dir migrations postgres "$DATABASE_URL" up

# Rollback if needed
go tool goose -dir migrations postgres "$DATABASE_URL" down
```

### Viewing Migration Status

```bash
# List applied migrations
go tool goose -dir migrations postgres "$DATABASE_URL" status

# Rollback to previous version
go tool goose -dir migrations postgres "$DATABASE_URL" down
```

---

## Building Services

### Local Development Build

```bash
# Build all services
make build

# Build single service
cd services/core && go build ./cmd/server && cd ../..
```

### Production Build (Docker)

```bash
# Build Docker images
docker build -t myrmex-core:latest ./services/core
docker build -t myrmex-hr:latest ./services/module-hr
docker build -t myrmex-subject:latest ./services/module-subject
docker build -t myrmex-timetable:latest ./services/module-timetable
```

### Cross-Platform Build

```bash
# Build for Linux (for Docker)
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o server ./cmd/server

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o server.exe ./cmd/server
```

---

## Running Services

### Development Mode (Hot Reload)

Using `air` for hot reload:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with auto-reload
cd services/core
air

# Modify code → service restarts automatically
```

### Production Mode

```bash
# Terminal 1: Core
./services/core/server &

# Terminal 2: HR
./services/module-hr/server &

# Terminal 3: Subject
./services/module-subject/server &

# Terminal 4: Timetable
./services/module-timetable/server &

# Check if running
ps aux | grep server
```

### Systemd Service (Linux)

Create `/etc/systemd/system/myrmex-core.service`:

```ini
[Unit]
Description=Myrmex Core Service
After=network.target postgresql.service nats.service

[Service]
Type=simple
User=myrmex
WorkingDirectory=/opt/myrmex
ExecStart=/opt/myrmex/bin/core-server
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

Environment="DATABASE_URL=postgres://myrmex:myrmex_dev@postgres:5432/myrmex?sslmode=disable"
Environment="NATS_URL=nats://nats:4222"
Environment="CORE_JWT_SECRET=your-secret-key"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable myrmex-core
sudo systemctl start myrmex-core
sudo systemctl status myrmex-core
sudo journalctl -u myrmex-core -f
```

---

## Docker Compose Deployment

### Single Command Start

```bash
# Start all services (infrastructure)
make up

# Verify
docker ps
docker compose -f deploy/docker/compose.yml logs

# Stop all services
make down
```

### Docker Compose File Structure

```yaml
# deploy/docker/compose.yml
services:
  postgres:16-alpine
    # PostgreSQL database
    ports: 5432:5432
    volumes: postgres_data

  nats:2.10-alpine
    # Message bus
    ports: 4222:4222, 8222:8222
    volumes: nats_data

  redis:7-alpine
    # Cache (reserved for future use)
    ports: 6379:6379

volumes:
  postgres_data
  nats_data
```

### Adding Services to Docker Compose

To run Go services in Docker:

```yaml
# deploy/docker/compose.yml
services:
  core:
    build:
      context: .
      dockerfile: services/core/Dockerfile
    ports: 8000:8000, 50051:50051
    environment:
      DATABASE_URL: postgres://myrmex:myrmex_dev@postgres:5432/myrmex?sslmode=disable
      NATS_URL: nats://nats:4222
    depends_on:
      postgres:
        condition: service_healthy
      nats:
        condition: service_started

  module-hr:
    build:
      context: .
      dockerfile: services/module-hr/Dockerfile
    ports: 50052:50052
    environment:
      DATABASE_URL: postgres://myrmex:myrmex_dev@postgres:5432/myrmex?sslmode=disable
      NATS_URL: nats://nats:4222
    depends_on: [postgres, nats]

  # Similar for module-subject, module-timetable

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports: 3000:3000
    environment:
      VITE_API_URL: http://core:8000
    depends_on: [core]
```

Start all services with:

```bash
docker compose -f deploy/docker/compose.yml up -d

# View logs
docker compose -f deploy/docker/compose.yml logs -f

# Stop all
docker compose -f deploy/docker/compose.yml down
```

---

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8000

# Kill process
kill -9 <PID>

# Or use alternative port
CORE_HTTP_PORT=8001 go run ./cmd/server
```

### Database Connection Error

```bash
# Test PostgreSQL connection
psql postgres://myrmex:myrmex_dev@localhost:5432/myrmex

# Check PostgreSQL is running
docker ps | grep postgres

# View PostgreSQL logs
docker logs <postgres-container-id>

# Restart PostgreSQL
make down
make up
```

### NATS Connection Error

```bash
# Test NATS connection
curl http://localhost:8222/varz

# Check NATS is running
docker ps | grep nats

# View NATS logs
docker logs <nats-container-id>

# Check firewall
telnet localhost 4222
```

### Proto Changes Not Reflected

```bash
# Regenerate Go code
make proto

# Rebuild service
cd services/core && go build ./cmd/server && cd ../..

# Restart service
go run ./cmd/server
```

### Frontend API 404 Errors

```bash
# Verify API gateway is running
curl http://localhost:8000/api/health

# Check CORS headers
curl -H "Origin: http://localhost:3000" http://localhost:8000/api/health

# Check frontend API_URL
cat frontend/.env.local | grep VITE_API_URL
```

### Migration Rollback Issues

```bash
# Check migration status
go tool goose -dir migrations postgres "$DATABASE_URL" status

# Rollback one step
go tool goose -dir migrations postgres "$DATABASE_URL" down

# Rollback all
go tool goose -dir migrations postgres "$DATABASE_URL" reset

# Re-run migrations
go tool goose -dir migrations postgres "$DATABASE_URL" up
```

### Service Health Checks

```bash
# Core HTTP gateway
curl http://localhost:8000/api/health

# Core gRPC health
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# HR gRPC health
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check

# NATS info
curl http://localhost:8222/varz

# Database connectivity
psql "$DATABASE_URL" -c "SELECT 1"
```

---

## Production Deployment (Future)

### Requirements
- Kubernetes 1.24+ or Docker Swarm
- PostgreSQL 16 with replication (master-slave)
- NATS JetStream cluster (3+ nodes)
- Redis cluster (optional)
- Monitoring: Prometheus + Grafana
- Log aggregation: ELK Stack or Loki
- Load balancer: Nginx or Traefik

### Deployment Strategy
1. **Container Registry**: Push images to Docker Hub / ECR / GCR
2. **Orchestration**: Deploy via Helm charts (Kubernetes) or Docker Swarm
3. **Database**: RDS (AWS) / Cloud SQL (GCP) / managed PostgreSQL
4. **NATS**: JetStream cluster with 3+ nodes
5. **Monitoring**: Prometheus scrape gRPC metrics
6. **Logging**: Fluent-bit → Elasticsearch
7. **Secrets**: HashiCorp Vault / AWS Secrets Manager
8. **CI/CD**: GitHub Actions → Docker build → Kubernetes deploy

### High Availability Setup
- **Load Balancer**: Nginx/Traefik with health checks
- **Database Replication**: PostgreSQL streaming replication (RTO: 1min, RPO: 5s)
- **NATS HA**: 3-node cluster with persistent storage
- **Service Replicas**: Minimum 2 per service (auto-scaling 1-10)
- **Health Checks**: Liveness + readiness probes per service

### Scaling Strategy
- **Horizontal**: Auto-scale services based on CPU/memory
- **Vertical**: Increase pod resources for data-intensive services (CSP solver)
- **Database**: Connection pooling (PgBouncer), read replicas for analytics
- **Cache**: Redis cache-aside for frequently accessed data
- **CDN**: Static assets (frontend) via CloudFront/Cloudflare

---

## Quick Reference

| Task | Command |
|------|---------|
| **Start infrastructure** | `make up` |
| **Run migrations** | `make migrate` |
| **Build services** | `make build` |
| **Start Core** | `cd services/core && go run ./cmd/server` |
| **Start HR** | `cd services/module-hr && go run ./cmd/server` |
| **Start frontend** | `cd frontend && npm run dev` |
| **Generate protos** | `make proto` |
| **Run tests** | `make test` |
| **Lint code** | `make lint` |
| **Stop infrastructure** | `make down` |
| **View logs** | `docker logs <container>` |
| **Reset database** | `make down && make up && make migrate` |

---

## Deployment Checklist

- [ ] Database credentials updated for production
- [ ] JWT secret changed to strong random value (min 32 chars)
- [ ] LLM API key (Claude/OpenAI) configured
- [ ] All services passing unit tests
- [ ] Load tests completed (1000+ concurrent users)
- [ ] Logging configured (level: info or warn)
- [ ] Monitoring enabled (Prometheus, Grafana)
- [ ] Backup strategy in place
- [ ] Disaster recovery plan documented
- [ ] Security audit completed
- [ ] HTTPS/TLS certificates configured
- [ ] Rate limiting enabled
- [ ] CORS origins configured for production domain
- [ ] Database connection pooling tuned
- [ ] NATS JetStream persistence enabled
- [ ] Docker images tagged with version
- [ ] Rollback plan documented
- [ ] Post-deployment health checks passing
