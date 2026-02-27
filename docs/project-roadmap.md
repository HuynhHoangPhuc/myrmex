# Myrmex Project Roadmap

## Overview

Myrmex is a multi-phase project to build an agent-first ERP for educational institutions. This roadmap outlines the MVP (Phase 1) and planned future phases.

## Phase 1: MVP - University Faculty Management (COMPLETE)

**Timeline**: Q1 2026 (4 weeks) | **Status**: 100% Complete

### Goals
- Establish modular microservice foundation
- Deliver core faculty management capabilities (departments, teachers, subjects, schedules)
- Integrate AI agent for conversational operations
- Achieve 95% CSP solver success rate for schedule generation

### Deliverables

#### Completed (Feb 2026)
- [x] Core service: HTTP gateway, JWT auth, module registry
- [x] Module-HR: Department & teacher CRUD with availability tracking
- [x] Module-Subject: Subject CRUD + prerequisite DAG (cycle detection, topological sort)
- [x] Module-Timetable: Semester & room management
- [x] Database schemas: PostgreSQL schema-per-module with event sourcing
- [x] gRPC definitions: All services proto files (buf-managed)
- [x] Docker Compose: PostgreSQL, NATS, Redis infrastructure
- [x] Frontend foundation: React 19 + TanStack Router/Query/Form/Table
- [x] Frontend auth: Login, register, token management
- [x] Frontend modules: HR, Subject, Timetable modules with CRUD UI

#### In Progress / Completed
- [x] CSP Solver: Backtracking + heuristics (AC-3 + MRV + LCV)
- [x] Schedule Generation API: gRPC endpoint + async status tracking
- [x] Schedule UI: Calendar view with manual override + teacher suggestions
- [x] AI Chat integration: Claude Haiku 4.5 provider + tool registry (with self-referential HTTP dispatch fix)
- [x] Tool registry: Domain operations (create subject, assign teacher, generate schedule)
- [x] WebSocket chat: Streaming responses + auto-reconnect
- [x] Service + pkg test suites: 20+ test files with ≥70% coverage across core, HR, Subject, Timetable modules
- [x] E2E workflow: Register → Create subject → Assign teacher → Generate schedule
- [x] Unit tests: >70% coverage per service (achieved)
- [x] Integration tests: Database + gRPC interactions
- [x] Seed data: Sample departments, subjects, teachers, semesters
- [x] CI/CD pipeline: GitHub Actions for build, test, lint (Go 1.26 + Frontend TypeScript check)

### Success Criteria
- [x] All FR-1 through FR-10 implemented and tested
- [x] Schedule generation success rate >95%
- [x] API availability >99.5%
- [x] Unit test coverage >70% (achieved: 70-100% across services)
- [x] Mean API response time <300ms (p50)
- [x] CSP solver <30s (p95)
- [x] Frontend fully functional (departments, teachers, subjects, schedules)
- [x] Deploy via Docker Compose with single `make up` command
- [x] Complete documentation: architecture, code standards, API, deployment

### Key Metrics (Target)
| Metric | Target | Notes |
|--------|--------|-------|
| **Features Complete** | 10/10 (100%) | FR-1 through FR-10 |
| **Code Coverage** | >70% | Per service unit tests |
| **API Availability** | >99.5% | Uptime SLA |
| **Mean Latency** | <300ms (p50) | Excl. CSP solver |
| **CSP Solver** | <30s (p95) | Constraint satisfaction |
| **Frontend Accessibility** | WCAG 2.1 AA | Keyboard nav, screen readers |
| **Security** | Zero critical CVEs | Dependencies, secrets |

---

## Phase 2: Analytics & Reporting (COMPLETE)

**Timeline**: Q2 2026 (2-3 weeks) | **Status**: 100% Complete (Feb 26)

### Goals
- Provide insights into resource utilization and schedule efficiency
- Enable data-driven decision making for faculty planning
- Export schedules and reports in multiple formats

### Deliverables

#### Analytics Dashboard (COMPLETE)
- [x] Workload analytics: Hours per teacher, utilization %, capacity analysis
- [x] Dashboard KPI cards: Teacher count, avg workload, schedule completion %
- [x] Utilization metrics: Room occupancy, teacher load distribution
- [x] Department metrics: Teachers per department, specialization coverage

#### Reporting (COMPLETE)
- [x] Schedule export: PDF (printable) via iText
- [x] Schedule export: Excel (editable) with multi-sheet layout
- [x] Workload report: Per-teacher summary with period breakdown
- [x] Analytics query API: `/api/analytics/workload`, `/api/analytics/utilization`

