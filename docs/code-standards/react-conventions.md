# React Conventions

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| **Files** | kebab-case | `use-chat.ts`, `chat-panel.tsx`, `query-client.ts` |
| **Components** | PascalCase files, export PascalCase | `ChatPanel.tsx` exports `ChatPanel` |
| **Hooks** | `use` prefix, camelCase | `useChat`, `useCurrentUser`, `useTeachers` |
| **Types/Interfaces** | PascalCase | `ChatMessage`, `WsServerEvent`, `CreateTeacherForm` |
| **Variables/Functions** | camelCase | `fetchTeachers`, `isLoading`, `handleSubmit`, `pageSize` |
| **Constants** | UPPER_SNAKE_CASE (globals), camelCase (module-level) | `API_BASE_URL`, `defaultPageSize` |
| **CSS Classes** | kebab-case (Tailwind) | `text-sm`, `bg-blue-500`, `rounded-lg` |

## Directory Structure

```
frontend/src/
├── main.tsx                    # Entry point
├── index.css                   # Global Tailwind styles
│
├── config/                     # App-level configuration
│   ├── query-client.ts         # TanStack Query defaults
│   └── router.ts               # TanStack Router definition
│
├── lib/                        # Shared utilities & hooks
│   ├── api/
│   │   ├── client.ts           # Axios instance + interceptors
│   │   ├── endpoints.ts        # API route constants
│   │   └── types.ts            # Shared API response types
│   ├── hooks/
│   │   ├── use-auth.ts
│   │   ├── use-current-user.ts
│   │   └── use-toast.ts
│   ├── stores/
│   │   └── auth-store.ts
│   └── utils/
│       ├── cn.ts               # Tailwind classname merge
│       └── format-date.ts
│
├── components/                 # Reusable shared components
│   ├── layouts/
│   │   ├── app-layout.tsx
│   │   ├── sidebar-nav.tsx
│   │   └── top-bar.tsx
│   ├── shared/
│   │   ├── confirm-dialog.tsx
│   │   ├── data-table.tsx
│   │   ├── form-field.tsx
│   │   ├── loading-spinner.tsx
│   │   └── page-header.tsx
│   └── ui/                     # Shadcn/ui components
│
├── chat/                       # AI chat feature
│   ├── components/
│   │   ├── chat-panel.tsx
│   │   ├── chat-message.tsx
│   │   ├── chat-input.tsx
│   │   └── chat-toggle-button.tsx
│   ├── hooks/
│   │   └── use-chat.ts
│   └── types.ts
│
├── modules/                    # Feature modules (co-located)
│   ├── hr/
│   │   ├── components/         # HR-specific UI
│   │   ├── hooks/              # HR-specific hooks
│   │   └── types.ts            # HR types
│   ├── subject/
│   │   ├── components/
│   │   ├── hooks/
│   │   └── types.ts
│   └── timetable/
│       ├── components/
│       ├── hooks/
│       └── types.ts
│
└── routes/                     # File-based routing
    ├── __root.tsx
    ├── index.tsx
    ├── login.tsx
    ├── register.tsx
    └── _authenticated/
        ├── dashboard.tsx
        ├── hr/
        ├── subjects/
        └── timetable/
```

## File Organization

### Component Files

**Single-component file**:
```tsx
// components/shared/loading-spinner.tsx
export function LoadingSpinner() {
  return <div className="spinner">Loading...</div>
}
```

**Multi-component file** (if tightly coupled):
```tsx
// components/shared/confirm-dialog.tsx
function ConfirmDialogContent() { ... }
function ConfirmDialogTrigger() { ... }

export function ConfirmDialog() {
  return (
    <Dialog>
      <ConfirmDialogTrigger />
      <ConfirmDialogContent />
    </Dialog>
  )
}
```

### Hook Files

Keep hooks focused and exportable:

```tsx
// modules/hr/hooks/use-teachers.ts
export function useTeachers(filters?: TeacherFilters) {
  return useQuery({ ... })
}

export function useCreateTeacher() {
  return useMutation({ ... })
}

export function useUpdateTeacher(id: string) {
  return useMutation({ ... })
}

export function useDeleteTeacher(id: string) {
  return useMutation({ ... })
}
```

### Type Files

Group related types in single module file:

```tsx
// modules/hr/types.ts
export interface Teacher {
  id: string
  name: string
  email: string
  departmentId: string
  specializations: string[]
}

export interface CreateTeacherRequest {
  name: string
  email: string
  departmentId: string
}

export type TeacherFilters = {
  departmentId?: string
  search?: string
}
```

## Component Best Practices

### Functional Components Only

Use modern React 19 functional components:

```tsx
// Good: Functional component
export function ChatMessage({ message }: { message: ChatMessage }) {
  return <div className="message">{message.content}</div>
}

// Avoid: Class components
class ChatMessage extends React.Component { ... }

// Avoid: React.FC type
const ChatMessage: React.FC<{ message: ChatMessage }> = ({ message }) => ...
```

### Props Typing

Type props explicitly:

```tsx
// Good: Inline interface
interface DataTableProps<T> {
  data: T[]
  columns: ColumnDef<T>[]
  isLoading?: boolean
  onRowClick?: (row: T) => void
}

export function DataTable<T>(props: DataTableProps<T>) {
  return <table>...</table>
}

// Good: JSX.IntrinsicAttributes for HTML props
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary'
}

export function Button({ variant = 'primary', className, ...props }: ButtonProps) {
  return <button className={cn('btn', `btn-${variant}`, className)} {...props} />
}
```

### Event Handlers

Use `handle{Event}` naming convention:

