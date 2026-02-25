# Myrmex: Agent-First Modular ERP System

**Myrmex** is an open-source, agent-first ERP system built with Go microservices. Designed for educational institutions, it enables intelligent scheduling, resource management, and AI-powered operations.

## Overview

- **Type**: Modular microservice ERP (like Odoo, but Go-native)
- **AI Integration**: ChatGPT/Claude-powered agent for conversational operations
- **MVP**: University faculty management (HR, Subjects, Timetable modules)
- **Tech Stack**: Go 1.26 + gRPC + NATS JetStream + PostgreSQL + React

## Quick Start

### Docker Demo (Recommended)

Requires only Docker & Docker Compose.

```bash
# Optional: set LLM_API_KEY for AI chat feature
cp .env.example .env
# edit .env and set LLM_API_KEY=your-key

# Start everything (builds images, runs migrations, seeds demo data)
make demo

# Open http://localhost:3000
```

To stop: `make demo-down` | To wipe data and restart fresh: `make demo-reset`

### Local Development Setup

```bash
# Prerequisites: Go 1.26+, Node.js 18+, Docker

# Start infrastructure (PostgreSQL, NATS, Redis)
make up

# Run database migrations
export DATABASE_URL="postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable"
make migrate

# Generate protobuf code (if .proto files changed)
make proto

# Build all services
make build

# Start services in separate terminals
cd services/core && go run ./cmd/server
cd services/module-hr && go run ./cmd/server
cd services/module-subject && go run ./cmd/server
cd services/module-timetable && go run ./cmd/server

# In another terminal, start frontend (Vite dev server proxies /api and /ws)
cd frontend && npm install && npm run dev
```

### Accessing the System
- **Frontend**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **gRPC Services**: localhost:50051-50054
- **NATS JetStream**: localhost:4222
- **PostgreSQL**: localhost:5432 (user: myrmex, pass: myrmex_dev)

## Project Structure

```
myrmex/
├── services/              # Go microservices
│   ├── core/             # HTTP gateway + auth + AI chat
│   ├── module-hr/        # Department & teacher management
│   ├── module-subject/   # Subject & prerequisites (DAG)
│   └── module-timetable/ # Schedule generation (CSP solver)
├── pkg/                  # Shared packages (logger, config, eventstore, nats)
├── proto/                # Protobuf definitions
├── frontend/             # React + TypeScript + TanStack
├── deploy/               # Docker Compose configuration
├── docs/                 # Comprehensive documentation
└── Makefile              # Build targets
```

## Key Features

### Backend
- **Modular Architecture**: Each service is independent; schemas, migrations, repositories per module
- **Clean Architecture**: Domain layer → Application (CQRS) → Infrastructure → Interface (gRPC/HTTP)
- **Event Sourcing**: All write operations captured as events in PostgreSQL event store
- **gRPC**: Type-safe inter-service communication via protobuf
- **Async Messaging**: NATS JetStream for event publishing & consumption
- **Authentication**: JWT (access + refresh tokens, 15min + 7days)

### AI Agent Features
- **Multi-LLM Support**: Claude Haiku 4.5 (default) or OpenAI (configurable)
- **Tool Registry**: Dynamic registration of domain operations as callable tools
- **Conversational Operations**: Chat-driven scheduling, teacher assignments, subject creation
- **WebSocket**: Real-time streaming responses

### Frontend
- **Modern Stack**: React 19 + TypeScript 5.7 + TanStack Router/Query/Form/Table
- **File-Based Routing**: Auto-generated route tree (TanStack Router)
- **Co-Located API Hooks**: Modules own their data-fetching logic
- **Real-Time Chat**: WebSocket integration with auto-reconnect & exponential backoff
- **Responsive Design**: Shadcn/ui + Tailwind CSS

### Scheduling (Timetable Module)
- **CSP Solver**: Constraint satisfaction with backtracking + heuristics (MRV, LCV)
- **Hard Constraints**: No time conflicts, specialization match, room capacity
- **Soft Constraints**: Teacher preferences, workload balance
- **Timeout-Safe**: Context cancellation returns best partial solution
- **Manual Override**: Teachers can be manually assigned with suggestions

