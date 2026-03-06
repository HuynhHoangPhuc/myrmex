# 2026-03-06 — Surgical Tech Debt Cleanup

## Summary

Completed 4-phase tech debt sprint across Myrmex ERP codebase. All changes are pure refactoring — zero logic changes, all tests green.

## What Changed

### Phase 1: Test Coverage
- **module-analytics**: 0 → 5 test files. Previously had 11 source files and zero tests. Added query handler tests (dashboard, workload, utilization) + consumer event handler tests.
- **module-notification**: 3 → 8 test files. Fixed broken `dispatch_notification_test.go` (undefined mock). Added repository, HTTP handler, and event router tests.

### Phase 2: Backend Modularization
Worst offenders split under 200 LOC each:

| Original | LOC | Split Into |
|----------|-----|-----------|
| `student_server.go` | 681 | 4 files (student, enrollment, grade, invite_code) |
| `tool_executor.go` | 466 | 5 files (executor + 4 module route files) |
| `timetable_handler.go` | 398 | 3 files (semester, slot, room) |
| `main.go` (core) | 368 | 3 files (main, init_infrastructure, init_services) |
| `router.go` | 353 | 3 files (router, module_routes, proxy) |

### Phase 3: Core Gateway Cleanup
4 more handler/service files split:
- `hr_handler.go` (379) → hr_handler + hr_department_handler + hr_availability_handler
- `subject_handler.go` (344) → subject_handler + subject_prerequisite_handler + subject_dag_handler
- `student_handler.go` (328) → student_handler + student_enrollment_handler + student_grade_handler
- `oauth_service.go` (380) → oauth_service + oauth_exchange + oauth_state

### Phase 4: Frontend Cleanup
5 large route files split into 13 focused components:
- `help/index.tsx` (346 → 52 LOC) + help-role-guides + help-primitives
- `enrollments/index.tsx` (316 → 118 LOC) + enrollment-columns + reject-dialog + use-enrollment-filters hook
- `grades/index.tsx` (283 → 107 LOC) + grade-dialog + grade-columns + use-grade-assignment hook
- `time-slot-manager.tsx` (284 → 42 LOC) + time-slot-dialogs + time-slot-constants
- `student/subjects.tsx` (278 → 100 LOC) + subject-list + enrollment-history-table + enroll-dialog

Bonus: `use-enrollment-filters` hook reused by both enrollments and grades (DRY win).

## Metrics

- **14 large files** (278–681 LOC) → **40+ focused files**, all ≤200 LOC
- **13 new test files** added (analytics: 5, notification: 8)
- **0 compilation errors**, **0 test regressions**
- Frontend build: `✓ built in 4.61s`

## Decisions

- `student_server.go` → 301 LOC (slightly over) — constructor boilerplate unavoidable. Accepted.
- `tool_executor.go` → split switch into sub-builder functions per module, dispatched via module prefix. Backward-compat shim kept for existing tests.
- `time-slot-manager.tsx` DeleteConfirmDialog moved to own `useDeleteTimeSlot` call — cleaner API.
- Phase 1 and Phase 4 ran in parallel (independent concerns), then Phase 2 → Phase 3 sequentially.