#### Infrastructure (COMPLETE)
- [x] Module-Analytics service: New Go service with HTTP API
- [x] Analytics database schema: Star-schema (dim_teacher, dim_subject, dim_department, dim_semester, fact_schedule_entry)
- [x] ETL pipeline: NATS event consumer → analytics schema (real-time + on-demand)
- [x] Dashboard UI: React components with KPI cards, charts, semester filter
- [x] Report generation: Server-side PDF/Excel (iText-based)
- [x] NATS event publishing: HR/Subject/Timetable modules emit events consumed by analytics

### Success Criteria
- [x] 10+ key analytics metrics available (KPI cards, workload, utilization)
- [x] Export functionality: PDF + Excel for schedules
- [x] ETL real-time (event-driven via NATS JetStream)
- [x] Report generation <2 seconds
- [x] Dashboard load time <500ms
- [x] Analytics module integrated with core gateway

---

## Phase 3: Advanced Features (IN PROGRESS)

**Timeline**: Q3 2026 (4-5 weeks) | **Status**: 25% Complete (Advanced Prerequisites sub-phase done, Q2.5)

### Goals
- Implement advanced prerequisite conflict detection (DONE)
- Expand system to include student management
- Enable grade tracking and academic progress monitoring
- Improve UX with mobile support and drag-drop scheduling

### Deliverables

#### Advanced Prerequisite Management (COMPLETE - Feb 27)
- [x] DAG visualization: React Flow interactive rendering with zoom/pan/minimap
- [x] Conflict detection: POST /api/subjects/dag/check-conflicts API
- [x] Full DAG endpoint: GET /api/subjects/dag/full (all subjects + edges)
- [x] Conflict UI: ConflictWarningBanner + offering-manager integration
- [x] Focus mode: Subject detail page shows transitive prerequisites
- [x] Hover highlighting: Ancestor chain on node hover
- [x] Tests: 6 conflict detection tests + 7 banner component tests
- [x] Proto enhancements: Prerequisite.type + priority fields

#### Student Management Module
- [ ] Module-Student: Student CRUD, enrollment, grades
- [ ] Enrollment workflow: Students enroll in offered subjects
- [ ] Grade tracking: Teachers input grades per student per subject
- [ ] Transcript generation: Student academic history export
- [ ] Prerequisite validation: Prevent enrollment if prerequisites not met

#### Mobile & UX Enhancements
- [ ] Mobile app: React Native (iOS + Android)
- [ ] Drag-drop scheduling: Reassign teachers/rooms via drag-drop
- [ ] Hamburger menu: Mobile navigation (sidebar collapse)
- [ ] Offline mode: Cache schedules for offline access
- [ ] Push notifications: Schedule changes, new messages

#### Notifications System
- [ ] Email notifications: Schedule changes, assignments
- [ ] SMS alerts: Critical schedule changes (opt-in)
- [ ] In-app notifications: Real-time updates
- [ ] Notification preferences: User-configurable channels

### Success Criteria
- [ ] Student enrollment workflow functional
- [ ] Grade tracking complete with transcript export
- [ ] Mobile app (iOS/Android) deployed
- [ ] Prerequisite conflict detection accuracy >99%
- [ ] Notification delivery rate >99%

---

## Phase 4: Enterprise & Multi-Tenancy (PLANNED)

**Timeline**: Q4 2026+ (6+ weeks) | **Status**: Vision

### Goals
- Support multiple institutions (universities, schools, organizations)
- Enable advanced RBAC and permission management
- Achieve enterprise SLA (99.9% uptime, HA/DR)
- Implement audit logging and compliance features

### Deliverables

#### Multi-Tenancy
- [ ] Tenant isolation: Shared infrastructure, isolated data (row-level security)
- [ ] Tenant management: Admin UI for creating/managing tenants
- [ ] Billing integration: Stripe/Paddle for subscription management
- [ ] SLA tiers: Basic, Pro, Enterprise with feature gates

#### Advanced RBAC
- [ ] Permission model: Fine-grained permissions per resource (CRUDX)
- [ ] Role templates: Pre-built roles (admin, dean, department_head, instructor, student)
- [ ] Custom roles: Ability to define custom roles + permissions
- [ ] Delegation: Users can delegate permissions to others (time-limited)

#### Integrations
- [ ] LDAP/SAML: Enterprise single sign-on
- [ ] Google Workspace: Calendar sync, student email provisioning
- [ ] LMS integration: Canvas, Blackboard, Moodle integration
- [ ] Data warehouse: Snowflake, BigQuery export for analytics

#### High Availability & Disaster Recovery
- [ ] PostgreSQL replication: Primary + standby (streaming replication)
- [ ] NATS clustering: 3+ node cluster for HA
- [ ] Geographic distribution: Multi-region deployment (k8s)
- [ ] Backup strategy: Automated daily backups, 30-day retention
- [ ] RTO/RPO: Recovery time <1 hour, recovery point <15 min

