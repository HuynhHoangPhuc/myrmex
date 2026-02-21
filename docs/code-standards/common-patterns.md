# Common Patterns (Backend + Frontend)

## Pagination

### Backend (Go + sqlc)

**Query Definition** (`services/module-hr/internal/sql/queries/teachers.sql`):

```sql
-- name: ListTeachers :many
SELECT * FROM teachers
WHERE department_id = $1
  AND ($2::text IS NULL OR name ILIKE $2)
ORDER BY name
LIMIT $3 OFFSET $4;

-- name: CountTeachers :one
SELECT COUNT(*) FROM teachers
WHERE department_id = $1
  AND ($2::text IS NULL OR name ILIKE $2);
```

**Handler** (`services/module-hr/internal/application/query/list_teachers_handler.go`):

```go
type ListTeachersQuery struct {
    DepartmentID string
    SearchTerm   string
    PageSize     int
    PageNumber   int // 1-indexed
}

type TeacherDTO struct {
    ID    string
    Name  string
    Email string
}

type ListTeachersResponse struct {
    Items      []TeacherDTO
    Total      int64
    PageSize   int
    PageNumber int
}

type ListTeachersHandler struct {
    repo domain.TeacherRepository
}

func (h *ListTeachersHandler) Handle(ctx context.Context, q ListTeachersQuery) (*ListTeachersResponse, error) {
    // Convert 1-indexed to 0-indexed offset
    offset := (q.PageNumber - 1) * q.PageSize

    // Fetch teachers
    teachers, err := h.repo.ListByDepartment(ctx, q.DepartmentID, q.SearchTerm, q.PageSize, offset)
    if err != nil {
        return nil, err
    }

    // Fetch total count
    total, err := h.repo.CountByDepartment(ctx, q.DepartmentID, q.SearchTerm)
    if err != nil {
        return nil, err
    }

    // Convert to DTOs
    dtos := make([]TeacherDTO, len(teachers))
    for i, t := range teachers {
        dtos[i] = TeacherDTO{ID: t.ID, Name: t.Name, Email: t.Email}
    }

    return &ListTeachersResponse{
        Items:      dtos,
        Total:      total,
        PageSize:   q.PageSize,
        PageNumber: q.PageNumber,
    }, nil
}
```

### Frontend (React + TanStack Query)

**Query Hook**:

```tsx
// modules/hr/hooks/use-teachers.ts
const TEACHERS_QUERY_KEY = ['teachers']

interface UseTeachersOptions {
  departmentId: string
  page?: number
  pageSize?: number
  search?: string
}

export function useTeachers({
  departmentId,
  page = 1,
  pageSize = 10,
  search,
}: UseTeachersOptions) {
  return useQuery({
    queryKey: [...TEACHERS_QUERY_KEY, { departmentId, page, pageSize, search }],
    queryFn: async () => {
      const { data } = await apiClient.get('/teachers', {
        params: {
          department_id: departmentId,
          page,
          page_size: pageSize,
          search,
        },
      })
      return data as {
        items: Teacher[]
        total: number
        pageSize: number
        pageNumber: number
      }
    },
    staleTime: 30 * 1000,
  })
}
```

**Component**:

```tsx
export function TeachersPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const pageSize = 10

  const { data, isLoading, error } = useTeachers({
    departmentId: 'dept-123',
    page,
    pageSize,
    search,
  })

  const pageCount = Math.ceil((data?.total ?? 0) / pageSize)

  return (
    <div>
      <input
        placeholder="Search..."
        value={search}
        onChange={(e) => {
          setSearch(e.target.value)
          setPage(1) // Reset to first page on search
        }}
      />

      <DataTable data={data?.items ?? []} columns={columns} isLoading={isLoading} />

      <div className="pagination">
        <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1}>
          Previous
        </button>
        <span>
          Page {page} of {pageCount}
        </span>
        <button onClick={() => setPage((p) => p + 1)} disabled={page === pageCount}>
          Next
        </button>
      </div>
    </div>
  )
}
```

## Error Translation

### Backend (Domain → gRPC)

**Domain Error**:

```go
// internal/errorhandling/errors.go
var (
    ErrTeacherNotFound = errors.New("teacher not found")
    ErrDuplicateEmail = errors.New("email already in use")
)
```

**gRPC Handler**:

