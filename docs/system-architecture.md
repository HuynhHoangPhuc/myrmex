# System Architecture

## High-Level Overview

Myrmex is a microservice architecture with modular services communicating via gRPC and event streaming through NATS JetStream. The HTTP gateway (Core) proxies requests to gRPC services; the AI chat agent orchestrates operations via a dynamic tool registry.

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
│                    Core Service (port 8000)                             │
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
│  │  ├─ ANY  /api/analytics/*          → Module-Analytics HTTP    │   │
│  │  └─ WebSocket /ws/chat?token=X    → ChatGateway (Streaming)   │   │
│  │                                                                │   │
│  │  Middleware:                                                   │   │
│  │  ├─ CORS (origin validation)                                  │   │
│  │  ├─ Auth (JWT extraction + validation)                        │   │
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
│  ├─ Tool registry (dynamic registration)                               │
│  └─ Event streaming to frontend                                        │
└─────────────────────────────────────────────────────────────────────────┘
              │          │           │              │              │
              │ gRPC     │ gRPC      │ gRPC         │ gRPC         │ HTTP
              │          │           │              │              │
┌─────────────▼──┐  ┌────▼──────┐  ┌─▼──────────┐  ┌┴──────────────┐  ┌──────────────┐
│  Module-HR     │  │  Module-  │  │ Module-    │  │ User/Auth     │  │ Module-      │
│  Service       │  │ Subject   │  │ Timetable  │  │ (in Core)     │  │ Analytics    │
│  (port 50052)  │  │ (50053)   │  │ (50054)    │  │ (50051)       │  │ (8080 HTTP)  │
│                │  │           │  │            │  │               │  │              │
│ Department     │  │ Subject   │  │ Semester   │  │ User Mgmt     │  │ Star Schema  │
│ Teacher CRUD   │  │ DAG       │  │ Room       │  │ JWT Auth      │  │ Dashboard KPIs
│ Availability   │  │ Prereq    │  │ Schedule   │  │ Refresh Token │  │ Workload     │
│                │  │ Service   │  │ CSP Solver │  │               │  │ Utilization  │
│ Domain:        │  │ Service   │  │ Service    │  │ Domain:       │  │ Export PDF/XL│
│ ├─ Department  │  │           │  │            │  │ ├─ User       │  │              │
│ ├─ Teacher     │  │ Domain:   │  │ Domain:    │  │ └─ Session    │  │ Dimensions:  │
│ ├─ Availability│  │ ├─ Subject│  │ ├─ Semester│  │               │  │ ├─ Teacher   │
│ └─ Specialization│ │ ├─ Prereq│  │ ├─ Schedule│  │               │  │ ├─ Subject   │
└────────┬───────┘  │ └─ DAG    │  │ ├─ Room    │  └───────────────┘  │ ├─ Department│
         │          │           │  │ └─ TimeSlot│                     │ ├─ Semester  │
         │          └─────┬──────┘  │            │                     │ └─ Facts     │
         │                │         └─────┬──────┘                     │ (schedules)  │
         │                │               │                           │              │
         └────────────────┼───────────────┼───────────────────────────┘
         │          │           │  │ └─ TimeSlot│
         │          └─────┬──────┘  │            │
         │                │         └─────┬──────┘
         │                │               │
         └────────────────┼───────────────┘
                          │
         NATS JetStream (Event Bus, port 4222)
                          │
         ┌────────────────┴────────────────────────────────────┐
         │                                                     │
    ┌────▼─────┐  ┌──────────┐  ┌────────────┐  ┌──────────┐
    │ Event    │  │ Async    │  │ Frontend   │  │ Analytics│
    │ Store    │  │ Consumers│  │ WebSocket  │  │ Consumer │
    │ (Append) │  │ (Listen) │  │ (Stream)   │  │ (ETL)    │
    └──────────┘  └──────────┘  └────────────┘  └──────────┘
         │                                            │
┌────────▼─────────────────────────────────────────────────────────┐
│     PostgreSQL (Shared Database)                                  │
│     localhost:5432 (myrmex / myrmex_dev)                          │
│                                                                   │
│ Operational Schemas:                                             │
│ ├─ core                                                          │
│ │  ├─ users, module_registry, conversations, event_store        │
│ ├─ hr (departments, teachers, availability, specializations)    │
│ ├─ subject (subjects, prerequisites, event_store)               │
│ └─ timetable (semesters, offerings, rooms, schedules, events)   │
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
    │ (spare)│  │ Storage │
    └────────┘  └─────────┘
```

## Service Topology

### Core Service (HTTP Gateway + Auth + AI Chat)

**Purpose**: Entry point for all HTTP requests; manages authentication and delegates to gRPC services.

**Ports**:
- HTTP: `:8080`
- gRPC: `:50051`
- Metrics: `:9090` (future)

**Key Responsibilities**:
1. **HTTP Gateway**: Proxy requests to gRPC services
2. **Authentication**: JWT token generation + validation
3. **User Management**: User CRUD, roles
4. **Module Registry**: Service discovery + health checks
5. **AI Chat**: WebSocket endpoint, tool execution, LLM integration

**Outbound Dependencies**:
- PostgreSQL (core schema: users, event_store, conversations)
- NATS JetStream (publish events)
- LLM API (OpenAI, Claude, or Gemini for LLM inference)

**Inbound Dependencies**:
- Module-HR (gRPC)
- Module-Subject (gRPC)
- Module-Timetable (gRPC)

### Module-HR (Department & Teacher Management)

**Purpose**: Faculty data management.

**Port**: gRPC `:50052`

**Key Entities**:
- **Department**: id, name, created_at, deleted_at
- **Teacher**: id, name, email, department_id, specializations[], availability[]
- **Availability**: {day_of_week, start_time, end_time} (value object) — Time slots in RFC3339 format (HH:MM)
  - Example: {day_of_week: "MONDAY", start_time: "07:00", end_time: "19:00"}
  - Periods 1-6 mapped to: 07:00-09:00, 09:00-11:00, 11:00-13:00, 13:00-15:00, 15:00-17:00, 17:00-19:00

**Key Operations**:
- CRUD: Create/Read/Update/Delete department
- CRUD: Create/Read/Update/Delete teacher (includes availability in response)
- **GetTeacher**: Returns `availability: [{day_of_week, start_time, end_time}]` as time strings
- **GetTeacherAvailability**: Fetch availability with time-based representation
- **UpdateTeacherAvailability**: Accept `{available_slots: [{day_of_week, start_time, end_time}]}`, store as period integers
- **Assign specializations**: Link teachers to subject specializations

**Domain Services**:
- Availability validation
- Soft delete logic

**Event Types**:
- `teacher.created`, `teacher.updated`, `teacher.deleted`
- `department.created`, `department.updated`, `department.deleted`
- `teacher.availability_updated`

**Outbound**:
- PostgreSQL (hr schema)
- NATS JetStream (publish events)

### Module-Subject (Subject & Prerequisite DAG)

**Purpose**: Course structure with prerequisite management.

**Port**: gRPC `:50053`

**Key Entities**:
- **Subject**: id, code, name, credits, weekly_hours, department_id, is_active
- **Prerequisite**: subject_id, prerequisite_subject_id, type (strict/recommended/corequisite), priority (1-5 soft: -2 to +2 hard)
- **DAG**: In-memory representation for cycle detection + topological sort
- **DAGNode**: Subject with metadata for visualization (id, code, name, credits, department_id)
- **DAGEdge**: Prerequisite link with type and priority for conflict detection

**Key Operations**:
- CRUD: Subject
- Add prerequisite (validates no cycles via DFS 3-color)
- Remove prerequisite
- Topological sort (all DAGs starting from a subject)
- Validate DAG (full cycle check)
- **GetFullDAG**: Returns all subjects + prerequisite edges (for DAG visualization)
- **CheckPrerequisiteConflicts**: Detects missing prerequisites in a subject set

**Domain Services**:
- **DAGService**: Cycle detection (DFS with 3 colors: white/gray/black)
- **DAGService**: Topological sort (BFS layers)

**Event Types**:
- `subject.created`, `subject.updated`, `subject.deleted`
- `prerequisite.added`, `prerequisite.removed`
- `prerequisite_dag.validated` (async cycle check)

**Outbound**:
- PostgreSQL (subject schema)
- NATS JetStream (publish events)

### Module-Analytics (Analytics & Reporting)

**Purpose**: Business intelligence and reporting on resource utilization.

**Port**: HTTP `:8080` (reverse-proxied via Core gateway at `/api/analytics/`)

**Key Entities**:
- **Dimensions**: Teacher, Subject, Department, Semester (denormalized)
- **Facts**: ScheduleEntry (star-schema fact table with measures: hours, utilization)

**Key Operations**:
- GetDashboardSummary: KPI aggregates (teachers count, avg workload, schedule completion %)
- GetWorkloadAnalytics: Per-teacher workload summary with weekly breakdown
- GetUtilizationAnalytics: Resource utilization metrics (rooms, teachers, semesters)
- ExportSchedulePDF: PDF export of semester schedule
- ExportScheduleExcel: Excel export with multi-sheet layout

**Domain Services**:
- **AnalyticsRepository**: Query dimension & fact tables
- **ExportService**: PDF/Excel report generation (iText-based)

**Event Types**:
- Consumes: `teacher.created`, `teacher.updated`, `department.created`, `subject.created`, `schedule.generation_completed`
- Triggers ETL via NATS consumer (nightly or on-demand)

**Outbound**:
- PostgreSQL (analytics schema)
- NATS JetStream (consumes events)

### Module-Timetable (Schedule Generation & Management)

**Purpose**: Semester + schedule management with CSP-based generation.

**Port**: gRPC `:50054`

**Key Entities**:
- **Semester**: id, name, year, term, start_date, end_date, offered_subject_ids[], academic_year (computed), is_active (computed)
- **Room**: id, code, capacity
- **TimeSlot**: day_of_week, period_of_day (reference data)
- **Schedule**: id, semester_id, status (pending/generating/completed/failed), entries[]
- **ScheduleEntry**: schedule_id, subject_id, teacher_id, room_id, day, period, week, subject_name, teacher_name, room_name (denormalized)

**Key Operations**:
- CRUD: Semester + offerings
- CRUD: Room
- **ListTimeSlots**: Fetch reference time slot data (day_of_week, period, time range) — RPC: `ListTimeSlots`
- **ListRooms**: Fetch available rooms (id, code, capacity) — RPC: `ListRooms`
- **GenerateSchedule**: Trigger CSP solver (async) — returns status `generating` → `completed` or `failed`
- **ListSchedules**: Paginated list with optional semester filter
- **GetSchedule**: Fetch generated schedule with status + enriched entries
- **UpdateEntry**: Manual teacher assignment with validation
- **SuggestTeachers**: Ranking by specialization match + availability

**Domain Services**:
- **CSPService**: Constraint satisfaction problem solver
  - Variables: (teacher, room, day, period) per subject
  - **Hard Constraints**:
    - No teacher time conflicts
    - Teacher specialization matches subject
    - Room capacity ≥ subject class size
  - **Soft Constraints**:
    - Minimize teacher workload imbalance
    - Respect teacher availability preferences
  - **Algorithm**: AC-3 (arc consistency) + Backtracking
    - MRV heuristic: Choose most constrained variable
    - LCV heuristic: Choose least constraining value
    - Context timeout (30s): Return best partial solution
  - **Output**: Ordered list of schedule entries

**Event Types**:
- `semester.created`, `semester.updated`
- `schedule.generation_started`, `schedule.generation_completed`, `schedule.generation_failed`
- `schedule.entry_assigned`, `schedule.entry_updated`

**Outbound**:
- PostgreSQL (timetable schema)
- Module-Subject (gRPC): Fetch subject details, validate prerequisites
- Module-HR (gRPC): Fetch teacher details, validate specializations + availability
- NATS JetStream (publish events)

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
| POST | `/api/auth/login` | Core | Returns access_token + refresh_token |
| POST | `/api/auth/register` | Core | Creates user with email/password |
| POST | `/api/auth/refresh` | Core | Refresh access token |
| GET | `/api/auth/me` | Core | Current user profile |
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
| GET | `/api/timetable/semesters/:id` | Module-Timetable | Single semester (includes offered_subject_ids, year, term, academic_year, is_active, time_slots, rooms) |
| POST | `/api/timetable/semesters/:id/offered-subjects` | Module-Timetable | Add subject offering (body: subject_id) |
| DELETE | `/api/timetable/semesters/:id/offered-subjects/:subjectId` | Module-Timetable | Remove subject offering |
| POST | `/api/timetable/semesters/:id/generate` | Module-Timetable | Trigger CSP schedule generation; returns status `generating` → `completed`/`failed` |
| GET | `/api/timetable/time-slots` | Module-Timetable | Reference time slots (day_of_week, period, start_time, end_time); gRPC: ListTimeSlots |
| GET | `/api/timetable/rooms` | Module-Timetable | List available rooms; gRPC: ListRooms |
| GET | `/api/timetable/schedules` | Module-Timetable | Paginated list |
| GET | `/api/timetable/schedules/:id` | Module-Timetable | Single schedule with enriched entries (subject_name, teacher_name, room_name) |
| PUT | `/api/timetable/schedules/:id/entries/:entryId` | Module-Timetable | Manual teacher assignment (body: teacher_id) |
| GET | `/api/timetable/suggest-teachers` | Module-Timetable | Query: subject_id, day_of_week, start_period, end_period; returns array |
| GET | `/api/timetable/schedules/:id/stream` | Module-Timetable | SSE stream of schedule generation progress |
| **Analytics** | | | |
| GET | `/api/analytics/dashboard-summary` | Module-Analytics | KPI cards: teacher count, avg workload, schedule completion % |
| GET | `/api/analytics/workload` | Module-Analytics | Workload analytics per teacher with period breakdown |
| GET | `/api/analytics/utilization` | Module-Analytics | Resource utilization metrics (rooms, teachers, semesters) |
| GET | `/api/analytics/department-metrics` | Module-Analytics | Department-level metrics (teachers per dept, specialization coverage) |
| GET | `/api/analytics/schedule-metrics` | Module-Analytics | Schedule metrics (completion rate, conflicts, constraints) |
| GET | `/api/analytics/schedule-heatmap` | Module-Analytics | Schedule density heatmap (day/period utilization) |
| GET | `/api/analytics/export/pdf?semester_id=:id` | Module-Analytics | PDF schedule export |
| GET | `/api/analytics/export/excel?semester_id=:id` | Module-Analytics | Excel schedule export |
| **Chat** | | | |
| WebSocket | `/ws/chat?token=ACCESS_TOKEN` | Core | Streaming chat interface |

## Database Schema (Logical)

### Core Schema
```sql
users (
  id: uuid primary key,
  email: string unique,
  password_hash: string,
  role: enum(admin, faculty_coordinator, academician, scheduler),
  created_at: timestamp,
  deleted_at: timestamp nullable
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

### Timetable Schema
```sql
semesters (
  id: uuid primary key,
  name: string,
  year: int,
  term: int [1..3],  -- 1=fall, 2=spring, 3=summer (internal enum)
  start_date: date,
  end_date: date,
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

### Request-Response (Synchronous)
- **HTTP ↔ Core**: Standard REST request/response
- **Core ↔ Module gRPC**: Blocking call with timeout (5s default)
- **Module ↔ Module gRPC**: Blocking call with timeout

**Example: Prerequisite Conflict Detection**
```
1. Frontend: POST /api/subjects/dag/check-conflicts
   { subject_ids: ["math101", "physics201", "cs302"] }
2. Core routes to Module-Subject gRPC
3. Module-Subject checks prerequisites:
   - physics201 requires math101 ✓
   - cs302 requires math101 ✓
4. Returns: { has_conflicts: false, conflicts: [] }
   Or: { has_conflicts: true, conflicts: [{subject: "advanced-calc", missing: "linear-algebra"}] }
5. Frontend renders ConflictWarningBanner with "Add missing" button if conflicts exist
```

### Event-Driven (Asynchronous)
- **NATS JetStream**: Durability + ordering
- **Consumers**: Log handlers, cache invalidators, email notifiers
- **Frontend**: WebSocket subscribed to relevant events

### Polling (Frontend)
- **Schedule Status**: Poll every 3s until completed/failed
- **Module Health**: Periodic health check (30s)

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
- **Boundary**: JWT token validation at Core gateway + gRPC interceptor
- **Token Lifetime**: 15min access + 7day refresh
- **Refresh Flow**: Client sends refresh token → Core issues new access token

### Authorization (RBAC)
- **Roles**: admin, faculty_coordinator, academician, scheduler
- **Enforcement**: Per gRPC method in interceptor (future: fine-grained)
- **Admin-Only**: Module management, user deletion

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
| **Core Down** | All API calls fail | Restart Core; requests queue client-side |
| **Module-HR Down** | HR operations fail; other modules work | Restart HR; circuit breaker returns cached data |
| **Module-Subject Down** | Subject ops fail; timetable can't fetch subjects | Restart Subject; cache prerequisites in Module-Timetable |
| **PostgreSQL Down** | All data access fails | Failover to replica (future); requests buffer in client |
| **NATS Down** | Events not persisted; real-time features fail | Restart NATS; replay events from event_store (future) |
| **LLM API Down** | Chat unavailable | Graceful error message; queue requests (future) |
| **CSP Timeout** | Schedule generation incomplete | Return best partial solution; mark as "partial" |

## Performance Characteristics

| Component | Target | Actual (MVP) | Notes |
|-----------|--------|--------------|-------|
| API Latency (p95) | <500ms | ~300ms | Excl. CSP solver |
| CSP Solver | <30s (p95) | ~20s for 100-subject semesters | Context cancellation → partial |
| DB Query | <100ms (p95) | ~50ms | Via sqlc + optimized indexes |
| gRPC Call | <200ms (p95) | ~100ms | Local network |
| WebSocket Latency | <200ms | ~50ms | Direct connection |
| Frontend Bundle | <100KB gzipped | ~80KB | Vite tree-shaking |
| Memory per Service | <200MB | ~100MB | Go efficiency |

## Future Enhancements

1. **Multi-Tenancy**: Isolate data per institution (Phase 4)
2. **HA Setup**: PostgreSQL replication, NATS clustering, Core redundancy (Phase 2)
3. **Distributed Tracing**: OpenTelemetry integration (Phase 3)
4. **Service Mesh**: Istio for advanced traffic management (Phase 4)
5. **Caching Layer**: Redis cache-aside pattern for frequently accessed data (Phase 2)
6. **Search**: Elasticsearch for full-text search (Phase 3)
7. **Analytics**: Data warehouse + BI tools (Phase 4)