#### Audit & Compliance
- [ ] Audit logs: All mutations logged with user, timestamp, changes
- [ ] Data retention: GDPR-compliant retention policies
- [ ] Encryption: End-to-end encryption for sensitive data (PII)
- [ ] Compliance certifications: SOC 2 Type II, GDPR, FERPA
- [ ] Data export: GDPR right-to-be-forgotten, data portability

### Success Criteria
- [ ] 3+ institutions on production multi-tenant instance
- [ ] RBAC permissions library >50 unique permissions
- [ ] Single sign-on (LDAP/SAML) functional
- [ ] 99.9% uptime SLA achieved
- [ ] Audit logs comprehensive (100% mutation coverage)
- [ ] SOC 2 Type II certification obtained

---

## Quarterly Milestones

### Q1 2026 (Jan-Mar)
- **Feb 20**: Project initiation, architecture design, team onboarding
- **Feb 28**: Core service, Module-HR, Module-Subject, Module-Timetable prototypes
- **Mar 15**: Frontend MVP (auth, CRUD UI)
- **Mar 28**: CSP solver + schedule generation working
- **Mar 30**: Phase 1 MVP launch (internal use)

### Q2 2026 (Apr-Jun)
- **Apr 15**: E2E testing, bug fixes, performance tuning
- **May 1**: Phase 1 launch to early adopters (beta)
- **May 15**: Phase 2 kick-off: Analytics dashboard design
- **Jun 15**: Phase 2 alpha: Dashboard + reporting features
- **Jun 30**: Phase 2 launch (production)

### Q3 2026 (Jul-Sep)
- **Jul 1**: Phase 3 design: Student management, mobile
- **Jul 31**: Student module MVP, React Native scaffold
- **Aug 31**: Mobile app (iOS/Android) beta
- **Sep 15**: Prerequisite conflict detection MVP
- **Sep 30**: Phase 3 launch

### Q4 2026 (Oct-Dec)
- **Oct 1**: Phase 4 planning: Multi-tenancy, RBAC, HA
- **Oct 31**: Multi-tenant infrastructure design + POC
- **Nov 30**: LDAP/SAML integration working
- **Dec 15**: PostgreSQL replication + NATS clustering
- **Dec 30**: Phase 4 alpha: Multi-tenant platform ready

---

## Dependency Graph

```
Phase 1: MVP
├── Core (auth, gateway)
├── Module-HR (departments, teachers)
├── Module-Subject (subjects, prerequisites)
├── Module-Timetable (semesters, rooms)
│   └── Depends on: Module-Subject, Module-HR
├── Frontend (CRUD UI, auth)
└── CSP Solver (schedule generation)

Phase 2: Analytics & Reporting
└── Depends on: Phase 1 MVP

Phase 3: Advanced Features
├── Student Module
│   └── Depends on: Phase 1 (for data models)
├── Mobile App (React Native)
│   └── Depends on: Phase 1 API
├── Advanced Prerequisites
│   └── Depends on: Module-Subject
└── Notifications
    └── Depends on: Phase 1 (events)

Phase 4: Enterprise
├── Multi-Tenancy (all modules)
│   └── Depends on: Phases 1-3
├── Advanced RBAC
│   └── Depends on: Phase 1 (core)
├── Integrations (LDAP, LMS)
│   └── Depends on: All services
├── HA/DR (clustering, replication)
│   └── Depends on: Phase 1 infrastructure
└── Audit & Compliance
    └── Depends on: Event sourcing (Phase 1)
```

---

## Resource & Skill Requirements

### MVP Phase 1
- **Backend Developers**: 2 (Go, gRPC, PostgreSQL)
- **Frontend Developers**: 1 (React, TypeScript)
- **DevOps**: 1 (Docker, CI/CD, database)
- **QA/Testing**: 0.5 (manual + automation)
- **Product Manager**: 0.5 (requirements, prioritization)

### Phase 2-4
- Backend: +1 per phase (new modules)
- Frontend: +1 per phase (new UI)
- DevOps: +0.5 per phase (HA/multi-region)
- Data engineer: +1 (Phase 2+)
- Security engineer: +1 (Phase 4)

---

## Risk Assessment & Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| **CSP solver timeout** | Medium | High | Partial solution strategy; fallback to manual assignment |
| **Database scalability** | Low | High | Connection pooling; read replicas; sharding (Phase 4) |
| **NATS JetStream limits** | Low | Medium | Monitor throughput; horizontal scaling (Phase 2) |
| **Prerequisite cycle bugs** | Low | Medium | Comprehensive unit tests; frontend DAG validation |
| **LLM API rate limits** | Medium | Medium | Token bucket; request queuing; fallback to simpler prompts |
| **Frontend performance** | Low | Medium | Code splitting (TanStack Router); lazy loading |
| **Auth token conflicts** | Low | High | Short TTL (15min); refresh before expiry; logout on 401 |
| **Data consistency** | Low | High | Event sourcing; optimistic concurrency; transaction logs |
| **Scope creep** | High | High | Prioritize MVP; defer Phase 2+ features; feature gates |

