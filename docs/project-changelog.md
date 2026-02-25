# Myrmex Project Changelog

All notable changes to the Myrmex project are documented here.

## [2026-02-25] — Demo-in-a-Box: Docker One-Liner Deployment

**Status**: Complete

### Summary
Completed Phase 1 deployment polish: containerized entire system (4 Go services + frontend) into single-command Docker Compose setup with auto-migration and seed data. `make demo` now spins up full Myrmex at localhost:3000 (UI) + localhost:8080 (API) with zero manual configuration.

### Phase 1: Configuration & Environment Overrides
- **SetEnvKeyReplacer**: Added `v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))` to all 4 service main.go files (core, module-hr, module-subject, module-timetable)
  - Enables env var override of nested config keys (e.g., `DATABASE_URL` → `database.url`)
- **Database Credentials**: Fixed module-hr, module-subject, module-timetable config/local.yaml:
  - Changed from `postgres:postgres` to `myrmex:myrmex_dev` (matches compose)
  - Added schema-specific `search_path` (hr, subject, timetable per module)
- **SelfURL Configuration**: Made core service selfURL configurable:
  - Added `server.self_url` key to core config/local.yaml
  - Reads from env var `SERVER_SELF_URL` with fallback to `http://localhost:{port}`
  - Enables Docker internal networking (core:8080 instead of localhost:8080)

### Phase 2: Dockerfiles & Docker Compose
- **Module-HR Dockerfile**: Created new workspace-aware Dockerfile (mirrors module-subject/timetable pattern)
- **Core Dockerfile Fix**: Fixed to use workspace context (copy pkg/, gen/, go.work before service dir)
- **Frontend Nginx Config**: Created nginx-docker.conf with reverse proxy:
  - `/api/` → http://core:8080/api/ (eliminates CORS + CSP issues)
  - `/ws/` → http://core:8080/ws/ (WebSocket proxying)
  - Static assets cached, SPA fallback to /index.html
- **Docker Compose Expansion**: Expanded deploy/docker/compose.yml:
  - Services: postgres, nats, redis (infrastructure)
  - Services: migrate (goose + seed runner)
  - Services: core, module-hr, module-subject, module-timetable (Go services)
  - Services: frontend (React UI via nginx)
  - Environment vars: DATABASE_URL, NATS_URL, gRPC addresses, LLM_API_KEY override
  - Dependencies: Proper service health checks and startup ordering

### Phase 3: Makefile & Developer Experience
- **Makefile Targets**:
  - `make demo` — Start entire system (docker compose up --build -d)
  - `make demo-down` — Stop services (preserves data)
  - `make demo-logs` — Tail logs from all services
  - `make demo-reset` — Wipe database and restart fresh
- **Frontend API URLs**: Changed defaults to relative paths:
  - client.ts: `/api` (instead of hardcoded localhost:8080)
  - use-schedule-stream.ts: `/api` (same pattern)
  - use-chat.ts: `window.location.host` for WebSocket (dynamic)
  - Enables same-origin requests through nginx proxy
- **Vite Dev Server**: Added proxy config for local development:
  - `/api` → http://localhost:8080 (Go services)
  - `/ws` → ws://localhost:8080 (WebSocket)
  - Works seamlessly with existing `npm run dev` workflow
- **.env.example**: Created template for optional LLM_API_KEY configuration
- **README Updates**: Added Docker Demo section, fixed port references (8000 → 8080)

### Quality Metrics
- All services compile: `go build ./...` ✓
- Docker images build: `docker compose -f deploy/docker/compose.yml build` ✓
- Services start: `docker compose up` → all healthy ✓
- Migrations run automatically ✓
- Seed data loads (3 depts, 8 teachers, 10 subjects, 5 rooms) ✓
- Frontend accessible at localhost:3000 ✓
- API accessible at localhost:8080/api/health ✓
- No breaking changes to existing APIs

### Deployment Complexity Reduced
- **Before**: ~10 manual steps across 5+ terminals (postgres, nats, 4 services, frontend)
- **After**: `git clone && make demo` (one command, fully automated)

