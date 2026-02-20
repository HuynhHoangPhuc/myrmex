import * as React from 'react'
import {
  type ColumnDef,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { LoadingSpinner } from '@/components/shared/loading-spinner'

interface PaginationInfo {
  page: number
  pageSize: number
  total: number
}

interface DataTableProps<T> {
  columns: ColumnDef<T>[]
  data: T[]
  isLoading?: boolean
  pagination?: PaginationInfo
  onPageChange?: (page: number) => void
  onSortChange?: (sort: SortingState) => void
  toolbar?: React.ReactNode
  emptyMessage?: string
}

// Generic TanStack Table wrapper with Shadcn/ui Table primitives
// Supports server-side pagination + sorting, loading skeleton, empty state
export function DataTable<T>({
  columns,
  data,
  isLoading = false,
  pagination,
  onPageChange,
  onSortChange,
  toolbar,
  emptyMessage = 'No results found.',
}: DataTableProps<T>) {
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({})

  const table = useReactTable({
    data,
    columns,
    state: { sorting, columnVisibility },
    onSortingChange: (updater) => {
      const next = typeof updater === 'function' ? updater(sorting) : updater
      setSorting(next)
      onSortChange?.(next)
    },
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    // Pagination is managed server-side
    manualPagination: true,
    manualSorting: true,
  })

  return (
    <div className="space-y-4">
      {toolbar && <div className="flex items-center justify-between">{toolbar}</div>}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {isLoading ? (
              // Loading skeleton rows
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {columns.map((_, j) => (
                    <TableCell key={j}>
                      <div className="h-4 w-full animate-pulse rounded bg-muted" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : table.getRowModel().rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                  {emptyMessage}
                </TableCell>
              </TableRow>
            ) : (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() ? 'selected' : undefined}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {pagination && (
        <DataTablePagination pagination={pagination} onPageChange={onPageChange} isLoading={isLoading} />
      )}
    </div>
  )
}

interface DataTablePaginationProps {
  pagination: PaginationInfo
  onPageChange?: (page: number) => void
  isLoading?: boolean
}

function DataTablePagination({ pagination, onPageChange, isLoading }: DataTablePaginationProps) {
  const { page, pageSize, total } = pagination
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1
  const end = Math.min(page * pageSize, total)

  return (
    <div className="flex items-center justify-between px-2">
      <p className="text-sm text-muted-foreground">
        {total === 0 ? 'No results' : `Showing ${start}–${end} of ${total}`}
      </p>
      <div className="flex items-center gap-2">
        <button
          className="rounded border px-3 py-1 text-sm disabled:opacity-50"
          onClick={() => onPageChange?.(page - 1)}
          disabled={page <= 1 || isLoading}
        >
          Previous
        </button>
        <span className="text-sm">
          {isLoading ? <LoadingSpinner size="sm" /> : `Page ${page} of ${totalPages}`}
        </span>
        <button
          className="rounded border px-3 py-1 text-sm disabled:opacity-50"
          onClick={() => onPageChange?.(page + 1)}
          disabled={page >= totalPages || isLoading}
        >
          Next
        </button>
      </div>
    </div>
  )
}
