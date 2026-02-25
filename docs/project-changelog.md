# Myrmex Project Changelog

All notable changes to the Myrmex project are documented here.

## [Unreleased]

### Added
- **Backend: Schedule enrichment** — Added denormalized fields to `ScheduleEntry` (subject_name, subject_code, teacher_name, room_name) via migration `007_enrich_schedule_entries.sql` for efficient API responses
- **Backend: ListSchedules RPC** — Implemented complete `ListSchedules` gRPC service with optional semester filtering and pagination support
- **Backend: HTTP handlers** — Fixed `ListSchedules` and `GetSchedule` HTTP endpoints in core gateway to properly return enriched schedule data
- **Frontend: Schedule Calendar** — Built interactive weekly timetable grid view with day columns (Mon–Sat) and period rows, color-coded by department
- **Frontend: Period utilities** — Created `period-to-time.ts` mapping periods to human-readable time strings (08:00–21:45)
- **Frontend: Manual override modal** — Integrated teacher suggestion system with assign modal for schedule adjustments
- **Database: Demo seed data** — Created `deploy/docker/seed.sql` with deterministic demo data (3 departments, 8 teachers, 10 subjects, 5 rooms, 30 time slots)
- **Build: Seed target** — Added `make seed` and `make reset-db` Makefile targets for database population

### Fixed
- **Proto: ScheduleEntry** — Extended message with enriched display fields (subject_name, subject_code, teacher_name, room_name)
- **Frontend: ScheduleEntry type** — Replaced aspirational start_time/end_time strings with actual start_period/end_period integers from API
- **SQL queries** — Updated `ListEntriesBySchedule` with JOIN to fetch time_slot and room details in single query
- **SubjectInfo model** — Added Name field to enable subject name lookups in timetable service

### Technical Details
- All 3 services compile cleanly (module-timetable, core, module-hr)
- TypeScript type check passes across frontend
- Enriched API responses support full schedule visualization without additional API calls per entry
- Seed data covers Mon–Fri, periods 1–6; easily extensible for additional time slots

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
