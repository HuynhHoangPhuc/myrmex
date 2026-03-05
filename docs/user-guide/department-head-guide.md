# Department Head Guide

## Dashboard Overview

The Dashboard (`/dashboard`) shows a summary card set for your department:

| Card | Content |
|------|---------|
| Teachers | Total teachers in your department |
| Subjects | Active subjects this semester |
| Schedule Coverage | % of offered subjects with assigned teachers |
| Workload | Average teaching hours per teacher this semester |

---

## Managing Department Teachers

Navigate to **HR → Teachers** (`/hr/teachers`).

### View Teachers
- The list is pre-filtered to your department.
- Columns: name, email, employee ID, teaching hours this semester.

### Add an Existing User to Your Department
1. Click **Assign Teacher**.
2. Search by name or email (user must already have a `teacher` role).
3. Select the user → click **Assign**.

### Remove a Teacher from Department
1. Click the teacher row → **Edit**.
2. Clear the department field → **Save**.

> Removing a teacher from the department does not delete their account or cancel active schedule sessions.

---

## Subject Management

Navigate to **Subjects** (`/subjects`).

### Create a Subject
1. Click **New Subject**.
2. Fill in: code (e.g. `CS101`), name, credits, description, department.
3. Click **Save**.

### Edit a Subject
1. Click the subject row → **Edit**.
2. Modify fields → **Save**.

### Manage Prerequisites (DAG)
1. Navigate to **Subjects → Prerequisites** (`/subjects/prerequisites`).
2. Select a subject.
3. Click **Add Prerequisite** → search and select the prerequisite subject.
4. The system validates: adding a prerequisite that creates a cycle is rejected automatically.
5. Prerequisites are displayed as a directed acyclic graph (DAG) — zoom and pan to explore.

> Cycle detection runs automatically. If a prerequisite is rejected, the system will indicate the conflicting path.

---

## Semester Setup

Navigate to **Timetable → Semesters** (`/timetable/semesters`).

### Create a Semester
1. Click **New Semester**.
2. Enter: name (e.g. `2024-2025 Fall`), start date, end date, registration deadline.
3. Click **Save**.

### Configure Rooms
1. Open the semester → **Rooms** tab.
2. Click **Add Room** → enter room code, building, capacity.
3. Save each room. Rooms are shared across all departments.

### Configure Time Slots
1. Open the semester → **Time Slots** tab.
2. Click **Add Time Slot** → select day of week, start time, end time.
3. Repeat for all available slots. Standard setup: Mon–Sat, 4–6 slots per day.

### Add Offered Subjects
1. Open the semester → **Offerings** tab.
2. Click **Add Offering** → select subject, set max enrollment cap.
3. Save. Offerings define which subjects are available for enrollment this semester.

---

## Schedule Generation

Navigate to **Timetable → Generate** (`/timetable/generate`).

The scheduler uses a CSP solver (backtracking + AC-3) to assign rooms and time slots to all offered subjects while respecting teacher availability and room capacity.

### Run Generation
1. Select the target semester.
2. Review the pre-flight checklist (all items must be green):
   - At least one room configured
   - At least one time slot configured
   - All offerings have a teacher assigned
   - Teacher availability preferences submitted
3. Click **Generate Schedule**.
4. The solver runs in the background. A progress indicator shows status.
5. When complete, you are redirected to the schedule view.

> If the solver times out (complex constraints), it returns the best partial schedule found. Unresolved sessions are marked **Unscheduled** — handle them manually.

---

## Manual Schedule Adjustments

Navigate to **Timetable → Schedules** (`/timetable/schedules`).

1. Filter by semester.
2. Click a session row → **Edit**.
3. Change: room, time slot, or assigned teacher.
4. The system checks for conflicts (room double-booking, teacher overlap) in real time.
5. If no conflicts, click **Save**.

Conflict types:
- **Room conflict** — same room, same time slot, different session.
- **Teacher conflict** — same teacher, same time slot, different session.
- **Capacity conflict** — enrolled students exceed room capacity.

---

## Analytics

Navigate to **Analytics** (`/analytics`).

| Report | Description |
|--------|-------------|
| Workload Distribution | Teaching hours per teacher; highlights over/under-loaded teachers |
| Room Utilization | % of time slots used per room; identifies bottlenecks |
| Enrollment Trends | Subject enrollment counts over semesters |
| Prerequisite Completion | % of students completing prerequisites on time |

- Use the **Semester** filter to compare across terms.
- Click **Export CSV** on any report to download raw data.
