# Code Standards & Conventions

Myrmex follows Domain-Driven Design (DDD) + Clean Architecture principles for backend and React/TypeScript best practices for frontend.

## Backend Standards

Go services follow strict architectural layers and patterns:

- **[Go Conventions](./go-conventions.md)** - Naming, directory structure, file organization
- **[Go Architecture](./go-architecture.md)** - DDD, Clean Architecture, CQRS, error handling
- **[Go Patterns](./go-patterns.md)** - Testing, logging, config management, concurrency

## Frontend Standards

React/TypeScript follows modern patterns and co-location principles:

- **[React Conventions](./react-conventions.md)** - Naming, directory structure, component patterns
- **[React Patterns](./react-patterns.md)** - API hooks, forms, state management, WebSocket integration
- **[React Advanced](./react-advanced.md)** - Performance, testing, Tailwind styling

## Common Patterns

- **[Pagination](./common-patterns.md#pagination)** - Backend + frontend pagination
- **[Optimistic Updates](./common-patterns.md#optimistic-updates)** - Smooth UX with rollback
- **[Query Key Strategy](./common-patterns.md#query-key-strategy)** - Cache invalidation

## Linting & Formatting

- **Backend**: `go vet ./...`, `go tool buf lint`
- **Frontend**: ESLint (via Vite), TypeScript strict mode
- **Commits**: Conventional commit format, no credentials

---

**Quick Reference**:
- Go file naming: `snake_case` (e.g., `teacher_repository_impl.go`)
- React file naming: `kebab-case` (e.g., `create-teacher-form.tsx`)
- Go types: `PascalCase` (e.g., `Teacher`, `CreateTeacherCommand`)
- React hooks: `use` prefix + camelCase (e.g., `useTeachers`)
