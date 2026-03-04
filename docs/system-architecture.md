# System Architecture

## High-Level Overview

Myrmex is a microservice architecture with modular services communicating via gRPC and event streaming through NATS JetStream. The HTTP gateway (Core) proxies requests to gRPC services; the AI chat agent orchestrates operations via a dynamic tool registry. Audit logging captures mutations asynchronously via NATS → PostgreSQL for compliance and forensics.

### Audit Logging Pipeline

Audit logging is implemented as a fire-and-forget NATS pipeline with persistent storage:

1. **Middleware Capture** (audit_middleware.go):
   - Post-handler Gin middleware intercepts responses
   - Derives action from HTTP method + endpoint pattern: POST→Create, PATCH→Update, DELETE→Delete, GET→Read (skipped)
   - Skips internal service calls (X-Internal-Service header)
   - Publishes to NATS subject `AUDIT.logs` with user_id, resource_type, action, old/new values
   - Fire-and-forget; failures don't block API response

2. **Async Consumer** (audit_consumer.go):
   - Durable JetStream consumer listening on `AUDIT.logs` stream
   - Receives audit events in order (NATS ordering guarantee)
   - Writes to `core.audit_logs` partitioned table with user_id, resource_type, action, timestamp
   - Acknowledges on successful insert; nack on error (retry with exponential backoff)

3. **Storage Layer** (migration/008):
   - Partitioned table `core.audit_logs` with 12 monthly partitions (2026-03 through 2027-02)
   - Columns: id, user_id, resource_type, action, old_value, new_value, timestamp
   - Indexes: BRIN (timestamp range), B-tree (user_id, resource_type) for fast filtering
   - Queries scan individual partitions via constraint exclusion (efficient for date ranges)

4. **Query API** (audit_handler.go):
   - Endpoint: GET `/api/audit-logs` (admin/super_admin only)
   - Filters: user_id, resource_type, action, date range (start_date, end_date)
   - Pagination: limit, offset (default 100 records per page)
   - Response: Array of audit entries with human-readable action labels

5. **Frontend** (/admin/audit-logs):
   - Table UI with sortable columns: User, Resource Type, Action, Timestamp
   - Row expansion to view old/new value diffs (JSON diff rendering)
   - Filter toolbar: User selector, resource type dropdown, action checkboxes, date picker
   - Pagination controls with total count

**Optional Configuration**: Audit middleware is no-op if NATS not configured; graceful degradation for testing/dev.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Client Layer                                   │
│  ┌──────────────────────┐  ┌──────────────────────────────────────┐    │
│  │   React Frontend     │  │   External API (if needed)           │    │
│  │   (localhost:3000)   │  │                                      │    │
│  └──────────┬───────────┘  └──────────────────────────────────────┘    │
└─────────────┼──────────────────────────────────────────────────────────┘
              │
              │ HTTP/WebSocket
              │
