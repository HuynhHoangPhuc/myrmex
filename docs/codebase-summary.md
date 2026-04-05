# Myrmex Codebase Summary

## Overview

Myrmex is a Go monorepo with 9 modules using `go.work`:
- `gen/go` - Generated protobuf code (buf generate)
- `pkg` - Shared packages (logger, config, cache Redis abstraction, eventstore, nats, middleware)
- `services/core` - HTTP gateway, auth, module registry, AI chat, audit logging, WebSocket notifications relay
- `services/module-hr` - Department & teacher management
- `services/module-subject` - Subject & prerequisite DAG
- `services/module-timetable` - Semester, room, schedule management & CSP solver
- `services/module-student` - Student CRUD, enrollment workflow, grades, transcripts
- `services/module-analytics` - Analytics dashboard, KPIs, reporting (PDF/Excel export)
- `services/module-notification` - Email + in-app notifications, preferences, event routing, email queue (NEW)

**Total Codebase**: ~280K tokens (350+ files, 1.1M+ chars)

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
│   ├── student/v1/
│   ├── subject/v1/
│   └── timetable/v1/
│
├── pkg/                    # Shared packages
│   ├── logger/             # Zap logger factory (NewLogger)
│   ├── config/             # Viper config (file + env overlay)
│   ├── cache/              # Cache interface + RedisCache impl (JSON values, SCAN invalidation)
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
│   ├── student/v1/
│   │   └── student.proto   # StudentService (CRUD foundation)
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
│   ├── module-student/     # Student management: CRUD, enrollment, grades, transcript
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/      # Student, Enrollment, Grade aggregates + repositories
│   │   │   ├── application/ # CQRS handlers (create student, request/approve enrollment, assign grade)
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/ # sqlc + repository impls
│   │   │   │   ├── messaging/ # NATS event publishers
│   │   │   │   └── cache/  # Redis-backed prerequisite caching
│   │   │   ├── interface/grpc/ # StudentService + EnrollmentService + GradeService gRPC
│   │   │   ├── migrations/  # Student schema + enrollment + grades tables
│   │   │   ├── sql/queries/
│   │   │   └── config/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── module-analytics/    # Analytics, reporting, dashboards
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── application/
│   │   │   │   ├── query/ # GetWorkloadHandler, GetUtilizationHandler, GetDashboardSummaryHandler
│   │   │   │   └── export/ # PDF/Excel generators
│   │   │   ├── infrastructure/
│   │   │   │   ├── persistence/ # AnalyticsRepository (star-schema queries)
│   │   │   │   └── messaging/ # NATS consumer for ETL
│   │   │   ├── interface/
│   │   │   │   ├── grpc/ # AnalyticsService gRPC
│   │   │   │   └── http/ # Dashboard + export HTTP handlers
│   │   │   ├── migrations/
│   │   │   ├── sql/queries/
│   │   │   └── config/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── module-notification/  # Email + in-app notifications (NEW)
│       ├── cmd/server/main.go (port 8056)
│       ├── internal/
│       │   ├── domain/       # Notification, Preference aggregates, repositories
│       │   ├── application/  # DispatchNotification, ListNotifications commands
│       │   ├── infrastructure/
│       │   │   ├── persistence/ # notification, preference, email_queue repositories
│       │   │   ├── messaging/   # NATS publisher, event consumer, event router
│       │   │   ├── email/       # SMTP service, MJML template renderer
│       │   │   └── recipient/   # Cross-schema recipient resolver (HR, Student, Analytics)
│       │   ├── interface/http/   # NotificationHandler, PreferenceHandler, AnnouncementHandler
│       │   ├── migrations/       # notification, preference, email_queue schemas
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
│   │   │   ├── components/ # ChatPanel (fixed right-side panel with expand/fullscreen), ChatMessage, ChatInput
│   │   │   ├── hooks/      # use-chat.ts (WebSocket + auto-reconnect)
│   │   │   └── types.ts    # WsServerEvent, WsClientMessage
│   │   ├── notifications/  # In-app notifications (NEW)
│   │   │   ├── components/ # notification-panel.tsx, notification-toast.tsx, notification-preferences.tsx
│   │   │   └── hooks/      # use-notifications.ts, use-notification-ws.ts
│   │   ├── modules/
│   │   │   ├── hr/         # Teacher + Department (components, hooks, types)
│   │   │   ├── subject/    # Subject + Prerequisites (React Flow DAG viz, conflict detection)
│   │   │   │   ├── components/
│   │   │   │   │   ├── prerequisite-dag.tsx        # React Flow DAG canvas + controls
│   │   │   │   │   ├── dag-subject-node.tsx        # Custom node rendering
│   │   │   │   │   ├── conflict-warning-banner.tsx # Conflict display + auto-fix
│   │   │   │   │   ├── offering-manager.tsx        # Semester offerings with conflict checks
│   │   │   │   │   ├── prereq-chip.tsx             # Prerequisite code styling with tooltips
│   │   │   │   │   └── subject-columns.tsx         # Table columns including prerequisites
│   │   │   │   ├── utils/
│   │   │   │   │   ├── dag-layout.ts     # Dagre layout helper
│   │   │   │   │   └── dept-color.ts     # Deterministic dept color mapping
│   │   │   │   └── hooks/use-subjects.ts # useFullDAG, useCheckConflicts
│   │   │   ├── timetable/  # Semester + Schedule (CSP trigger, calendar, room assignment)
│   │   │   │   ├── components/
│   │   │   │   │   ├── room-manager.tsx         # Multi-select room UI
│   │   │   │   │   └── room-assignment-dialog.tsx # Room picker dialog
│   │   │   │   ├── hooks/
│   │   │   │   │   └── use-rooms.ts             # Query global room list
│   │   │   │   └── types.ts                     # AssignRoomInput type
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
    ├── deployment-guide.md
    ├── oauth-provider-setup.md      # Google + Microsoft OAuth provider configuration
    ├── journals/                     # Work logs
    │   └── 2026-04-05-fix-oauth-login.md  # OAuth login fix root cause analysis
    └── project-changelog.md
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
| **Cache** | Redis | 7-alpine | Shared cache abstraction in `pkg/cache` (JSON values + SCAN invalidation) |
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
| **Graph Viz** | React Flow + Dagre | 12.10.1 + 2.0.4 | Interactive DAG visualization |
| **Form** | TanStack Form + Zod | 1.28.3 + 3.24.1 | Form validation |
| **Table** | TanStack Table | 8.21.3 | Data table with pagination, sorting |
| **UI Framework** | Shadcn/ui | Latest | Radix UI + Tailwind CSS 4 |
| **UI Tooltips** | Radix UI Tooltip | Latest | Interactive tooltips component |
| **HTTP Client** | Axios | 1.7.9 | API requests + interceptors |
| **Icons** | Lucide React | 0.575.0 | Icon library |