### Prerequisites Management (Subject Module)
- **DAG Model**: Directed acyclic graph of course prerequisites
- **Types**: Strict, recommended, corequisite (priority 1-5)
- **Validation**: Cycle detection (DFS 3-color), topological sort
- **Frontend**: Interactive DAG visualization

## Documentation

See `/docs` for comprehensive guides:
- **project-overview-pdr.md**: Vision, goals, stakeholders, requirements
- **codebase-summary.md**: High-level codebase overview
- **code-standards.md**: Go & frontend conventions, patterns, file naming
- **system-architecture.md**: Service topology, data flows, design decisions
- **project-roadmap.md**: MVP status, Phase 2-4 planned features
- **deployment-guide.md**: Local dev setup, Docker Compose, environment variables

## Development Workflow

1. **Implement**: Follow DDD + Clean Architecture (see `/docs/code-standards.md`)
2. **Test**: Run `make test` for all services
3. **Lint**: Run `make lint` (buf, go vet)
4. **Migrate**: Proto changes → `make proto` → test
5. **Commit**: Conventional commits, no credentials in git
6. **Document**: Update relevant docs when adding features

## API Overview

### Authentication
```bash
POST /api/auth/register          # Create account
POST /api/auth/login             # Get access + refresh tokens
POST /api/auth/refresh           # Refresh access token
GET  /api/users/me               # Get current user (requires JWT)
```

### HR Module
```bash
GET/POST   /api/hr/teachers      # List/create teachers
GET/PATCH  /api/hr/teachers/{id} # Get/update teacher
DELETE     /api/hr/teachers/{id} # Delete teacher
GET/POST   /api/hr/departments   # Manage departments
```

### Subject Module
```bash
GET/POST   /api/subjects         # List/create subjects
GET/PATCH  /api/subjects/{id}    # Get/update subject
DELETE     /api/subjects/{id}    # Delete subject
GET/POST   /api/subjects/{id}/prerequisites  # Manage DAG
```

### Timetable Module
```bash
GET/POST   /api/timetable/semesters        # Manage semesters
POST       /api/timetable/semesters/{id}/generate  # Trigger CSP solver
GET        /api/timetable/schedules        # Get schedules
PUT        /api/timetable/schedules/{id}/entries/{entryId}  # Manual assignment
```

### AI Chat
```bash
WebSocket  /ws/chat?token=ACCESS_TOKEN  # Stream chat responses
```

## Environment Variables

**Core Service** (services/core/config/local.yaml):
```yaml
server:
  http_port: 8080
  grpc_port: 50051

auth:
  jwt_secret: your-secret-key
  access_token_ttl: 15m
  refresh_token_ttl: 7d

llm:
  provider: claude  # or openai
  model: claude-haiku-4-5-20251001
  api_key: ${CLAUDE_API_KEY}

nats:
  url: nats://localhost:4222

database:
  url: postgres://myrmex:myrmex_dev@localhost:5432/myrmex?sslmode=disable
```

Similar configs for each module (adjust ports & database schemas).

## Running Tests

```bash
# All services
make test

# Single service
cd services/module-hr && go test ./...

# With coverage
go test -cover ./...
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Port already in use | Kill process: `lsof -i :8080` then `kill -9 <PID>` |
| Database connection error | Check `docker ps`, ensure PostgreSQL is running, verify DATABASE_URL |
| Proto changes not reflected | Run `make proto` before rebuild |
| Frontend API 404 | Ensure API gateway is running on :8080 and CORS is configured |
| NATS connection error | Check `docker logs nats`, ensure port 4222 is open |

## Contributing

1. Create a feature branch from `main`
2. Follow code standards (see `/docs/code-standards.md`)
3. Write tests for new functionality
4. Update relevant documentation
5. Submit PR with clear description

## License

MIT

## Support

For questions or issues, open a GitHub issue or check `/docs` for detailed guides.
