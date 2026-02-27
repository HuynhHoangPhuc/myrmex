// Reusable banner for displaying prerequisite conflicts in the offering manager.
// Shows which subjects have missing hard prerequisites and offers an auto-fix button.

import { AlertTriangle } from 'lucide-react'
import type { ConflictDetail } from '../types'

interface ConflictWarningBannerProps {
  conflicts: ConflictDetail[]
  onAddMissing?: (subjectIds: string[]) => void
}

export function ConflictWarningBanner({ conflicts, onAddMissing }: ConflictWarningBannerProps) {
  if (conflicts.length === 0) return null

  const allMissingIds = [...new Set(conflicts.flatMap((c) => c.missing.map((m) => m.id)))]

  return (
    <div className="rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/20 p-3 space-y-2">
      <div className="flex items-center gap-2 text-amber-700 dark:text-amber-400">
        <AlertTriangle className="h-4 w-4 shrink-0" />
        <span className="text-sm font-medium">
          {conflicts.length} subject(s) have missing prerequisites
        </span>
      </div>
      <ul className="text-xs text-amber-600 dark:text-amber-500 space-y-1 pl-6 list-disc">
        {conflicts.map((c) => (
          <li key={c.subject_id}>
            <strong>{c.subject_name}</strong> needs:{' '}
            {c.missing.map((m) => m.code).join(', ')}
          </li>
        ))}
      </ul>
      {onAddMissing && allMissingIds.length > 0 && (
        <button
          onClick={() => onAddMissing(allMissingIds)}
          className="text-xs text-amber-700 dark:text-amber-400 underline hover:no-underline"
        >
          Add {allMissingIds.length} missing prerequisite(s)
        </button>
      )}
    </div>
  )
}