## Key Files & Their Purposes

### Backend Entry Points
- `services/core/cmd/server/main.go` - HTTP gateway + gRPC server (port 8080/50051) — auth, module registry, chat, audit, rate limiting
- `services/module-hr/cmd/server/main.go` - gRPC (port 50052) — departments, teachers, availability
- `services/module-subject/cmd/server/main.go` - gRPC (port 50053) — subjects, prerequisites, DAG validation
- `services/module-timetable/cmd/server/main.go` - gRPC (port 50054) — semesters, rooms, CSP scheduling
- `services/module-student/cmd/server/main.go` - gRPC (port 50055) — student CRUD, enrollment, grades, transcripts
- `services/module-analytics/cmd/server/main.go` - HTTP (port 8055) — analytics KPIs, workload, utilization, export
- `services/module-notification/cmd/server/main.go` - HTTP (port 8056) — notifications, preferences, email queue, NATS consumer

### Gateway Proxies
- `services/core/internal/interface/http/student_handler.go` - Admin-only `/api/students` CRUD proxy to module-student
- `services/core/cmd/server/module_clients.go` - Wires `student.grpc_addr` into the core gateway

### Critical Domain Logic
- `services/module-hr/internal/domain/entity/teacher.go` - Teacher aggregate
- `services/module-subject/internal/domain/service/dag_service.go` - Cycle detection + topological sort
- `services/module-timetable/internal/domain/service/csp_solver.go` - Constraint satisfaction with AC-3 + backtracking

