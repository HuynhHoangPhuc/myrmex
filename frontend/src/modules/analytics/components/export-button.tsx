import { Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { apiClient } from '@/lib/api/client'
import { ENDPOINTS } from '@/lib/api/endpoints'

type ReportType = 'workload' | 'utilization' | 'schedule'

interface ExportButtonProps {
  /** Which analytics report to export */
  type: ReportType
  /** Optional semester UUID — omit to export all semesters */
  semesterId?: string
}

async function downloadExport(format: 'pdf' | 'xlsx', type: ReportType, semesterId?: string) {
  const params: Record<string, string> = { format, type }
  if (semesterId) params.semester_id = semesterId

  const { data } = await apiClient.get(ENDPOINTS.analytics.export, {
    params,
    responseType: 'blob',
  })

  const ext = format === 'pdf' ? 'pdf' : 'xlsx'
  const blob = new Blob([data])
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${type}-report.${ext}`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

/**
 * ExportButton renders two download buttons (PDF and Excel) for an analytics report.
 * Uses apiClient so the auth token is included in the request.
 */
export function ExportButton({ type, semesterId }: ExportButtonProps) {
  const handleExport = (format: 'pdf' | 'xlsx') => {
    downloadExport(format, type, semesterId)
  }

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => handleExport('pdf')}
        aria-label={`Export ${type} report as PDF`}
      >
        <Download className="mr-1 h-4 w-4" />
        PDF
      </Button>
      <Button
        variant="outline"
        size="sm"
        onClick={() => handleExport('xlsx')}
        aria-label={`Export ${type} report as Excel`}
      >
        <Download className="mr-1 h-4 w-4" />
        Excel
      </Button>
    </div>
  )
}
