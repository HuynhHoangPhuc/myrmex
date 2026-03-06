# Phase Implementation Report

## Executed Phase
- Phase: Phase 2 — Split top 5 largest backend files
- Plan: /Users/phuc/Developer/myrmex/260306-2218-surgical-tech-debt
- Status: completed

## Files Modified / Created

### 2A — student_server.go (681 LOC → 4 files)
| File | LOC |
|------|-----|
| `services/module-student/internal/interface/grpc/student_server.go` (modified) | 301 |
| `services/module-student/internal/interface/grpc/enrollment_server.go` (created) | 233 |
| `services/module-student/internal/interface/grpc/grade_server.go` (created) | 120 |
| `services/module-student/internal/interface/grpc/invite_code_server.go` (created) | 76 |

### 2B — tool_executor.go (466 LOC → 5 files)
| File | LOC |
|------|-----|
| `services/core/internal/infrastructure/agent/tool_executor.go` (modified) | 256 |
| `services/core/internal/infrastructure/agent/tool_routes_hr.go` (created) | 51 |
| `services/core/internal/infrastructure/agent/tool_routes_subject.go` (created) | 62 |
| `services/core/internal/infrastructure/agent/tool_routes_timetable.go` (created) | 81 |
| `services/core/internal/infrastructure/agent/tool_routes_student_analytics.go` (created) | 99 |

### 2C — timetable_handler.go (398 LOC → 3 files)
| File | LOC |
|------|-----|
| `services/core/internal/interface/http/timetable_handler.go` (modified) | 241 |
| `services/core/internal/interface/http/timetable_slot_handler.go` (created) | 123 |
| `services/core/internal/interface/http/timetable_room_handler.go` (created) | 52 |

### 2D — main.go (368 LOC → 3 files)
| File | LOC |
|------|-----|
| `services/core/cmd/server/main.go` (modified) | 247 |
| `services/core/cmd/server/init_infrastructure.go` (created) | 130 |
| `services/core/cmd/server/init_services.go` (created) | 102 |

### 2E — router.go (353 LOC → 3 files)
| File | LOC |
|------|-----|
| `services/core/internal/interface/http/router.go` (modified) | 141 |
| `services/core/internal/interface/http/router_module_routes.go` (created) | 142 |
| `services/core/internal/interface/http/router_proxy.go` (created) | 86 |

## Tasks Completed
- [x] 2A: student_server.go split into 4 files (struct/CRUD, enrollment, grade, invite_code)
- [x] 2B: tool_executor.go split into 5 files (executor core + 4 route sub-builders)
- [x] 2C: timetable_handler.go split into 3 files (semester, slot, room)
- [x] 2D: main.go split into 3 files (main, init_infrastructure, init_services)
- [x] 2E: router.go split into 3 files (router, module_routes, proxy)
- [x] Build verified after each split: `go build ./...` passes for both services
- [x] Tests pass: `go test ./...` for module-student (all ok) and core (all ok)

## Tests Status
- Build (module-student): pass
- Build (core): pass
- Unit tests (module-student): all ok — command, query, domain/entity, persistence, grpc packages
- Unit tests (core): all ok — agent, auth, llm, notification, persistence, command, query packages

## Implementation Notes
- `student_server.go` is 301 LOC (slightly over 200) because the struct + constructor alone takes ~85 lines; no logic is extractable without artificial splitting
- `timetable_handler.go` is 241 LOC because struct, constructor, `buildSubjectMap`, `semesterToJSON`, and 4 handlers are tightly coupled
- `tool_executor.go` at 256 LOC includes the backward-compat package-level `buildEndpoint` shim required by existing tests (original signature: `buildEndpoint(baseURL, module, method, args) (url, method, body)`)
- `buildEndpoint` was refactored from a monolithic switch into a method on `ToolExecutor` dispatching to per-module sub-builders; the test shim preserves old call sites without test changes
- NATS nil-safety handled via `toNATSPublisher` and `toNATSJS` helpers in `init_infrastructure.go`

## Issues Encountered
- Existing agent test called package-level `buildEndpoint(baseURL, module, method, args)` — resolved by adding a thin wrapper that delegates to the new method-based implementation
- `init_services.go` initially had a `parseDurationConfig` stub — removed; using `time.ParseDuration(v.GetString(...))` directly
- NATS types: `*nats.Conn` needed explicit import; nil-safe accessors added to avoid interface{} casting in main

## Next Steps
- Phase 3: Split core gateway handlers and OAuth service (328–380 LOC)
