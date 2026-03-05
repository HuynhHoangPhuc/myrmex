import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

export interface ImportError {
  row: number
  message: string
}

export interface ImportResult {
  total: number
  created: number
  skipped: number
  errors: ImportError[]
}

interface ImportResultsProps {
  result: ImportResult
  /** Label used in the error CSV filename, e.g. "teachers" or "students" */
  typeLabel: string
}

function downloadErrorCsv(errors: ImportError[], typeLabel: string) {
  const header = 'Row,Error Message'
  const rows = errors.map((e) => `${e.row},"${e.message.replace(/"/g, '""')}"`)
  const csv = [header, ...rows].join('\n')
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${typeLabel}-import-errors.csv`
  a.click()
  URL.revokeObjectURL(url)
}

// Displays import summary badges and, if errors exist, an error table with download option
export function ImportResults({ result, typeLabel }: ImportResultsProps) {
  const hasErrors = result.errors.length > 0

  return (
    <div className="space-y-4">
      {/* Summary badges */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-1.5">
          <span className="text-sm text-muted-foreground">Total:</span>
          <Badge variant="outline">{result.total}</Badge>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-sm text-muted-foreground">Created:</span>
          <Badge variant="default">{result.created}</Badge>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-sm text-muted-foreground">Skipped:</span>
          <Badge variant="secondary">{result.skipped}</Badge>
        </div>
        {hasErrors && (
          <div className="flex items-center gap-1.5">
            <span className="text-sm text-muted-foreground">Errors:</span>
            <Badge variant="destructive">{result.errors.length}</Badge>
          </div>
        )}
      </div>

      {/* Error table */}
      {hasErrors && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-destructive">
              {result.errors.length} row{result.errors.length !== 1 ? 's' : ''} failed
            </p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => downloadErrorCsv(result.errors, typeLabel)}
            >
              Download Error Report
            </Button>
          </div>

          <div className="rounded-md border border-destructive/30">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[80px]">Row</TableHead>
                  <TableHead>Error Message</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.errors.map((err, i) => (
                  <TableRow key={i}>
                    <TableCell className="font-mono text-xs">{err.row}</TableCell>
                    <TableCell className="text-xs text-destructive">{err.message}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      )}

      {!hasErrors && (
        <p className="text-sm text-muted-foreground">All rows imported successfully.</p>
      )}
    </div>
  )
}
