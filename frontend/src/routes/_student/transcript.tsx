import { createFileRoute } from '@tanstack/react-router'
import { Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useMyTranscript, downloadTranscript } from './-hooks/use-my-transcript'

export const Route = createFileRoute('/_student/transcript')({
  component: TranscriptPage,
})

const GRADE_VARIANT = {
  A: 'default',
  B: 'secondary',
  C: 'secondary',
  D: 'outline',
  F: 'destructive',
} as const

function TranscriptPage() {
  const { data, isLoading } = useMyTranscript()

  if (isLoading) return <LoadingSpinner />
  if (!data) return <p className="text-muted-foreground">No transcript available.</p>

  // Group entries by semester_id
  const bySemester = data.entries.reduce<Record<string, typeof data.entries>>((acc, e) => {
    const key = e.semester_id
    acc[key] = acc[key] ? [...acc[key], e] : [e]
    return acc
  }, {})

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">Academic Transcript</h1>
          <p className="text-sm text-muted-foreground">
            {data.student.full_name} · {data.student.student_code}
          </p>
        </div>
        <Button variant="outline" onClick={downloadTranscript}>
          <Download className="mr-2 h-4 w-4" /> Download PDF
        </Button>
      </div>

      {/* Summary */}
      <div className="grid gap-4 sm:grid-cols-3 rounded-lg border p-4">
        <Stat label="Overall GPA" value={data.gpa.toFixed(2)} />
        <Stat label="Total Credits" value={String(data.total_credits)} />
        <Stat label="Passed Credits" value={String(data.passed_credits)} />
      </div>

      {/* Semester groups */}
      {Object.entries(bySemester).map(([semId, entries]) => {
        const semGpa = entries
          .filter((e) => e.grade_numeric != null)
          .reduce((sum, e, _, arr) => sum + (e.grade_numeric ?? 0) / arr.length, 0)
        const semCredits = entries.reduce((s, e) => s + (e.credits ?? 0), 0)
        return (
          <div key={semId}>
            <h2 className="mb-2 text-sm font-semibold text-muted-foreground">
              Semester {semId.slice(0, 8)}…
            </h2>
            <div className="rounded-md border">
              <table className="w-full text-sm">
                <thead className="border-b bg-muted/50">
                  <tr>
                    <th className="px-4 py-2 text-left font-medium">Code</th>
                    <th className="px-4 py-2 text-left font-medium">Subject</th>
                    <th className="px-4 py-2 text-right font-medium">Credits</th>
                    <th className="px-4 py-2 text-right font-medium">Grade</th>
                    <th className="px-4 py-2 text-right font-medium">Letter</th>
                  </tr>
                </thead>
                <tbody>
                  {entries.map((e) => (
                    <tr key={e.enrollment_id} className="border-b last:border-0">
                      <td className="px-4 py-2 font-mono text-xs">{e.subject_code || '—'}</td>
                      <td className="px-4 py-2">{e.subject_name || '—'}</td>
                      <td className="px-4 py-2 text-right">{e.credits ?? '—'}</td>
                      <td className="px-4 py-2 text-right">
                        {e.grade_numeric != null ? e.grade_numeric.toFixed(1) : '—'}
                      </td>
                      <td className="px-4 py-2 text-right">
                        {e.grade_letter ? (
                          <Badge variant={GRADE_VARIANT[e.grade_letter as keyof typeof GRADE_VARIANT] ?? 'outline'}>
                            {e.grade_letter}
                          </Badge>
                        ) : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p className="mt-1 text-right text-xs text-muted-foreground">
              GPA: {semGpa.toFixed(2)} · Credits: {semCredits}
            </p>
          </div>
        )
      })}
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-0.5 text-xl font-bold">{value}</p>
    </div>
  )
}
