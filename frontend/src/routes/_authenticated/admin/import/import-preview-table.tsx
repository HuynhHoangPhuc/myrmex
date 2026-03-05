import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface ImportPreviewTableProps {
  /** Raw CSV text content */
  csvText: string
}

// Parses a single CSV line respecting quoted fields
function parseCsvLine(line: string): string[] {
  const fields: string[] = []
  let current = ''
  let inQuotes = false

  for (let i = 0; i < line.length; i++) {
    const ch = line[i]
    if (ch === '"') {
      if (inQuotes && line[i + 1] === '"') {
        current += '"'
        i++
      } else {
        inQuotes = !inQuotes
      }
    } else if (ch === ',' && !inQuotes) {
      fields.push(current)
      current = ''
    } else {
      current += ch
    }
  }
  fields.push(current)
  return fields
}

// Shows header row + first 5 data rows from CSV text
export function ImportPreviewTable({ csvText }: ImportPreviewTableProps) {
  const lines = csvText
    .split('\n')
    .map((l) => l.trim())
    .filter(Boolean)

  if (lines.length === 0) return null

  const headers = parseCsvLine(lines[0])
  // Show up to 5 data rows (skip header at index 0)
  const previewRows = lines.slice(1, 6).map(parseCsvLine)

  if (previewRows.length === 0) return null

  return (
    <div className="rounded-md border">
      <div className="px-4 py-2 text-xs text-muted-foreground border-b">
        Preview — first {previewRows.length} row{previewRows.length !== 1 ? 's' : ''} of {lines.length - 1} data rows
      </div>
      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              {headers.map((h, i) => (
                <TableHead key={i} className="text-xs whitespace-nowrap">
                  {h}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {previewRows.map((row, ri) => (
              <TableRow key={ri}>
                {row.map((cell, ci) => (
                  <TableCell key={ci} className="text-xs py-2 max-w-[160px] truncate" title={cell}>
                    {cell || <span className="text-muted-foreground/50">—</span>}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
