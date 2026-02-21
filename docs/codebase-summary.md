# Myrmex Codebase Summary

## Overview

Myrmex is a Go monorepo with 5 modules using `go.work`:
- `gen/go` - Generated protobuf code (buf generate)
- `pkg` - Shared packages (logger, config, eventstore, nats, middleware)
- `services/core` - HTTP gateway, auth, module registry, AI chat
- `services/module-hr` - Department & teacher management
- `services/module-subject` - Subject & prerequisite DAG
- `services/module-timetable` - Semester, room, schedule management & CSP solver

**Total Codebase**: ~254K tokens (321 files, 985K chars)

## Repomix Snapshot

- `repomix-output.xml` was generated via `repomix` on 2026-02-21 and captures a compact representation of the repository, including token counts and a security scan that excludes certain docs and config files flagged as sensitive.
- Use the compaction output when navigating large/generated files such as `gen/go/*` or proto artifacts that are otherwise too heavy to load directly.

## Monorepo Structure

```
myrmex/
├── go.work                 # Monorepo manifest (1.26)
├── go.work.sum
├── Makefile                # Build targets: proto, build, test, lint, up, down, migrate
├── buf.yaml                # Protobuf linter/code gen config
├── buf.gen.yaml            # Code generation rules (Go, gRPC)
├── repomix-output.xml      # Codebase snapshot
│
├── gen/go/                 # Generated protobuf code (prod: ignore)
│   ├── core/v1/
│   ├── hr/v1/
│   ├── subject/v1/
│   └── timetable/v1/
│
├── pkg/                    # Shared packages
│   ├── logger/             # Zap logger factory (NewLogger)
│   ├── config/             # Viper config (file + env overlay)
│   ├── eventstore/         # PostgreSQL event store (interface + impl)
│   ├── nats/               # NATS JetStream connect/publish/subscribe
│   ├── middleware/         # gRPC auth interceptor (ValidateJWT)
│   └── go.mod
│
├── proto/                  # Protobuf definitions
│   ├── core/v1/
│   │   ├── auth.proto      # AuthService (Login, Register, RefreshToken)
│   │   ├── user.proto      # User message + UserService (CRUD)
│   │   ├── module.proto    # ModuleRegistryService
│   │   └── common.proto    # Shared types (PaginationRequest/Response)
│   ├── hr/v1/
│   │   ├── teacher.proto   # TeacherService (CRUD + Availability)
│   │   └── department.proto # DepartmentService
│   ├── subject/v1/
│   │   ├── subject.proto   # SubjectService
│   │   └── prerequisite.proto # PrerequisiteService
│   └── timetable/v1/
│       ├── timetable.proto # TimetableService (Generate, Get, UpdateEntry, Suggest)
│       └── semester.proto  # SemesterService + Room + TimeSlot
│
├── services/
│   ├── core/               # HTTP gateway + auth + chat
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/     # User aggregate, auth/llm domain services
│   │   │   ├── application/ # CQRS handlers
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/ # User repository, event store impl
│   │   │   │   ├── auth/   # JWT, bcrypt
│   │   │   │   ├── llm/    # Claude + OpenAI provider adapters
│   │   │   │   ├── agent/  # Tool registry + executor
│   │   │   │   └── messaging/ # NATS publisher
│   │   │   ├── interface/
│   │   │   │   ├── grpc/   # Auth, Module Registry, User gRPC servers
│   │   │   │   ├── http/   # Gin router, middleware, handlers
│   │   │   │   └── middleware/ # CORS, rate limit, auth
│   │   │   ├── migrations/ # Goose SQL migrations
│   │   │   ├── sql/queries/ # sqlc query definitions
│   │   │   └── config/     # local.yaml (JWT secret, LLM key, ports)
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── module-hr/          # Department + Teacher management
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   │   ├── entity/ # Department, Teacher (aggregate), Availability (VO)
│   │   │   │   ├── repository/ # DepartmentRepository, TeacherRepository interfaces
│   │   │   │   └── service/ # DomainService for business logic
│   │   │   ├── application/ # CQRS: CreateTeacher, ListTeachers, UpdateAvailability, etc.
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/ # sqlc + repository impls
│   │   │   │   └── messaging/ # NATS publishers
│   │   │   ├── interface/grpc/ # DepartmentServer, TeacherServer
│   │   │   ├── migrations/
│   │   │   ├── sql/queries/
│   │   │   └── config/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── module-subject/     # Subject + Prerequisite DAG
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   │   ├── entity/ # Subject (aggregate), Prerequisite, PrerequisiteType (VO)
│   │   │   │   ├── repository/
│   │   │   │   └── service/ # DAGService (cycle detection, topological sort)
│   │   │   ├── application/ # CQRS: CreateSubject, AddPrerequisite, ValidateDAG, etc.
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/ # sqlc + repository impls
│   │   │   │   └── messaging/
│   │   │   ├── interface/grpc/
│   │   │   ├── migrations/
│   │   │   ├── sql/queries/
│   │   │   └── config/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── module-timetable/   # Semester + Schedule + CSP solver
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── domain/
│       │   │   ├── entity/ # Semester, Schedule (aggregate), ScheduleEntry, Room, TimeSlot
│       │   │   ├── repository/
│       │   │   └── service/ # CSP solver with AC-3, backtracking, MRV, LCV heuristics
│       │   ├── application/ # CQRS: GenerateSchedule, GetSchedule, UpdateEntry, etc.
│       │   ├── infrastructure/
│       │   │   ├── persistence/
│       │   │   └── messaging/
│       │   ├── interface/grpc/
│       │   ├── migrations/
│       │   ├── sql/queries/
│       │   └── config/
│       ├── go.mod
│       └── Dockerfile
│
├── frontend/               # React + TypeScript
│   ├── src/
│   │   ├── main.tsx
│   │   ├── index.css
│   │   ├── config/
│   │   │   ├── query-client.ts # TanStack Query defaults (30s stale, 5min gc)
│   │   │   └── router.ts       # TanStack Router + route tree
│   │   ├── lib/
│   │   │   ├── api/
│   │   │   │   ├── client.ts   # Axios + JWT interceptor + 401 logout
│   │   │   │   ├── endpoints.ts # API route constants
│   │   │   │   └── types.ts    # Shared API types
│   │   │   ├── hooks/ # use-auth, use-current-user, use-toast
│   │   │   ├── stores/ # auth-store.ts (localStorage JWT)
│   │   │   └── utils/  # cn(), format-date()
│   │   ├── components/
│   │   │   ├── layouts/    # AppLayout, SidebarNav, TopBar
│   │   │   ├── shared/     # DataTable, FormField, PageHeader, ConfirmDialog
│   │   │   └── ui/         # Shadcn/ui primitives (11 components)
│   │   ├── chat/
│   │   │   ├── components/ # ChatPanel (FAB), ChatMessage, ChatInput
│   │   │   ├── hooks/      # use-chat.ts (WebSocket + auto-reconnect)
│   │   │   └── types.ts    # WsServerEvent, WsClientMessage
│   │   ├── modules/
│   │   │   ├── hr/         # Teacher + Department (components, hooks, types)
│   │   │   ├── subject/    # Subject + Prerequisites (DAG viz)
│   │   │   └── timetable/  # Semester + Schedule (CSP trigger, calendar)
│   │   └── routes/         # File-based routing (auto-routed by TanStack)
│   │       ├── __root.tsx
│   │       ├── index.tsx, login.tsx, register.tsx
│   │       └── _authenticated/ (auth guard)
│   │           ├── dashboard.tsx
│   │           ├── hr/
│   │           ├── subjects/
│   │           └── timetable/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── tailwind.config.ts
│
├── deploy/
│   └── docker/
│       ├── compose.yml      # PostgreSQL 16, NATS 2.10, Redis 7
│       └── init.sql         # Schema initialization
│
└── docs/                    # This documentation
    ├── README.md
    ├── project-overview-pdr.md
    ├── codebase-summary.md (this file)
    ├── code-standards.md
    ├── system-architecture.md
    ├── project-roadmap.md
    └── deployment-guide.md
```

