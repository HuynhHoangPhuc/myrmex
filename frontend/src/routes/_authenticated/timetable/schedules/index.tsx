import * as React from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { z } from 'zod'
import type { ColumnDef } from '@tanstack/react-table'
import { Eye } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/shared/page-header'
import { DataTable } from '@/components/shared/data-table'
import { useSchedules } from '@/modules/timetable/hooks/use-schedules'
import type { Schedule, ScheduleStatus } from '@/modules/timetable/types'

const searchSchema = z.object({
  page: z.number().catch(1),
  pageSize: z.number().catch(25),
  semesterId: z.string().optional().catch(undefined),
})

const STATUS_VARIANT: Record<ScheduleStatus, 'secondary' | 'outline' | 'destructive'> = {
  pending: 'outline',
  generating: 'secondary',
  completed: 'secondary',
  failed: 'destructive',
}

export const Route = createFileRoute('/_authenticated/timetable/schedules/')({
  validateSearch: (s) => searchSchema.parse(s),
  component: ScheduleListPage,
})

function ScheduleListPage() {
  const { page, pageSize, semesterId } = Route.useSearch()
  const navigate = Route.useNavigate()

  const { data, isLoading } = useSchedules({ page, pageSize, semesterId })

  const columns = React.useMemo<ColumnDef<Schedule>[]>(
    () => [
      {
        id: 'index',
        header: '#',
        cell: ({ row }) => (
          <span className="text-muted-foreground text-sm">
            {(page - 1) * pageSize + row.index + 1}
          </span>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => {
          const status = row.getValue<ScheduleStatus>('status')
          const label = status.charAt(0).toUpperCase() + status.slice(1)
          return <Badge variant={STATUS_VARIANT[status]}>{label}</Badge>
        },
      },
      {
        accessorKey: 'score',
        header: 'Score',
        cell: ({ row }) => row.original.status === 'completed'
          ? row.original.score.toFixed(2)
          : '—',
      },
      {
        accessorKey: 'hard_violations',
        header: 'Hard Violations',
        cell: ({ row }) => row.original.status === 'completed'
          ? <Badge variant={row.original.hard_violations === 0 ? 'secondary' : 'destructive'}>
              {row.original.hard_violations}
            </Badge>
          : '—',
      },
      {
        accessorKey: 'created_at',
        header: 'Generated',
        cell: ({ row }) => new Date(row.getValue('created_at')).toLocaleString(),
      },
      {
        id: 'actions',
        cell: ({ row }) => (
          <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
            <Link to="/timetable/schedules/$id" params={{ id: row.original.id }}>
              <Eye className="h-4 w-4" />
            </Link>
          </Button>
        ),
      },
    ],
    [page, pageSize],
  )

  return (
    <div>
      <PageHeader
        title="Schedules"
        description="View all generated timetable schedules."
        actions={
          <Button asChild>
            <Link to="/timetable/generate">Generate New</Link>
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        pagination={{ page, pageSize, total: data?.total ?? 0 }}
        onPageChange={(p) => void navigate({ search: { page: p, pageSize, semesterId } })}
      />
    </div>
  )
}