```tsx
export function CreateTeacherForm() {
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // Submit form
  }

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value)
  }

  return (
    <form onSubmit={handleSubmit}>
      <input onChange={handleNameChange} />
    </form>
  )
}
```

## Hooks Conventions

### Data Fetching Hooks

Co-locate with modules; use query key factory pattern:

```tsx
// modules/hr/hooks/use-teachers.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { API_ROUTES } from '@/lib/api/endpoints'
import type { Teacher, CreateTeacherRequest } from '../types'

// Query key factory
const QUERY_KEY = ['teachers']

interface UseTeachersOptions {
  departmentId?: string
}

export function useTeachers(options?: UseTeachersOptions) {
  return useQuery({
    queryKey: [...QUERY_KEY, options],
    queryFn: async () => {
      const { data } = await apiClient.get<Teacher[]>(
        API_ROUTES.HR.TEACHERS,
        { params: options }
      )
      return data
    },
    staleTime: 30 * 1000, // 30 seconds
  })
}

export function useCreateTeacher() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateTeacherRequest) => {
      const res = await apiClient.post<Teacher>(API_ROUTES.HR.TEACHERS, data)
      return res.data
    },
    onSuccess: () => {
      // Invalidate list queries
      queryClient.invalidateQueries({ queryKey: QUERY_KEY })
    },
  })
}
```

### Custom Hooks

Keep hooks pure and reusable:

```tsx
// lib/hooks/use-debounce.ts
export function useDebounce<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delayMs)

    return () => clearTimeout(handler)
  }, [value, delayMs])

  return debouncedValue
}
```

### Auth Hooks

Centralized auth state:

```tsx
// lib/hooks/use-auth.ts
export function useAuth() {
  const navigate = useNavigate()
  const store = useAuthStore()

  const login = async (email: string, password: string) => {
    const { accessToken, refreshToken } = await apiClient.post(
      API_ROUTES.AUTH.LOGIN,
      { email, password }
    )
    store.setTokens(accessToken, refreshToken)
    navigate('/')
  }

  const logout = () => {
    store.clearTokens()
    navigate('/login')
  }

  return { login, logout, isAuthenticated: !!store.accessToken }
}
```

## State Management

### Query State (Server State)

Use TanStack Query for all server state:

```tsx
// Queries: Read from server
const { data: teachers, isLoading } = useTeachers()

// Mutations: Write to server
const { mutate: createTeacher, isPending } = useCreateTeacher()
```

### Auth State (Client State)

Use Zustand with localStorage persistence:

```tsx
// lib/stores/auth-store.ts
import create from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  user: User | null
  setTokens: (access: string, refresh: string) => void
  clearTokens: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      setTokens: (access, refresh) => set({ accessToken: access, refreshToken: refresh }),
      clearTokens: () => set({ accessToken: null, refreshToken: null, user: null }),
    }),
    {
      name: 'auth-store',
      storage: localStorage,
    }
  )
)
```

### Local UI State

Use `useState` for local component state:

```tsx
export function TeacherFilters() {
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedDept, setSelectedDept] = useState<string | null>(null)

  return (
    <div>
      <input
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
      />
    </div>
  )
}
```

## TypeScript Patterns

### Strict Mode

Enable TypeScript strict mode in `tsconfig.json`:

```json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "noImplicitThis": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true
  }
}
```

### Union Types

Use unions for flexible prop handling:

```tsx
type AlertVariant = 'success' | 'error' | 'warning' | 'info'

interface AlertProps {
  variant: AlertVariant
  message: string
}
```

### Generic Components

Create reusable generics:

```tsx
interface DataTableProps<T> {
  data: T[]
  columns: ColumnDef<T>[]
}

export function DataTable<T extends { id: string }>({
  data,
  columns,
}: DataTableProps<T>) {
  return <table>...</table>
}
```

## Import Organization

```tsx
import {
  // React + React DOM
  useState,
  useEffect,
  ReactNode,
} from 'react'

// TanStack libraries
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'

// Third-party libraries
import { Button } from '@/components/ui/button'

// Internal: lib
import { cn } from '@/lib/utils'
import { apiClient } from '@/lib/api/client'

// Internal: domain
import type { Teacher } from '../types'
import { useTeachers } from '../hooks/use-teachers'
```

## Code Formatting

### Line Length

Keep lines under 100 characters for readability.

### JSX Formatting

```tsx
// Good: Readable JSX
export function TeacherCard({ teacher }: { teacher: Teacher }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{teacher.name}</CardTitle>
      </CardHeader>
      <CardContent>
        <p>{teacher.email}</p>
      </CardContent>
    </Card>
  )
}

// Avoid: Cramped JSX
export const TeacherCard = ({ t }: { t: Teacher }) => <Card><CardHeader><CardTitle>{t.name}</CardTitle></CardHeader><CardContent><p>{t.email}</p></CardContent></Card>
```

## Accessibility (a11y)

### ARIA Labels

Always add labels for interactive elements:

```tsx
<button aria-label="Delete teacher" onClick={handleDelete}>
  <TrashIcon />
</button>

<input
  type="text"
  aria-label="Search teachers"
  placeholder="Search..."
/>
```

### Semantic HTML

Use semantic elements:

```tsx
// Good: Semantic
<header>
  <h1>Teachers</h1>
</header>
<main>
  <table>...</table>
</main>

// Avoid: Non-semantic
<div>
  <div>Teachers</div>
  <div>
    <div>...</div>
  </div>
</div>
```

### Keyboard Navigation

Ensure all interactions work with keyboard:

```tsx
<button
  onClick={handleDelete}
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      handleDelete()
    }
  }}
>
  Delete
</button>
```
