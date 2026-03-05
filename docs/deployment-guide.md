# Deployment Guide

## Table of Contents

1. [Docker Demo (Recommended)](#docker-demo-recommended)
2. [Local Development Setup](#local-development-setup)
3. [Environment Variables](#environment-variables)
4. [Database Migrations](#database-migrations)
5. [Building Services](#building-services)
6. [Running Services](#running-services)
7. [Docker Compose Deployment](#docker-compose-deployment)
8. [GCP Cloud Run Deployment (Production)](#gcp-cloud-run-deployment-production)
9. [Staging Environment (Phase 6)](#staging-environment-phase-6)
10. [HCMUS Data Migration (Phase 6)](#hcmus-data-migration-phase-6)
11. [Troubleshooting](#troubleshooting)

---

## Docker Demo (Recommended)

### Quickstart: One-Command Setup

The fastest way to experience Myrmex is via Docker:

```bash
# Clone the repository
git clone https://github.com/yourusername/myrmex.git
cd myrmex

# (Optional) Set up LLM API key for AI chat feature
cp .env.example .env
# Edit .env and add your Claude/OpenAI API key if desired

# Start the entire system with one command
make demo

# Open in browser
# Frontend: http://localhost:3000
# API Gateway: http://localhost:8080
```

That's it! All services (core, HR, Subject, Timetable, Student, Analytics), databases (PostgreSQL, NATS, Redis), and frontend start automatically. Migrations run, seed data is loaded.

### Common Commands

```bash
# View logs from all services
make demo-logs

# Stop all services (preserves data)
make demo-down

# Stop and reset database (wipe all data, start fresh)
make demo-reset

# View running containers
docker ps
```

### What Gets Started

The `make demo` command starts:
- **PostgreSQL 16**: Database (port 5432)
- **NATS 2.10**: Message bus (port 4222)
- **Redis 7**: Cache (port 6379)
- **Core Service**: HTTP gateway + JWT auth (port 8080)
- **Module-HR**: Department & teacher management (port 50052)
- **Module-Subject**: Subject & prerequisite management (port 50053)
- **Module-Timetable**: Schedule generation & management (port 50054)
- **Module-Student**: Student CRUD foundation (port 50055)
- **Module-Analytics**: Analytics dashboard + query endpoints (port 8055)
- **Frontend**: React UI served via nginx (port 3000)

All services communicate via Docker network. Migrations and seed data run automatically. Analytics module subscribes to NATS events for real-time ETL.

### Troubleshooting Docker Demo

**Ports already in use:**
```bash
# Free up ports or change in docker-compose
# Default: 3000 (frontend), 8080 (API), 5432 (DB)
docker ps  # See what's running
docker kill <container-id>
```

**Database issues:**
```bash
# Reset database completely
make demo-reset

# Check database logs
docker compose -f deploy/docker/compose.yml logs postgres
```

**Inspect services:**
```bash
# Check if API is responding
curl http://localhost:8080/api/health

# List all running containers
docker compose -f deploy/docker/compose.yml ps
```

### Next Steps After Demo

1. **Register a user**: Visit http://localhost:3000 and click "Register"
2. **Create departments, teachers, subjects**: Use the UI
3. **Generate a schedule**: Add semesters, rooms, and time slots, then use the CSP solver
4. **Try AI chat** (requires LLM_API_KEY): Click chat icon to ask the agent to create data or generate schedules

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
CORE_HTTP_PORT=8080
CORE_GRPC_PORT=50051
CORE_LLM_PROVIDER="claude"         # "openai" | "claude" | "gemini" | "mock"
CORE_LLM_MODEL="claude-haiku-4-5-20251001"

# LLM API Key (add one based on provider)
CLAUDE_API_KEY="sk-ant-..."         # If using Claude
# OPENAI_API_KEY="sk-..."           # If using OpenAI
# GEMINI_API_KEY="AIzaSy..."        # If using Gemini (free tier)

# OAuth/SSO (Optional)
GOOGLE_CLIENT_ID="xxxxx.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="GOCSPX-xxxxx"
MICROSOFT_CLIENT_ID="xxxxx-xxxxx-xxxxx"
MICROSOFT_CLIENT_SECRET="xxxxx~xxxxx"
MICROSOFT_TENANT_ID="xxxxx-xxxxx-xxxxx"

# Notifications (Optional - Phase 4.4)
SMTP_HOST="smtp.sendgrid.net"
SMTP_PORT=587
SMTP_USERNAME="apikey"
SMTP_PASSWORD="SG.xxxxx"
NOTIFICATION_FROM_EMAIL="noreply@myrmex.local"

# HR Service
HR_GRPC_PORT=50052

# Subject Service
SUBJECT_GRPC_PORT=50053

# Timetable Service
TIMETABLE_GRPC_PORT=50054

# Student Service
STUDENT_GRPC_PORT=50055

# Analytics Service
ANALYTICS_HTTP_ADDR="http://localhost:8055"

# Frontend
VITE_API_URL="http://localhost:8080"
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
# Should list schemas: public, core, hr, subject, timetable, student, analytics
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
cd services/module-student && go build ./cmd/server && cd ../..
cd services/module-analytics && go build ./cmd/server && cd ../..
```

### Step 7: Run Services

Open 6 separate terminal tabs/windows and start each service:

**Terminal 1: Core Service**
```bash
cd services/core
go run ./cmd/server
# Output: Server listening on :8080 (HTTP) and :50051 (gRPC)
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

**Terminal 5: Student Service**
```bash
cd services/module-student
go run ./cmd/server
# Output: Server listening on :50055 (gRPC)
```

**Terminal 6: Analytics Service**
```bash
cd services/module-analytics
go run ./cmd/server
# Output: Server listening on :8055 (HTTP)
# Consumes NATS events for ETL
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

1. **HTTP Gateway**: `curl http://localhost:8080/api/health`
   - Should return `200 OK`

2. **Frontend**: Open http://localhost:3000
   - Should show login page

3. **Register & Login**:
   - Click "Register" → Create account
   - Login with email/password
   - Should see dashboard

---

## Environment Variables

### Core Service
```bash
CORE_JWT_SECRET="your-32-char-min-secret!!"
CORE_HTTP_PORT=8080
CORE_GRPC_PORT=50051
CORE_LLM_PROVIDER="claude"
CORE_LLM_MODEL="claude-haiku-4-5-20251001"
CORE_LLM_API_KEY="sk-ant-..."
GOOGLE_CLIENT_ID="xxxxx.apps.googleusercontent.com"  # Optional
MICROSOFT_CLIENT_ID="xxxxx"                           # Optional
SMTP_HOST="smtp.sendgrid.net"                         # Optional
```

### Module Services
- `HR_GRPC_PORT=50052`
- `SUBJECT_GRPC_PORT=50053`
- `TIMETABLE_GRPC_PORT=50054`
- `STUDENT_GRPC_PORT=50055`
- `ANALYTICS_HTTP_ADDR=http://localhost:8055`
- `DATABASE_URL=postgres://user:pass@host:5432/myrmex`
- `NATS_URL=nats://localhost:4222`

---

## Database Migrations

### Running Migrations

```bash
export DATABASE_URL="postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable"
make migrate  # Runs all service migrations
```

### Creating Migrations

```bash
cd services/module-hr
go tool goose create migration_name sql  # Creates migration file
# Edit migration with up/down SQL
go tool goose -dir migrations postgres "$DATABASE_URL" up
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
docker build -t myrmex-student:latest ./services/module-student
docker build -t myrmex-analytics:latest ./services/module-analytics
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

### Student Gateway Wiring

The core service now expects a student gRPC address in both local and Docker setups:

- Local config key: `student.grpc_addr` (default `localhost:50055`)
- Docker env override: `STUDENT_GRPC_ADDR=module-student:50055`
- Exposed gateway surface: admin-only `/api/students` CRUD routes proxied from core to module-student

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

# Terminal 5: Student
./services/module-student/server &

# Terminal 6: Analytics
./services/module-analytics/server &

# Check if running
ps aux | grep server
```

### Systemd Service (Linux Production)

Create `/etc/systemd/system/myrmex-core.service` with proper environment variables and enable via systemctl.

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

### Docker Compose Structure

`deploy/docker/compose.yml` includes:
- Infrastructure: postgres, nats, redis
- Services: core, module-hr, module-subject, module-timetable, module-student, module-analytics
- Frontend: nginx-based React UI

All services use environment variable overrides for configuration (DATABASE_URL, NATS_URL, gRPC addresses).

```bash
# Start all services
docker compose -f deploy/docker/compose.yml up -d

# View logs
docker compose -f deploy/docker/compose.yml logs -f

# Stop all
docker compose -f deploy/docker/compose.yml down
```

---

## GCP Cloud Run Deployment (Production)

### Architecture Overview

Production deployment uses Google Cloud Platform with:
- **Cloud SQL**: PostgreSQL 16 (managed, SSL enforced)
- **Memorystore**: Redis 7 (managed cache)
- **Cloud Run**: Serverless compute (7 services with auto-scaling)
- **Pub/Sub**: Managed message broker (replaces NATS)
- **Artifact Registry**: Docker image repository
- **Secret Manager**: Centralized secrets
- **VPC**: Network isolation + Cloud NAT
- **Cloud Monitoring**: Observability + alerts

### Prerequisites

- GCP project with billing enabled
- `gcloud` CLI installed and authenticated
- Terraform 1.5+ installed
- GitHub repository with secrets configured (for WIF auth)

### Step 1: Terraform Setup

```bash
cd deploy/terraform
terraform init
terraform plan -var="project_id=YOUR_GCP_PROJECT" -var="region=us-central1"
terraform apply -var="project_id=YOUR_GCP_PROJECT" -var="region=us-central1"
```

Creates: VPC, Cloud NAT, Cloud SQL (PostgreSQL 16), Memorystore (Redis 7), Artifact Registry, Cloud Run services, Pub/Sub topics, monitoring.

### Step 2: Populate Secret Manager

```bash
gcloud secrets create myrmex-db-password --data-file=- <<< "your-secure-password"
gcloud secrets create myrmex-jwt-secret --data-file=- <<< "your-32-char-min-jwt-secret"
gcloud secrets create myrmex-claude-api-key --data-file=- <<< "sk-ant-xxxxx"
gcloud secrets create myrmex-google-client-id --data-file=- <<< "xxxxx.apps.googleusercontent.com"
gcloud secrets create myrmex-google-client-secret --data-file=- <<< "GOCSPX-xxxxx"
gcloud secrets create myrmex-microsoft-client-id --data-file=- <<< "xxxxx"
gcloud secrets create myrmex-microsoft-client-secret --data-file=- <<< "xxxxx~xxxxx"
gcloud secrets create myrmex-microsoft-tenant-id --data-file=- <<< "xxxxx"
```

Grant Cloud Run service account access:

```bash
PROJECT_ID=$(gcloud config get-value project)
SERVICE_ACCOUNT="myrmex-cloud-run@${PROJECT_ID}.iam.gserviceaccount.com"

# Grant secret accessor role
gcloud secrets add-iam-policy-binding myrmex-db-password \
  --member="serviceAccount:${SERVICE_ACCOUNT}" \
  --role="roles/secretmanager.secretAccessor"
# ... repeat for other secrets
```

### Step 3: Workload Identity Federation (WIF) Setup

Enable GitHub Actions to deploy without long-lived secrets:

```bash
PROJECT_ID=$(gcloud config get-value project)

# Create OIDC workload identity pool
gcloud iam workload-identity-pools create github --project="${PROJECT_ID}" --location=global
gcloud iam workload-identity-pools providers create-oidc github \
  --project="${PROJECT_ID}" --location=global --workload-identity-pool=github \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor"

# Create service account
gcloud iam service-accounts create github-actions --project="${PROJECT_ID}"
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/run.admin"
```

Add to GitHub Actions `.env`: `GCP_PROJECT_ID`, `GCP_WORKLOAD_IDENTITY_PROVIDER`, `GCP_SERVICE_ACCOUNT`

### Step 4: Run Database Migrations

```bash
# Create Cloud Run Job for migrations
gcloud run jobs create myrmex-migrate \
  --image="us-central1-docker.pkg.dev/${PROJECT_ID}/myrmex/myrmex-core:latest" \
  --task-count=1 \
  --set-env-vars="DATABASE_URL=postgres://...@/myrmex" \
  --service-account="myrmex-cloud-run@${PROJECT_ID}.iam.gserviceaccount.com"

# Execute migrations
gcloud run jobs execute myrmex-migrate

# Monitor status
gcloud run jobs describe myrmex-migrate
```

### Step 5: Deploy Services

Automated via `.github/workflows/deploy.yml`: WIF auth → Docker build/push → migrations → Cloud Run deployment → smoke tests.

Manual deployment:
```bash
docker build -t us-central1-docker.pkg.dev/${PROJECT_ID}/myrmex/myrmex-core:latest ./services/core
docker push us-central1-docker.pkg.dev/${PROJECT_ID}/myrmex/myrmex-core:latest
gcloud run deploy myrmex-core \
  --image="us-central1-docker.pkg.dev/${PROJECT_ID}/myrmex/myrmex-core:latest" \
  --region=us-central1 --memory=512Mi --cpu=1 --max-instances=10 \
  --service-account="myrmex-cloud-run@${PROJECT_ID}.iam.gserviceaccount.com"
```

### Step 6: Verify Deployment

```bash
curl https://myrmex-core-xxxxx.run.app/health
gcloud run services list --region=us-central1
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=myrmex-core" --limit=50
```

### Load Testing

Verify deployment with k6 scripts:

```bash
k6 run deploy/load-tests/auth-flow.js        # 100 VUs auth flow
k6 run deploy/load-tests/api-crud.js         # 200 VUs CRUD ops
k6 run deploy/load-tests/mixed-workload.js   # 500 VUs realistic traffic
```

### Production Checklist

- [ ] Secrets in Secret Manager, WIF configured
- [ ] Terraform applied + migrations executed
- [ ] All 7 Cloud Run services deployed + health checks passing
- [ ] Cloud Monitoring alerts configured (email + Slack channels set in `alert_email`/`alert_slack_webhook_url`)
- [ ] Load tests pass (p95 <500ms)
- [ ] SSL + custom domain configured (`frontend_domain`/`api_domain` in tfvars)
- [ ] Backups enabled (Cloud SQL PITR + 7-day retention active)
- [ ] SENTRY_DSN populated in Secret Manager
- [ ] All services min_instances=1 (no cold starts for 200+ users)
- [ ] DB connections confirmed below 150 (pool: core=30, student/notif=20, others=15)

---

## Staging Environment (Phase 6)

### Overview

Staging mirrors production in the same GCP project with `staging-` prefixed service names and a separate Cloud SQL instance (`myrmex-postgres-staging`).

### CD Pipeline (Tag-Based Releases)

| Trigger | Action |
|---------|--------|
| Push to `main` | Auto-deploy all services to `staging-*` Cloud Run services |
| Push git tag `v*` | Deploy to production Cloud Run services |
| `workflow_dispatch` | Manual deploy to either environment |

### Provision Staging

```bash
# Set staging secrets
gcloud secrets versions add DATABASE_URL_STAGING --data-file=- <<< "postgres://..."
gcloud secrets versions add JWT_SECRET_STAGING --data-file=- <<< "staging-secret-32chars"

# Terraform creates staging Cloud SQL + Cloud Run services automatically
# (staging-cloud-run.tf provisions staging-core, staging-module-*, staging-frontend)
terraform apply
```

### Seed / Reset Staging

```bash
# Initial seed
./deploy/scripts/seed-staging.sh "postgres://user:pass@staging-host/myrmex"

# Full wipe + re-seed (prompts for confirmation)
./deploy/scripts/reset-staging.sh "postgres://..." myrmex-gcp-project asia-southeast1
```

---

## HCMUS Data Migration (Phase 6)

### Pipeline Overview

```
HCMUS Excel/CSV → Transform scripts → Validation → API import → Verification
```

### Step 1: Transform Source Data

```bash
pip install pandas openpyxl
python3 deploy/migration/transform-teachers.py --input hcmus-teachers.xlsx --output data/teachers.csv
python3 deploy/migration/transform-students.py --input hcmus-students.xlsx --output data/students.csv
```

### Step 2: Pre-flight Validation

```bash
python3 deploy/migration/validate-data.py --input data/ --report validation-report.md
# Must exit 0 (PASS) before proceeding
```

### Step 3: Bootstrap Admin

```bash
ADMIN_PASSWORD="SecurePass@2026" \
  ./deploy/migration/bootstrap-admin.sh "https://api.myrmex.hcmus.edu.vn" "postgres://..."
# Exports ADMIN_TOKEN on success
```

### Step 4: Import Data

```bash
./deploy/migration/import-data.sh "https://api.myrmex.hcmus.edu.vn" "$ADMIN_TOKEN" data/
```

### Step 5: Verify

```bash
./deploy/migration/verify-import.sh "postgres://..."
# Must exit 0 (PASS) before declaring go-live
```

### Rollback (Emergency)

```bash
./deploy/migration/rollback.sh "postgres://..."
# Prompts: type 'ROLLBACK' to confirm schema wipe
```

See `deploy/migration/` for full runbook and script documentation.

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Port already in use | `lsof -i :8080` to find process, `kill -9 <PID>` to kill |
| Database connection error | Check PostgreSQL running: `docker ps \| grep postgres`; verify DATABASE_URL |
| NATS connection error | Check NATS: `curl http://localhost:8222/varz` |
| Proto changes not reflected | Run `make proto` then rebuild: `go build ./cmd/server` |
| Frontend API 404 | Verify gateway: `curl http://localhost:8080/api/health` |
| Migration issues | Check status: `go tool goose -dir migrations postgres "$DATABASE_URL" status` |

### Health Checks

```bash
curl http://localhost:8080/api/health          # HTTP gateway
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check  # Core gRPC
curl http://localhost:8222/varz                 # NATS
psql "$DATABASE_URL" -c "SELECT 1"              # Database
```

---

## Production Deployment (Future)

### Key Requirements
- Kubernetes 1.24+ with Helm
- PostgreSQL 16 with replication
- NATS JetStream cluster (3+ nodes)
- Monitoring (Prometheus) + Logging (ELK/Loki)
- Secrets management (Vault / AWS Secrets Manager)

### HA & Scaling
- Load balancer (Nginx/Traefik) with health checks
- PostgreSQL streaming replication
- NATS 3-node HA cluster
- Auto-scaling services (min 2, max 10 per service)
- Redis cache-aside for frequently accessed data

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
