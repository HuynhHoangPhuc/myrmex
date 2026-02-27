// Prerequisite chip: shows subject code as a navigation link.
// On hover (desktop) or long-press (mobile), shows a floating card
// with the subject's basic info (code, name, prerequisite type).

import { Link } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { PrereqInfo } from '../hooks/use-subjects'

interface PrereqChipProps {
  prereq: PrereqInfo
}

export function PrereqChip({ prereq }: PrereqChipProps) {
  const isHard = prereq.type === 'hard'

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Link
          to="/subjects/$id"
          params={{ id: prereq.id }}
          className="font-mono text-xs font-medium text-primary hover:underline"
          onClick={(e) => e.stopPropagation()}
        >
          {prereq.code}
        </Link>
      </TooltipTrigger>
      <TooltipContent
        side="top"
        className="rounded-lg border bg-popover px-3 py-2.5 text-popover-foreground shadow-md"
      >
        <p className="font-mono text-sm font-bold text-primary">{prereq.code}</p>
        <p className="mt-0.5 max-w-48 text-xs text-muted-foreground">{prereq.name}</p>
        <Badge
          variant={isHard ? 'destructive' : 'outline'}
          className="mt-1.5 text-xs"
        >
          {isHard ? 'Hard' : 'Soft'} prerequisite
        </Badge>
      </TooltipContent>
    </Tooltip>
  )
}
