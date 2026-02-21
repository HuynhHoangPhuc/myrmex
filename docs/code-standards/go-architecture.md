# Go Architecture Patterns

## Domain-Driven Design (DDD)

Myrmex strictly enforces DDD with 4 clean architectural layers.

### Domain Layer (Pure Business Logic)

**Location**: `internal/domain/{entity,valueobject,service,repository}`

**Characteristics**:
- Zero dependencies on infrastructure (no `sql`, `gin`, `grpc` imports)
- Contains aggregates, entities, value objects, repository interfaces
- Repository interfaces defined here; implementations in infrastructure layer
- All business rules codified as methods on domain objects

**Example: Teacher Aggregate**

```go
// internal/domain/entity/teacher.go
package entity

type Teacher struct {
    ID              string
    Name            string
    Email           string
    DepartmentID    string
    Specializations []string
    Availability    map[DayOfWeek][]PeriodOfDay
}

// Business rule: Cannot have empty availability
func (t *Teacher) UpdateAvailability(day DayOfWeek, periods []PeriodOfDay) error {
    if len(periods) == 0 {
        return ErrInvalidAvailability // Domain error
    }
    t.Availability[day] = periods
    return nil
}

// Value object: Immutable availability period
type Availability struct {
    DayOfWeek DayOfWeek
    Periods   []PeriodOfDay
}

func NewAvailability(day DayOfWeek, periods []PeriodOfDay) (*Availability, error) {
    if len(periods) == 0 {
        return nil, ErrInvalidAvailability
    }
    return &Availability{day, periods}, nil
}
```

**Repository Interface (defined in domain)**:

```go
// internal/domain/repository/teacher_repository.go
package repository

type TeacherRepository interface {
    Save(ctx context.Context, teacher *entity.Teacher) error
    GetByID(ctx context.Context, id string) (*entity.Teacher, error)
    ListByDepartment(ctx context.Context, deptID string) ([]*entity.Teacher, error)
}
```

### Application Layer (CQRS)

**Location**: `internal/application/{command,query}`

**Characteristics**:
- Orchestrates domain + infrastructure
- Separates reads (queries) from writes (commands)
- Commands: Create, Update, Delete → persist + publish events
- Queries: Get, List, Search → return DTOs, no side effects

**Command Handler Example**:

```go
// internal/application/command/create_teacher_handler.go
package command

type CreateTeacherCommand struct {
    Name         string
    Email        string
    DepartmentID string
}

type CreateTeacherHandler struct {
    repo      domain.TeacherRepository
    publisher infrastructure.EventPublisher
}

func (h *CreateTeacherHandler) Handle(ctx context.Context, cmd CreateTeacherCommand) (string, error) {
    // Validate
    if cmd.Email == "" {
        return "", errors.New("email required")
    }

    // Create aggregate
    teacher := domain.NewTeacher(cmd.Name, cmd.Email, cmd.DepartmentID)

    // Persist (within transaction)
    if err := h.repo.Save(ctx, teacher); err != nil {
        return "", fmt.Errorf("save teacher: %w", err)
    }

    // Publish event
    if err := h.publisher.Publish(ctx, "teacher.created", map[string]any{
        "teacher_id": teacher.ID,
        "name": teacher.Name,
        "email": teacher.Email,
    }); err != nil {
        // Log warning; don't fail the operation
        logger.WithError(err).Warn("Failed to publish teacher.created event")
    }

    return teacher.ID, nil
}
```

**Query Handler Example**:

```go
// internal/application/query/list_teachers_handler.go
package query

type ListTeachersQuery struct {
    DepartmentID string
    PageSize     int
    PageNumber   int
}

type TeacherDTO struct {
    ID              string
    Name            string
    Email           string
    Specializations []string
}

type ListTeachersHandler struct {
    repo domain.TeacherRepository
}

func (h *ListTeachersHandler) Handle(ctx context.Context, q ListTeachersQuery) ([]TeacherDTO, error) {
    teachers, err := h.repo.ListByDepartment(ctx, q.DepartmentID)
    if err != nil {
        return nil, fmt.Errorf("list teachers: %w", err)
    }

    // Convert to DTO
    dtos := make([]TeacherDTO, len(teachers))
    for i, t := range teachers {
        dtos[i] = TeacherDTO{
            ID:              t.ID,
            Name:            t.Name,
            Email:           t.Email,
            Specializations: t.Specializations,
        }
    }
    return dtos, nil
}
```

### Infrastructure Layer (Data Access & External Services)