```go
// internal/interface/grpc/teacher_server.go
func (s *TeacherServer) GetTeacher(
    ctx context.Context,
    req *pb.GetTeacherRequest,
) (*pb.Teacher, error) {
    teacher, err := s.getTeacherHandler.Handle(ctx, req.Id)
    if err != nil {
        return nil, s.translateError(err)
    }
    return toProto(teacher), nil
}

func (s *TeacherServer) translateError(err error) error {
    if errors.Is(err, domain.ErrTeacherNotFound) {
        return status.Error(codes.NotFound, "teacher not found")
    }
    if errors.Is(err, domain.ErrDuplicateEmail) {
        return status.Error(codes.AlreadyExists, "email already in use")
    }
    // Default: internal error (don't expose internal details)
    logger.WithError(err).Error("Unexpected error")
    return status.Error(codes.Internal, "internal server error")
}
```

### Frontend (API → UI)

**API Client**:

```tsx
// lib/api/client.ts
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
})

// 401: Unauthorized (token expired)
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().clearTokens()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
```

**Error Handling in Components**:

```tsx
export function TeacherDetail({ id }: { id: string }) {
  const { data: teacher, error, isError } = useTeacher(id)

  if (isError) {
    if (error?.response?.status === 404) {
      return <ErrorAlert title="Not Found" message="Teacher not found" />
    }
    if (error?.response?.status === 403) {
      return <ErrorAlert title="Forbidden" message="You don't have permission to view this" />
    }
    return <ErrorAlert title="Error" message={error?.message ?? 'Something went wrong'} />
  }

  return <TeacherCard teacher={teacher} />
}
```

## Logging Strategy

### Backend (Go + Zap)

**Service Startup**:

```go
logger.Info("Starting service",
    zap.String("service", "module-hr"),
    zap.String("version", version),
    zap.Int("grpc_port", 50052),
)
```

**Operation Success**:

```go
logger.Info("Teacher created",
    zap.String("teacher_id", teacher.ID),
    zap.String("email", teacher.Email),
    zap.Duration("elapsed", time.Since(start)),
)
```

**Operation Error**:

```go
logger.WithError(err).Error("Failed to create teacher",
    zap.String("email", cmd.Email),
    zap.String("operation", "CreateTeacher"),
)
```

**Debug Logging** (only in dev):

```go
if logger.Level() == zap.DebugLevel {
    logger.Debug("Query executed",
        zap.String("query", "SELECT * FROM teachers"),
        zap.Duration("duration", elapsed),
    )
}
```

### Frontend (Console + Error Tracking)

**Info Log**:

```tsx
console.log('[INFO] Teachers loaded', { count: teachers.length })
```

**Error Log** (with context):

```tsx
if (error) {
    console.error('[ERROR] Failed to fetch teachers', {
        error: error.message,
        status: error.response?.status,
        url: error.config?.url,
    })
}
```

**Never Log**:
- Passwords
- API keys
- Credit card numbers
- PII (personal identifiable information)

## Testing Patterns

### Backend (Go)

**Table-Driven Tests**:

```go
func TestTeacherAvailability(t *testing.T) {
    tests := []struct {
        name      string
        day       DayOfWeek
        periods   []PeriodOfDay
        wantErr   bool
        errType   error
    }{
        {"Valid single period", Monday, []PeriodOfDay{1}, false, nil},
        {"Valid multiple periods", Tuesday, []PeriodOfDay{1, 2, 3}, false, nil},
        {"Empty periods", Wednesday, []PeriodOfDay{}, true, ErrInvalidAvailability},
        {"Duplicate period", Thursday, []PeriodOfDay{1, 1}, true, ErrDuplicatePeriod},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            teacher := domain.NewTeacher("John", "john@example.com", "dept-1")
            err := teacher.UpdateAvailability(tt.day, tt.periods)

            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, want %v", err, tt.wantErr)
            }

            if tt.wantErr && !errors.Is(err, tt.errType) {
                t.Errorf("got error type %T, want %T", err, tt.errType)
            }
        })
    }
}
```

**Mocking**:

```go
// Use gomock for interface mocks
type MockTeacherRepository struct {
    SaveFunc func(ctx context.Context, teacher *domain.Teacher) error
}

func (m *MockTeacherRepository) Save(ctx context.Context, teacher *domain.Teacher) error {
    return m.SaveFunc(ctx, teacher)
}

func TestCreateTeacherHandler(t *testing.T) {
    mockRepo := &MockTeacherRepository{
        SaveFunc: func(ctx context.Context, teacher *domain.Teacher) error {
            return nil // Success case
        },
    }

    handler := application.NewCreateTeacherHandler(mockRepo)
    // Test handler...
}
```

