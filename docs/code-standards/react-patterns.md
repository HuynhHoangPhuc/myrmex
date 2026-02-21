# React Patterns & Best Practices

## Forms with TanStack Form + Zod

Combine TanStack Form for logic + Zod for validation:

```tsx
// modules/hr/components/create-teacher-form.tsx
import { useForm } from '@tanstack/react-form'
import { zodValidator } from '@tanstack/zod-form-adapter'
import { z } from 'zod'
import { useCreateTeacher } from '../hooks/use-teachers'

const schema = z.object({
  name: z.string().min(1, 'Name required'),
  email: z.string().email('Invalid email'),
  departmentId: z.string().min(1, 'Department required'),
})

type CreateTeacherFormData = z.infer<typeof schema>

export function CreateTeacherForm() {
  const { mutate, isPending } = useCreateTeacher()

  const form = useForm({
    defaultValues: {
      name: '',
      email: '',
      departmentId: '',
    },
    onSubmit: async (values) => {
      mutate(values)
    },
    validatorAdapter: zodValidator(),
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        e.stopPropagation()
        form.handleSubmit()
      }}
    >
      <form.Field
        name="name"
        children={(field) => (
          <div>
            <label htmlFor={field.name}>Name</label>
            <input
              id={field.name}
              name={field.name}
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
            />
            {field.state.meta.errors?.length > 0 && (
              <span className="text-red-600">{field.state.meta.errors[0]}</span>
            )}
          </div>
        )}
      />

      {/* Other fields... */}

      <button type="submit" disabled={isPending}>
        {isPending ? 'Creating...' : 'Create'}
      </button>
    </form>
  )
}
```

## Query Key Management

Hierarchical query keys for cache invalidation:

```tsx
// modules/hr/hooks/use-teachers.ts
const TEACHERS_QUERY_KEY = ['teachers']

export function useTeachers(filters?: TeacherFilters) {
  return useQuery({
    queryKey: [...TEACHERS_QUERY_KEY, filters],
    queryFn: () => fetchTeachers(filters),
    staleTime: 30 * 1000,
  })
}

export function useTeacher(id: string) {
  return useQuery({
    queryKey: [...TEACHERS_QUERY_KEY, id],
    queryFn: () => fetchTeacher(id),
  })
}

export function useCreateTeacher() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createTeacher,
    onSuccess: () => {
      // Invalidate all teacher queries
      queryClient.invalidateQueries({ queryKey: TEACHERS_QUERY_KEY })
    },
  })
}

export function useUpdateTeacher(id: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: Partial<Teacher>) => updateTeacher(id, data),
    onSuccess: (updated) => {
      // Update specific teacher detail
      queryClient.setQueryData([...TEACHERS_QUERY_KEY, id], updated)
      // Invalidate list (may have changed ordering, dept, etc.)
      queryClient.invalidateQueries({ queryKey: TEACHERS_QUERY_KEY })
    },
  })
}
```

## Optimistic Updates

Improve UX by updating UI immediately while request is pending:

```tsx
export function useDeleteTeacher(id: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (teacherId: string) => apiClient.delete(`/teachers/${teacherId}`),

    // Optimistic: Update before server responds
    onMutate: async (teacherId) => {
      // Cancel outgoing queries
      await queryClient.cancelQueries({ queryKey: TEACHERS_QUERY_KEY })

      // Snapshot previous data
      const previousTeachers = queryClient.getQueryData(TEACHERS_QUERY_KEY)

      // Optimistically update UI
      queryClient.setQueryData(TEACHERS_QUERY_KEY, (old: Teacher[]) =>
        old.filter((t) => t.id !== teacherId)
      )

      return { previousTeachers }
    },

    // Rollback on error
    onError: (err, variables, context) => {
      if (context?.previousTeachers) {
        queryClient.setQueryData(TEACHERS_QUERY_KEY, context.previousTeachers)
      }
    },

    // Confirm on success
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TEACHERS_QUERY_KEY })
    },
  })
}

// Usage
export function DeleteTeacherButton({ teacherId }: { teacherId: string }) {
  const { mutate, isPending } = useDeleteTeacher(teacherId)

  return (
    <button
      onClick={() => mutate(teacherId)}
      disabled={isPending}
    >
      {isPending ? 'Deleting...' : 'Delete'}
    </button>
  )
}
```

## Data Table (TanStack Table)

Reusable table component with sorting, filtering, pagination:

```tsx
// components/shared/data-table.tsx
import {
  useReactTable,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  ColumnDef,
} from '@tanstack/react-table'

interface DataTableProps<T> {
  data: T[]
  columns: ColumnDef<T>[]
  isLoading?: boolean
  pageSize?: number
  onPaginationChange?: (pageIndex: number) => void
}

export function DataTable<T>({
  data,
  columns,
  isLoading,
  pageSize = 10,
  onPaginationChange,
}: DataTableProps<T>) {
  const [pagination, setPagination] = React.useState({ pageIndex: 0, pageSize })

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    state: { pagination },
    onPaginationChange: setPagination,
  })

  return (
    <div>
      {isLoading && <LoadingSpinner />}

      <table>
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th
                  key={header.id}
                  onClick={header.column.getToggleSortingHandler()}
                >
                  {flexRender(header.column.columnDef.header, header.getContext())}
                </th>
              ))}
            </tr>
          ))}
        </thead>

        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <td key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>

      {/* Pagination controls */}
      <div>
        <button onClick={() => table.previousPage()}>Previous</button>
        <span>Page {pagination.pageIndex + 1}</span>
        <button onClick={() => table.nextPage()}>Next</button>
      </div>
    </div>
  )
}

// Usage
export function TeachersTable() {
  const { data: teachers } = useTeachers()

  const columns: ColumnDef<Teacher>[] = [
    { accessorKey: 'name', header: 'Name' },
    { accessorKey: 'email', header: 'Email' },
    {
      id: 'actions',
      cell: ({ row }) => (
        <Button onClick={() => editTeacher(row.original.id)}>Edit</Button>
      ),
    },
  ]

  return <DataTable data={teachers ?? []} columns={columns} />
}
```

## WebSocket Integration

Real-time chat with auto-reconnect:

```tsx
// chat/hooks/use-chat.ts
import { useCallback, useEffect, useRef, useState } from 'react'
import { useAuthStore } from '@/lib/stores/auth-store'

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
}

interface WsServerEvent {
  type: 'message' | 'typing' | 'error' | 'connected'
  message?: ChatMessage
  error?: string
}

export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [isReconnecting, setIsReconnecting] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const MAX_RECONNECT_ATTEMPTS = 5
  const token = useAuthStore((s) => s.accessToken)

  const connect = useCallback(() => {
    if (!token) return

    const wsUrl = `${import.meta.env.VITE_CHAT_WS_URL}?token=${token}`

    try {
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        setIsConnected(true)
        setIsReconnecting(false)
        reconnectAttemptsRef.current = 0
      }

      ws.onmessage = (event) => {
        const data: WsServerEvent = JSON.parse(event.data)

        if (data.type === 'message' && data.message) {
          setMessages((prev) => [...prev, data.message])
        } else if (data.type === 'error') {
          console.error('Chat error:', data.error)
        }
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        setIsConnected(false)
      }

      ws.onclose = () => {
        setIsConnected(false)
        attemptReconnect()
      }

      wsRef.current = ws
    } catch (error) {
      console.error('Failed to connect:', error)
      attemptReconnect()
    }
  }, [token])

  const attemptReconnect = useCallback(() => {
    if (reconnectAttemptsRef.current >= MAX_RECONNECT_ATTEMPTS) {
      console.error('Max reconnection attempts reached')
      return
    }

    reconnectAttemptsRef.current += 1
    setIsReconnecting(true)

    // Exponential backoff: 1s, 2s, 4s, 8s, 16s
    const delay = Math.pow(2, reconnectAttemptsRef.current - 1) * 1000
    setTimeout(() => connect(), delay)
  }, [connect])

  const send = useCallback((content: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected')
      return
    }

    wsRef.current.send(JSON.stringify({ type: 'message', content }))
    setMessages((prev) => [...prev, { role: 'user', content }])
  }, [])

  useEffect(() => {
    connect()

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [connect])

  return {
    messages,
    send,
    isConnected,
    isReconnecting,
  }
}

// Usage
export function ChatPanel() {
  const { messages, send, isConnected } = useChat()
  const [input, setInput] = useState('')

  const handleSend = () => {
    if (input.trim()) {
      send(input)
      setInput('')
    }
  }

  return (
    <div className="chat-panel">
      <div className="messages">
        {messages.map((msg, i) => (
          <ChatMessage key={i} message={msg} />
        ))}
      </div>

      {!isConnected && <div className="error">Disconnected</div>}

      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && handleSend()}
        disabled={!isConnected}
        placeholder="Type message..."
      />
      <button onClick={handleSend} disabled={!isConnected}>
        Send
      </button>
    </div>
  )
}
```

