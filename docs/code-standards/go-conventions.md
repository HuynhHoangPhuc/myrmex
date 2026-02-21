# Go Conventions

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| **Packages** | lowercase, no underscores | `persistence`, `config`, `handler` |
| **Files** | snake_case, descriptive | `teacher_repository_impl.go`, `create_teacher_handler.go` |
| **Types** | PascalCase | `Teacher`, `CreateTeacherCommand`, `DepartmentRepository` |
| **Interfaces** | PascalCase, often `{Name}er` or `{Name}Service` | `TeacherRepository`, `EventStore`, `AuthService` |
| **Functions/Methods** | PascalCase (exported), camelCase (unexported) | `CreateTeacher()`, `validateEmail()` |
| **Variables** | camelCase | `teacherID`, `departmentName`, `isActive` |
| **Constants** | UPPER_SNAKE_CASE | `MAX_RETRIES`, `DEFAULT_TIMEOUT` |
| **Private fields** | camelCase, lowercase first | `teacherID`, `mu` (mutex) |

## Directory Structure (Per Service)

```
services/{service}/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point, config + service initialization
├── internal/
│   ├── domain/
│   │   ├── entity/                 # Aggregates + entities (business logic)
│   │   │   └── {aggregate}.go      # E.g., teacher.go (aggregate root)
│   │   ├── valueobject/            # Immutable value objects
│   │   │   └── {vo}.go             # E.g., availability.go
│   │   ├── repository/             # Repository interfaces (DIP)
│   │   │   └── {aggregate}_repository.go
│   │   └── service/                # Domain services (business logic not on entity)
│   │       └── {domain}_service.go # E.g., dag_service.go
│   ├── application/
│   │   ├── command/                # CQRS write handlers
│   │   │   └── {operation}_handler.go # E.g., create_teacher_handler.go
│   │   └── query/                  # CQRS read handlers
│   │       └── {operation}_handler.go # E.g., list_teachers_handler.go
│   ├── infrastructure/
│   │   ├── persistence/            # Data access implementation
│   │   │   ├── sqlc/               # Generated sqlc code (queries.sql.go, models.go)
│   │   │   ├── {aggregate}_repository_impl.go
│   │   │   └── event_store_impl.go
│   │   ├── messaging/              # NATS publishers
│   │   │   └── event_publisher.go
│   │   ├── auth/                   # Auth (core only: JWT, bcrypt)
│   │   │   ├── jwt_manager.go
│   │   │   └── password_hasher.go
│   │   ├── llm/                    # LLM providers (core only)
│   │   │   ├── claude_provider.go
│   │   │   └── openai_provider.go
│   │   └── agent/                  # Tool registry (core only)
│   │       ├── tool_registry.go
│   │       └── tool_executor.go
│   ├── interface/
│   │   ├── grpc/                   # gRPC server implementations
│   │   │   └── {service}_server.go # E.g., teacher_server.go
│   │   ├── http/                   # HTTP handlers (core only)
│   │   │   ├── router.go
│   │   │   ├── middleware.go       # CORS, auth, rate limit
│   │   │   └── handlers/
│   │   │       └── {resource}_handler.go
│   │   └── middleware/             # gRPC/HTTP middleware
│   │       └── auth_interceptor.go
│   ├── config/
│   │   └── {service}.yaml          # Config file (default values)
│   ├── migrations/                 # Goose SQL migrations (001_initial.sql)
│   ├── sql/
│   │   └── queries/                # sqlc query definitions ({resource}.sql)
│   └── errorhandling/              # Custom error types
│       └── errors.go
├── go.mod
├── go.sum
└── Dockerfile
```

## Key Principles

### Domain Layer Excellence
- **Zero Infrastructure Dependencies**: No `sql`, `gin`, `grpc` imports
- **Pure Business Logic**: All domain rules encoded in entities + value objects
- **Repository Interfaces**: Define in domain, implement in infrastructure (DIP)
- **Clear Boundaries**: Domain exports only what external layers need

### File Naming Strategy
- **Clarity**: Filename should describe purpose without reading content
- **Type-Specific Suffixes**: `_repository.go`, `_handler.go`, `_impl.go`
- **Consistency**: All services follow same naming pattern

