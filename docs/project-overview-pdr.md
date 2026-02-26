# Myrmex Project Overview & PDR

## Executive Summary

**Myrmex** is an agent-first, modular ERP system designed for educational institutions. Built on Go microservices with AI-powered operations, it enables intelligent scheduling, resource management, and conversational workflows. MVP delivers university faculty management (HR, Subjects, Timetable).

## Vision & Goals

### Vision
Create an extensible, AI-native ERP that lets institutions operate via conversation. Unlike traditional ERPs (Odoo, SAP), Myrmex prioritizes agent-driven operations and microservice modularity.

### Goals
1. **Intelligent Scheduling**: Auto-generate conflict-free timetables with CSP solver; teacher-friendly manual overrides
2. **Modular Architecture**: Add modules without redeploying core; independent data schemas
3. **Conversational Operations**: Use ChatGPT/Claude to create subjects, assign teachers, generate schedules
4. **Scalable Design**: gRPC + event sourcing + async messaging support 1000+ concurrent users per service

## Product Requirements Document (PRD)

### Stakeholders
| Role | Needs |
|------|-------|
| **Admin** | Module management, user roles, audit logs, system configuration |
| **Faculty Coordinator** | Create departments, manage teachers, assign specializations |
| **Academician** | Define subjects, set prerequisites (DAG), manage course structure |
| **Scheduler** | Generate timetables, override assignments, manage rooms & semesters |
| **AI User** | Chat-driven subject creation, schedule generation, teacher suggestions |

### Functional Requirements (MVP Phase 1)

#### FR-1: Authentication & Authorization
- User registration & login (email/password, hashed bcrypt)
- JWT tokens: 15min access + 7day refresh
- Role-based access control (admin, faculty_coordinator, academician, scheduler)
- Token refresh endpoint with automatic logout on 401

#### FR-2: Department Management (HR Module)
- CRUD operations for departments
- Soft delete (retention for audit)
- Listing with pagination
- Query by department_id or name

#### FR-3: Teacher Management (HR Module)
- CRUD for teachers with:
  - Name, email, phone
  - Department affiliation
  - Specializations (many-to-many)
  - Availability (day_of_week + period_of_day)
- Soft delete
- Listing with filtering by department
- Pagination support

#### FR-4: Subject Management (Subject Module)
- CRUD for subjects with:
  - Code, name, credits, weekly_hours
  - Description, department_id
  - Prerequisite management
- Listing with pagination
- Query by department

#### FR-5: Prerequisite DAG (Subject Module)
- Add prerequisites to subjects (strict, recommended, corequisite)
- Priority levels (1-5 soft: -2 to +2 hard)
- Cycle detection & validation
- Topological sort for prerequisites
- Removal of prerequisite links
- Frontend DAG visualization

#### FR-6: Semester Management (Timetable Module)
- CRUD for semesters (name, year, term, date_range, offered_subjects)
- Associate subjects with semester
- Add/remove offerings
- Soft delete

#### FR-7: Room Management (Timetable Module)
- CRUD for rooms (code, capacity)
- Listing with filtering

#### FR-8: Schedule Generation (Timetable Module)
- Trigger CSP solver for semester
- Input: semester, offered subjects, available teachers, rooms
- Output: conflict-free schedule entries
- Constraints:
  - **Hard**: No teacher time conflicts, specialization match, room capacity
  - **Soft**: Teacher preferences, workload balance
- Timeout: 30s context cancellation → return best partial solution
- Status tracking: pending → generating → completed/failed

#### FR-9: Schedule Management (Timetable Module)
- View generated schedules
- Manual override (drag-drop in UI, API PATCH)
- Teacher suggestion ranking
- Validation on override (hard constraints only)

#### FR-10: AI Chat Interface (Core Service)
- WebSocket endpoint `/ws/chat?token=ACCESS_TOKEN`
- Tool registry for domain operations:
  - Create subject, add teacher, generate schedule, etc.
