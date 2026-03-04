import * as React from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'
import { PageHeader } from '@/components/shared/page-header'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export const Route = createFileRoute('/_authenticated/admin/audit-logs/')({
  component: AuditLogsPage,
})

interface AuditLogRow {
  id: string
  user_id: string
  user_role: string
  action: string
  resource_type: string
  resource_id?: string
  old_value?: unknown
  new_value?: unknown
  ip_address?: string
  status_code?: number
  created_at: string
}

interface AuditLogsResponse {
  data: AuditLogRow[]
  total: number
  page: number
  page_size: number
}

const RESOURCE_TYPES = ['', 'teacher', 'department', 'subject', 'student', 'enrollment', 'grade', 'user', 'role', 'semester', 'schedule']

function statusBadge(code?: number) {
  if (!code) return null
  const variant = code < 300 ? 'default' : code < 500 ? 'secondary' : 'destructive'
  return <Badge variant={variant}>{code}</Badge>
}

function useAuditLogs(filters: Record<string, string>, page: number, pageSize: number) {
  return useQuery({
    queryKey: ['audit-logs', filters, page, pageSize],
    queryFn: async () => {
      const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
      Object.entries(filters).forEach(([k, v]) => { if (v) params.set(k, v) })
      const { data } = await apiClient.get<AuditLogsResponse>(
        `${ENDPOINTS.auditLogs.list}?${params.toString()}`,
      )
      return data
    },
  })
}

function AuditLogsPage() {
  const [filters, setFilters] = React.useState<Record<string, string>>({
    resource_type: '',
    action: '',
    user_id: '',
  })
  const [page, setPage] = React.useState(1)
  const [expandedRow, setExpandedRow] = React.useState<string | null>(null)
  const pageSize = 20

  const { data, isLoading } = useAuditLogs(filters, page, pageSize)
  const logs = data?.data ?? []
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  function setFilter(key: string, value: string) {
    setFilters((prev) => ({ ...prev, [key]: value }))
    setPage(1)
  }

  return (
    <div className="space-y-6">
      <PageHeader title="Audit Logs" description="Immutable record of all system mutations" />

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3">
        <Select value={filters.resource_type} onValueChange={(v) => setFilter('resource_type', v === '_all' ? '' : v)}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="Resource type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_all">All resources</SelectItem>
            {RESOURCE_TYPES.filter(Boolean).map((r) => (
              <SelectItem key={r} value={r}>{r}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Input
          placeholder="Filter by action..."
          className="w-48"
          value={filters.action}
          onChange={(e) => setFilter('action', e.target.value)}
        />

        <Input
          placeholder="Filter by user ID..."
          className="w-64"
          value={filters.user_id}
          onChange={(e) => setFilter('user_id', e.target.value)}
        />

        {(filters.resource_type || filters.action || filters.user_id) && (
          <Button variant="ghost" size="sm" onClick={() => { setFilters({ resource_type: '', action: '', user_id: '' }); setPage(1) }}>
            Clear
          </Button>
        )}
      </div>

      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>Resource</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>IP</TableHead>
              <TableHead className="w-[60px]" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground py-8">Loading...</TableCell>
              </TableRow>
            ) : logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground py-8">No audit logs found</TableCell>
              </TableRow>
            ) : logs.map((log) => (
              <React.Fragment key={log.id}>
                <TableRow className="cursor-pointer hover:bg-muted/50" onClick={() => setExpandedRow(expandedRow === log.id ? null : log.id)}>
                  <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
                    {new Date(log.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell className="font-mono text-xs max-w-[120px] truncate" title={log.user_id}>
                    {log.user_id.slice(0, 8)}…
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-xs">{log.user_role}</Badge>
                  </TableCell>
                  <TableCell className="text-xs font-mono">{log.action}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{log.resource_type}</TableCell>
                  <TableCell>{statusBadge(log.status_code)}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{log.ip_address ?? '—'}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">{expandedRow === log.id ? '▲' : '▼'}</TableCell>
                </TableRow>
                {expandedRow === log.id && (
                  <TableRow>
                    <TableCell colSpan={8} className="bg-muted/30 p-4">
                      <div className="grid grid-cols-2 gap-4 text-xs">
                        <div>
                          <p className="font-semibold mb-1 text-muted-foreground">Before</p>
                          <pre className="bg-background rounded p-2 overflow-auto max-h-40 text-xs">
                            {log.old_value ? JSON.stringify(log.old_value, null, 2) : '(none)'}
                          </pre>
                        </div>
                        <div>
                          <p className="font-semibold mb-1 text-muted-foreground">After</p>
                          <pre className="bg-background rounded p-2 overflow-auto max-h-40 text-xs">
                            {log.new_value ? JSON.stringify(log.new_value, null, 2) : '(none)'}
                          </pre>
                        </div>
                      </div>
                      {log.resource_id && (
                        <p className="mt-2 text-xs text-muted-foreground">Resource ID: {log.resource_id}</p>
                      )}
                    </TableCell>
                  </TableRow>
                )}
              </React.Fragment>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <span>{total} total entries</span>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Previous
          </Button>
          <span>Page {page} / {totalPages}</span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