Example:
- `teacher_repository.go` - Interface definition
- `teacher_repository_impl.go` - PostgreSQL implementation
- `create_teacher_handler.go` - CQRS command handler
- `list_teachers_handler.go` - CQRS query handler

### Type Organization
```go
// Good: Clear that this is a repository interface
type TeacherRepository interface {
    Save(ctx context.Context, teacher *Teacher) error
    GetByID(ctx context.Context, id string) (*Teacher, error)
}

// Good: Implementation is explicit
type TeacherRepositoryImpl struct {
    db *sql.DB
    queries *sqlc.Queries
}

// Avoid: Unclear what "TeacherStore" does
type TeacherStore interface { ... }
```

## Constants & Magic Numbers

```go
// Define at package level for shared values
const (
    MaxTeachersPerDepartment = 100
    DefaultPageSize = 20
    MaxPageSize = 100
)

// Use enums for domain concepts
type DayOfWeek int

const (
    Monday DayOfWeek = iota
    Tuesday
    Wednesday
    Thursday
    Friday
    Saturday
    Sunday
)

// Avoid inline magic numbers
// Bad: period > 5 (what does 5 mean?)
// Good: period > MaxPeriodsPerDay
```

## Error Handling Conventions

Define custom errors in `internal/errorhandling/errors.go`:

```go
var (
    ErrTeacherNotFound = errors.New("teacher not found")
    ErrDuplicateEmail = errors.New("email already in use")
    ErrInvalidAvailability = errors.New("availability period cannot be empty")
)

// Custom error type with context
type ValidationError struct {
    Field string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}
```

Use `errors.Is()` for matching:

```go
if err := createTeacher(...); err != nil {
    if errors.Is(err, ErrDuplicateEmail) {
        // Handle duplicate email
    }
}
```

## Logging Conventions

Use Zap logger with structured fields:

```go
logger.Info("Server started",
    zap.String("service", "module-hr"),
    zap.Int("port", 50052),
)

logger.WithError(err).Error("Database connection failed",
    zap.String("host", dbHost),
    zap.Int("port", 5432),
)
```

**Never log**: Passwords, API keys, credit card numbers, PII

## Configuration Conventions

Use Viper for YAML + environment overlay:

Config file (`services/{service}/config/default.yaml`):
```yaml
server:
  grpc_port: 50052

database:
  url: postgres://myrmex:myrmex_dev@localhost:5432/hr?sslmode=disable
  max_connections: 20

nats:
  url: nats://localhost:4222
```

Environment overrides (use env var prefixes):
```bash
export HR_SERVER_GRPC_PORT=50053
export HR_DATABASE_URL="postgres://..."
```

Code initialization:
```go
config := viper.New()
config.SetConfigName("default")
config.SetConfigType("yaml")
config.AddConfigPath("./config")
config.BindEnv("server.grpc_port", "HR_SERVER_GRPC_PORT")

var cfg struct {
    Server struct {
        GrpcPort int
    }
}
config.Unmarshal(&cfg)
```

## Code Organization Guidelines

### Package Organization
```go
// types.go - Define domain types
type Teacher struct { ... }
type Department struct { ... }

// repository.go - Define repository interface
type TeacherRepository interface { ... }

// service.go - Domain business logic (not on aggregate)
type TeacherService struct { ... }
```

### Import Organization
```go
import (
    // Std library
    "context"
    "fmt"

    // External dependencies
    "github.com/google/uuid"

    // Internal packages
    "myrmex/pkg/logger"
    "myrmex/services/module-hr/internal/domain"
)
```

### Visibility Rules
- **Exported (PascalCase)**: Part of package's public API
- **Unexported (camelCase)**: Internal implementation details
- **Avoid**: Exporting package variables; use functions + getters

```go
// Good: Interface with getter methods
type Teacher struct {
    id string // unexported
    name string
}

func (t *Teacher) ID() string {
    return t.id
}

// Avoid: Public fields that shouldn't be modified
type Teacher struct {
    ID string // Easy to accidentally mutate
    Name string
}
```