- Multi-LLM support: Claude (default), OpenAI (configurable)
- Message persistence to PostgreSQL
- Streaming responses via WebSocket

### Non-Functional Requirements

| Requirement | Target | Notes |
|-------------|--------|-------|
| **Availability** | 99.5% uptime | Per module |
| **Response Time** | <500ms (p95) | API endpoints excl. CSP solver |
| **CSP Solver Time** | <30s (p95) | Config: AC-3 + backtracking |
| **Scalability** | 1000 concurrent users/service | Horizontal scaling via k8s |
| **Database** | PostgreSQL 16+, event sourcing | Shared DB, schema-per-module |
| **Security** | JWT + HTTPS + bcrypt | No plain-text passwords |
| **Monitoring** | Logs (Zap), metrics (Prometheus ready) | Per module, JSON structured logs |
| **Testability** | >70% unit test coverage | Per module |

### Data Models (MVP)

**Core DB Schema:**
- users (id, email, password_hash, created_at, role)
- module_registry (id, name, version, health_check_url, grpc_addr)
- conversations (id, user_id, messages, created_at)
- event_store (aggregate_id, aggregate_type, event_type, payload, timestamp)

**HR DB Schema:**
- departments (id, name, created_at, deleted_at)
- teachers (id, name, email, department_id, created_at, deleted_at)
- teacher_availability (teacher_id, day_of_week, period_of_day)
- teacher_specializations (teacher_id, subject_id)
- event_store (same pattern)

**Subject DB Schema:**
- subjects (id, code, name, credits, weekly_hours, department_id, created_at)
- prerequisites (id, subject_id, prerequisite_subject_id, type, priority)
- event_store

**Timetable DB Schema:**
- semesters (id, name, year, term, start_date, end_date, created_at)
- semester_offerings (semester_id, subject_id)
- rooms (id, code, capacity)
- schedules (id, semester_id, status, created_at, completed_at)
- schedule_entries (id, schedule_id, subject_id, teacher_id, room_id, day_of_week, period_of_day, week_of_semester)
- time_slots (day_of_week, period_of_day) [reference data]
- event_store

### Success Metrics

| Metric | Target | How Measured |
|--------|--------|--------------|
| **Schedule Generation Success Rate** | >95% | (successful CSP completions) / (total runs) |
| **Prerequisite DAG Accuracy** | 100% | Zero undetected cycles in validation tests |
| **API Availability** | 99.5% | (uptime minutes) / (total minutes) |
| **Mean Response Time (API)** | <300ms | p50 latency across all endpoints |
| **Unit Test Coverage** | >70% | Per-module coverage report |
| **User Onboarding Time** | <15 min | Time to first schedule generation |
| **CSP Solver Efficiency** | <30s (p95) | Duration of schedule generation |

## Acceptance Criteria

### Phase 1 (MVP) - University Faculty Management
- [x] All FR-1 through FR-10 implemented and tested
- [x] Frontend UI complete for all modules (departments, teachers, subjects, semesters, schedules)
- [x] Docker Compose local dev environment fully functional
- [x] CI/CD pipeline: proto lint, build, test on PR
- [x] Documentation: code standards, system architecture, deployment guide, API docs
- [x] Seed data: sample departments, subjects, prerequisites, teachers
- [x] E2E test: register → login → create subject → assign teacher → generate schedule
- [x] API response format standardization (Feb 26)
- [x] Proto field additions for enriched entity models (Feb 26)

### Phase 2 (Post-MVP) - Analytics & Reporting
- [ ] Workload analytics: hours per teacher, utilization metrics
- [ ] Conflict reports: prerequisites, capacity violations
- [ ] Export schedules (PDF, Excel)
- [ ] Dashboard with KPIs

### Phase 3 - Advanced Features
- [ ] Mobile app (React Native)
- [ ] Student enrollment & grades
- [ ] Automatic prerequisite conflict detection
- [ ] Real-time schedule collaboration