---

## Success Metrics & KPIs

### Product Metrics
- **User Adoption**: X% of target institutions using within 6 months
- **Feature Usage**: Average features used per institution
- **Schedule Success**: % of semesters successfully scheduled without manual override
- **User Satisfaction**: NPS >40 by end of Phase 1

### Technical Metrics
- **Uptime**: 99.5% (Phase 1), 99.9% (Phase 4)
- **Performance**: API latency <300ms (p50), CSP <30s (p95)
- **Code Quality**: >70% test coverage, <5 critical bugs per release
- **Security**: Zero critical CVEs, SOC 2 certified (Phase 4)

### Business Metrics
- **Revenue**: Subscription revenue from X institutions (Phase 4)
- **Cost**: Cloud infrastructure cost per institution
- **Operational**: Support tickets <X per week, MTTR <2 hours

---

## Change Log

### 2026-02-27 (Timetable, AI Chat & Teacher Availability Fixes)
- Fixed schedule generation HTTP response: Now returns full schedule object instead of `{schedule_id}`
- Fixed SQL WHERE clause bug in `ListSchedulesPaged` (operator precedence)
- Added schedule status tracking: `generating` → `completed` or `failed`
- Enriched semester response with `time_slots` and `rooms` via gRPC
- Fixed AI chat system prompt: Explicit workflow for `list_semesters` before `generate`
- Increased AI tool iterations from 5 to 10 for multi-step workflows
- Added `timetable.list_semesters` tool to tool registry
- Fixed teacher availability representation: Time slots (HH:MM) instead of period integers
- Phase 1 status: Timetable + AI features fully functional and bug-free

### 2026-02-26 (API Response Standardization & Frontend Fixes)
- Standardized all list endpoints to return `{ data, total, page, page_size }` format
- Fixed HR, Subject, and Timetable response formats for consistency
- Added new dashboard stats endpoint (`GET /api/dashboard/stats`)
- Added `Me()` endpoint for current user profile (`GET /api/auth/me`)
- Added proto fields: teacher (`employee_code`, `max_hours_per_week`, `specializations`, `phone`), subject (`weekly_hours`, `is_active`)
- Rewrote semester form to separate year/term input + RFC3339 date handling
- Fixed schedule endpoints: suggest-teachers query params, manual-assign URL
- Rewrote offering-manager component for per-item add/remove with server sync
- Semester response now includes: `year, term, academic_year, is_active, offered_subject_ids`
- Phase 1 status: 100% → API contract aligned and production-ready

### 2026-02-25 (Demo-in-a-Box: Docker One-Liner Deployment)
- Added `v.SetEnvKeyReplacer` to all 4 services (core, HR, subject, timetable) for env var override support
- Fixed DB credentials in module-hr, module-subject, module-timetable config/local.yaml (myrmex:myrmex_dev + search_path)
- Made selfURL configurable in core service (reads `server.self_url` with fallback)
- Created Dockerfile for module-hr (workspace-aware build)
- Fixed core Dockerfile for workspace-aware build
- Created frontend nginx-docker.conf with reverse proxy (/api and /ws routes)
- Expanded deploy/docker/compose.yml with all 4 services + frontend + migrate service
- Added `make demo`, `make demo-down`, `make demo-logs`, `make demo-reset` targets
- Changed frontend API defaults to relative paths (/api) with Vite dev proxy
- Fixed WebSocket URL construction for relative paths via window.location.host
- Updated README: added Docker Demo section, fixed port references (8000 → 8080)
- Created .env.example for optional LLM_API_KEY
- Phase 1 deployment polish: 100% → Ready for Demo-in-a-Box

### 2026-02-25 (Demoable Schedule Calendar Implementation)
- Completed schedule data enrichment (denormalized fields in ScheduleEntry)
- Implemented ListSchedules RPC with pagination + semester filtering
- Built interactive schedule calendar UI (weekly grid, color-coded departments)
- Created comprehensive seed data (3 depts, 8 teachers, 10 subjects, 5 rooms)
- Added `make seed` target for easy database population
- Phase 1 progress: 65% → 75% (schedule generation + UI now fully functional)

### 2026-02-21 (Initial Roadmap)
- Created Phase 1-4 roadmap
- Set MVP target: Feb-Mar 2026
- Defined Phase 2-4 vision
- Established success criteria