## URL-Driven Pagination

Keep pagination state in URL for bookmarkability:

```tsx
// lib/api/hooks/use-paginated-query.ts
import { useNavigate } from '@tanstack/react-router'
import { useSearch } from '@tanstack/react-router'

const paginationSchema = z.object({
  page: z.number().int().positive().default(1),
  pageSize: z.number().int().positive().default(10),
  search: z.string().optional(),
})

export function usePaginatedTeachers() {
  const navigate = useNavigate()
  const search = useSearch({ from: '/teachers' })
  const { page = 1, pageSize = 10, search: searchTerm } = paginationSchema.parse(search)

  const { data } = useQuery({
    queryKey: ['teachers', { page, pageSize, searchTerm }],
    queryFn: () => fetchTeachers({ page, pageSize, search: searchTerm }),
  })

  const onPageChange = (newPage: number) => {
    navigate({
      search: { page: newPage, pageSize, search: searchTerm },
    })
  }

  return {
    teachers: data?.items ?? [],
    pageCount: Math.ceil((data?.total ?? 0) / pageSize),
    currentPage: page,
    onPageChange,
  }
}

// URL: /teachers?page=2&pageSize=20&search=john
```

## Component Patterns

### Render Props (for complex state logic)

```tsx
interface RenderProps<T> {
  data: T | null
  isLoading: boolean
  error: Error | null
  refetch: () => void
}

interface QueryBoundaryProps<T> {
  queryFn: () => Promise<T>
  children: (props: RenderProps<T>) => React.ReactNode
}

export function QueryBoundary<T>({
  queryFn,
  children,
}: QueryBoundaryProps<T>) {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['query-boundary'],
    queryFn,
  })

  return children({ data: data ?? null, isLoading, error: error as Error | null, refetch })
}

// Usage
<QueryBoundary
  queryFn={() => fetchTeachers()}
  children={({ data, isLoading, error }) => (
    <>
      {isLoading && <Spinner />}
      {error && <ErrorAlert error={error} />}
      {data && <TeachersList teachers={data} />}
    </>
  )}
/>
```

### Composition Pattern

Build complex UIs by composing smaller components:

```tsx
// Good: Composable
<Card>
  <CardHeader>
    <CardTitle>Teachers</CardTitle>
  </CardHeader>
  <CardContent>
    <DataTable columns={columns} data={teachers} />
  </CardContent>
  <CardFooter>
    <Button onClick={handleCreate}>Add Teacher</Button>
  </CardFooter>
</Card>

// Avoid: Monolithic component
<TeachersList /> {/* Does everything internally */}
```

## Performance Optimization

### Memoization (Use Sparingly)

Only memoize expensive computations or frequently-re-rendering components:

```tsx
// Expensive computation
const expensiveValue = useMemo(() => {
  return teachers.reduce((acc, t) => {
    // Complex calculation
    return acc + t.hours
  }, 0)
}, [teachers])

// Expensive callback
const handleComplexLogic = useCallback(() => {
  // Complex operations
}, [dependency1, dependency2])

// Component memoization (if receiving many props)
const TeacherRow = memo(({ teacher }: { teacher: Teacher }) => (
  <tr>
    <td>{teacher.name}</td>
  </tr>
))
```

### Code Splitting (TanStack Router)

Routes auto-split per route:

```tsx
// routes/teachers/$id.tsx is lazy-loaded when route matches
export const Route = createFileRoute('/teachers/$id')({
  component: TeacherDetailPage,
})

// No explicit code splitting needed
```

### Image Lazy Loading

```tsx
<img
  src="/teacher-photo.jpg"
  loading="lazy"
  alt="Teacher photo"
  width={200}
  height={200}
/>
```

## Error Handling

### Error Boundary

Catch rendering errors:

```tsx
// components/error-boundary.tsx
export class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error: Error | null }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }

  render() {
    if (this.state.hasError) {
      return <div className="error">Something went wrong: {this.state.error?.message}</div>
    }

    return this.props.children
  }
}
```

### Query Error Handling

```tsx
export function TeachersTable() {
  const { data: teachers, error, isError } = useTeachers()

  if (isError) {
    return <ErrorAlert error={error} retryFn={refetch} />
  }

  return <table>{/* ... */}</table>
}
```
