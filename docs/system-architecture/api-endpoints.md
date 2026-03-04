# API Endpoints (Complete Reference)

## Authentication & User Management

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

## HR Module

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
| GET | `/api/hr/teachers` | Module-HR | Paginated list: `{ data, total, page, page_size }` |
| POST | `/api/hr/teachers` | Module-HR | Create teacher |
| GET | `/api/hr/teachers/:id` | Module-HR | Single teacher |
| PATCH | `/api/hr/teachers/:id` | Module-HR | Update teacher |
| DELETE | `/api/hr/teachers/:id` | Module-HR | Soft delete |
| GET | `/api/hr/teachers/:id/availability` | Module-HR | Availability schedule: `{ availability: [{day_of_week, start_time, end_time}] }` (time strings) |
| PUT | `/api/hr/teachers/:id/availability` | Module-HR | Update availability (body: `{ available_slots: [{day_of_week, start_time, end_time}] }`) |
| GET | `/api/hr/departments` | Module-HR | Paginated list |
| POST | `/api/hr/departments` | Module-HR | Create department |

## Subject Module

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
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

## Timetable Module

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
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

## Student Module

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
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

## Analytics Module

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
| GET | `/api/analytics/dashboard` | Module-Analytics | KPI cards: teacher count, avg workload, schedule completion % |
| GET | `/api/analytics/workload` | Module-Analytics | Workload analytics per teacher with period breakdown |
| GET | `/api/analytics/utilization` | Module-Analytics | Resource utilization metrics (rooms, teachers, semesters) |
| GET | `/api/analytics/department-metrics` | Module-Analytics | Department-level metrics (teachers per dept, specialization coverage) |
| GET | `/api/analytics/schedule-metrics` | Module-Analytics | Schedule metrics (completion rate, conflicts, constraints) |
| GET | `/api/analytics/schedule-heatmap` | Module-Analytics | Schedule density heatmap (day/period utilization) |
| GET | `/api/analytics/export` | Core proxy route reserved for future analytics export surface |

## Notifications Module

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
| GET | `/api/notifications` | Module-Notification | Paginated list of user's notifications; filters by read status, event type |
| POST | `/api/notifications/:id/mark-read` | Module-Notification | Mark single notification as read |
| GET | `/api/notifications/preferences` | Module-Notification | Fetch current user's 12×2 preference matrix (event_type × channel) |
| PATCH | `/api/notifications/preferences` | Module-Notification | Update user notification preferences; bulk matrix update |
| POST | `/api/notifications/announcement` | Module-Notification | Admin-only; broadcast announcement to all users (body: title, message) |

## Audit & Compliance

| Method | Endpoint | Service | Notes |
|--------|----------|---------|-------|
| GET | `/api/audit-logs` | Core | Admin/super_admin only: paginated audit logs with filters (user_id, resource_type, action, date range) |

## Chat

| Protocol | Endpoint | Service | Notes |
|----------|----------|---------|-------|
| WebSocket | `/ws/chat?token=ACCESS_TOKEN` | Core | Streaming chat interface; tool execution, LLM responses, markdown rendering |
