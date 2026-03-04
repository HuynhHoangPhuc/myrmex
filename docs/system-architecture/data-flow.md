# Data Flow & Interaction Patterns

## Request Flow (HTTP → gRPC)

### Example: Create Teacher

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

## Event Flow (Async)

### Example: Schedule Generation Completion

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

## Student Registration Flow (Invite Code)

### Example: Student registration with invite code

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

## Chat Agent Flow (WebSocket + Tool Execution)

### Example: "Create a subject called Math 101"

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
```

### Example: "Generate a schedule for the current semester" (Multi-step workflow)

```
1. LLM receives message: "Generate a schedule for the current semester"
2. Step 1: LLM calls timetable.list_semesters → Returns semester UUIDs
3. Step 2: LLM calls timetable.generate with returned semester UUID → Triggers CSP solver
4. Returns: "Schedule generation started for Semester X. Please check back in 30 seconds."
```

Note: maxToolIterations=10 enables complex workflows requiring multiple tool calls.

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
