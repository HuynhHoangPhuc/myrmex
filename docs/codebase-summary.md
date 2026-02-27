# Myrmex Codebase Summary

## Overview

Myrmex is a Go monorepo with 6 modules using `go.work`:
- `gen/go` - Generated protobuf code (buf generate)
- `pkg` - Shared packages (logger, config, eventstore, nats, middleware)
- `services/core` - HTTP gateway, auth, module registry, AI chat
- `services/module-hr` - Department & teacher management
- `services/module-subject` - Subject & prerequisite DAG
- `services/module-timetable` - Semester, room, schedule management & CSP solver
- `services/module-analytics` - Analytics dashboard, KPIs, reporting (PDF/Excel export)

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
│   │   │   │   ├── llm/    # OpenAI, Claude, Gemini provider adapters
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
│   ├── module-timetable/   # Semester + Schedule + CSP solver
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   │   ├── entity/ # Semester, Schedule (aggregate), ScheduleEntry, Room, TimeSlot
│   │   │   │   ├── repository/
│   │   │   │   └── service/ # CSP solver with AC-3, backtracking, MRV, LCV heuristics
│   │   │   ├── application/ # CQRS: GenerateSchedule, GetSchedule, UpdateEntry, etc.
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/
│   │   │   │   └── messaging/
│   │   │   ├── interface/grpc/
│   │   │   ├── migrations/
│   │   │   ├── sql/queries/
│   │   │   └── config/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── module-analytics/    # Analytics, reporting, dashboards
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── application/
│       │   │   ├── query/ # GetWorkloadHandler, GetUtilizationHandler, GetDashboardSummaryHandler
│       │   │   └── export/ # PDF/Excel generators
│       │   ├── infrastructure/
│       │   │   ├── persistence/ # AnalyticsRepository (star-schema queries)
│       │   │   └── messaging/ # NATS consumer for ETL
│       │   ├── interface/
│       │   │   ├── grpc/ # AnalyticsService gRPC
│       │   │   └── http/ # Dashboard + export HTTP handlers
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
│   │   │   ├── timetable/  # Semester + Schedule (CSP trigger, calendar)
│   │   │   └── analytics/  # Dashboard KPIs, workload/utilization charts, exports
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
│   ├── tailwind.config.ts
│   ├── vitest.config.ts (integrated in vite.config.ts)
│   ├── playwright.config.ts
│   └── src/
│       ├── test-setup.ts           # Vitest globals setup
│       ├── **/*.test.ts(x)         # Unit tests (Vitest + React Testing Library)
│       └── e2e/                    # E2E tests (Playwright)
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
| **AI** | Claude 4.5 / OpenAI / Gemini | Configurable | Conversational operations |
| **Frontend Testing** | Vitest + React Testing Library | Latest | Unit tests, ~70% coverage |
| **E2E Testing** | Playwright | Latest | Browser automation tests |
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
| Total Files | 350+ |
| Total Tokens | ~300K+ |
| Total Characters | ~1.1M+ |
| Largest Files | Protobuf generated (teacher.pb.go: 9.8K tokens) |
| Services | 5 (core, hr, subject, timetable, analytics) |
| Shared Packages | 5 (logger, config, eventstore, nats, middleware) |
| Go Modules | 6 (gen, pkg, core, hr, subject, timetable, analytics) |
| Frontend Components | 11 Shadcn/ui + 5+ custom |
| Proto Definitions | 10 files across 4 services |

## Dependencies per Service

```
Core → (nothing)
Module-HR → pkg, NATS, PostgreSQL
Module-Subject → pkg, NATS, PostgreSQL
Module-Timetable → pkg, Module-Subject (gRPC), NATS, PostgreSQL
Module-Analytics → pkg, NATS, PostgreSQL (consumes events)
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

# Build all services (including module-analytics)
make build

# Run backend tests (Go)
make test

# Run backend tests with coverage
make test-cover

# Lint code (buf + go vet)
make lint

# Start infrastructure (Docker Compose)
make up

# Run database migrations (all services)
make migrate

# Run services (each in separate terminal)
cd services/core && go run ./cmd/server
cd services/module-hr && go run ./cmd/server
cd services/module-subject && go run ./cmd/server
cd services/module-timetable && go run ./cmd/server
cd services/module-analytics && go run ./cmd/server

# Start frontend
cd frontend && npm install && npm run dev

# Frontend unit tests (Vitest)
cd frontend && npm run test

# Frontend unit tests in watch mode
cd frontend && npm run test:watch

# Frontend coverage report
cd frontend && npm run test:coverage

# Frontend E2E tests (Playwright)
cd frontend && npm run test:e2e

# Frontend E2E tests with UI
cd frontend && npm run test:e2e:ui

# One-command demo (all services + infra)
make demo
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

## Analytics & Testing Infrastructure (Feb 26)

### Module-Analytics Service
- New service: `services/module-analytics` for business intelligence
- Star-schema analytics database: `dim_teacher`, `dim_subject`, `dim_department`, `dim_semester`, `fact_schedule_entry`
- Dashboard APIs: `/api/analytics/dashboard-summary`, `/api/analytics/workload`, `/api/analytics/utilization`
- Export functionality: PDF/Excel schedule generation via `export_handler.go`
- NATS event consumer: Processes NATS events (hr.teacher.*, subject.*, schedule.generation_completed) for ETL
- All operations via HTTP (reverse-proxied by core gateway at `/api/analytics/*`)

### Frontend Testing Infrastructure
- **Vitest + React Testing Library**: Unit tests integrated in vite.config.ts
  - Run: `npm run test` / `npm run test:watch` / `npm run test:coverage`
  - Current test files: 4+ tests covering auth store, date formatting, API endpoints, period utilities
- **Playwright**: E2E test framework configured
  - Run: `npm run test:e2e` / `npm run test:e2e:ui`
  - Config: `playwright.config.ts`

### Go Testing
- `make test`: Runs all backend tests (Go 1.26)
- `make test-cover`: Generates coverage reports per service
- Coverage: >70% across core, module-hr, module-subject, module-timetable, module-analytics
- All services now in Makefile SERVICES list for automated testing

### Mock LLM Provider
- `LLM_PROVIDER=mock` option for testing without real API keys
- Enables E2E tests and CI/CD pipelines to run without API credentials

## Infrastructure Updates (Feb 26)

### WebSocket Fix
- Switched from `coder/websocket` to `gorilla/websocket` in HTTP chat handler
- Old library was incompatible with Gin's response writer; gorilla/websocket resolves compatibility issues

### Multi-LLM Provider Support
- Added Gemini provider alongside existing OpenAI and Claude
- New file: `services/core/internal/infrastructure/llm/gemini_provider.go`
- `LLMProvider` interface: `ChatWithTools()` + `StreamChat()` methods
- Config: `LLM_PROVIDER` (default: openai), `LLM_MODEL=gemini-3-flash-preview` for Gemini free tier
- `ToolCall.ProviderMeta` field for provider-specific metadata (e.g., Gemini's `thoughtSignature`)
- ThinkingBudget disabled in Gemini to avoid signature requirements on non-thinking usage

### Docker Compose Enhancement
- All docker compose targets now use `--env-file .env` for root .env pickup
- New `COMPOSE` variable defined in Makefile
- `LLM_PROVIDER` and `LLM_MODEL` env vars now passed to core container

### Frontend Chat Enhancement
- Added `react-markdown` to `chat-message.tsx` for markdown rendering in AI chat bubble

## Proto & API Updates (Feb 26)

### Teacher Proto Enhancements
- `employee_code: string` — Institutional employee identifier
- `max_hours_per_week: int32` — Workload constraint
- `specializations: []string` — Subject specializations (from many-to-many join)
- `phone: string` — Contact information

### Subject Proto Enhancements
- `weekly_hours: int32` — Contact hours per week (constraint for CSP)
- `is_active: bool` — Offering status (defaults true)

## Bug Fixes & Enhancements (Feb 27)

### Schedule Generation & HTTP Response
- Fixed `GenerateSchedule` HTTP response to return full schedule object (was returning only `{schedule_id}`)
- Added schedule status constants: `generating`, `completed`, `failed`
- Schedule status tracks generation state: starts as `generating`, transitions to `completed` or `failed`

### SQL Query Bug Fix
- Fixed `ListSchedulesPaged` WHERE clause operator precedence bug
- Corrected: `($1 = '000...'::uuid OR semester_id = $1)` (was missing proper grouping)
- Prevents incorrect filtering when `semester_id` parameter is NULL

### Semester Response Enrichment
- `GetSemester` now fetches and includes `time_slots` (reference data) and `rooms` via gRPC
- Added `ListTimeSlots` and `ListRooms` RPCs to timetable proto
- Enables frontend to render schedules with full time slot and room context

### AI Chat Agent Improvements
- System prompt: Added explicit workflow instruction for semester-dependent operations
  - Always call `timetable.list_semesters` first to get semester UUID
  - Then call `timetable.generate` with the UUID
- Increased `maxToolIterations` from 5 to 10 for complex multi-step workflows
- Added `timetable.list_semesters` tool to tool registry
- Fixed `timetable.suggest_teachers`: Removed unused `semester_id` required field

### Teacher Availability & Time Representation
- Teacher availability now represents time slots as RFC3339 time strings (HH:MM format)
- Period-to-time conversion: 6 periods mapped to 07:00–19:00 in 2-hour increments
- `GetTeacher` HTTP response includes `availability: [{day_of_week, start_time, end_time}]`
- `UpdateTeacherAvailability` accepts `{available_slots: [{day_of_week, start_time, end_time}]}`
- Conversion helpers: `hrSlotStart()`, `hrSlotEnd()`, `hrTimeToSlot()` for seamless backend storage
