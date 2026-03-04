# Service Topology & APIs

Detailed service descriptions for Myrmex microservices.

## Core Service (HTTP Gateway + Auth + AI Chat)

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

## Module-HR (Department & Teacher Management)

**Purpose**: Faculty data management and availability scheduling.

**Port**: gRPC `:50052`

**Key Entities**:
- **Department**: id, name, created_at, deleted_at
- **Teacher**: id, name, email, department_id, specializations[], availability[]
- **Availability**: {day_of_week, start_time, end_time} — Time slots in HH:MM format

**Key Operations**:
- CRUD: Department, Teacher (with availability + specializations)
- GetTeacher: Returns availability as time strings
- UpdateTeacherAvailability: Accepts {available_slots: [{day_of_week, start_time, end_time}]}
- Soft delete for both departments and teachers

**Event Types**: `teacher.created/updated/deleted`, `department.created/updated/deleted`, `teacher.availability_updated`

## Module-Subject (Subject & Prerequisite DAG)

**Purpose**: Course structure with prerequisite management and cycle detection.

**Port**: gRPC `:50053`

**Key Entities**:
- **Subject**: id, code, name, credits, weekly_hours, department_id, is_active
- **Prerequisite**: subject_id, prerequisite_subject_id, type (strict/recommended/corequisite), priority
- **DAG**: In-memory for cycle detection + topological sort

**Key Operations**:
- CRUD: Subject
- Add/Remove prerequisites (cycle detection via DFS 3-color)
- **GetFullDAG**: All subjects + edges for visualization
- **CheckPrerequisiteConflicts**: Detect missing prerequisites in subject set
- Topological sort for enrollment planning

**Domain Services**: DAGService (cycle detection, topological sort via BFS)

## Module-Student (Student Management + Enrollment + Grades)

**Purpose**: Complete student lifecycle: invite code generation, registration, enrollment request/approval, grade assignment, transcript generation.

**Port**: gRPC `:50055`

**Key Entities**:
- **Student**: id, student_code, user_id, full_name, email, department_id, enrollment_year, status, is_active
- **InviteCode**: code (SHA-256 hash), created_by_user_id, used_at (nullable), expires_at
- **Enrollment**: id, student_id, offered_subject_id, semester_id, status
- **Grade**: id, enrollment_id, grade_numeric (0-10), grade_letter (auto-derived A-F)

**Key Operations**:
- **Student CRUD**: Create, get, list (paginated), update, soft-delete
- **CreateInviteCode**: Admin generates code; hashed in DB
- **RequestEnrollment**: Student requests subject enrollment
- **ApproveEnrollment**: Admin approves with prerequisite validation
- **AssignGrade**: Teacher/admin assigns numeric grade → letter auto-derived
- **GetTranscript**: Student's full academic history + GPA calculation

**Domain Services**:
- **PrerequisiteValidator**: Checks student has completed prerequisites (Redis-cached)
- **GradeComputer**: Auto-derives letter grade
- **TranscriptBuilder**: Aggregates approved enrollments + grades

## Module-Analytics (Analytics & Reporting)

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

**Event Types**: Consumes `teacher.created/updated`, `department.created`, `subject.created`, `schedule.generation_completed`

## Module-Notification (Email + In-App Notifications)

**Purpose**: Email + in-app notification delivery with user preference management.

**Port**: HTTP `:8056` (reverse-proxied via Core at `/api/notifications/`)

**Key Entities**:
- **Notification**: id, user_id, event_type, payload, is_read, created_at, read_at
- **Preference**: id, user_id, event_type (12 types), channel (email|in_app), enabled
- **EmailQueue**: id, to_email, subject, body, status, retry_count, next_retry_at

**Key Operations**:
- **ListNotifications**: Paginated list (filters: read status, event type)
- **MarkNotificationRead**: Mark single notification as read
- **GetPreferences**: Fetch 12×2 preference matrix for current user
- **UpdatePreferences**: Bulk update (PATCH) user preferences
- **DispatchNotification**: Route events to user's preferred channels
- **BroadcastAnnouncement**: Admin-only; sends announcement to all users

**Domain Services**:
- **NotificationDispatcher**: Routes events to email + in-app channels
- **EmailRenderer**: MJML template engine for subject-specific emails
- **RecipientResolver**: Cross-schema lookup for accurate recipient matching
- **RetryScheduler**: Exponential backoff for failed emails

**Event Types**: Consumes all domain events (schedule.*, enrollment.*, grade.*, new_announcement, role_updated, user.deleted)

**Configuration**:
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD` (SMTP server credentials)
- `NOTIFICATION_FROM_EMAIL` (sender address for emails)

## Module-Timetable (Schedule Generation & Management)

**Purpose**: Semester and schedule management with CSP-based generation.

**Port**: gRPC `:50054`

**Key Entities**:
- **Semester**: id, name, year, term, start_date, end_date, offered_subject_ids[], room_ids[]
- **Room**: id, code, capacity, type (classroom|lab|lecture_hall), features[]
- **Schedule**: id, semester_id, status (pending/generating/completed/failed), entries[]
- **ScheduleEntry**: schedule_id, subject_id, teacher_id, room_id, day, period, week

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