### Shared Infrastructure
- `pkg/logger/logger.go` - Zap logger initialization
- `pkg/config/config.go` - Viper configuration loading
- `pkg/cache/cache.go` - Cache interface + cache miss sentinel
- `pkg/cache/redis_cache.go` - Redis JSON cache with cursor-based pattern invalidation
- `pkg/eventstore/event_store.go` - PostgreSQL event sourcing interface
- `pkg/nats/nats.go` - JetStream connection + pubsub
- `pkg/middleware/auth_interceptor.go` - gRPC JWT validation

### Protobuf Definitions
- `proto/core/v1/auth.proto` - Auth service RPC definitions
- `proto/hr/v1/teacher.proto` - Teacher CRUD RPC
- `proto/student/v1/student.proto` - Student CRUD RPC foundation
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
| Total Files | 380+ |
| Total Tokens | ~320K+ |
| Total Characters | ~1.2M+ |
| Largest Files | Protobuf generated (teacher.pb.go: 9.8K tokens) |
| Services | 7 (core, hr, subject, timetable, student, analytics, notification) |
| Shared Packages | 8 (logger, config, cache, eventstore, nats, middleware, messaging) |
| Go Modules | 9 (gen, pkg, core, hr, subject, timetable, student, analytics, notification) |
| Frontend Components | 11 Shadcn/ui + 8+ custom |
| Proto Definitions | 11 files across 5 services |
| Terraform Modules | 8 (networking, cloud-sql, artifact-registry, cloud-run, pubsub, iam, cicd-iam, monitoring) |
| CI/CD Workflows | 3 (.github/workflows/ci.yml, deploy.yml, test.yml) |

## Dependencies per Service