**Location**: `internal/infrastructure/{persistence,messaging,auth,llm}`

**Characteristics**:
- Implements repository interfaces (defined in domain)
- Database access via sqlc (type-safe queries)
- Event publishing via NATS
- External service integration (LLM, auth)
- Depends on domain; never exports to domain

**Repository Implementation**:

```go
// internal/infrastructure/persistence/teacher_repository_impl.go
package persistence

type TeacherRepositoryImpl struct {
    db      *sql.DB
    queries *sqlc.Queries
}

func (r *TeacherRepositoryImpl) Save(ctx context.Context, teacher *domain.Teacher) error {
    // Use sqlc-generated method (type-safe)
    err := r.queries.CreateTeacher(ctx, sqlc.CreateTeacherParams{
        ID:           teacher.ID,
        Name:         teacher.Name,
        Email:        teacher.Email,
        DepartmentID: teacher.DepartmentID,
    })
    if err != nil {
        if isDuplicateKey(err) {
            return domain.ErrDuplicateEmail
        }
        return fmt.Errorf("create teacher: %w", err)
    }
    return nil
}

func (r *TeacherRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Teacher, error) {
    row, err := r.queries.GetTeacher(ctx, id)
    if err == sql.ErrNoRows {
        return nil, domain.ErrTeacherNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get teacher: %w", err)
    }

    // Convert DB row to domain entity
    teacher := &domain.Teacher{
        ID:           row.ID,
        Name:         row.Name,
        Email:        row.Email,
        DepartmentID: row.DepartmentID,
    }
    return teacher, nil
}
```

**Event Publisher**:

```go
// internal/infrastructure/messaging/event_publisher.go
package messaging

type EventPublisher struct {
    conn *nats.Conn
}

func (p *EventPublisher) Publish(ctx context.Context, eventType string, payload any) error {
    data, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal payload: %w", err)
    }

    subject := fmt.Sprintf("domain.%s", eventType) // e.g., "domain.teacher.created"
    return p.conn.PublishMsg(&nats.Msg{
        Subject: subject,
        Data:    data,
        Header:  nats.Header{"Event-Type": []string{eventType}},
    })
}
```

### Interface Layer (gRPC/HTTP Handlers)

**Location**: `internal/interface/{grpc,http}`

**Characteristics**:
- Receives external requests (gRPC proto, HTTP JSON)
- Converts to application commands/queries
- Calls handlers
- Translates domain errors to gRPC/HTTP status codes
- Returns proto/JSON responses

**gRPC Server Implementation**:

```go
// internal/interface/grpc/teacher_server.go
package grpc

type TeacherServer struct {
    pb.UnimplementedTeacherServiceServer
    createTeacherHandler application.CreateTeacherHandler
    listTeachersHandler  application.ListTeachersHandler
}

func (s *TeacherServer) CreateTeacher(ctx context.Context, req *pb.CreateTeacherRequest) (*pb.Teacher, error) {
    // Validate input
    if err := req.Validate(); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    // Call application handler
    cmd := application.CreateTeacherCommand{
        Name:         req.Name,
        Email:        req.Email,
        DepartmentID: req.DepartmentId,
    }
    id, err := s.createTeacherHandler.Handle(ctx, cmd)
    if err != nil {
        // Translate domain error to gRPC status
        return nil, s.translateError(err)
    }

    // Return proto response
    return &pb.Teacher{
        Id:           id,
        Name:         req.Name,
        Email:        req.Email,
        DepartmentId: req.DepartmentId,
    }, nil
}

func (s *TeacherServer) translateError(err error) error {
    switch {
    case errors.Is(err, domain.ErrTeacherNotFound):
        return status.Error(codes.NotFound, "teacher not found")
    case errors.Is(err, domain.ErrDuplicateEmail):
        return status.Error(codes.AlreadyExists, "email already in use")
    case errors.Is(err, domain.ErrInvalidAvailability):
        return status.Error(codes.InvalidArgument, "invalid availability")
    default:
        return status.Error(codes.Internal, "internal server error")
    }
}
```

## CQRS (Command Query Responsibility Segregation)

Separate read and write operations for clarity and scalability.

### Commands (Writes)

**Structure**:
```go
type CreateTeacherCommand struct {
    Name         string
    Email        string
    DepartmentID string
}

type CreateTeacherHandler struct {
    repo      domain.TeacherRepository
    publisher infrastructure.EventPublisher
}

func (h *CreateTeacherHandler) Handle(ctx context.Context, cmd CreateTeacherCommand) (string, error) {
    // ... implementation
}
```