---

## [2026-02-25] — AI Chat Tool Executor Fix + Comprehensive Tests & CI/CD

**Status**: Complete

### Summary
Completed Phase 1 MVP: Fixed critical tool executor dispatch bug, implemented internal JWT token generation for self-referential HTTP calls, added 20+ test files achieving ≥70% coverage across all services, and created GitHub Actions CI/CD pipeline. Phase 1 status: 75% → 100%.

### Track A: AI Chat Tool Executor Fix
- **Tool Executor Refactor**: Fixed critical dispatch bug where `buildEndpoint()` now correctly returns `(url, method string)` tuple
- **HTTP Method Routing**: Implemented proper HTTP method routing:
  - `timetable.generate` → POST (triggers async generation)
  - `timetable.suggest_teachers` → GET (query-based filtering)
  - Read operations → GET; mutations → POST
- **Query Parameter Encoding**: Implemented `buildQueryParams()` using `url.Values` for proper RFC 3986 encoding
- **Internal JWT Generation**: Added `GenerateInternalToken()` in `jwt_service.go` with fixed 24h TTL for service-to-service communication
- **Self-Referential Architecture**: Tool executor now uses `selfURL` (core's HTTP base) + `internalJWT` token instead of direct gRPC; allows correct router middleware execution
- **Core Initialization**: Updated `cmd/server/main.go` to generate internal JWT at startup and pass to `NewToolExecutor()`

### Track B: Comprehensive Tests & CI/CD
- **Test Coverage**: Added 20+ test files across services:
  - Core: `auth/`, `handler/`, `gateway/` test suites
  - Module-HR: Department, teacher, availability tests
  - Module-Subject: Subject, prerequisite, DAG tests
  - Module-Timetable: Semester, room, schedule, CSP tests
- **Coverage Achievement**: All critical packages now ≥70% (many at 80-100%)
- **Bug Fix**: Fixed timetable gRPC build failure in `semester_server_test.go`
- **CI/CD Pipeline**: Created `.github/workflows/ci.yml`:
  - Go 1.26 build, vet, test (120s timeout)
  - Frontend TypeScript check (Node 20)
  - Runs on push to main + pull requests
  - All tests pass on clean build

### Quality Metrics
- Go build: `go build ./...` ✓
- Go vet: `go vet ./...` ✓
- Test coverage: >70% per service ✓
- TypeScript check: `tsc --noEmit` ✓
- CI/CD: GitHub Actions pipeline operational ✓
- No breaking changes to existing APIs

### Phase 1 Status Update
- **Overall**: 75% → 100% complete
- **All success criteria**: Achieved
- **Timeline**: On track (Q1 2026)
- **Ready for**: Phase 2 (Analytics & Reporting)

---

## [Unreleased]

### Added
- **Backend: Schedule enrichment** — Added denormalized fields to `ScheduleEntry` (subject_name, subject_code, teacher_name, room_name) via migration `007_enrich_schedule_entries.sql` for efficient API responses
- **Backend: ListSchedules RPC** — Implemented complete `ListSchedules` gRPC service with optional semester filtering and pagination support
- **Backend: HTTP handlers** — Fixed `ListSchedules` and `GetSchedule` HTTP endpoints in core gateway to properly return enriched schedule data
- **Frontend: Schedule Calendar** — Built interactive weekly timetable grid view with day columns (Mon–Sat) and period rows, color-coded by department
- **Frontend: Period utilities** — Created `period-to-time.ts` mapping periods to human-readable time strings (08:00–21:45)
- **Frontend: Manual override modal** — Integrated teacher suggestion system with assign modal for schedule adjustments
- **Database: Demo seed data** — Created `deploy/docker/seed.sql` with deterministic demo data (3 departments, 8 teachers, 10 subjects, 5 rooms, 30 time slots)
- **Build: Seed target** — Added `make seed` and `make reset-db` Makefile targets for database population

### Fixed
- **Proto: ScheduleEntry** — Extended message with enriched display fields (subject_name, subject_code, teacher_name, room_name)
- **Frontend: ScheduleEntry type** — Replaced aspirational start_time/end_time strings with actual start_period/end_period integers from API
- **SQL queries** — Updated `ListEntriesBySchedule` with JOIN to fetch time_slot and room details in single query
- **SubjectInfo model** — Added Name field to enable subject name lookups in timetable service

### Technical Details
- All 3 services compile cleanly (module-timetable, core, module-hr)
- TypeScript type check passes across frontend
- Enriched API responses support full schedule visualization without additional API calls per entry
- Seed data covers Mon–Fri, periods 1–6; easily extensible for additional time slots

---

## [2026-02-25] — Demoable Schedule Calendar Implementation

**Status**: Complete | **PR**: TBD

### Summary
Closed gap between structurally complete core and fully demoable end-to-end demo. Implemented backend schedule data enrichment, ListSchedules RPC, schedule detail view with interactive calendar grid, and comprehensive seed data for demo purposes.

### Phase 1: Backend — Enrich ScheduleEntry + ListSchedules RPC
- Created migration: `services/module-timetable/migrations/007_enrich_schedule_entries.sql`
- Updated: `entity.ScheduleEntry` with denorm + display fields (SubjectName, SubjectCode, TeacherName, DepartmentID, DayOfWeek, StartPeriod, EndPeriod, RoomName)
- Updated: SQL queries with JOIN to fetch enriched data efficiently
- Updated: `generate_schedule_handler.go` to populate denormalized names at write time
- Extended: Proto `ScheduleEntry` message with enriched fields
- Added: `ListSchedules` RPC with optional semester filter + pagination
- Implemented: `list_schedules_handler.go` query handler
- Fixed: HTTP handlers `ListSchedules` (was stub) and `GetSchedule` (JSON shape)
- Wired: Handlers in module-timetable `main.go` and core gateway

### Phase 2: Frontend — Schedule Calendar Grid
- Updated: `ScheduleEntry` type (start_period/end_period instead of time strings)
- Created: `utils/period-to-time.ts` — period → HH:MM conversion (08:00–21:45)
- Created: `utils/dept-color.ts` — deterministic dept color assignment
- Created: `components/schedule-entry-card.tsx` — cell card showing subject/teacher/room
- Created: `components/schedule-grid.tsx` — CSS Grid calendar (days × periods)
- Created: `components/assign-teacher-modal.tsx` — teacher suggestion + assignment
- Implemented: `routes/_authenticated/timetable/schedules/$id.tsx` — full schedule detail view
- Verified: `useSchedule` hook returns enriched API response

### Phase 3: Seed Data + make seed
- Created: `deploy/docker/seed.sql` — idempotent demo seed (3 depts, 8 teachers, 10 subjects, 5 rooms, 30 time slots)
- Seed structure: Departments → Teachers → Subjects → Prerequisites → Semester → Rooms → Time slots
- Teacher availability: All 8 teachers available Mon–Fri periods 1–6 (240 availability records)
- Added: `make seed` target (runs seed.sql via psql)
- Added: `make reset-db` target (goose reset → migrate → seed)

### Quality Metrics
- All services compile: `go build ./...` ✓
- TypeScript check: `npx tsc --noEmit` ✓
- Existing tests pass: `make test` ✓
- No breaking changes to existing APIs
- Seed data idempotent: safe to run multiple times

---

## [2026-02-23] — Core Infrastructure + CSP Solver Foundation

**Status**: Complete

### Summary
Established modular microservice foundation with gRPC APIs, PostgreSQL schemas, and CSP constraint solver. Implemented core auth gateway and foundational UI components.

---

## [2026-02-20] — Project Initiation

**Status**: Complete

### Summary
Repo setup, documentation structure, design patterns, development rules, and initial architecture decisions.

---

## Versioning

Myrmex uses date-based versioning: `YYYY-MM-DD` reflects implementation completion date. Semantic versioning to follow at Phase 1 GA (1.0.0).
