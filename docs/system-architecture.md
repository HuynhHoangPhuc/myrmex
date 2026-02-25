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
│  │  ├─ GET  /api/users/me             → UserService gRPC         │   │
│  │  ├─ ANY  /api/hr/*                 → Module-HR gRPC (proxy)   │   │
│  │  ├─ ANY  /api/subjects/*           → Module-Subject gRPC      │   │
│  │  ├─ ANY  /api/timetable/*          → Module-Timetable gRPC    │   │
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
              │          │           │              │
              │ gRPC     │ gRPC      │ gRPC         │ gRPC
              │          │           │              │
┌─────────────▼──┐  ┌────▼──────┐  ┌─▼──────────┐  ┌┴──────────────┐
│  Module-HR     │  │  Module-  │  │ Module-    │  │ User/Auth     │
│  Service       │  │ Subject   │  │ Timetable  │  │ (in Core)     │
│  (port 50052)  │  │ (50053)   │  │ (50054)    │  │ (50051)       │
│                │  │           │  │            │  │               │
│ Department     │  │ Subject   │  │ Semester   │  │ User Mgmt     │
│ Teacher CRUD   │  │ DAG       │  │ Room       │  │ JWT Auth      │
│ Availability   │  │ Prereq    │  │ Schedule   │  │ Refresh Token │
│                │  │ Service   │  │ CSP Solver │  │               │
│ Domain:        │  │ Service   │  │ Service    │  │               │
│ ├─ Department  │  │           │  │            │  │ Domain:       │
│ ├─ Teacher     │  │ Domain:   │  │ Domain:    │  │ ├─ User       │
│ ├─ Availability│  │ ├─ Subject│  │ ├─ Semester│  │ └─ Session    │
│ └─ Specialization│ │ ├─ Prereq│  │ ├─ Schedule│  │               │
└────────┬───────┘  │ └─ DAG    │  │ ├─ Room    │  └───────────────┘
         │          │           │  │ └─ TimeSlot│
         │          └─────┬──────┘  │            │
         │                │         └─────┬──────┘
         │                │               │
         └────────────────┼───────────────┘
                          │
         NATS JetStream (Event Bus, port 4222)
                          │
         ┌────────────────┴────────────────┐
         │                                 │
    ┌────▼─────┐  ┌──────────┐  ┌────────▼──┐
    │ Event    │  │ Async    │  │ Frontend  │
    │ Store    │  │ Consumers│  │ WebSocket │
    │ (Append) │  │ (Listen) │  │ (Stream)  │
    └──────────┘  └──────────┘  └───────────┘
         │
┌────────▼──────────────────────────────────┐
│     PostgreSQL (Shared Database)           │
│     localhost:5432 (myrmex / myrmex_dev)   │
│                                            │
│ Schemas (per service):                    │
│ ├─ core                                    │
│ │  ├─ users                                │
│ │  ├─ module_registry                      │
│ │  ├─ conversations                        │
│ │  └─ event_store                          │
│ ├─ hr                                      │
│ │  ├─ departments                          │
│ │  ├─ teachers                             │
│ │  ├─ teacher_availability                 │
│ │  ├─ teacher_specializations              │
│ │  └─ event_store                          │
│ ├─ subject                                 │
│ │  ├─ subjects                             │
│ │  ├─ prerequisites                        │
│ │  └─ event_store                          │
│ └─ timetable                               │
│    ├─ semesters                            │
│    ├─ semester_offerings                   │
│    ├─ rooms                                │
│    ├─ schedules                            │
│    ├─ schedule_entries                     │
│    ├─ time_slots                           │
│    └─ event_store                          │
└────────────────────────────────────────────┘
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
- HTTP: `:8000`
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
- Claude/OpenAI API (LLM inference)

**Inbound Dependencies**:
- Module-HR (gRPC)
- Module-Subject (gRPC)
- Module-Timetable (gRPC)

### Module-HR (Department & Teacher Management)

**Purpose**: Faculty data management.

**Port**: gRPC `:50052`

**Key Entities**:
- **Department**: id, name, created_at, deleted_at
- **Teacher**: id, name, email, department_id, specializations[], availability{}
- **Availability**: day_of_week → periods[] (value object)

**Key Operations**:
- CRUD: Create/Read/Update/Delete department
- CRUD: Create/Read/Update/Delete teacher
- Update teacher availability (days + periods per day)
- Assign specializations to teachers

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
- **Subject**: id, code, name, credits, weekly_hours, department_id
- **Prerequisite**: subject_id, prerequisite_subject_id, type (strict/recommended/corequisite), priority (1-5)
- **DAG**: In-memory representation for cycle detection + topological sort

**Key Operations**:
- CRUD: Subject
- Add prerequisite (validates no cycles via DFS 3-color)
- Remove prerequisite
- Topological sort (all DAGs starting from a subject)
- Validate DAG (full cycle check)

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

### Module-Timetable (Schedule Generation & Management)

**Purpose**: Semester + schedule management with CSP-based generation.

**Port**: gRPC `:50054`

**Key Entities**:
- **Semester**: id, name, year, term, start_date, end_date, offered_subjects[]
- **Room**: id, code, capacity
- **TimeSlot**: day_of_week, period_of_day (reference data)
- **Schedule**: id, semester_id, status (pending/generating/completed/failed), entries[]
- **ScheduleEntry**: schedule_id, subject_id, teacher_id, room_id, day, period, week

**Key Operations**:
- CRUD: Semester + offerings
- CRUD: Room
- **GenerateSchedule**: Trigger CSP solver (async)
- GetSchedule: Fetch generated schedule with status
- UpdateEntry: Manual assignment with validation
- SuggestTeachers: Ranking by specialization match + availability

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

3. LLM (Claude/OpenAI)
   - Receives system prompt: "You are a university scheduling assistant"
   - Receives tool schema:
     [
       {
         name: "create_subject",
         description: "Create a new subject",
         parameters: {name, credits, department_id, ...}
       },
       ...
     ]
   - Analyzes message, calls tool: create_subject(name="Math 101", credits=3, ...)

4. Tool Executor (Self-Referential HTTP)
   - Receives LLM tool call
   - Validates parameters
   - Dispatches via HTTP to core's own API (selfURL + internal JWT token)
     Example: POST /api/timetable/semesters/{id}/generate (with internal JWT header)
   - Tool endpoint routes to appropriate module gRPC (Module-Subject, Module-HR, Module-Timetable)
   - Returns result to LLM
   - Note: selfURL is core's HTTP base URL (e.g., "http://localhost:8000")
   - Note: internalJWT is a service-level JWT with 24h TTL generated at startup

5. LLM Response
   - Streams confirmation: "I've created subject Math 101 with 3 credits"
   - WebSocket streaming to client

6. Client
   - Displays chat message
   - Invalidates /subjects query
   - Refreshes subject list
```

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
  term: enum(fall, spring, summer),
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
  week_of_semester: int
)

event_store (same pattern)
```

## Interaction Patterns

### Request-Response (Synchronous)
- **HTTP ↔ Core**: Standard REST request/response
- **Core ↔ Module gRPC**: Blocking call with timeout (5s default)
- **Module ↔ Module gRPC**: Blocking call with timeout

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