```
Core → (nothing)
Module-HR → pkg, NATS, PostgreSQL
Module-Subject → pkg, NATS, PostgreSQL
Module-Timetable → pkg, Module-Subject (gRPC), NATS, PostgreSQL
Module-Student → pkg, NATS, PostgreSQL
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
cd services/module-student && go run ./cmd/server
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

## Production Deployment Status (Mar 5, 2026)

### Phase 5 Production Pilot — All 8 phases complete:
- **Phase 1 - Messaging Abstraction**: `pkg/messaging/` backend-agnostic Publisher/Consumer (NATS/Pub/Sub)
- **Phase 2 - GCP Terraform IaC**: Full Cloud SQL, Memorystore, Artifact Registry, Cloud Run, Pub/Sub, VPC, Secret Manager setup
- **Phase 3 - CI/CD Pipeline**: `.github/workflows/` with WIF auth, parallel Docker build/deploy, migration job
- **Phase 4 - Security Hardening**: CORS control, SSL enforced, rate limiting, frontend envsubst runtime injection
- **Phase 5 - CSV Bulk Import**: POST /api/admin/import/teachers/students with row-level error reporting
- **Phase 6 - Observability**: `/health` endpoint with dependency checks, X-Request-ID propagation, monitoring alerts
- **Phase 7 - Quality**: gosec fixes, npm audit, k6 load tests (3 scripts: auth-flow, api-crud, mixed-workload)
- **Phase 8 - User Guides**: 5 markdown guides + in-app help page with tabbed UI

## Known Gaps & Limitations

1. **Auth**: No token rotation; no 2FA; password reset workflow pending
2. **Chat**: Message storage in PostgreSQL (will migrate to MongoDB for non-critical data)
3. **Scale**: NATS single-instance (needs clustering for HA in Phase 4)
4. **Tenancy**: Single-tenant MVP; multi-tenant planned for future phases
5. **Mobile**: React Native app deferred to future releases

## Analytics & Testing Infrastructure (Feb 26)

### Module-Analytics Service
- New service: `services/module-analytics` for business intelligence
- Star-schema analytics database: `dim_teacher`, `dim_subject`, `dim_department`, `dim_semester`, `fact_schedule_entry`
- Dashboard APIs: `/api/analytics/dashboard`, `/api/analytics/workload`, `/api/analytics/utilization`, `/api/analytics/department-metrics`, `/api/analytics/schedule-metrics`, `/api/analytics/schedule-heatmap`
- Export functionality: PDF/Excel schedule generation via `export_handler.go`
- NATS event consumer: Processes NATS events (hr.teacher.*, subject.*, schedule.generation_completed) for ETL
- All operations via HTTP (reverse-proxied by core gateway at `/api/analytics/*`)

### Frontend UX Enhancements (Feb 27)
- **Breadcrumb Entity Name Resolution**: Dynamic entity name resolution via React Query
  - Subjects: Fetches subject names from `/api/subjects`
  - Teachers: Fetches teacher names from `/api/hr/teachers`
  - Semesters: Fetches semester names from `/api/timetable/semesters`
  - Provides context-aware navigation across modules
- **Tooltip Component**: Radix UI-based tooltip for interactive hints on UI elements
- **AI Assistant Toggle**: Added to top bar for easy access to chat features
- **Chat Panel**: Fixed right-side panel (380px wide) with expand/fullscreen support and clear messages button

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
- Coverage: >70% across core, module-hr, module-subject, module-timetable, module-student, module-analytics
- All services in Makefile SERVICES list for automated testing

### Notifications System (Mar 4 - Phase 4.4 COMPLETE)
- **Module-Notification Service**: New HTTP microservice on port 8056 with 50+ endpoints
- **Email Notifications**: SMTP backend (go-mail) with MJML templating engine
- **In-App Notifications**: WebSocket push via NATS JetStream consumer in core service
- **Preference Matrix**: 12 events × 2 channels (email + in-app) per user, stored in PostgreSQL
- **Event Routing**: 10 event types (new_announcement, schedule.*, enrollment.*, grade.*, role_updated, user.deleted)
- **Email Queue**: PostgreSQL-backed queue with exponential backoff (5 attempts, 24h max retry window)
- **Cross-Schema Resolver**: Smart recipient lookup across HR (teachers), Student, and Analytics schemas
- **Frontend Notifications**: Pagination + filters, preferences matrix UI, WS toast component, sidebar nav item
- **234+ Backend Tests**: All passing (event consumer, email queue, preferences, routing)

### RBAC Implementation (Mar 4)
- **6 Roles**: super_admin, admin, dean, dept_head, teacher, student
- **Department Scoping**: dept_head + teacher roles bound to department_id in JWT
- **Two-Tier Enforcement**: Middleware (RequireDeptScope) + Handler checks
- **JWT Claims**: Extended with department_id + teacher_id for O(1) permission lookups
- **Admin UI**: /admin/roles page for role management with batch assignment
- **Route Guards**: Protected HR/Subject mutations based on user scope
- **Audit Integration**: All role changes logged to core.audit_logs via middleware

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

## Phase 5 Production Pilot — Deployment & Operations (Mar 5, 2026)

### Messaging Abstraction Layer
- `pkg/messaging/publisher.go` — Pluggable backend (NATS/Pub/Sub/NoopPublisher)
- `pkg/messaging/consumer.go` — Consumer interface with Subscribe/Receive pattern
- `MESSAGING_BACKEND=pubsub|nats` env var controls backend across all 7 services
- Enables seamless migration from NATS (dev) to Google Cloud Pub/Sub (production)

### GCP Cloud Run Deployment (Terraform IaC)
- `deploy/terraform/` — 8 modules: networking, cloud-sql, artifact-registry, cloud-run, pubsub, secret-manager, iam, monitoring
- Cloud SQL: PostgreSQL 16 with SSL enforced (ssl_mode=ENCRYPTED_ONLY)
- Memorystore: Redis 7 for cache backend
- Artifact Registry: Docker image registry (us-central1)
- Cloud Run: 7 services with auto-scaling (min 1, max 10 replicas)
- Pub/Sub: Fully managed message broker (replaces NATS in production)
- Secret Manager: Centralized secrets (API keys, DB credentials, JWT secret)
- VPC: Custom VPC with Cloud NAT for egress control
- IAM: Fine-grained service account roles (least privilege)

### CI/CD Pipeline
- `.github/workflows/ci.yml` — Lint + test all 7 services on PR
- `.github/workflows/deploy.yml` — Full CD pipeline:
  - WIF authentication (Workload Identity Federation) — no long-lived secrets
  - Parallel Docker build for 8 images (core, 6 modules, frontend)
  - Push to Artifact Registry
  - Cloud Run Job: myrmex-migrate (goose migrations)
  - Parallel Cloud Run deployment (7 services)
  - Smoke test: `GET /health` health check validation
- `.github/workflows/test.yml` — Extended testing with coverage reports

### Security Hardening
- `cors_middleware.go` — CORS_ALLOWED_ORIGINS env var (comma-separated), fallback to wildcard in dev
- Auth rate limit: 10 req/min (reduced from 100 for password brute-force protection)
- API rate limit: 100 req/min (per IP)
- SSL enforcement on Cloud SQL + TLS in transit
- Frontend: `nginx-cloudrun.conf` template with envsubst runtime injection of `${CORE_SERVICE_URL}`
- `frontend/docker-entrypoint.sh` — Runtime environment variable substitution

### CSV Bulk Import
- `services/core/internal/interface/http/import_handler.go`:
  - POST /api/admin/import/teachers (admin/super_admin only)
  - POST /api/admin/import/students (admin/super_admin only)
  - GET /api/admin/import/template/teachers, /api/admin/import/template/students
- Row-level error reporting: Track which rows failed + reason (duplicate, validation)
- Skip duplicates gracefully (by email for teachers, by student_code for students)
- Frontend: `frontend/src/routes/_authenticated/admin/import/` with:
  - File upload with CSV preview
  - Results summary + error CSV download
  - Progress indicators

### Observability & Monitoring
- `/health` endpoint with dependency checks: DB + Redis connectivity
- `request_id_middleware.go` — X-Request-ID propagation across services (for tracing)
- `deploy/terraform/monitoring.tf`:
  - Uptime check for `/health` (5-minute intervals)
  - 5xx error rate alert (>1% triggers notification)
  - Cloud SQL connection count alert (>80% capacity)
  - Custom dashboard for latency, error rate, CPU usage

### Quality Assurance
- **gosec fixes**: Fixed G109 (int32 overflow via strconv.ParseInt)
- **npm audit**: Applied all critical/high fixes (frontend dependencies)
- **k6 load tests**: 3 scripts in `deploy/load-tests/`:
  - `auth-flow.js` — 100 VUs, login → create subject → logout
  - `api-crud.js` — 200 VUs, CRUD operations across modules
  - `mixed-workload.js` — 500 VUs, realistic traffic pattern (ramp-up/down)

### User Guides & In-App Help
- `docs/user-guide/` — 5 markdown files:
  - `index.md` — Overview + navigation
  - `admin-guide.md` — System administration, user management, audit logs
  - `teacher-guide.md` — Subject creation, availability, schedule view
  - `student-guide.md` — Enrollment, transcript, prerequisites
  - `department-head-guide.md` — Department management, workload analytics
- Frontend: `frontend/src/routes/_authenticated/help/index.tsx` — Tabbed in-app help page
- Sidebar: Navigation link to help + documentation

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

## Room Assignment Feature (Mar 1)

### Backend Additions
- **Proto**: `room_ids: []string` field added to Semester message, `SetSemesterRooms` RPC
- **Database**: Migration adds `room_ids UUID[]` column to timetable.semesters
- **Repository**: `SetRoomIDs` method in semester repository for persistence
- **gRPC Handler**: `SetSemesterRooms` RPC handler validates and stores room assignments
- **HTTP Endpoints**: `GET /api/timetable/rooms` (ListRooms), `POST /api/timetable/semesters/:id/rooms` (SetSemesterRooms)
- **Schedule Generation**: CSP solver now respects semester `room_ids` constraint

### Frontend Additions
- **Components**: `room-manager.tsx` (multi-select checkbox UI), `room-assignment-dialog.tsx` (room picker + confirm)
- **Hooks**: `use-rooms.ts` (query room list), `useSetSemesterRooms()`, `useAssignRoom()` mutations
- **Types**: `AssignRoomInput` for timetable operations
- **Integration**: Semester wizard step 2 includes room selection; schedule detail adds "Change Room" action
- **User Flow**: Select rooms during semester setup → CSP respects constraints → Manual override via dialog

## Audit Logging & Compliance (Mar 4)

### Backend Audit System
- **Middleware Capture**: Post-handler Gin middleware derives action from HTTP method (POST→Create, PATCH→Update, DELETE→Delete)
- **Async NATS Pipeline**: Fire-and-forget publish to AUDIT.logs stream (non-blocking)
- **Durable Consumer** (audit_consumer.go): JetStream consumer with ack/nack retry, preserves event ordering
- **Audit Repository**: Raw pgx + sqlc for Insert + paginated List (nullable filters)
- **Monthly Partitions**: core.audit_logs with 12 partitions (2026-03 through 2027-02)
- **Indexes**: BRIN (timestamp), B-tree (user_id, resource_type, action) for efficient queries
- **Query Filtering**: user_id, resource_type, action, date range (start_date, end_date) support
- **Constraint Exclusion**: Monthly partition pruning for date-range query optimization
- **Admin Enforcement**: GET /api/audit-logs restricted to admin/super_admin roles

### Frontend Audit Logs UI
- **Route**: /admin/audit-logs (admin/super_admin only)
- **Table Columns**: User, Resource Type, Action, Timestamp (sortable)
- **Row Expansion**: View old/new value diffs with JSON diff rendering
- **Filters**: User selector (dropdown), resource type, action checkboxes, date picker
- **Pagination**: Previous/next controls, total count display
- **Non-blocking**: Graceful degradation if NATS not configured (testing/dev)

## Agent Tool Registry Expansion (Mar 2)

### Tool Registry — 50+ Tools Across 5 Modules
- **Module Naming**: `module.action` pattern (e.g., `hr.list_teachers`)
- **Thread-Safe Implementation**: RWMutex-protected tool map with concurrent access
- **Tool Dispatch**: HTTP self-referential dispatch via internal JWT token
- **HR Tools (10)**: list_teachers, get_teacher, list_departments, get_teacher_availability, create_teacher, update_teacher, delete_teacher, update_teacher_availability, create_department, list_departments
- **Subject Tools (11+)**: list_subjects, get_subject, create_subject, update_subject, delete_subject, get_prerequisites, add_prerequisite, remove_prerequisite, check_conflicts, validate_dag, topological_sort, full_dag
- **Timetable Tools (14)**: list_semesters, generate, get_semester, get_schedule, list_schedules, suggest_teachers, list_rooms, create_semester, set_semester_rooms, create_time_slot, delete_time_slot, apply_time_slot_preset, add_offered_subject, remove_offered_subject, manual_assign
- **Student Tools (10)**: list, get, transcript, create, update, delete, list_enrollments (with subject_id filter), review_enrollment
- **Analytics Tools (6)**: workload, utilization, dashboard, department_metrics, schedule_metrics, schedule_heatmap

### UUID-to-Name Enrichment Pattern
- **Subject Handler**: `buildSubjectMap()` helper enriches prerequisites/topological sort responses with subject names + codes
- **Timetable Handler**: `buildSubjectMap()` helper enriches semester offerings with subject names + codes
- **Purpose**: Make agent responses human-readable instead of raw UUIDs for improved conversational UX

### Student Enrollment Filtering
- **ListEnrollments** API now accepts optional `subject_id` query parameter
- **Implementation**: Fetches 500-item batch server-side, filters client-side for accurate count
- **Agent Tool Update**: `student.list_enrollments` now includes subject_id filtering capability

### Frontend Silent Token Refresh
- **API Client** (`frontend/src/lib/api/client.ts`): Automatic 401 handling with refresh token retry
- **Request Queuing**: Pending requests queued while refresh in progress, all retried on success
- **Graceful Fallback**: Only redirects to login if refresh fails or no refresh token available
- **Improved UX**: Seamless user experience without interruption on token expiry

### Agent Guidelines & Chat Improvements
- **Enhanced Tool Descriptions**: Include UUID resolution workflow hints (e.g., "call list_departments first")
- **Multi-Step Workflow Support**: maxToolIterations=10 allows complex agent sequences
- **System Prompt Guidance**: Explicit instructions for semester-dependent operations (list_semesters before generate)
- **Error Message Security**: Generic error messages prevent internal error leakage to users

### UI & UX Enhancements
- **Collapsible Thinking Toggle**: Optional expanded thinking display in chat tool execution
- **Dark Mode Visibility**: Improved contrast for chat panel and navigation in dark mode
- **Error Handling**: Consistent error messaging across all modules

## OAuth Login Fix (Apr 5, 2026)

### Critical Fixes Across All Environments

**Commit**: `c00eef6 fix(oauth): resolve Docker Compose env vars, Terraform redirect URLs, cookie Secure flag, and URL encoding`

Fixed OAuth login failures in Docker Compose (local dev), Terraform (staging/prod), and Cloud Run by patching 4 distinct root causes:

1. **Docker Compose Missing Env Vars**
   - Added OAuth variables to `deploy/docker/compose.yml` (OAUTH_ENABLED, GOOGLE_CLIENT_*, MICROSOFT_CLIENT_*)
   - OAuth routes now initialize correctly in development environment

2. **Terraform Missing Redirect URLs**
   - Added `OAUTH_GOOGLE_REDIRECT_URL`, `OAUTH_MICROSOFT_REDIRECT_URL`, `OAUTH_FRONTEND_CALLBACK_URL` to Cloud Run env vars
   - URLs templated from domain variables for automatic domain-based configuration
   - Fixes OAuth provider validation rejection in staging/production

3. **Cookie Secure Flag Behind TLS Proxy**
   - Extended secure cookie detection in `auth_handler.go` to check `X-Forwarded-Proto: https` header
   - Resolves issue where `Request.TLS == nil` behind load balancer/nginx proxy, causing cookies to be marked insecure
   - Browsers now accept session cookies on HTTPS in proxied environments

4. **URL Encoding Bug in Error Redirects**
   - Wrapped authorization code in `url.QueryEscape()` in `oauth_handler.go` before constructing callback URL
   - Prevents malformed redirect URLs when codes contain special characters (&, =, etc.)

### Documentation & Configuration
- Created `docs/oauth-provider-setup.md` — Complete Google and Microsoft OAuth provider setup walkthrough with screenshots
- Updated `docs/deployment-guide.md` — Added OAuth environment variable references and configuration sections
- See `docs/journals/2026-04-05-fix-oauth-login.md` for detailed root cause analysis and lessons learned

### Test Coverage
- 87 tests pass (full test suite, no regressions)
- OAuth callback flow tests validated
- Domain validation tests (hd/tid claims) verified
- Proxy scenario tests added for TLS termination edge cases

### Related Files
- `services/core/internal/interface/http/auth_handler.go` — Secure cookie detection
- `services/core/internal/interface/http/oauth_handler.go` — URL encoding fix
- `deploy/docker/compose.yml` — OAuth environment variables
- `deploy/terraform/main.tf` — Redirect URL templating
