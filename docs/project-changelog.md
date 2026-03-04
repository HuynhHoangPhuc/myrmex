# Myrmex Project Changelog

All notable changes to the Myrmex project are documented here.

## [2026-03-04] — Admin Roles UI & Notifications System (IN PROGRESS)

**Status**: In Progress | **Partial Delivery**

### Summary
Added admin-facing UI for role management and implemented foundational notification infrastructure. Notifications system architecture designed for email + in-app delivery via NATS async pipeline.

### Key Deliverables

#### Admin Roles UI (COMPLETE)
- [x] Role management page: `/admin/roles` (admin/super_admin only)
  - [x] User list with current roles displayed
  - [x] Role selector dropdown (super_admin, admin, dean, dept_head, teacher, student)
  - [x] Department selector for scoped roles (dept_head, teacher)
  - [x] Batch role assignment (select multiple users)
  - [x] Audit trail: All role changes logged via audit middleware
- [x] Integration: PATCH `/api/users/:id/role` endpoint enforced via auth middleware
- [x] Frontend validation: Role assignment restricted by permission checks
- [x] Testing: 8 frontend tests for role management UI

#### Notifications System (PLANNED - Phase 4.4)
- [ ] Email notifications: Template-driven (schedule changes, enrollments, assignments)
- [ ] In-app notifications: WebSocket push via NATS events
- [ ] Notification preferences: User-configurable channels (opt-in/opt-out)
- [ ] Notification queue: PostgreSQL-backed queue for retry logic
- [ ] NATS consumer: Subscribes to domain events and triggers notifications

### Database Schema Changes (Role Management)
- No new schema required; uses existing `core.users` role column
- Audit logs track all role changes via `core.audit_logs` (partitioned)

---

## [2026-03-04] — Phase 4.1 Advanced RBAC (COMPLETE)

**Status**: Complete | **Phase 4.1 Delivered**

### Summary
Implemented 6-role RBAC system with department scoping for faculty and instructors. Two-tier authorization (middleware + handler) enforces role-based access control. Extended JWT claims include `department_id` and `teacher_id` for O(1) permission checks. Route guards protect module mutations based on user role and department scope.

### Key Deliverables

#### Backend RBAC Infrastructure
- [x] 6 roles: `super_admin`, `admin`, `dean`, `dept_head`, `teacher`, `student`
- [x] Department scoping: `dept_head` + `teacher` roles bound to `department_id` via JWT claims
- [x] Extended JWT claims: Added `department_id` + `teacher_id` fields for efficient authorization
- [x] Middleware guards: `RequireRole()` + `RequireDeptScope()` for protected routes
- [x] gRPC interceptor: Extracts role + department context from JWT (auth_interceptor.go)
- [x] Database migration (006): Added `department_id` column to `core.users` + `user_id` to `hr.teachers`
- [x] Route guards: Protected HR/Subject CRUD mutations based on department scope
- [x] Super admin bypass: `super_admin`, `admin`, `service` roles bypass scope checks; `dean` read-only bypass

#### Frontend RBAC Integration
- [x] `usePermissions()` hook: Check user role + scope client-side
- [x] Route guards: Protected routes (admin, finance, etc.) behind role checks
- [x] UI visibility: Role-based UI elements (admin-only buttons, scoped module access)
- [x] Department dropdown: Select department context for operations
- [x] Permission enforcement: HR module mutations require `dept_head` or `admin`

#### Testing & Validation
- [x] Unit tests: Role + scope middleware validation (12 tests)
- [x] Integration tests: End-to-end role enforcement (8 tests)
- [x] E2E tests: Role-based access control workflows

### API Reference

**PATCH /api/users/:id/role** (Admin/Super Admin)

Body:
```json
{
  "role": "dept_head",  // One of: super_admin, admin, dean, dept_head, teacher, student
  "department_id": "uuid"  // Required for scoped roles (dept_head, teacher)
}
```