## Tech Stack Summary

| Layer | Technology | Version | Purpose |
|-------|-----------|---------|---------|
| **Runtime** | Go | 1.26 | All backend services |
| **RPC** | gRPC + Protobuf | Latest (buf managed) | Inter-service communication |
| **DB** | PostgreSQL | 16-alpine | All data + event store |
| **ORM** | sqlc | Latest | Type-safe query generation |
| **Migration** | goose | Latest | Schema management |
| **Message Bus** | NATS JetStream | 2.10-alpine | Event streaming + persistence |
| **Cache** | Redis | 7-alpine | (Reserved, not yet used) |
| **HTTP Gateway** | Gin | Latest | Core service HTTP API |
| **Config** | Viper | Latest | YAML + env config |
| **Logging** | Zap | Latest | Structured JSON logs |
| **Auth** | JWT + bcrypt | Latest | Access/refresh tokens, password hashing |
| **AI** | Claude 4.5 / OpenAI | Configurable | Conversational operations |
| **Frontend** | React | 19 | SPA UI |
| **Router** | TanStack Router | 1.161.3 | File-based routing |
| **State** | TanStack Query | 5.90.21 | Server state management |
| **Form** | TanStack Form + Zod | 1.28.3 + 3.24.1 | Form validation |
| **Table** | TanStack Table | 8.21.3 | Data table with pagination, sorting |
| **UI Framework** | Shadcn/ui | Latest | Radix UI + Tailwind CSS 4 |
| **HTTP Client** | Axios | 1.7.9 | API requests + interceptors |
| **Icons** | Lucide React | 0.575.0 | Icon library |

## Key Files & Their Purposes