**Characteristics**:
- One command handler per write operation
- File: `internal/application/command/{operation}_handler.go`
- Return: Error or result ID
- Side effect: Persist to database + publish event

### Queries (Reads)

**Structure**:
```go
type ListTeachersQuery struct {
    DepartmentID string
    PageSize     int
    PageNumber   int
}

type ListTeachersHandler struct {
    repo domain.TeacherRepository
}

func (h *ListTeachersHandler) Handle(ctx context.Context, q ListTeachersQuery) ([]TeacherDTO, error) {
    // ... implementation
}
```

**Characteristics**:
- One query handler per read operation
- File: `internal/application/query/{operation}_handler.go`
- Return: DTO (data transfer object), not domain entities
- No side effects

## Error Handling Strategy

### Domain Errors

Define custom errors in domain for business rule violations:

```go
// internal/errorhandling/errors.go
var (
    ErrTeacherNotFound        = errors.New("teacher not found")
    ErrDuplicateEmail         = errors.New("email already in use")
    ErrInvalidAvailability    = errors.New("availability period cannot be empty")
    ErrPrerequisiteCycleFound = errors.New("prerequisite cycle detected")
)
```

### Error Translation

At interface layer, translate domain errors to external status codes:

```go
func (s *TeacherServer) translateError(err error) error {
    switch {
    case errors.Is(err, domain.ErrTeacherNotFound):
        return status.Error(codes.NotFound, "teacher not found")
    case errors.Is(err, domain.ErrDuplicateEmail):
        return status.Error(codes.AlreadyExists, "email already in use")
    default:
        return status.Error(codes.Internal, "internal server error")
    }
}
```

### Logging Errors

Always log with context:

```go
if err := h.repo.Save(ctx, teacher); err != nil {
    logger.WithError(err).With(
        zap.String("teacher_id", teacher.ID),
        zap.String("operation", "CreateTeacher"),
    ).Error("Failed to save teacher")
    return err
}
```

## Event Sourcing

All mutations generate events in event store for audit trail + replay capability.

**Event Structure**:
```go
type Event struct {
    AggregateID   string      // e.g., teacher-id-123
    AggregateType string      // e.g., "teacher"
    EventType     string      // e.g., "created", "availability_updated"
    Payload       interface{} // Event-specific data
    Timestamp     time.Time
    Version       int         // Optimistic concurrency
}
```

**Publishing in Handler**:
```go
event := Event{
    AggregateID:   teacher.ID,
    AggregateType: "teacher",
    EventType:     "created",
    Payload: map[string]any{
        "name": teacher.Name,
        "email": teacher.Email,
    },
}

// Persist to event store (within same transaction as domain write)
if err := h.eventStore.Append(ctx, event); err != nil {
    return fmt.Errorf("append event: %w", err)
}

// Publish to NATS for async consumers
if err := h.publisher.Publish(ctx, event.EventType, event.Payload); err != nil {
    // Log warning; don't fail operation
    logger.WithError(err).Warn("Failed to publish event")
}
```

## Concurrency & Transactions

### Database Transactions

Always use transactions for atomic operations:

```go
func (h *CreateTeacherHandler) Handle(ctx context.Context, cmd CreateTeacherCommand) (string, error) {
    // Begin transaction
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return "", fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback() // Rollback if error

    // Save teacher within transaction
    if err := h.repo.SaveTx(ctx, tx, teacher); err != nil {
        return "", fmt.Errorf("save: %w", err)
    }

    // Append event within transaction
    if err := h.eventStore.AppendTx(ctx, tx, event); err != nil {
        return "", fmt.Errorf("append event: %w", err)
    }

    // Commit on success
    if err := tx.Commit(); err != nil {
        return "", fmt.Errorf("commit: %w", err)
    }

    return teacher.ID, nil
}
```

### Optimistic Concurrency

Use version column to detect concurrent writes:

```sql
UPDATE teachers
SET name = $1, version = version + 1
WHERE id = $2 AND version = $3;
```

If version mismatch, retry or return conflict error.

### Goroutines & Context

Always use context-aware patterns:

```go
// Good: Context-aware
go func(ctx context.Context) {
    select {
    case <-ctx.Done():
        return
    case result := <-ch:
        // Process result
    }
}(ctx)

// Avoid: Bare goroutine, no cancellation
go func() {
    // This runs forever, no way to stop
}()
```