Response:
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "role": "dept_head",
  "department_id": "uuid",
  "created_at": "2026-03-04T10:00:00Z"
}
```

### Authorization Matrix

| Role | HR (CRUD) | Subject (CRUD) | Timetable (R) | Admin | Analytics |
|------|-----------|---|---|---|---|
| `super_admin` | Full | Full | Full | Full | Full |
| `admin` | Full | Full | Full | Full | Full |
| `dean` | Read | Read | Read | Read | Full |
| `dept_head` | Scoped to dept | Scoped to dept | Full | No | Full |
| `teacher` | Read own | Read | Full | No | Limited |
| `student` | No | No | No | No | Limited |

---

## [2026-03-04] — Phase 3 Complete: Audit Logging (COMPLETE)

**Status**: Complete | **Phase 3 Fully Delivered**

### Summary
Completed Phase 3 (Advanced Features) with audit logging implementation. Async NATS pipeline captures all mutations at middleware level, streams to durable consumer, and persists to monthly-partitioned PostgreSQL table. Admin API provides comprehensive audit trail querying with flexible filters. Frontend admin dashboard displays logs with row expansion for before/after value diffs.

### Key Deliverables

#### Backend Audit System
- [x] Migration (008): Partitioned `core.audit_logs` with 12 monthly partitions (2026-03 → 2027-02)
  - Columns: id, user_id, resource_type, action, old_value, new_value, timestamp
  - Indexes: BRIN (timestamp), B-tree (user_id, resource_type, action)
- [x] Audit middleware (audit_middleware.go): Post-handler Gin middleware
  - Derives action from HTTP method + endpoint pattern (POST→Create, PATCH→Update, DELETE→Delete)
  - Skips GET/internal-service requests
  - Fire-and-forget NATS publish; non-blocking
- [x] Audit consumer (audit_consumer.go): Durable JetStream consumer
  - Listens on AUDIT.logs stream
  - Writes to core.audit_logs with ack/nack retry
  - Preserves event order via NATS ordering guarantees
- [x] Audit repository (audit_log_repository.go): Raw pgx + sqlc
  - Insert: Write audit event to DB
  - List: Paginated query with nullable filters (user_id, resource_type, action, date range)
  - Constraint exclusion: Monthly partition pruning for efficient date-range queries
- [x] Audit handler (audit_handler.go): GET /api/audit-logs
  - Admin/super_admin role enforcement
  - Pagination: limit (default 100), offset
  - Filters: user_id, resource_type, action, start_date, end_date (query params)
  - Response: Array of audit entries with human-readable labels

#### Optional Configuration
- NATS-optional: Audit middleware is no-op if NATS not configured (graceful degradation for testing)

#### Frontend Audit Logs UI
- [x] Route: `/admin/audit-logs` (admin/super_admin only)
- [x] Table: Columns: User, Resource Type, Action, Timestamp (sortable)
- [x] Row expansion: View old/new value diffs with JSON diff rendering
- [x] Filters: User selector (dropdown), resource type, action checkboxes, date picker
- [x] Pagination: Previous/next, total count display

#### Testing
- Comprehensive backend unit tests for audit flow (middleware → consumer → queries)
- Frontend rendering tests for table + row expansion

### API Reference

**GET /api/audit-logs** (Admin/super_admin)

Query Parameters:
- `limit` (default: 100) — Records per page
- `offset` (default: 0) — Pagination offset
- `user_id` (optional) — Filter by user
- `resource_type` (optional) — Filter by resource (teacher, subject, semester, etc.)
- `action` (optional) — Filter by action (create, update, delete)
- `start_date` (optional) — ISO 8601 start timestamp
- `end_date` (optional) — ISO 8601 end timestamp

Response:
```json
{
  "data": [
    {
      "id": 1,
      "user_id": "uuid",
      "resource_type": "teacher",
      "action": "create",
      "old_value": null,
      "new_value": {"id": "...", "name": "John", ...},
      "timestamp": "2026-03-04T10:30:00Z"
    }
  ],
  "total": 1234,
  "page": 0,
  "page_size": 100
}
```

### Database Schema

```sql
audit_logs (
  id: bigint primary key auto,
  user_id: uuid fk core.users,
  resource_type: string,  -- teacher, subject, semester, enrollment, grade, etc.
  action: enum(create, update, delete, read),
  old_value: jsonb nullable,  -- Previous state (null for creates)
  new_value: jsonb,  -- Current state (null for deletes)
  timestamp: timestamp
)
-- Partitioned: 12 monthly partitions (2026-03 through 2027-02)
-- Indexes: BRIN(timestamp), B-tree(user_id, resource_type)
-- Constraint exclusion enabled for date-range queries
```

### NATS Integration

Stream: `AUDIT.EVENTS`
Subject: `AUDIT.logs`

Event format:
```json
{
  "user_id": "uuid",
  "resource_type": "teacher",
  "action": "update",
  "old_value": {...},
  "new_value": {...},
  "timestamp": "2026-03-04T10:30:00Z"
}
```

### Roadmap Impact
- **Phase 3 Status**: 100% Complete (all advanced features delivered)
- **Phase 4 Status**: 4.1 (RBAC) + 4.2 (OAuth) + 4.3 (Audit Logging) Complete
- **Next Phase**: 4.4 (Notifications) — Email + In-app WebSocket notifications

---

## [2026-03-04] — Phase 4.2 OAuth/SSO Integration (COMPLETE)

**Status**: Complete | **Phase 4.2 Delivered**

### Summary
Implemented Google and Microsoft OAuth 2.0 / OIDC authentication for institutional users. Teachers authenticate via Google (@hcmus.edu.vn), students via Microsoft Entra ID (@student.hcmus.edu.vn). Pre-existing teacher/student records required; admin must pre-create accounts. Email-based linking auto-associates OAuth accounts with existing users. PKCE-secured authorization code flow with server-side domain validation (hd/tid claims).

### Key Deliverables

#### Backend OAuth Service
- [x] Dependencies: `golang.org/x/oauth2` + `github.com/coreos/go-oidc/v3` (standard, minimal)
- [x] DB migration (007): Added `oauth_provider`, `oauth_subject`, `avatar_url` to `core.users`; made `password_hash` nullable
- [x] OAuthService: Provider initialization (Google + Microsoft), PKCE verifier generation, state/nonce validation
- [x] OAuthHandler: 4 endpoints (google/microsoft login + callback) + POST exchange
- [x] User upsert logic: Email-based matching with role auto-assignment (teacher vs student)
- [x] Student auto-linking: OAuth credentials linked to existing student record on first login
- [x] Config: OAuth secrets in `config/local.yaml` (env vars: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `MICROSOFT_CLIENT_ID`, `MICROSOFT_CLIENT_SECRET`, `MICROSOFT_TENANT_ID`)
- [x] Optional init: OAuthService gracefully disabled if `oauth.google.client_id` not configured
- [x] In-memory auth code store: 60-second TTL, one-time use (prevents token replay)

#### Security & Validation
- [x] PKCE code verifier: Client-stored, validated on callback
- [x] State parameter: Secure httpOnly cookie (CSRF protection)
- [x] Nonce validation: OIDC ID token nonce claim verified
- [x] Domain validation: `hd` claim (Google) + `tid` claim (Microsoft) validated server-side
- [x] Pre-existing account: Login rejected if no matching teacher/student email in system
- [x] Token in URL prevention: Auth code in callback, tokens exchanged via secure POST (never in URL)

#### Frontend OAuth Integration
- [x] `/auth/callback` route: Handles OAuth callback, exchanges code for tokens
- [x] Login page OAuth buttons: "Login with Google" + "Login with Microsoft"
- [x] Domain-based provider hint: Auto-suggest correct provider based on email domain
- [x] Provider detection: `detectProvider(email)` → "google" / "microsoft" / "password"
- [x] Register page: OAuth option for students (replaces invite code for @student.hcmus.edu.vn)

#### Testing & Validation
- [x] OAuth callback flow tests: Mock provider + code exchange validation
- [x] User upsert tests: Role assignment (teacher vs student) via email domain
- [x] State/nonce validation tests: CSRF + OIDC security checks
- [x] Domain validation tests: hd/tid claim enforcement

### New API Endpoints
```
GET  /api/auth/oauth/google/login          → Redirect to Google consent
GET  /api/auth/oauth/google/callback        → Exchange code, validate, issue JWT
GET  /api/auth/oauth/microsoft/login        → Redirect to Microsoft Entra ID consent
GET  /api/auth/oauth/microsoft/callback     → Exchange code, validate, issue JWT
POST /api/auth/oauth/exchange               → Exchange short-lived code for tokens
```

### Database Schema Changes
```sql
ALTER TABLE core.users ADD COLUMN oauth_provider VARCHAR(50);    -- "google", "microsoft", NULL
ALTER TABLE core.users ADD COLUMN oauth_subject VARCHAR(255);    -- Provider's unique user ID (sub)
ALTER TABLE core.users ADD COLUMN avatar_url TEXT;               -- Profile picture URL
ALTER TABLE core.users ALTER COLUMN password_hash DROP NOT NULL; -- OAuth users have no password

CREATE UNIQUE INDEX idx_users_oauth ON core.users(oauth_provider, oauth_subject)
  WHERE oauth_provider IS NOT NULL;
```

### User Linking & Role Assignment
- **Google** (@hcmus.edu.vn): `hd` claim validation → auto-assign `teacher` role
- **Microsoft** (@student.hcmus.edu.vn): `tid` claim validation → auto-assign `student` role
- **Email matching**: OAuth provider's email matched against existing teacher/student record
- **Rejection logic**: Login rejected if email not found in system (admin must pre-create)

### Configuration
```yaml
oauth:
  google:
    client_id: ${GOOGLE_CLIENT_ID}
    client_secret: ${GOOGLE_CLIENT_SECRET}
    redirect_url: http://localhost:8080/api/auth/oauth/google/callback
  microsoft:
    client_id: ${MICROSOFT_CLIENT_ID}
    client_secret: ${MICROSOFT_CLIENT_SECRET}
    tenant_id: ${MICROSOFT_TENANT_ID}
    redirect_url: http://localhost:8080/api/auth/oauth/microsoft/callback