### Backend Entry Points
- `services/core/cmd/server/main.go` - HTTP gateway + gRPC server (port 8000/50051)
- `services/module-hr/cmd/server/main.go` - gRPC (port 50052)
- `services/module-subject/cmd/server/main.go` - gRPC (port 50053)
- `services/module-timetable/cmd/server/main.go` - gRPC (port 50054)

### Critical Domain Logic
- `services/module-hr/internal/domain/entity/teacher.go` - Teacher aggregate
- `services/module-subject/internal/domain/service/dag_service.go` - Cycle detection + topological sort
- `services/module-timetable/internal/domain/service/csp_solver.go` - Constraint satisfaction with AC-3 + backtracking

### Shared Infrastructure
- `pkg/logger/logger.go` - Zap logger initialization
- `pkg/config/config.go` - Viper configuration loading
- `pkg/eventstore/event_store.go` - PostgreSQL event sourcing interface
- `pkg/nats/nats.go` - JetStream connection + pubsub
- `pkg/middleware/auth_interceptor.go` - gRPC JWT validation

### Protobuf Definitions
- `proto/core/v1/auth.proto` - Auth service RPC definitions
- `proto/hr/v1/teacher.proto` - Teacher CRUD RPC
- `proto/subject/v1/prerequisite.proto` - Prerequisite DAG RPC
- `proto/timetable/v1/timetable.proto` - Schedule generation RPC

### Database Migrations (per service)
- `services/{service}/migrations/` - Goose-managed SQL migrations
- `services/{service}/internal/sql/queries/` - sqlc query definitions (*.sql)

### Frontend Key Files
- `frontend/src/config/router.ts` - Route definitions + TanStack Router
- `frontend/src/lib/api/client.ts` - Axios + interceptors
- `frontend/src/chat/hooks/use-chat.ts` - WebSocket chat integration
- `frontend/src/modules/*/hooks/` - Module-specific API hooks (co-located)
- `frontend/src/routes/` - File-based route components

## Code Metrics

| Metric | Value |
|--------|-------|
| Total Files | 321 |
| Total Tokens | ~254K |
| Total Characters | ~985K |
| Largest Files | Protobuf generated (teacher.pb.go: 9.8K tokens) |
| Services | 4 (core, hr, subject, timetable) |
| Shared Packages | 5 (logger, config, eventstore, nats, middleware) |
| Go Modules | 5 (gen, pkg, core, hr, subject, timetable) |
| Frontend Components | 11 Shadcn/ui + 5+ custom |
| Proto Definitions | 10 files across 4 services |

## Dependencies per Service

```
Core → (nothing)
Module-HR → pkg, NATS, PostgreSQL
Module-Subject → pkg, NATS, PostgreSQL
Module-Timetable → pkg, Module-Subject (gRPC), NATS, PostgreSQL
Frontend → Core gRPC gateway (HTTP/JSON)
```

## Configuration Files

All services use Viper config (file + env overlay):
- `services/{service}/config/local.yaml` - Local dev config (Git-ignored)
- `services/{service}/config/default.yaml` - Default values (in repo)
- Environment variables override YAML (e.g., `DATABASE_URL`, `NATS_URL`)

## Build & Run

```bash
# Generate protobuf code
make proto

# Build all services
make build

# Run tests
make test

# Lint (buf + go vet)
make lint

# Start infrastructure (Docker Compose)
make up

# Run database migrations
make migrate

# Run services (each in separate terminal)
cd services/core && go run ./cmd/server
cd services/module-hr && go run ./cmd/server
cd services/module-subject && go run ./cmd/server
cd services/module-timetable && go run ./cmd/server

# Start frontend
cd frontend && npm install && npm run dev
```

## Security Considerations

- **Credentials**: `local.yaml` files are Git-ignored; use `default.yaml` for non-secret defaults
- **JWT**: 15min access token + 7day refresh token
- **Password**: bcrypt hashing (salt rounds: 10)
- **CORS**: Configured in core Gin router
- **gRPC Auth**: JWT interceptor validates all internal service calls

## Performance Characteristics

- **API Latency**: <500ms p95 (excl. CSP solver)
- **CSP Solver**: <30s p95 (context cancellation → partial solution)
- **Database Queries**: Type-safe via sqlc (no ORM overhead)
- **Event Store**: Optimistic concurrency (version column)
- **Frontend SPA**: Vite + tree-shaking; ~80KB gzipped initial bundle

## Known Gaps & Limitations

1. **Frontend**: No E2E tests, no mobile hamburger menu, no drag-drop yet
2. **Auth**: No token rotation; no 2FA
3. **Chat**: Message storage in PostgreSQL (will migrate to MongoDB)
4. **Monitoring**: Prometheus metrics not yet integrated
5. **Scale**: NATS single-instance (needs clustering for HA)
6. **Tenancy**: Single-tenant MVP; multi-tenant planned for Phase 4
