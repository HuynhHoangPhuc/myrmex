# Database Schema (Logical)

PostgreSQL database with schema-per-module isolation + shared audit logs and notifications.

## Core Schema

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

audit_logs (
  id: bigint primary key auto,
  user_id: uuid fk users,
  resource_type: string,  -- "teacher", "subject", "semester", "enrollment", "grade", etc.
  action: enum(create, read, update, delete),
  old_value: jsonb nullable,  -- Previous state (null for creates)
  new_value: jsonb,  -- Current state (null for deletes)
  timestamp: timestamp,
  -- 12 monthly partitions: 2026-03 through 2027-02
  -- BRIN index on timestamp, B-tree on (user_id, resource_type) for filtering
)
```

## HR Schema

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

## Subject Schema

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

## Student Schema

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
```

## Timetable Schema

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

## Notification Schema

```sql
notifications (
  id: uuid primary key,
  user_id: uuid fk core.users,
  event_type: string,  -- new_announcement, schedule.generated, enrollment.approved, grade.assigned, etc.
  payload: jsonb,  -- Event-specific data (e.g., {schedule_id, semester_name} for schedule.generated)
  is_read: bool default false,
  created_at: timestamp,
  read_at: timestamp nullable,
  index: (user_id, is_read) for inbox queries
)

preferences (
  id: uuid primary key,
  user_id: uuid fk core.users,
  event_type: string,  -- One of 12 event types
  channel: enum(email, in_app),  -- Delivery channel
  enabled: bool default true,  -- User can opt-out per channel+event combination
  created_at: timestamp,
  updated_at: timestamp,
  unique index: (user_id, event_type, channel)
)

email_queue (
  id: uuid primary key,
  to_email: string,  -- Recipient email address
  subject: string,  -- Email subject
  body: string,  -- Email body (HTML from MJML template)
  status: enum(pending, sent, failed),  -- Current status
  retry_count: int default 0,  -- Number of failed attempts
  next_retry_at: timestamp nullable,  -- When to retry (exponential backoff)
  created_at: timestamp,
  failed_at: timestamp nullable,  -- When permanently failed (after 5 retries)
  index: (status, next_retry_at) for queue processing
)
```

## Analytics Schema (Star Schema)

```sql
dim_teacher (
  teacher_id: uuid primary key,
  name: string,
  email: string,
  department_id: uuid,
  department_name: string,
  is_active: bool,
  created_at: timestamp,
  updated_at: timestamp
)

dim_subject (
  subject_id: uuid primary key,
  code: string,
  name: string,
  credits: int,
  weekly_hours: int,
  department_id: uuid,
  is_active: bool,
  created_at: timestamp,
  updated_at: timestamp
)

dim_department (
  department_id: uuid primary key,
  name: string,
  created_at: timestamp,
  updated_at: timestamp
)

dim_semester (
  semester_id: uuid primary key,
  name: string,
  year: int,
  term: int,
  start_date: date,
  end_date: date,
  created_at: timestamp,
  updated_at: timestamp
)

fact_schedule_entry (
  id: uuid primary key,
  schedule_id: uuid,
  semester_id: uuid fk dim_semester,
  subject_id: uuid fk dim_subject,
  teacher_id: uuid fk dim_teacher,
  department_id: uuid fk dim_department,
  day_of_week: string,
  period_of_day: int,
  hours: decimal,  -- weekly_hours denormalized for workload calculations
  utilization_pct: decimal,  -- room utilization percentage
  created_at: timestamp,
  updated_at: timestamp
)
```

## Key Patterns

- **Soft Deletes**: `deleted_at` timestamp for logical deletion (queries filter WHERE deleted_at IS NULL)
- **Event Sourcing**: Each module maintains an `event_store` table; events immutable, append-only
- **Denormalization**: `schedule_entries` includes subject_name, teacher_name, room_name for efficient API responses
- **Partitioning**: `core.audit_logs` partitioned monthly for fast date-range queries
- **Indexes**: Strategic indexes on frequently filtered columns (user_id, resource_type, status, is_read, etc.)
- **Optimistic Concurrency**: `event_store.version` for conflict detection during concurrent updates