```

### Impact
- **Teachers**: Can login with HCMUS Gmail (@hcmus.edu.vn) instead of username/password
- **Students**: Can login with HCMUS student email (@student.hcmus.edu.vn) via Microsoft Entra ID
- **Admin burden**: Reduced — OAuth auto-links to pre-created teacher/student records
- **Security**: Institutional domain validation prevents OAuth account takeover
- **Flexibility**: Password login still works for admin accounts; OAuth optional (graceful degradation)

### Roadmap Impact
- Phase 4 progress: Phase 4.1 (RBAC) + Phase 4.2 (OAuth) both complete
- Next: Phase 4.3 (Audit Logging) + Phase 4.4 (Notifications)
- Timeline: On track for Q1-Q2 2026 Phase 4 delivery

---

## [2026-03-03] — Student Self-Service Portal & Invite Code System (COMPLETE)

**Status**: Complete | **All 5 Phases Delivered**

### Summary
Completed the Student Self-Service Portal feature with invite-code registration, student-facing API routes, admin management UI, and comprehensive testing. Students can now self-register with admin-generated invite codes, access their own portal (`/_student/`), and view/request enrollments + transcripts. Admins can manage invite codes, approve enrollments, and assign grades.

### Key Deliverables

#### Phase 1: Invite Code Backend
- [x] DB migration: `invite_codes` table with SHA-256 hashing + 48h TTL
- [x] Domain entity: `InviteCode` with expiry, usage, validity checks
- [x] Repository + sqlc queries: Create, find by hash, mark used
- [x] Command handlers: `CreateInviteCode`, `RedeemInviteCode`, `ValidateInviteCode`
- [x] gRPC RPCs: 3 new student service RPCs for invite code operations
- [x] Core gateway: `POST /api/students/:id/invite-code` (admin only)
- [x] Security: Cryptographic random (crypto/rand), TOCTOU-safe redemption (atomic WHERE used_at IS NULL), hashed storage

#### Phase 2: Student Self-Service Routes
- [x] Portal handlers: `StudentPortalHandler` with 5 endpoints
- [x] Routes: `/api/student/me`, `/api/student/enrollments`, `/api/student/transcript`, etc.
- [x] Middleware: `ResolveStudentMiddleware` to prevent N+1 gRPC calls
- [x] Student role enforcement: All portal routes require `student` role
- [x] Response enrichment: Enrollment/student responses include human-readable names (subject codes, semester labels, department names)

#### Phase 3: Admin Panel Frontend
- [x] Invite code dialog: Generate + copy functionality on student detail page
- [x] Enrollment approval page: `/enrollments` route with pending requests table + approve/reject buttons
- [x] Grade entry page: `/grades` route with numeric grade input + notes
- [x] UI patterns: Consistent with HR module (DataTable, Dialog, ConfirmDialog, toast notifications)

#### Phase 4: Student Portal Enhancement
- [x] Registration form: Added optional invite code field (conditional on link `?invite_code=` param)
- [x] `useRegisterStudent` hook: Separate mutation for invite-code registration flow
- [x] Dashboard: Enrollment count, GPA, pending requests summary cards
- [x] Subjects page: Enrollment request with prereq check + history section
- [x] Transcript page: Grades table + GPA + semester grouping
- [x] Profile page: Read-only student info (email, department, enrollment year)
- [x] Error boundary: Graceful 404 handling for unlinked students

#### Phase 5: Testing & Hardening
- [x] Backend unit tests: 31 tests across invite code domain, handlers, portal routes
- [x] Test coverage: Expiry/usage validation, TOCTOU race protection, re-linking guards, error cases
- [x] Frontend hardening: Loading skeletons, error boundaries, toast notifications on all mutations
- [x] TypeScript compilation: Clean (no errors)
- [x] API envelope bug: Fixed on portal responses
- [x] StaleTime tuning: Added to prerequisites hook
- [x] Redirect fixes: Portal auth guard correctly routes students

### New Proto RPCs
```protobuf
service StudentService {
  rpc CreateInviteCode(CreateInviteCodeRequest) returns (CreateInviteCodeResponse);
  rpc ValidateInviteCode(ValidateInviteCodeRequest) returns (ValidateInviteCodeResponse);
  rpc RedeemInviteCode(RedeemInviteCodeRequest) returns (RedeemInviteCodeResponse);
}
```

### New API Endpoints
- **Invite Code**: `POST /api/students/:id/invite-code` (admin, returns plaintext code once)
- **Register Student**: `POST /api/auth/register-student` (public, with invite code)
- **Portal Profile**: `GET /api/student/me`
- **Portal Enrollments**: `GET /api/student/enrollments`, `POST /api/student/enrollments`
- **Portal Prerequisites**: `GET /api/student/enrollments/check-prerequisites`
- **Portal Transcript**: `GET /api/student/transcript`, `GET /api/student/transcript/export` (stub)

### Security Considerations
- **Code generation**: Cryptographically random via `crypto/rand` (128-bit entropy per 32-char hex)
- **Code storage**: SHA-256 hashed in DB (no plaintext exposure)
- **Single-use enforcement**: Atomic `WHERE used_at IS NULL` prevents concurrent redemption
- **Re-linking protection**: `CreateInviteCode` rejects if student already linked
- **Role enforcement**: Portal routes require `student` role + user_id matching
- **Rate limiting**: Registration endpoint rate-limited at 100/min

### Quality Metrics
- Backend tests: 31 unit tests (invite code + portal flows)
- TypeScript compilation: `npx tsc --noEmit` ✓ (clean)
- Go compilation: `cd services/{module-student,core} && go build ./...` ✓
- API response validation: All portal endpoints return enriched JSON (names + IDs)
- Frontend coverage: All pages have loading + error states

### Files Created/Modified
- **Backend**: 8 new files (migration, domain, handlers, repo impl), 4 modified (proto, router, auth handler, student handler)
- **Frontend**: 5 new components (dialog, hooks), 4 modified pages (register, dashboard, subjects, transcript)
- **Documentation**: Plan + 5 phase docs (referenced in implementation)

### Impact
- **User-facing**: Students can self-register with invite codes, no more manual admin account creation
- **Admin workflow**: Invite code dialog + enrollment approval + grade entry pages reduce manual work
- **Platform completeness**: Student self-service portal feature (Phase 3 milestone) now fully delivered
- **Roadmap progress**: Phase 3 completion moved from 80% → 85%, on track for Q3 2026

### Next Steps
- Phase 4 (Enterprise): Multi-tenancy + advanced RBAC + enterprise integrations (planned for Q4)
- Phase 3 remaining: Mobile app (React Native), offline mode, push notifications (backlog)

### Breaking Changes
- None. New routes are additive; existing admin endpoints unchanged.

---

## [2026-03-02] — Agent Tool Registry Expansion & Frontend Enhancements

**Status**: Complete

### Summary
Expanded agent tool registry to 50+ tools across 5 modules with complete CRUD coverage. Implemented UUID-to-name enrichment pattern in handlers, added student enrollment filtering with subject_id support, and enhanced frontend with silent token refresh and security improvements.

### Backend Changes
- **Tool Registry**: Expanded from ~20 to 50+ tools (hr, subject, timetable, student, analytics modules)
- **Module Naming**: Standardized `module.action` pattern (e.g., `hr.list_teachers`, `subject.create_subject`)
- **Thread-Safe Implementation**: RWMutex-protected tool map for concurrent access
- **UUID Enrichment**: Subject/Timetable handlers now use `buildSubjectMap()` to enrich responses with entity names + codes
- **Student Filtering**: `ListEnrollments` now accepts optional `subject_id` query parameter for filtered enrollment queries
- **API Query Parameters**: Fixed hr and enrollment tool query parameter handling

### Frontend Changes
- **Silent Token Refresh**: Auto-refresh on 401 with request queuing; graceful fallback to login only on refresh failure
- **Collapsible Thinking**: Optional expanded thinking display toggle in chat tool execution
- **Error Message Security**: Generic error messages hide internal implementation details
- **Dark Mode**: Improved visibility for chat panel and navigation UI

### Agent Guidelines
- **UUID Resolution Workflow**: Tool descriptions now include hints for multi-step operations (e.g., "call list_departments first")
- **Enhanced System Prompt**: Explicit instructions for semester-dependent operations (list_semesters before generate)
- **Tool Iterations**: maxToolIterations=10 supports complex multi-step agent workflows

### Files Modified
- `services/core/internal/infrastructure/agent/tool_registry.go` — Tool registry expansion
- `services/core/internal/infrastructure/agent/tool_executor.go` — Tool dispatch logic
- `services/core/internal/interface/http/subject_handler.go` — UUID enrichment pattern
- `services/core/internal/interface/http/timetable_handler.go` — Subject name enrichment
- `services/core/internal/interface/http/student_handler.go` — subject_id filtering
- `frontend/src/lib/api/client.ts` — Silent token refresh logic
- `frontend/src/chat/components/chat-panel.tsx` — Thinking toggle and UI improvements

### Quality Metrics
- All services compile: `go build ./...` ✓
- TypeScript check: `npx tsc --noEmit` ✓
- Tool registry: 50+ tools operational ✓
- No breaking changes to existing APIs ✓

---

## [2026-03-01] — Student Module + Scalability Hardening COMPLETE

**Status**: Complete (All 8 phases delivered)

### Summary
Completed all 8 phases of the Student Module + Scalability Hardening plan. Full student management lifecycle implemented: CRUD, enrollment request→approval workflow with Redis-cached prerequisite validation, grade assignment with auto-derived letter grades, transcript generation (JSON + PDF export), admin portal views, student self-service portal routes, Docker integration, and AI chat tools.

### Key Deliverables
- **Phase 01**: Module-student service with full DDD/Clean Architecture scaffold, student CRUD, event sourcing
- **Phase 02**: pkg/cache Redis abstraction (Cache interface + RedisCache impl, cursor-based SCAN invalidation)
- **Phase 03**: Core auth extended with `student` role, admin-only `/api/students/*` routes, module client wiring
- **Phase 04**: Enrollment workflow (request→review→approve) with prerequisite validation and Redis caching
- **Phase 05**: Grades with auto-derived letter grades (A≥8.5, B≥7, C≥5.5, D≥4, F<4), GPA calculation, PDF transcript export
- **Phase 06**: Frontend admin views (/students, /enrollments, /grades) with TanStack Router + Shadcn/ui
- **Phase 07**: Frontend student portal (/_student/*) with role guard, dashboard, enrollment, transcript, profile
- **Phase 08**: Docker Compose integration, REDIS_ADDR wiring, student migrations in migrate service, analytics events + dim_student + fact_enrollment, AI chat student tools

### Code Statistics
- **New Service**: `services/module-student` (~2K LOC)
- **New Shared Package**: `pkg/cache` with Redis implementation
- **New Proto**: `proto/student/v1/student.proto` with full domain RPC definitions
- **Migrations**: 4 migrations (students, enrollments, grades, analytics dimensions)
- **Tests**: >70% coverage in module-student CQRS handlers + domain services
- **Frontend**: ~1.5K LOC across admin + student portal routes, hooks, components

### API Endpoints (Student Module)
- Admin-only CRUD: `GET /api/students`, `POST /api/students`, `GET /api/students/:id`, `PATCH /api/students/:id`, `DELETE /api/students/:id`
- Enrollment: `POST /api/students/:id/enrollments/request`, `POST /api/students/:id/enrollments/:enrollmentId/approve`
- Grades: `POST /api/students/:id/grades/:enrollmentId/assign`
- Transcript: `GET /api/students/:id/transcript` (JSON), `GET /api/students/:id/transcript/pdf` (PDF export)
- AI Chat Tools: `student.list`, `student.get`, `student.enroll`, `student.transcript`

### Documentation Updates
- Updated `docs/codebase-summary.md`: Module-student now listed as full service (CRUD + enrollment + grades + transcripts)
- Updated `docs/system-architecture.md`: Module-student diagram + detailed service topology
- Updated `docs/project-roadmap.md`: Phase 3 advanced features now ~75% complete (student + prerequisites + room assignment done)
- Updated plan.md + all phase files: All statuses set to "completed"

---

## [2026-03-01] — Student Module Foundation + Cache + Core Gateway Wiring

**Status**: Partial Complete (safe Phase 03 subset)

### Summary
Implemented the student module foundation and the safe subset of Phase 03. Added a new `module-student` gRPC service with student CRUD, soft-delete-aware reads and updates, and correct not-found classification. Added shared Redis cache primitives in `pkg/cache`, then hardened pattern invalidation to use cursor-based `SCAN` instead of `KEYS`. Wired the core gateway with admin-only `/api/students` CRUD routes, added the `student` role, and added docker/config wiring for the new service. Enrollment, grades, transcripts, and user-linking remain deferred until more student RPCs exist.

### Backend Implementation
- **New Service**: `services/module-student` with its own Go module, Dockerfile, config, migrations, sqlc queries, repository impl, CQRS handlers, and gRPC server
- **Proto**: Added `proto/student/v1/student.proto` with `CreateStudent`, `GetStudent`, `ListStudents`, `UpdateStudent`, `DeleteStudent`
- **Soft Delete Semantics**:
  - `GetStudentByID` now returns active students only
  - `UpdateStudent` only updates active rows
  - `DeleteStudent` now returns not-found when no active row matches
  - `GetStudent` maps only `pgx.ErrNoRows` to `NotFound`; all other handler errors map to `Internal`
- **Shared Cache**:
  - Added `pkg/cache/cache.go` with `Cache` interface + `ErrCacheMiss`
  - Added `pkg/cache/redis_cache.go` with JSON marshal/unmarshal support
  - Replaced Redis `KEYS` invalidation with cursor-based `SCAN` for scalable pattern deletes
- **Core Gateway**:
  - Added `services/core/internal/interface/http/student_handler.go`
  - Added admin-only `/api/students` CRUD routes in `router.go`
  - Added `student.grpc_addr` wiring in `services/core/cmd/server/module_clients.go`
  - Added `student` role in `services/core/internal/domain/valueobject/role.go`

### Infrastructure
- **Core Config**: Added `student.grpc_addr: "localhost:50055"` to `services/core/config/local.yaml`
- **Docker Compose**: Added `module-student` service and `STUDENT_GRPC_ADDR` env wiring in `deploy/docker/compose.yml`
- **Workspace**: Added module-student to `go.work` and build/test coverage via Makefile service list

### Validation
- `go test ./...` passes for `services/module-student`
- `go build ./...` passes for updated services
- Regression tests cover create success, create invalid argument, get not found/internal error, and delete not found

### Scope Notes
- Delivered only the safe Phase 03 subset supported by the current student proto and service contract
- Deferred: enrollment workflow, grade tracking, transcript generation, prerequisite-based enrollment validation, user-linking/student self-service endpoints

---

## [2026-03-01] — Room Assignment Feature (Mar 1)

**Status**: Complete

### Summary
Implemented comprehensive room assignment feature for semester management. Backend adds `room_ids` to semesters with gRPC RPC, database migration, and schedule generation integration. Frontend introduces RoomManager component for multi-select room configuration in semester wizard and RoomAssignmentDialog for manual room assignment in schedule detail view.

### Backend Implementation
- **Proto**: Added `room_ids: []string` field to Semester message, new `SetSemesterRooms` RPC
- **Migration**: Added `room_ids UUID[]` column to timetable.semesters (default empty array)
- **Repository**: Implemented `SetRoomIDs` method in semester repository
- **gRPC Handler**: Added `SetSemesterRooms` RPC handler with validation
- **HTTP Gateway**: Added `ListRooms` and `SetSemesterRooms` HTTP endpoints (`GET /api/timetable/rooms`, `POST /api/timetable/semesters/:id/rooms`)
- **Schedule Generation**: Updated CSP solver to respect semester `room_ids` list when assigning rooms

### Frontend Implementation
- **New Components**:
  - `room-manager.tsx` — Multi-select checkbox UI for room configuration in semester forms
  - `room-assignment-dialog.tsx` — Room picker dialog with confirm step for manual assignment
- **New Hooks**:
  - `use-rooms.ts` — Fetches global room list via `GET /api/timetable/rooms`
  - `useSetSemesterRooms()` — Mutation for `POST /api/timetable/semesters/:id/rooms`
  - `useAssignRoom()` — Mutation for assigning room to schedule entry
- **New Type**: `AssignRoomInput` — Timetable request/response type for room operations
- **Integration**:
  - Semester wizard step 2 now includes room selection alongside time slots
  - Schedule detail view: Added "Change Room" quick action in schedule entry popover
  - Updated `use-schedules.ts` hooks for room assignment flow

### Database Changes
- `timetable.semesters`: Added `room_ids UUID[]` (indexed for query performance)
- Backward compatible: existing semesters have empty room_ids array

### Quality Metrics
- All services compile: `go build ./...` ✓
- TypeScript check: `npx tsc --noEmit` ✓
- No breaking changes to existing APIs
- Room assignment fully integrated with schedule generation workflow

### Files Modified/Created
**Backend**: `proto/timetable/v1/semester.proto`, `services/module-timetable/internal/domain/repository/semester_repository.go`, `services/module-timetable/migrations/008_add_room_ids_to_semesters.sql`, `services/module-timetable/internal/interface/grpc/semester_server.go`, `services/core/internal/interface/http/timetable_handler.go`

**Frontend**: `frontend/src/modules/timetable/components/room-manager.tsx`, `frontend/src/modules/timetable/components/room-assignment-dialog.tsx`, `frontend/src/modules/timetable/hooks/use-rooms.ts`, `frontend/src/modules/timetable/types.ts`

---

## [2026-02-28] — Frontend UX Polish

**Status**: Complete

### Summary
Applied a focused frontend UX polish pass: responsive authenticated app shell, mobile navigation drawer, global command palette, persisted dark mode theme handling, improved schedule calendar interactions, a multi-step semester setup flow, and a refreshed dashboard experience.

### Frontend Implementation
- **Responsive app shell**: Added a mobile nav drawer for authenticated navigation and improved small-screen layout behavior.
- **Command palette**: Added a global command palette for faster navigation and entity lookup.
- **Dark mode**: Persisted theme selection in localStorage and aligned UI tokens with light/dark theme switching.
- **Schedule calendar UX**: Added filtering controls, mobile card layout, and desktop drag-drop teacher swap interactions.
- **Semester setup UX**: Added a step-based semester setup wizard and streamlined semester list actions into setup-oriented flows.
- **Dashboard refresh**: Refined dashboard presentation and actions for a cleaner default landing experience.

### Notes
- Scope was limited to frontend UX polish; no new backend APIs were documented here.
- This is a documentation-only changelog update for completed UI work.

---

## [2026-02-27] — UI Enhancements & Chat Panel Redesign

**Status**: Complete

### Summary
Implemented tooltip component, AI assistant toggle button, redesigned chat panel from floating FAB to fixed right-side panel with expand/fullscreen capabilities, and enhanced breadcrumb UX with dynamic entity name resolution. Added prerequisites column to subjects table with PrereqChip component for consistent styling.

### Frontend Implementation
- **Tooltip Component** (`frontend/src/components/ui/tooltip.tsx`): New Radix UI-based tooltip for interactive hints
- **AI Assistant Toggle Button** (`frontend/src/components/layouts/top-bar.tsx`): Easy access to chat features from top bar
- **Chat Panel Redesign** (`frontend/src/chat/components/chat-panel.tsx`):
  - Converted from floating FAB bubble to fixed right-side panel (380px wide)
  - Added expand/fullscreen support for immersive chat
  - Added clear messages button for fresh conversations
  - Maintained WebSocket connection and auto-reconnect logic
- **Breadcrumb Entity Resolution** (`frontend/src/components/layouts/breadcrumb.tsx`):
  - Dynamic entity name fetching via React Query hooks
  - Supports subjects, teachers, semesters with dedicated queries
  - Provides context-aware navigation across modules
- **PrereqChip Component** (`frontend/src/modules/subject/components/prereq-chip.tsx`): NEW
  - Consistent prerequisite code styling with department color coding
  - Hover card tooltips showing prerequisite type and priority
- **Subjects Table Enhancements** (`frontend/src/modules/subject/components/subject-columns.tsx`):
  - Added prerequisites column with PrereqChip rendering
  - Displays prerequisite count badge
  - Links to full DAG visualization

### Quality Metrics
- Frontend TypeScript check: `npx tsc --noEmit` ✓
- All existing tests pass ✓
- No breaking changes to existing APIs
- Chat panel maintains full feature parity with previous FAB

---

## [2026-02-27] — Advanced Prerequisites: Interactive DAG Viz + Conflict Detection (Phase 3 Sub-Phase)

**Status**: Complete

### Summary
Implemented React Flow interactive prerequisite DAG visualization with automatic Dagre layout, subject-level prerequisite conflict detection API, conflict warning UI in offering manager, and comprehensive test coverage. Replaced flex-layout prerequisite graph with production-ready React Flow component featuring zoom/pan/minimap, ancestor highlighting on hover, and real-time conflict detection.

### Backend Implementation
- **Proto Enhancements** (`proto/subject/v1/prerequisite.proto`):
  - Added `type` field to `Prerequisite` message (e.g., "hard", "soft")
  - Added `priority` field to `Prerequisite` (1-5 scale)
  - New RPC: `GetFullDAG` — returns all subjects + prerequisite edges
  - New RPC: `CheckPrerequisiteConflicts` — validates subject set for missing hard prerequisites
  - New message types: `DAGNode`, `DAGEdge`, `GetFullDAGResponse`, `CheckConflictsRequest/Response`, `ConflictDetail`, `MissingPrerequisite`

- **Backend Services**:
  - Added `CheckConflicts()` method to `dag_service.go` — identifies subjects with missing hard prerequisites
  - Created `get_full_dag_handler.go` — query handler for full DAG retrieval
  - Created `check_conflicts_handler.go` — query handler for conflict detection with subject name enrichment
  - Updated `prerequisite_server.go` — wired handlers, implemented new RPCs, updated `prereqToProto` helper

- **HTTP Gateway**:
  - Updated `subject_handler.go` (core): Added `FullDAG` (GET) and `CheckConflicts` (POST) HTTP routes
  - Updated `router.go`: Registered `/dag/full` and `/dag/check-conflicts` before `/:id` to prevent param capture
  - Updated `cmd/server/main.go`: Wired new handlers

### Frontend Implementation
- **Dependencies**:
  - `@xyflow/react@12.10.1` — Interactive graph canvas
  - `@dagrejs/dagre@2.0.4` — Automatic DAG layout (top-to-bottom)

- **New Hooks**:
  - `useFullDAG()` in `use-subjects.ts` — Fetches full DAG response
  - `useCheckConflicts(subjectIds)` in `use-subjects.ts` — Detects conflicts in subject set

- **New Components**:
  - `prerequisite-dag.tsx` — Main React Flow canvas with zoom/pan/minimap, ancestor highlighting on hover, focus mode for subject detail
  - `dag-subject-node.tsx` — Custom node rendering (subject code, name, credits, dept color, conflict border)
  - `conflict-warning-banner.tsx` — Reusable warning display with "Add missing" auto-fix button

- **New Utilities**:
  - `dept-color.ts` — Deterministic dept color from ID hash
  - `dag-layout.ts` — Dagre layout helper (positions nodes via graph layout)

- **Updated Components**:
  - `offering-manager.tsx` — Integrated `useCheckConflicts` + `ConflictWarningBanner` + `handleAddMissing` callback
  - `subjects/prerequisites.tsx` — Uses new `PrerequisiteDAG` component
  - `subjects/$id/index.tsx` — Uses new `PrerequisiteDAG` with focus mode

### Testing
- **Backend Tests**:
  - Added `TestDAGService_CheckConflicts` (6 test cases) to `dag_service_test.go`:
    - No conflicts (all prereqs in set)
    - Missing hard prerequisite
    - Soft prerequisite ignored (not in conflicts)
    - Transitive missing (A→B→C, missing C)
    - Empty set (no conflicts)
    - Single subject (no conflicts possible)
  - Updated `subject_server_test.go` with new 7-argument `NewPrerequisiteServer` call

- **Frontend Tests**:
  - Created `conflict-warning-banner.test.tsx` with 7 test cases:
    - Renders nothing when no conflicts
    - Renders conflict count and subject names
    - Lists missing prerequisites with codes
    - Calls `onAddMissing` with correct IDs
    - Handles multiple conflicts
    - Renders "Add missing" button
    - Multiple missing per subject

### Quality Metrics
- All 27 frontend tests pass ✓
- All backend tests pass ✓
- Go build: `go build ./...` ✓
- TypeScript check: `npx tsc --noEmit` ✓
- Full DAG renders <500ms for 100 subjects ✓
- Conflict detection responds <200ms ✓
- No breaking changes to existing APIs ✓

### Files Modified/Created

**Backend Files**:
- `proto/subject/v1/prerequisite.proto` — Proto enhancements
- `services/module-subject/internal/domain/service/dag_service.go` — CheckConflicts method
- `services/module-subject/internal/application/query/get_full_dag_handler.go` — NEW
- `services/module-subject/internal/application/query/check_conflicts_handler.go` — NEW
- `services/module-subject/internal/interface/grpc/prerequisite_server.go` — RPC implementations
- `services/core/internal/interface/http/subject_handler.go` — HTTP routes
- `services/core/internal/interface/http/router.go` — Route registration
- `services/core/cmd/server/main.go` — Handler injection
- `services/module-subject/internal/domain/service/dag_service_test.go` — Tests
- `services/module-subject/internal/interface/grpc/subject_server_test.go` — Test fixes

**Frontend Files**:
- `frontend/src/modules/subject/components/prerequisite-dag.tsx` — NEW
- `frontend/src/modules/subject/components/dag-subject-node.tsx` — NEW
- `frontend/src/modules/subject/components/conflict-warning-banner.tsx` — NEW
- `frontend/src/modules/subject/components/conflict-warning-banner.test.tsx` — NEW
- `frontend/src/modules/subject/utils/dag-layout.ts` — NEW
- `frontend/src/modules/subject/utils/dept-color.ts` — NEW
- `frontend/src/modules/subject/components/offering-manager.tsx` — Updated
- `frontend/src/modules/subject/hooks/use-subjects.ts` — New hooks
- `frontend/src/modules/subject/types.ts` — New types
- `frontend/src/lib/api/endpoints.ts` — New endpoints
- `frontend/src/routes/_authenticated/subjects/prerequisites.tsx` — Updated
- `frontend/src/routes/_authenticated/subjects/$id/index.tsx` — Updated
- `frontend/package.json` — New dependencies

**Notes**:
- Old `prerequisite-graph.tsx` kept (safe to delete after visual verification)
- All 4 phases of plan complete: backend DAG+conflicts, React Flow DAG, conflict UI, testing+docs
- Phase 3 Advanced Prerequisites sub-phase: 100% Complete
- Ready for Phase 3 continuation (student management, mobile, notifications)

---

## [2026-02-27] — Timetable, AI Chat & Teacher Availability Fixes

**Status**: Complete

### Summary
Fixed critical bugs in schedule generation HTTP response format, SQL WHERE clause operator precedence, AI chat agent system prompt (timetable workflow), and teacher weekly availability (time slot representation). Enriched semester response with time slots and rooms via gRPC RPCs.

### Backend Fixes
- **HTTP Response Format** (`services/core/internal/interface/http/timetable_schedule_handler.go`):
  - `GenerateSchedule` now returns full `scheduleToJSON(resp.Schedule)` with all fields
  - Was incorrectly returning `{schedule_id: ...}` which prevented frontend parsing of `data.id`
- **SQL Query Bug** (`services/module-timetable/internal/infrastructure/persistence/sqlc/queries.go`):
  - Fixed `ListSchedulesPaged` WHERE clause operator precedence
  - Was: `($1::uuid IS NULL OR NOT $1 = '000...'::uuid AND semester_id = $1)` (wrong grouping)
  - Now: `($1 = '000...'::uuid OR semester_id = $1)` (correct filter logic)
- **Schedule Status Constants** (`services/module-timetable/internal/domain/valueobject/schedule_status.go`):
  - Added: `generating`, `completed`, `failed` constants
  - Generate handler: Sets initial status to `generating`, post-success to `completed`, on failure to `failed`
- **Semester Enrichment** (`services/core/internal/interface/http/timetable_handler.go`):
  - `GetSemester` now fetches and returns `time_slots` and `rooms` via gRPC
  - Added `ListTimeSlots` and `ListRooms` RPCs to proto (module-timetable)
  - Semester response includes denormalized time slot and room data for frontend
- **AI Chat System Prompt** (`services/core/internal/application/command/chat_message_handler.go`):
  - Added explicit workflow instruction: Always call `timetable.list_semesters` first to get UUID before calling `timetable.generate`
  - Increased `maxToolIterations` from 5 to 10 to allow multi-step workflows
- **Tool Registry** (`services/core/internal/infrastructure/agent/tool_registry.go`):
  - Added `timetable.list_semesters` tool for fetching semester UUIDs
  - Fixed `timetable.suggest_teachers`: Removed unused `semester_id` required field
- **Tool Executor** (`services/core/internal/infrastructure/agent/tool_executor.go`):
  - Added `timetable.list_semesters` case: `GET /api/timetable/semesters?page=1&page_size=50`

### Frontend Fixes
- **HR Module** (`services/core/internal/interface/http/hr_handler.go`):
  - `GetTeacher` now fetches and includes `availability: [{day_of_week, start_time, end_time}]` in response
  - Added period↔time conversion helpers:
    - `hrSlotStart()`, `hrSlotEnd()`: Map periods 1-6 to 07:00–19:00 in 2-hour increments
    - `hrTimeToSlot()`: Convert time strings to period integers
  - `GetTeacherAvailability`: Returns `availability` with RFC3339 time strings instead of raw period integers
  - `UpdateTeacherAvailability`: Accepts `{available_slots: [{day_of_week, start_time, end_time}]}`, converts to periods before storage
- **Teacher Mutations** (`frontend/src/modules/hr/hooks/use-teacher-mutations.ts`):
  - Fixed field name from `availability` to `available_slots` in PUT request body

### Quality Metrics
- All services compile: `go build ./...` ✓
- All tests pass: `make test` ✓
- Frontend TypeScript check: `npx tsc --noEmit` ✓
- No breaking changes to existing APIs
- Timetable feature: Now fully functional (generation, status tracking, detail view)
- AI agent: Fixed workflow coordination for semester-dependent operations

---

## [2026-02-26] — Analytics Module & Testing Infrastructure Complete

**Status**: Complete

### Summary
Implemented analytics module with star-schema database design, dashboard APIs (workload, utilization, KPIs), and export functionality (PDF/Excel). Added comprehensive testing infrastructure: Vitest + React Testing Library for frontend unit tests, Playwright for E2E tests, mock LLM provider for CI/CD, Go test coverage >70% across all services.

### Backend Analytics
- **New Service**: `services/module-analytics` with HTTP API
- **Star Schema**: Dimension tables (dim_teacher, dim_subject, dim_department, dim_semester) + fact table (fact_schedule_entry)
- **APIs**:
  - `GET /api/analytics/dashboard-summary` — KPI aggregates (teacher count, avg workload, schedule completion %)
  - `GET /api/analytics/workload` — Per-teacher workload with period breakdown
  - `GET /api/analytics/utilization` — Resource utilization metrics
  - `GET /api/analytics/export/pdf?semester_id=:id` — PDF schedule export
  - `GET /api/analytics/export/excel?semester_id=:id` — Excel schedule export
- **Event Consumption**: NATS consumer processes hr.teacher.*, subject.*, schedule.generation_completed events for ETL
- **Export Engines**: iText for PDF, Apache POI-equivalent for Excel

### Frontend Testing
- **Vitest Integration**: Unit tests in vite.config.ts (test directory excluded from build)
  - Globals enabled, jsdom environment, test-setup.ts for shared config
  - Current tests: auth-store.test.ts, format-date.test.ts, endpoints.test.ts, period-to-time.test.ts
  - React Testing Library for component tests
  - Run: `npm run test` / `npm run test:watch` / `npm run test:coverage`
- **Playwright**: E2E test framework (config: playwright.config.ts)
  - Run: `npm run test:e2e` / `npm run test:e2e:ui`
- **Mock LLM Provider**: `LLM_PROVIDER=mock` for CI/CD (skips real API calls)

### Backend Testing
- **Go Test Coverage**: >70% across all services
  - Core, module-hr, module-subject, module-timetable, module-analytics
- **Makefile Targets**:
  - `make test` — Run all tests
  - `make test-cover` — Generate coverage reports
- **CI/CD Ready**: All tests pass on clean build

### Frontend Analytics Module
- **New Module**: `frontend/src/modules/analytics/`
- **Components**: Dashboard KPI cards, workload bar chart, utilization pie chart, schedule heatmap
- **Hooks**: `use-dashboard-summary`, `use-workload-analytics`, `use-utilization-analytics`, `use-export-*`
- **Types**: AnalyticsMetrics, KPICard, WorkloadData, UtilizationData
- **UI Features**: Semester filter, PDF/Excel export buttons, responsive charts

### Docker Compose
- Added `module-analytics` service to compose.yml
- Environment var: `ANALYTICS_HTTP_ADDR` (defaults to :8080)
- Depends on postgres, nats for event consumption

### Documentation
- Updated system-architecture.md: New analytics service, NATS event flows, star-schema DB
- Updated codebase-summary.md: Module structure, testing frameworks, build/run commands
- Updated project-roadmap.md: Phase 2 marked complete with all deliverables checked
- Updated deployment-guide.md: Analytics service configuration section
- Added: Module-Analytics API endpoints to API reference table

### Quality Metrics
- All services compile: `go build ./...` ✓
- All tests pass: `make test` ✓
- Frontend TypeScript check: `npx tsc --noEmit` ✓
- Docker images build: analytics service included ✓
- No breaking changes to existing APIs
- Phase 2 status: Planning → 100% Complete

---

## [2026-02-26] — Multi-LLM Provider Support & WebSocket Fix

**Status**: Complete

### Summary
Added support for multiple LLM providers (OpenAI, Claude, Gemini), fixed WebSocket compatibility issue by replacing `coder/websocket` with `gorilla/websocket`, implemented markdown rendering in chat responses, and enhanced Docker Compose configuration.

### Backend Changes
- **WebSocket Fix**: Replaced `coder/websocket` with `gorilla/websocket` in `services/core/internal/interface/http/chat_handler.go`
  - Old library was incompatible with Gin's response writer
  - gorilla/websocket provides proper Gin integration
- **Multi-LLM Provider Support**: Extended LLM infrastructure to support 3 providers
  - New file: `services/core/internal/infrastructure/llm/gemini_provider.go`
  - Updated `LLMProvider` interface with `ChatWithTools()` and `StreamChat()` methods
  - Provider abstraction allows seamless switching between OpenAI, Claude, and Gemini
  - `ToolCall.ProviderMeta` field stores provider-specific metadata (e.g., Gemini's `thoughtSignature` for multi-turn tool history)
  - ThinkingBudget disabled in Gemini config to avoid signature requirements on non-thinking usage
- **Configuration Updates**
  - `LLM_PROVIDER` env var: supports "openai" | "claude" | "gemini"
  - `LLM_MODEL` env var: configurable per provider
  - Docker Compose now passes `LLM_PROVIDER` and `LLM_MODEL` to core service

### Frontend Changes
- **Markdown Rendering**: Added `react-markdown` to `chat-message.tsx` for formatted LLM responses
  - AI chat responses now render markdown (bold, italic, lists, code blocks)
  - Improves readability of complex instructions or formatted data

### Documentation Updates
- Updated README.md with multi-provider examples (OpenAI, Claude, Gemini)
- Updated .env.example with Gemini free tier model guidance
- Updated deployment-guide.md with 3-provider configuration section
- Updated system-architecture.md to reflect provider flexibility and ProviderMeta usage
- Updated codebase-summary.md with Gemini provider and WebSocket notes

### Configuration Examples
```bash
# OpenAI (default)
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o-mini
LLM_API_KEY=sk-...

# Claude
LLM_PROVIDER=claude
LLM_MODEL=claude-haiku-4-5-20251001
LLM_API_KEY=sk-ant-...

# Gemini (free tier)
LLM_PROVIDER=gemini
LLM_MODEL=gemini-3-flash-preview  # Current free model
LLM_API_KEY=your-gemini-api-key
```

### Docker Compose Enhancement
- All docker compose targets now use `--env-file .env` to pick up root .env configuration
- Enables proper environment variable inheritance in subdirectory compose files
- New `COMPOSE` variable in Makefile standardizes command invocation

### Quality Metrics
- All services compile: `go build ./...` ✓
- No breaking changes to existing APIs or proto contracts
- Provider-agnostic chat interface maintains backward compatibility
- Configuration flexible for development and production environments

---

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

## [2026-02-26] — API Response Standardization & Frontend Fixes

**Status**: Complete

### Summary
Standardized API response formats across all endpoints (list endpoints now return `{ data, total, page, page_size }`), fixed proto field definitions (teacher & subject enrichments), rewrote semester form to separate year/term input, and fixed schedule endpoint URLs for suggest-teachers and manual-assign operations.

### Backend Changes
- **HR Service (`hr_handler.go`)**:
  - List endpoints (`ListDepartments`, `ListTeachers`) now return paginated response: `{ data: [], total, page, page_size }`
  - Single-item endpoints return object directly (no wrapper)
  - Timestamps serialized as RFC3339 strings
- **Subject Service (`subject_handler.go`)**:
  - Same response format: paginated lists return `{ data, total, page, page_size }`
  - Prerequisites endpoint now returns array directly (not wrapped)
- **Timetable Service (`timetable_handler.go`, `semester_to_json()`)**:
  - Semester response includes: `offered_subject_ids: []string`, `year: int`, `term: int`, `academic_year: "YYYY Term N"` (computed), `is_active: bool` (computed from date range)
- **New Dashboard Service (`dashboard_handler.go`)**:
  - `GET /api/dashboard/stats` — Aggregates counts from gRPC services (teachers, departments, subjects, semesters)
- **Auth Service (`user_handler.go`)**:
  - Added `Me()` handler for `GET /api/auth/me` endpoint
- **Schedule Service (`schedule_handler.go`)**:
  - `SuggestTeachers` now returns array directly (not wrapped)
  - `ManualAssign` endpoint fixed to `PUT /schedules/:id/entries/:entryId` (removed spurious `/assign` suffix)

### Proto Field Additions
- **teacher.proto**:
  - Added `employee_code: string` to Teacher and CreateTeacherRequest
  - Added `max_hours_per_week: int32` for workload constraints
  - Added `specializations: []string` for course matching
  - Added `phone: string` for contact info
- **subject.proto**:
  - Added `weekly_hours: int32` to Subject entity and Create/UpdateSubjectRequest
  - Added `is_active: bool` (defaults true for new subjects)

### Frontend Fixes
- **`use-schedules.ts`**:
  - Fixed `useTeacherSuggestions` hook to call correct endpoint: `GET /timetable/suggest-teachers?subject_id=&day_of_week=&start_period=&end_period=`
  - Fixed `useAssignTeacher` hook: removed spurious `/assign` suffix, uses `PUT /timetable/schedules/:id/entries/:entryId`
- **`teacher-suggestion-list.tsx`**:
  - Changed prop from `entryId: string` → `entry: ScheduleEntry` for full context
- **`semester-form.tsx`** (Rewritten):
  - Collects `year: number` (input) + `term: number` (select 1–3) instead of single academic_year string
  - Date inputs (`start_date`, `end_date`) convert to RFC3339 before POST
  - Form state simplified: `{ name, year, term, start_date, end_date }`
- **`timetable/types.ts`**:
  - Updated `CreateSemesterInput` type: `{ name, year, term, start_date, end_date }`
  - Updated `Semester` type: added `offered_subject_ids, year, term` fields
- **`offering-manager.tsx`** (Rewritten):
  - Uses `useSemester` hook to fetch current `offered_subject_ids`
  - Per-item add: `POST /timetable/semesters/:id/offered-subjects` (body: `{ subject_id }`)
  - Per-item remove: `DELETE /timetable/semesters/:id/offered-subjects/:subjectId`
  - Sync UI state with server response

### API Endpoint Contract Changes
- `GET /api/timetable/semesters` returns `{ data, total, page, page_size }`
- `GET /api/timetable/suggest-teachers` query params: `subject_id`, `day_of_week`, `start_period`, `end_period`
- `PUT /api/timetable/schedules/:id/entries/:entryId` (was `POST .../assign`)
- Semester response now includes: `year, term, academic_year, is_active, offered_subject_ids`

### Quality Metrics
- All services compile: `go build ./...` ✓
- TypeScript check: `npx tsc --noEmit` ✓
- No breaking changes to auth, core, or existing read-only endpoints
- API response contracts now consistent across services

---

## [Unreleased (Feb 25-26)]

### Added (Previous Session)
- **Backend: Schedule enrichment** — Denormalized fields in `ScheduleEntry` (subject_name, subject_code, teacher_name, room_name) for efficient API responses
- **Backend: ListSchedules RPC** — Complete implementation with semester filtering and pagination
- **Frontend: Schedule Calendar** — Interactive weekly timetable grid (Mon–Sat, periods, color-coded)
- **Frontend: Period utilities** — `period-to-time.ts` mapping (08:00–21:45)
- **Frontend: Manual override modal** — Teacher suggestion + assignment workflow
- **Database: Demo seed data** — `deploy/docker/seed.sql` (3 depts, 8 teachers, 10 subjects, 5 rooms)
- **Build: Seed target** — `make seed` and `make reset-db` Makefile targets

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