┌─────────────▼──────────────────────────────────────────────────────────┐
│                      API Gateway Layer                                  │
│                    Core Service (port 8080)                             │
│  ┌────────────────────────────────────────────────────────────────┐   │
│  │  Gin HTTP Router                                               │   │
│  │  ├─ POST /api/auth/register        → AuthService gRPC         │   │
│  │  ├─ POST /api/auth/login           → AuthService gRPC         │   │
│  │  ├─ POST /api/auth/refresh         → AuthService gRPC         │   │
│  │  ├─ GET  /api/auth/me              → UserService gRPC         │   │
│  │  ├─ GET  /api/dashboard/stats      → Dashboard (aggregate)    │   │
│  │  ├─ ANY  /api/hr/*                 → Module-HR gRPC (proxy)   │   │
│  │  ├─ ANY  /api/subjects/*           → Module-Subject gRPC      │   │
│  │  ├─ ANY  /api/timetable/*          → Module-Timetable gRPC    │   │
│  │  ├─ ANY  /api/students/*           → Module-Student gRPC (admin CRUD + enrollment + grades) │   │
│  │  ├─ ANY  /api/analytics/*          → Module-Analytics HTTP    │   │
│  │  └─ WebSocket /ws/chat?token=X    → ChatGateway (Streaming)   │   │
│  │                                                                │   │
│  │  Middleware:                                                   │   │
│  │  ├─ CORS (origin validation)                                  │   │
│  │  ├─ Auth (JWT extraction + validation)                        │   │
│  │  ├─ Audit Logging (capture mutations → NATS)                  │   │
│  │  ├─ Rate Limiting (configurable per endpoint)                 │   │
│  │  └─ Request/Response logging                                  │   │
│  └────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  Core gRPC Server (port 50051)                                         │
│  ├─ AuthService (Login, Register, RefreshToken, ValidateToken)       │
│  ├─ UserService (CRUD)                                                 │
│  └─ ModuleRegistryService (Register, Unregister, List, Health)       │
│                                                                         │
│  AI Chat Gateway                                                       │
│  ├─ WebSocket handler + connection manager                             │
│  ├─ Message routing (user → LLM → tools → response)                   │
│  ├─ Tool Registry (50+ tools across 5 modules, thread-safe RWMutex)   │
│  │   └─ Dispatch: HTTP self-referential via internal JWT token       │
│  └─ Event streaming to frontend                                        │
└─────────────────────────────────────────────────────────────────────────┘
              │          │           │              │             │              │
              │ gRPC     │ gRPC      │ gRPC         │ gRPC        │ gRPC         │ HTTP
              │          │           │              │             │              │
┌─────────────▼──┐  ┌────▼──────┐  ┌─▼──────────┐  ┌▼───────────┐  ┌┴──────────────┐  ┌──────────────┐
│  Module-HR     │  │  Module-  │  │ Module-    │  │ Module-    │  │ User/Auth     │  │ Module-      │
│  Service       │  │ Subject   │  │ Timetable  │  │ Student    │  │ (in Core)     │  │ Analytics    │
│  (port 50052)  │  │ (50053)   │  │ (50054)    │  │ (50055)    │  │ (50051)       │  │ (8055 HTTP)  │
│                │  │           │  │            │  │               │  │              │
│ Department     │  │ Subject   │  │ Semester   │  │ Student          │  │ User Mgmt     │  │ Star Schema  │
│ Teacher CRUD   │  │ DAG       │  │ Room       │  │ Enrollment       │  │ JWT Auth      │  │ Dashboard KPIs
│ Availability   │  │ Prereq    │  │ Schedule   │  │ Grade            │  │ Refresh Token │  │ Workload     │
│                │  │ Service   │  │ CSP Solver │  │ Transcript       │  │               │  │ Utilization  │
│ Domain:        │  │ Service   │  │ Service    │  │ Domain:          │  │ Domain:       │  │ Export PDF/XL│
│ ├─ Department  │  │           │  │            │  │ ├─ Student       │  │ ├─ User       │  │              │
│ ├─ Teacher     │  │ Domain:   │  │ Domain:    │  │ ├─ Enrollment    │  │ └─ Session    │  │ Dimensions:  │
│ ├─ Availability│  │ ├─ Subject│  │ ├─ Semester│  │ └─ Grade         │  │               │  │ ├─ Teacher   │
│ └─ Specialization│ │ ├─ Prereq│  │ ├─ Schedule│  │                  │  │               │  │ ├─ Subject   │
└────────┬───────┘  │ └─ DAG    │  │ ├─ Room    │                   └───────────────┘  │ ├─ Department│
         │          │           │  │ └─ TimeSlot│                                      │ ├─ Semester  │
         │          └─────┬──────┘  │            │                                      │ └─ Facts     │
         │                │         └─────┬──────┘                                      │ (schedules)  │
         │                │               │                                            │              │
         └────────────────┼───────────────┼────────────────────────────────────────────┘
                          │
         NATS JetStream (Event Bus, port 4222)
                          │
         ┌────────────────┴────────────────────────────────────────────┐
         │                                                             │
    ┌────▼─────┐  ┌──────────┐  ┌────────────┐  ┌──────────┐  ┌────────┐
    │ Event    │  │ Async    │  │ Frontend   │  │ Analytics│  │ Audit  │
    │ Store    │  │ Consumers│  │ WebSocket  │  │ Consumer │  │Consumer│
    │ (Append) │  │ (Listen) │  │ (Stream)   │  │ (ETL)    │  │(Log DB)│
    └──────────┘  └──────────┘  └────────────┘  └──────────┘  └────────┘
         │                                            │
┌────────▼─────────────────────────────────────────────────────────┐
│     PostgreSQL (Shared Database)                                  │
│     localhost:5432 (myrmex / myrmex_dev)                          │
│                                                                   │
│ Operational Schemas:                                             │
│ ├─ core                                                          │
│ │  ├─ users, module_registry, conversations, event_store        │
│ │  └─ audit_logs (12 monthly partitions 2026-03→2027-02)        │
│ ├─ hr (departments, teachers, availability, specializations)    │
│ ├─ subject (subjects, prerequisites, event_store)               │
│ ├─ timetable (semesters, offerings, rooms, schedules, events)   │
│ └─ student (students, event_store)                              │
│                                                                   │
│ Analytics Schema:                                                │
│ └─ analytics                                                     │
│    ├─ dim_teacher (teacher dimension, denormalized)             │
│    ├─ dim_subject (subject dimension)                           │
│    ├─ dim_department (department dimension)                     │
│    ├─ dim_semester (semester dimension)                         │
│    └─ fact_schedule_entry (schedule fact table)                 │
└────────────────────────────────────────────────────────────────┘
         │
    ┌────▼───┐  ┌─────────┐
    │ Redis  │  │ Backup  │
    │ (cache)│  │ Storage │
    └────────┘  └─────────┘
```

## Service Topology

### Core Service (HTTP Gateway + Auth + AI Chat)

**Purpose**: Entry point for HTTP requests; manages authentication and delegates to gRPC services.

**Ports**: HTTP `:8080`, gRPC `:50051`, Metrics `:9090` (future)

**Key Responsibilities**:
1. **HTTP Gateway**: Proxy requests to gRPC modules (HR, Subject, Timetable, Student)
2. **Authentication**: JWT generation + validation (15min access, 7day refresh) with extended claims (role, department_id, teacher_id)
3. **User Management**: User CRUD, 6-role RBAC (super_admin, admin, dean, dept_head, teacher, student), role assignment API
4. **Authorization**: Two-tier enforcement (middleware + handler), department-scoped access for dept_head/teacher
5. **Module Registry**: Service discovery + health checks
6. **AI Chat**: WebSocket endpoint, tool execution (50+ tools), LLM integration

**Outbound**: PostgreSQL (core schema), NATS JetStream, LLM API (OpenAI/Claude/Gemini)

**Inbound Dependencies**: All module gRPC services

### Module-HR (Department & Teacher Management)

**Purpose**: Faculty data management and availability scheduling.

**Port**: gRPC `:50052`

**Key Entities**:
- **Department**: id, name, created_at, deleted_at
- **Teacher**: id, name, email, department_id, specializations[], availability[]
- **Availability**: {day_of_week, start_time, end_time} — Time slots in HH:MM format
  - Periods 1-6 mapped to: 07:00-09:00, 09:00-11:00, 11:00-13:00, 13:00-15:00, 15:00-17:00, 17:00-19:00

**Key Operations**:
- CRUD: Department, Teacher (with availability + specializations)
- GetTeacher: Returns availability as time strings
- UpdateTeacherAvailability: Accepts {available_slots: [{day_of_week, start_time, end_time}]}
- Soft delete for both departments and teachers

**Event Types**: `teacher.created/updated/deleted`, `department.created/updated/deleted`, `teacher.availability_updated`

**Outbound**: PostgreSQL (hr schema), NATS JetStream

### Module-Subject (Subject & Prerequisite DAG)

**Purpose**: Course structure with prerequisite management and cycle detection.

**Port**: gRPC `:50053`

**Key Entities**:
- **Subject**: id, code, name, credits, weekly_hours, department_id, is_active
- **Prerequisite**: subject_id, prerequisite_subject_id, type (strict/recommended/corequisite), priority (-2 to +5)
- **DAG**: In-memory for cycle detection + topological sort; DAGNode + DAGEdge for visualization

**Key Operations**:
- CRUD: Subject
- Add/Remove prerequisites (cycle detection via DFS 3-color)
- **GetFullDAG**: All subjects + edges for visualization
- **CheckPrerequisiteConflicts**: Detect missing prerequisites in subject set
- Topological sort for enrollment planning

**Domain Services**: DAGService (cycle detection, topological sort via BFS)

**Event Types**: `subject.created/updated/deleted`, `prerequisite.added/removed`, `prerequisite_dag.validated`

**Outbound**: PostgreSQL (subject schema), NATS JetStream

### Module-Student (Student Management + Enrollment + Grades + Invite Codes)

**Purpose**: Complete student lifecycle: invite code generation, registration, enrollment request/approval, grade assignment, transcript generation.

**Port**: gRPC `:50055`

**Key Entities**:
- **Student**: id, student_code, user_id, full_name, email, department_id, enrollment_year, status, is_active
- **InviteCode**: code (SHA-256 hash), created_by_user_id, used_at (nullable), expires_at; single-use TOCTOU-safe via WHERE used_at IS NULL
- **Enrollment**: id, student_id, offered_subject_id, semester_id, status (requested/approved/rejected/completed)
- **Grade**: id, enrollment_id, grade_numeric (0-10), grade_letter (auto-derived A-F)

**Key Operations**:
- **Student CRUD**: Create, get, list (paginated), update, soft-delete
- **CreateInviteCode**: Admin generates code; hashed in DB (redemption validation via constant-time compare)
- **ValidateInviteCode**: Check code exists + unused + not expired (used by registration flow)
- **RedeemInviteCode**: Atomically mark code used + link user to student (WHERE used_at IS NULL prevents double-redeem)
- **RequestEnrollment**: Student requests subject enrollment
- **ApproveEnrollment**: Admin approves with prerequisite validation
- **AssignGrade**: Teacher/admin assigns numeric grade → letter auto-derived
- **GetTranscript**: Student's full academic history + GPA calculation

**Domain Services**:
- **PrerequisiteValidator**: Checks student has completed prerequisites (Redis-cached)
- **GradeComputer**: Auto-derives letter grade
- **TranscriptBuilder**: Aggregates approved enrollments + grades

**Event Types**:
- `student.created`, `student.updated`, `student.deleted`
- `student.enrollment_requested`, `student.enrollment_approved`, `student.enrollment_rejected`
- `student.grade_assigned`
- `invite_code.created`, `invite_code.redeemed`

**Outbound**:
- PostgreSQL (student schema: students, enrollments, grades, invite_codes tables)
- NATS JetStream (publish events)
- Module-Subject (gRPC): For prerequisite validation
- Redis (pkg/cache): Cache prerequisites graph (TTL 1h)

### Module-Analytics (Analytics & Reporting)

**Purpose**: Business intelligence on resource utilization and performance metrics.

**Port**: HTTP `:8055` (reverse-proxied via Core at `/api/analytics/`)

**Key Entities**:
- **Dimensions**: Teacher, Subject, Department, Semester (denormalized)
- **Facts**: ScheduleEntry (star-schema with hours, utilization measures)

**Key Operations**:
- GetDashboardSummary: KPI aggregates (count, avg workload, completion %)
- GetWorkloadAnalytics: Per-teacher workload + weekly breakdown
- GetUtilizationAnalytics: Resource metrics (rooms, teachers, semesters)
- GetDepartmentMetrics / GetScheduleMetrics / GetScheduleHeatmap: Dashboard slices
- Export: PDF/Excel (future implementation)

**Event Types**: Consumes `teacher.created/updated`, `department.created`, `subject.created`, `schedule.generation_completed`

**Outbound**: PostgreSQL (analytics schema), NATS JetStream (ETL consumer)

### Module-Timetable (Schedule Generation & Management)

**Purpose**: Semester and schedule management with CSP-based generation.

**Port**: gRPC `:50054`

**Key Entities**:
- **Semester**: id, name, year, term, start_date, end_date, offered_subject_ids[], room_ids[], academic_year (computed), is_active
- **Room**: id, code, capacity, type (classroom|lab|lecture_hall), features[]
- **Schedule**: id, semester_id, status (pending/generating/completed/failed), entries[]
- **ScheduleEntry**: schedule_id, subject_id, teacher_id, room_id, day, period, week (denormalized with names)

**Key Operations**:
- CRUD: Semester, offerings, rooms
- **GenerateSchedule**: Trigger CSP solver (async) → `generating` → `completed`/`failed`
- **ListTimeSlots**: Reference time data (id, period, time range)
- **ListRooms**: Available rooms
- **GetSchedule**: Fetch with status + enriched entries
- **UpdateEntry**: Manual teacher assignment
- **SuggestTeachers**: Ranking by specialization + availability

**CSP Domain Service**:
- Variables: (teacher, room, day, period) per subject
- **Hard Constraints**: No conflicts, specialization match, room capacity
- **Soft Constraints**: Minimize workload imbalance, respect availability preferences
- **Algorithm**: AC-3 + Backtracking (MRV/LCV heuristics, 30s timeout)

**Event Types**: `semester.created/updated`, `schedule.generation_started/completed/failed`, `schedule.entry_assigned/updated`

**Outbound**: PostgreSQL (timetable schema), Module-Subject (gRPC), Module-HR (gRPC), NATS JetStream

## Data Flow

### Request Flow (HTTP → gRPC)

**Example: Create Teacher**

```
1. Client
   POST /api/hr/teachers
   {name: "John", email: "john@example.com", dept: "engineering"}
   Authorization: Bearer {token}

2. Core Service (HTTP Gateway)
   - Middleware: Extract + validate JWT
   - Middleware: Rate limit check
   - Handler: Parse JSON
   - Proxy: Forward to Module-HR gRPC

3. Module-HR (gRPC Server)
   - Validate input
   - Create Teacher aggregate
   - Save to PostgreSQL (hr.teachers)
   - Append event to hr.event_store
   - Publish NATS event "teacher.created"

4. Response
   200 OK
   {id: "t123", name: "John", email: "john@example.com", ...}
```

### Event Flow (Async)

**Example: Schedule Generation Completion**

```
1. Client
   POST /api/timetable/semesters/s123/generate
   → Core creates gRPC request
   → Module-Timetable starts CSP solver (async)
   → Returns status: "pending"

2. CSP Solver (in Module-Timetable)
   - Fetch semester + offerings from DB
   - Fetch teachers from HR (gRPC)
   - Fetch subjects from Subject (gRPC)
   - Solve constraints (30s timeout)
   - Append event to timetable.event_store: "schedule.generation_completed"
   - Publish NATS event "schedule.generation_completed"

3. Event Handlers (NATS subscribers)
   - Async logger: Log completion
   - Frontend notifier: Push notification via WebSocket (if connected)

4. Client Polling (every 3s)
   GET /api/timetable/schedules/sch456
   → Returns status: "completed"
   → Fetches entries
   → Renders calendar
```

### Student Registration Flow (Invite Code)

**Example: Student registration with invite code**

```
1. Admin
   POST /api/students/:id/invite-code
   → Module-Student generates code (e.g., "STUD-2024-ABC123")
   → Code hashed in DB (SHA-256), stored with expires_at + created_by
   → Returns plaintext code to admin (never stored unhashed)

2. Student
   POST /api/auth/register-student
   {
     "invite_code": "STUD-2024-ABC123",
     "email": "student@example.com",
     "password": "secure123",
     "full_name": "Jane Doe"
   }

3. Core Service
   - Step 1: Validate code → Module-Student.ValidateInviteCode (check unused + not expired)
   - Step 2: Create user with role=viewer (safe default)
   - Step 3: Redeem code atomically → Module-Student.RedeemInviteCode (WHERE used_at IS NULL prevents double-redeem)
   - Step 4: Upgrade user role to student
   - Rollback: If redemption fails, delete viewer user to avoid orphaned accounts
   → Returns access_token + refresh_token

4. Result
   - Invite code now linked to user_id, used_at=now()
   - User can now access student portal (/api/student/*)
   - Student record user_id_id now points to authenticated user
```

### Chat Agent Flow (WebSocket + Tool Execution)

**Example: "Create a subject called Math 101"**

```
1. Client
   WebSocket message:
   {
     type: "message",
     content: "Create a subject called Math 101 with 3 credits"
   }

2. Core Chat Gateway
   - Validate JWT (from query param)
   - Pass message to LLM with tool definitions
   - Support multiple providers (OpenAI, Claude, Gemini)
   - maxToolIterations: 10 (allows multi-step workflows)

3. LLM (OpenAI/Claude/Gemini)
   - Receives system prompt: "You are a university scheduling assistant"
   - Receives explicit workflow instructions: "Always call timetable.list_semesters first to get UUID before calling timetable.generate"
   - Receives tool schema:
     [
       {
         name: "timetable.list_semesters",
         description: "Fetch available semesters with UUIDs",
         parameters: {page, page_size}
       },
       {
         name: "timetable.generate",
         description: "Generate schedule for semester",
         parameters: {semester_id}
       },
       {
         name: "create_subject",
         description: "Create a new subject",
         parameters: {name, credits, department_id, ...}
       },
       ...
     ]
   - Analyzes message, calls tool: create_subject(name="Math 101", credits=3, ...)
   - Provider-specific metadata stored in ToolCall.ProviderMeta (e.g., Gemini's thoughtSignature)

4. Tool Executor (Self-Referential HTTP)
   - Receives LLM tool call (with ProviderMeta if needed for multi-turn history)
   - Validates parameters
   - Dispatches via HTTP to core's own API (selfURL + internal JWT token)
     Example: POST /api/timetable/semesters/{id}/generate (with internal JWT header)
   - Tool endpoint routes to appropriate module gRPC (Module-Subject, Module-HR, Module-Timetable)
   - Returns result to LLM
   - Note: selfURL is core's HTTP base URL (e.g., "http://localhost:8000" or "http://core:8080" in Docker)
   - Note: internalJWT is a service-level JWT with 24h TTL generated at startup

5. LLM Response
   - Streams confirmation: "I've created subject Math 101 with 3 credits"
   - Response markdown-rendered on frontend
   - WebSocket streaming to client

6. Client
   - Displays chat message with markdown formatting
   - Invalidates /subjects query
   - Refreshes subject list

**Example: "Generate a schedule for the current semester"** (Multi-step workflow)

```
1. LLM receives message: "Generate a schedule for the current semester"
2. Step 1: LLM calls timetable.list_semesters → Returns semester UUIDs
3. Step 2: LLM calls timetable.generate with returned semester UUID → Triggers CSP solver
4. Returns: "Schedule generation started for Semester X. Please check back in 30 seconds."
```
Note: maxToolIterations=10 enables complex workflows requiring multiple tool calls.

## API Endpoints (Complete Reference)

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
| POST | `/api/auth/login` | Core | Email/password login; returns access_token + refresh_token |
| POST | `/api/auth/register` | Core | Creates user with email/password |
| POST | `/api/auth/refresh` | Core | Refresh access token |
| GET | `/api/auth/me` | Core | Current user profile |
| GET | `/api/auth/oauth/google/login` | Core | Redirect to Google OAuth consent screen |
| GET | `/api/auth/oauth/google/callback` | Core | Google OAuth callback; validates code + exchanges for tokens |
| GET | `/api/auth/oauth/microsoft/login` | Core | Redirect to Microsoft OAuth consent screen |
| GET | `/api/auth/oauth/microsoft/callback` | Core | Microsoft OAuth callback; validates code + exchanges for tokens |
| POST | `/api/auth/oauth/exchange` | Core | Frontend exchanges short-lived auth code for access/refresh tokens |
| PATCH | `/api/users/:id/role` | Core | Admin/super_admin only: update user role + department_id |
| GET | `/api/dashboard/stats` | Core | Aggregate counts (teachers, departments, subjects) |
| **HR Module** | | | |
| GET | `/api/hr/teachers` | Module-HR | Paginated list: `{ data, total, page, page_size }` |
| POST | `/api/hr/teachers` | Module-HR | Create teacher |
| GET | `/api/hr/teachers/:id` | Module-HR | Single teacher |
| PATCH | `/api/hr/teachers/:id` | Module-HR | Update teacher |
| DELETE | `/api/hr/teachers/:id` | Module-HR | Soft delete |
| GET | `/api/hr/teachers/:id/availability` | Module-HR | Availability schedule: `{ availability: [{day_of_week, start_time, end_time}] }` (time strings) |
| PUT | `/api/hr/teachers/:id/availability` | Module-HR | Update availability (body: `{ available_slots: [{day_of_week, start_time, end_time}] }`) |
| GET | `/api/hr/departments` | Module-HR | Paginated list |
| POST | `/api/hr/departments` | Module-HR | Create department |
| **Subject Module** | | | |
| GET | `/api/subjects` | Module-Subject | Paginated list |
| POST | `/api/subjects` | Module-Subject | Create subject |
| GET | `/api/subjects/:id` | Module-Subject | Single subject |
| PATCH | `/api/subjects/:id` | Module-Subject | Update subject |
| DELETE | `/api/subjects/:id` | Module-Subject | Soft delete |
| GET | `/api/subjects/:id/prerequisites` | Module-Subject | Array of prerequisites |
| POST | `/api/subjects/:id/prerequisites` | Module-Subject | Add prerequisite |
| DELETE | `/api/subjects/:id/prerequisites/:prereqId` | Module-Subject | Remove prerequisite |
| GET | `/api/subjects/dag/full` | Module-Subject | Full DAG (nodes + edges) for all subjects |
| POST | `/api/subjects/dag/check-conflicts` | Module-Subject | Check prerequisite conflicts in subject set |
| **Timetable Module** | | | |
| GET | `/api/timetable/semesters` | Module-Timetable | Paginated list; tool: `timetable.list_semesters` |
| POST | `/api/timetable/semesters` | Module-Timetable | Create semester (body: name, year, term, start_date, end_date) |
| GET | `/api/timetable/semesters/:id` | Module-Timetable | Single semester (includes offered_subject_ids, room_ids, year, term, academic_year, is_active, time_slots, rooms) |
| POST | `/api/timetable/semesters/:id/offered-subjects` | Module-Timetable | Add subject offering (body: subject_id) |
| DELETE | `/api/timetable/semesters/:id/offered-subjects/:subjectId` | Module-Timetable | Remove subject offering |
| POST | `/api/timetable/semesters/:id/rooms` | Module-Timetable | Set semester rooms (body: room_ids[]) — gRPC: SetSemesterRooms |
| POST | `/api/timetable/semesters/:id/generate` | Module-Timetable | Trigger CSP schedule generation; returns status `generating` → `completed`/`failed` |
| GET | `/api/timetable/time-slots` | Module-Timetable | Reference time slots (day_of_week, period, start_time, end_time); gRPC: ListTimeSlots |
| GET | `/api/timetable/rooms` | Module-Timetable | List available rooms; gRPC: ListRooms |
| GET | `/api/timetable/schedules` | Module-Timetable | Paginated list |
| GET | `/api/timetable/schedules/:id` | Module-Timetable | Single schedule with enriched entries (subject_name, teacher_name, room_name) |
| PUT | `/api/timetable/schedules/:id/entries/:entryId` | Module-Timetable | Manual teacher assignment (body: teacher_id) |
| GET | `/api/timetable/suggest-teachers` | Module-Timetable | Query: subject_id, day_of_week, start_period, end_period; returns array |
| GET | `/api/timetable/schedules/:id/stream` | Module-Timetable | SSE stream of schedule generation progress |
| **Student Module** | | | |
| GET | `/api/students` | Module-Student | Admin-only paginated list; optional `department_id`, `status` filters |
| POST | `/api/students` | Module-Student | Admin-only create student |
| GET | `/api/students/:id` | Module-Student | Admin-only single active student |
| PATCH | `/api/students/:id` | Module-Student | Admin-only partial update |
| DELETE | `/api/students/:id` | Module-Student | Admin-only soft delete |
| POST | `/api/students/:id/invite-code` | Module-Student | Admin-only; generates single-use invite code |
| GET | `/api/students/:id/enrollments` | Module-Student | List enrollments for student; optional `subject_id` query param |
| GET | `/api/student/me` | Module-Student | Student self-service: current student profile |
| GET | `/api/student/enrollments` | Module-Student | Student self-service: list my enrollments |
| POST | `/api/student/enrollments` | Module-Student | Student self-service: request enrollment (semester_id, offered_subject_id) |
| GET | `/api/student/enrollments/check-prerequisites` | Module-Student | Student self-service: check prerequisites for subject (query: subject_id) |
| GET | `/api/student/transcript` | Module-Student | Student self-service: full transcript + GPA |
| POST | `/api/auth/register-student` | Core | Public: register student with invite code (code, email, password, full_name) |
| **Analytics** | | | |
| GET | `/api/analytics/dashboard` | Module-Analytics | KPI cards: teacher count, avg workload, schedule completion % |
| GET | `/api/analytics/workload` | Module-Analytics | Workload analytics per teacher with period breakdown |
| GET | `/api/analytics/utilization` | Module-Analytics | Resource utilization metrics (rooms, teachers, semesters) |
| GET | `/api/analytics/department-metrics` | Module-Analytics | Department-level metrics (teachers per dept, specialization coverage) |
| GET | `/api/analytics/schedule-metrics` | Module-Analytics | Schedule metrics (completion rate, conflicts, constraints) |
| GET | `/api/analytics/schedule-heatmap` | Module-Analytics | Schedule density heatmap (day/period utilization) |
| GET | `/api/analytics/export` | Core proxy route reserved for future analytics export surface |
| **Audit & Compliance** | | | |
| GET | `/api/audit-logs` | Core | Admin/super_admin only: paginated audit logs with filters (user_id, resource_type, action, date range) |
| **Chat** | | | |
| WebSocket | `/ws/chat?token=ACCESS_TOKEN` | Core | Streaming chat interface |

## Database Schema (Logical)

### Core Schema
```sql
users (
  id: uuid primary key,
  email: string unique,
  password_hash: string nullable,  -- NULL for OAuth-only accounts
  role: enum(super_admin, admin, dean, dept_head, teacher, student),
  department_id: uuid nullable,  -- Foreign ref to hr.departments (app-level validation)
  oauth_provider: string nullable,  -- "google" or "microsoft"
  oauth_subject: string nullable,  -- Provider's unique user ID (sub claim)
  avatar_url: text nullable,  -- Profile picture from OAuth provider
  created_at: timestamp,
  deleted_at: timestamp nullable,
  unique index: (oauth_provider, oauth_subject) where oauth_provider is not null
)

module_registry (
  id: uuid primary key,
  name: string unique,
  version: string,
  grpc_addr: string,
  health_check_url: string,
  status: enum(healthy, unhealthy),
  registered_at: timestamp
)

conversations (
  id: uuid primary key,
  user_id: uuid fk users,
  messages: jsonb [{ role, content, timestamp }],
  created_at: timestamp,
  updated_at: timestamp
)

event_store (
  id: bigint primary key auto,
  aggregate_id: uuid,
  aggregate_type: string, -- "user", "module", etc.
  event_type: string,
  payload: jsonb,
  timestamp: timestamp,
  version: bigint -- optimistic concurrency
)
```

### HR Schema
```sql
departments (
  id: uuid primary key,
  name: string,
  created_at: timestamp,
  deleted_at: timestamp nullable
)

teachers (
  id: uuid primary key,
  name: string,
  email: string unique,
  user_id: uuid nullable fk core.users,  -- Links teacher to authenticated user for login
  department_id: uuid fk departments,
  created_at: timestamp,
  deleted_at: timestamp nullable
)

teacher_availability (
  teacher_id: uuid fk teachers,
  day_of_week: enum(monday...sunday),
  periods: int[] [1,2,3,4,5], -- which periods available
  primary key (teacher_id, day_of_week)
)

teacher_specializations (
  teacher_id: uuid fk teachers,
  subject_id: uuid,
  primary key (teacher_id, subject_id)
)

event_store (same pattern as core)
```

### Subject Schema
```sql
subjects (
  id: uuid primary key,
  code: string unique,
  name: string,
  credits: int,
  weekly_hours: int,
  description: text,
  department_id: uuid fk hr.departments,
  created_at: timestamp,
  deleted_at: timestamp nullable
)

prerequisites (
  id: uuid primary key,
  subject_id: uuid fk subjects,
  prerequisite_subject_id: uuid fk subjects,
  prerequisite_type: enum(strict, recommended, corequisite),
  priority: int [1..5], -- soft: -2 to +2 hard
  created_at: timestamp
)

event_store (same pattern)
```

### Student Schema
```sql
students (
  id: uuid primary key,
  student_code: string unique,
  user_id: uuid nullable fk core.users,
  full_name: string,
  email: string unique,
  department_id: uuid fk hr.departments,
  enrollment_year: int,
  status: string,
  is_active: bool,
  created_at: timestamp,
  updated_at: timestamp
)

invite_codes (
  code: string primary key,  -- SHA-256 hash of plaintext code
  created_by_user_id: uuid fk core.users,
  used_at: timestamp nullable,  -- NULL = unused; TOCTOU protection via WHERE used_at IS NULL
  expires_at: timestamp,
  created_at: timestamp
)

enrollments (
  id: uuid primary key,
  student_id: uuid fk students,
  offered_subject_id: uuid,
  semester_id: uuid,
  status: string,
  created_at: timestamp
)

grades (
  id: uuid primary key,
  enrollment_id: uuid fk enrollments,
  grade_numeric: decimal,
  grade_letter: string,
  created_at: timestamp
)

event_store (same pattern)

audit_logs (
  id: bigint primary key auto,
  user_id: uuid fk core.users,
  resource_type: string,  -- "teacher", "subject", "semester", "enrollment", "grade", etc.
  action: enum(create, read, update, delete),
  old_value: jsonb nullable,  -- Previous state (null for creates)
  new_value: jsonb,  -- Current state (null for deletes)
  timestamp: timestamp,
  -- 12 monthly partitions: 2026-03 through 2027-02
  -- BRIN index on timestamp, B-tree on (user_id, resource_type) for filtering
)
```

### Timetable Schema
```sql
semesters (
  id: uuid primary key,
  name: string,
  year: int,
  term: int [1..3],  -- 1=fall, 2=spring, 3=summer (internal enum)
  start_date: date,
  end_date: date,
  room_ids: uuid[] default '{}',  -- Available rooms for this semester (empty = any room)
  created_at: timestamp,
  deleted_at: timestamp nullable
)

semester_offerings (
  semester_id: uuid fk semesters,
  subject_id: uuid fk subject.subjects,
  class_size: int,
  primary key (semester_id, subject_id)
)

rooms (
  id: uuid primary key,
  code: string unique,
  capacity: int,
  created_at: timestamp,
  deleted_at: timestamp nullable
)

time_slots (
  day_of_week: enum(monday...sunday),
  period_of_day: int [1..5],
  primary key (day_of_week, period_of_day)
)

schedules (
  id: uuid primary key,
  semester_id: uuid fk semesters,
  status: enum(pending, generating, completed, failed),
  created_at: timestamp,
  completed_at: timestamp nullable
)

schedule_entries (
  id: uuid primary key,
  schedule_id: uuid fk schedules,
  subject_id: uuid,
  teacher_id: uuid,
  room_id: uuid fk rooms,
  day_of_week: enum(monday...sunday),
  period_of_day: int [1..5],
  week_of_semester: int,
  -- Denormalized fields for efficient API responses
  subject_name: string,
  subject_code: string,
  teacher_name: string,
  room_name: string,
  department_id: uuid
)

event_store (same pattern)
```

## Interaction Patterns

**Request-Response (Sync)**:
- HTTP ↔ Core: Standard REST
- Core ↔ Module: gRPC (5s timeout)
- Example: POST /api/subjects/dag/check-conflicts returns {has_conflicts, conflicts[]}

**Event-Driven (Async)**:
- NATS JetStream: Durability + ordering
- Consumers: Logs, cache invalidation, notifications
- Frontend: WebSocket for real-time events

**Polling**:
- Schedule status: Every 3s until completed
- Module health: Every 30s

## Scalability & High Availability

### Horizontal Scaling
- **Stateless Services**: Each module is stateless; scale with k8s Deployment
- **Database**: PostgreSQL connection pooling (configurable per service)
- **NATS**: JetStream clusters (plan for Phase 2)

### Circuit Breaker (Future)
- Wrap gRPC calls to modules with retry + circuit breaker
- Fallback: Return cached data or partial response

### Monitoring (Future)
- **Prometheus Metrics**: Per service, exported on :9090
- **OpenTelemetry**: Distributed tracing across services
- **ELK Stack**: Centralized logging (Zap JSON output → Logstash)

## Security Boundaries

### Authentication

#### Traditional JWT Auth
- **Boundary**: JWT token validation at Core gateway + gRPC interceptor
- **Token Lifetime**: 15min access + 7day refresh
- **Refresh Flow**: Client sends refresh token → Core issues new access token
- **Claims Extension**: JWT includes `user_id`, `role`, `department_id` (scope), `teacher_id` for O(1) permission checks

#### OAuth 2.0 / OIDC Integration
- **Providers**: Google (teachers @hcmus.edu.vn), Microsoft Entra ID (students @student.hcmus.edu.vn)
- **Flow**: PKCE-secured authorization code flow
  1. Frontend: User clicks "Login with [Google/Microsoft]" → redirects to backend `/api/auth/oauth/{provider}/login`
  2. Backend: Generate state + PKCE verifier, store in secure httpOnly cookie, redirect to provider consent screen
  3. Provider: User grants permissions, redirects to callback with auth code
  4. Backend: Validate state, exchange code for ID token (PKCE), validate ID token (issuer, audience, nonce)
  5. Validate domain claim: `hd` (Google) or `tid` (Microsoft) server-side — reject if not institutional domain
  6. Lookup or upsert user by email: pre-existing teacher/student record required (admin must pre-create)
  7. Generate short-lived auth code (UUID, 60s TTL) in in-memory store, redirect frontend to `/auth/callback?code=X`
  8. Frontend: Exchange code via `POST /api/auth/oauth/exchange` → returns access + refresh tokens
- **User Linking**: Email-based matching; OAuth account linked to existing teacher/student on first login
- **Token Security**: Auth codes never in response body; all tokens in secure httpOnly cookies
- **Optional Initialization**: OAuthService only initialized if `oauth.google.client_id` set in config (graceful disable if not configured)
- **Config**: Stored in `config/local.yaml` (secrets via env vars: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `MICROSOFT_CLIENT_ID`, `MICROSOFT_CLIENT_SECRET`, `MICROSOFT_TENANT_ID`)

### Authorization (RBAC)
- **Roles**: 6 roles: `super_admin`, `admin`, `dean`, `dept_head`, `teacher`, `student`
- **Enforcement**: Two-tier — middleware (role + dept scope) + handler (resource ownership)
- **Scope Binding**: `dept_head` and `teacher` roles scoped to `department_id` in JWT claims
- **Admin-Only**: Module management, user deletion, role assignment
- **Middleware Guards**: `RequireRole()` + `RequireDeptScope()` on protected routes
- **Bypass**: `super_admin`, `admin`, `service` bypass scope checks; `dean` read-only bypass

### Data Isolation
- **Schema-per-Module**: Reduces blast radius if one module is compromised
- **Database Credentials**: Same user (myrmex) for all; future: service-specific users
- **Event Audit Trail**: All writes in event_store (compliance)

### Secrets Management
- **API Keys**: Stored in `config/local.yaml` (Git-ignored)
- **Database URL**: Env var (`DATABASE_URL`)
- **JWT Secret**: Env var (`JWT_SECRET`)
- **LLM Key**: Env var (`CLAUDE_API_KEY` or `OPENAI_API_KEY`)
- **Future**: HashiCorp Vault integration

## Failure Modes & Recovery

| Failure | Impact | Recovery |
|---------|--------|----------|
| **Core Down** | API calls fail | Restart; client buffering |
| **Module Down** | Operations fail | Restart; circuit breaker + cached data |
| **PostgreSQL Down** | Data access fails | Failover to replica (Phase 4) |
| **NATS Down** | Events lost | Restart; replay from event_store (Phase 4) |
| **CSP Timeout** | Schedule incomplete | Return best partial solution |

## Performance Targets (MVP)

- API Latency (p95): ~300ms (excl. CSP)
- CSP Solver: ~20s for 100-subject semesters (30s timeout)
- DB Query: ~50ms (sqlc optimized)
- gRPC Call: ~100ms (local network)
- WebSocket Latency: ~50ms
- Frontend Bundle: ~80KB gzipped (Vite tree-shaking)
- Memory per Service: ~100MB (Go efficiency)

## Future Enhancements

1. **Multi-Tenancy**: Institution isolation (Phase 4)
2. **HA Setup**: PostgreSQL replication, NATS clustering, Core redundancy (Phase 2)
3. **Distributed Tracing**: OpenTelemetry (Phase 3)
4. **Service Mesh**: Istio (Phase 4)
5. **Caching Layer**: Redis cache-aside pattern for frequently accessed data
6. **Full-Text Search**: Elasticsearch (Phase 3)
7. **Data Warehouse**: BI tools (Phase 4)