### Frontend (React)

**Component Tests**:

```tsx
// __tests__/CreateTeacherForm.test.tsx
import { render, screen } from '@testing-library/react'
import { userEvent } from '@testing-library/user-event'
import { vi } from 'vitest'
import { CreateTeacherForm } from '../components/create-teacher-form'

vi.mock('../hooks/use-teachers', () => ({
  useCreateTeacher: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}))

test('submits form with valid data', async () => {
  render(<CreateTeacherForm />)

  await userEvent.type(screen.getByLabelText('Name'), 'John Doe')
  await userEvent.type(screen.getByLabelText('Email'), 'john@example.com')
  await userEvent.selectOptions(screen.getByLabelText('Department'), 'dept-1')
  await userEvent.click(screen.getByText('Create'))

  expect(mockMutate).toHaveBeenCalledWith(
    expect.objectContaining({
      name: 'John Doe',
      email: 'john@example.com',
    })
  )
})

test('displays validation errors', async () => {
  render(<CreateTeacherForm />)

  await userEvent.click(screen.getByText('Create'))

  expect(screen.getByText('Name required')).toBeInTheDocument()
  expect(screen.getByText('Email required')).toBeInTheDocument()
})
```

## Security Patterns

### Backend

**Password Hashing** (use bcrypt, never plain text):

```go
// Hash password before storing
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
    return err
}

// Verify password on login
err = bcrypt.CompareHashAndPassword(storedHash, []byte(password))
if err != nil {
    return ErrInvalidPassword
}
```

**JWT Token Creation**:

```go
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "sub": userID,
    "exp": time.Now().Add(15 * time.Minute).Unix(),
    "iat": time.Now().Unix(),
})

signedToken, err := token.SignedString([]byte(jwtSecret))
```

**Input Validation**:

```go
// Always validate on boundary (gRPC/HTTP handler)
if err := req.Validate(); err != nil {
    return status.Error(codes.InvalidArgument, err.Error())
}

// Domain layer trusts validated input
if len(teacher.Email) == 0 {
    return ErrInvalidEmail
}
```

### Frontend

**Token Storage** (never localStorage for sensitive apps, but OK for MVP):

```tsx
// localStorage: Simple, persists across tabs, XSS-vulnerable
localStorage.setItem('accessToken', token)

// sessionStorage: Cleared on tab close, XSS-vulnerable
sessionStorage.setItem('accessToken', token)

// Memory: Secure, cleared on refresh
// Use HttpOnly cookies in production (requires backend support)
```

**Token in Requests** (via axios interceptor):

```tsx
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
```

**Never Expose Secrets**:

```tsx
// Bad: API key in client
const API_KEY = 'sk-secret-key-123' // Never do this!

// Good: Call backend endpoint, backend calls LLM
const response = await apiClient.post('/chat', { message })
```

## Monitoring & Observability

### Backend Metrics

```go
// Prometheus metrics (future integration)
var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"method", "path"},
    )
)
```

### Frontend Performance Metrics

```tsx
// Web Vitals (future integration)
import { onCLS, onFID, onFCP, onLCP, onTTFB } from 'web-vitals'

onCLS((metric) => console.log('CLS:', metric))
onFID((metric) => console.log('FID:', metric))
onFCP((metric) => console.log('FCP:', metric))
onLCP((metric) => console.log('LCP:', metric))
onTTFB((metric) => console.log('TTFB:', metric))
```

## Configuration Best Practices

### Backend (Viper)

```yaml
# Default config (in repo, no secrets)
server:
  grpc_port: 50052
  request_timeout: 30s

database:
  max_connections: 20

logging:
  level: info
  format: json
```

Environment override:
```bash
export MODULE_SERVER_GRPC_PORT=50053
export MODULE_DATABASE_URL="postgres://..."
```

### Frontend (Vite)

```env
# .env (in repo, public)
VITE_API_URL=http://localhost:8000

# .env.production
VITE_API_URL=https://api.myrmex.app

# .env.local (Git-ignored, secrets)
VITE_INTERNAL_API_KEY=sk-...
```

Access in code:
```tsx
const apiUrl = import.meta.env.VITE_API_URL
```