### Phase 4 - Enterprise
- [ ] Multi-tenant support
- [ ] LDAP/SAML integration
- [ ] Advanced RBAC & permissions
- [ ] Audit logging & compliance (GDPR, HIPAA)

## Technical Decisions

### Why Go?
- High performance, concurrent request handling
- Simple deployment (single binary per service)
- Strong typing, excellent error handling
- Native gRPC support

### Why Microservices?
- Independent scaling per module
- Isolated data schemas → schema evolution without global migration
- Clear separation of concerns (DDD boundaries)
- Easy to add new modules without redeploying core

### Why Event Sourcing?
- Complete audit trail of all write operations
- Replay capability for debugging & analytics
- Foundation for CQRS (eventual consistency)
- Foundation for multi-tenant separation (per-aggregate partition)

### Why CSP Solver (vs Greedy/Heuristic)?
- Guarantees better conflict resolution than greedy algorithms
- AC-3 preprocessing dramatically reduces search space
- Backtracking + heuristics (MRV, LCV) finds good solutions quickly
- Timeout-safe: partial solutions acceptable for large problems

### Why PostgreSQL (vs MongoDB)?
- **MVP**: Relational schema, ACID guarantees, event sourcing simplicity
- **Future**: MongoDB for business chat history (non-critical, high volume)
- Shared DB simplifies deployment (one Postgres instance for all modules)

### Why NATS JetStream (vs Kafka)?
- Lightweight, easy to operate (single binary)
- JetStream provides persistence & ordering
- Lower latency than Kafka for MVP scale
- Easy horizontal scaling later

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| CSP solver timeout causes incomplete schedules | Medium | High | Partial solution + UI warning; fallback to manual assignment |
| Prerequisite cycle in frontend DAG causes crashes | Low | Medium | Server-side cycle detection; frontend fallback to list view |
| Auth token expiry mid-session | Low | Medium | Auto-refresh on 401; clear localStorage on refresh failure |
| Database connection pooling exhaustion | Low | High | Monitor pool stats; configure per-service limits |
| AI chat tool registration conflicts | Low | High | Namespaced tool IDs; unit test all tool registrations |

## Dependencies

### External Services
- Claude API (or OpenAI) for LLM inference
- PostgreSQL 16 for data persistence
- NATS JetStream for async messaging

### Internal Modules
- Core (depends on: nothing)
- Module-HR (depends on: Core pkg, NATS, PostgreSQL)
- Module-Subject (depends on: Core pkg, NATS, PostgreSQL)
- Module-Timetable (depends on: Core pkg, Module-Subject gRPC, NATS, PostgreSQL)

## Timeline (Estimated)

| Phase | Duration | Status |
|-------|----------|--------|
| Phase 1: MVP (Departments, Teachers, Subjects, Timetable) | 4 weeks | In Progress |
| Phase 2: Analytics & Reporting | 2 weeks | Planned |
| Phase 3: Advanced Features | 4 weeks | Planned |
| Phase 4: Enterprise | 6 weeks | Planned |

## Glossary

| Term | Definition |
|------|-----------|
| **Aggregate** | DDD concept: root entity controlling consistency boundary (e.g., Teacher) |
| **CQRS** | Command Query Responsibility Segregation: separate read/write models |
| **DAG** | Directed Acyclic Graph: prerequisite structure with no cycles |
| **CSP** | Constraint Satisfaction Problem: assign values to variables under constraints |
| **Event Sourcing** | Store all changes as immutable events in event store |
| **gRPC** | RPC framework using HTTP/2 + protobuf for inter-service communication |
| **JetStream** | NATS persistence & ordering layer for event streaming |
| **Soft Delete** | Mark as deleted without removing from DB (retention, audit) |
| **Optimistic Concurrency** | Assume no conflicts; detect & retry on write conflicts |
