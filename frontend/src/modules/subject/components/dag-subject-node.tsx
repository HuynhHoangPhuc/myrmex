// Custom React Flow node displaying subject code, name, credits, and department color.
// The `hasConflict` flag switches the border to red (used by conflict detection).

import { useContext } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Badge } from '@/components/ui/badge'
import { getDeptColor } from '../utils/dept-color'
import { DagHoverContext } from './prerequisite-dag'

export interface SubjectNodeData {
  code: string
  name: string
  credits: number
  departmentId: string
  hasConflict?: boolean  // true = red border (conflict detection)
}

export function DagSubjectNode({ data, id }: NodeProps) {
  const d = data as unknown as SubjectNodeData
  // Read highlight state from context (not node data) to avoid React Flow
  // re-processing nodes on hover, which caused infinite blink loops.
  const highlightedIds = useContext(DagHoverContext)
  const isDimmed = highlightedIds !== null && !highlightedIds.has(id)
  const borderColor = d.hasConflict ? '#ef4444' : getDeptColor(d.departmentId)

  return (
    <>
      <Handle type="target" position={Position.Top} />
      <div
        className="rounded-lg border-2 bg-card p-2.5 shadow-sm cursor-pointer transition-opacity"
        style={{
          borderColor,
          width: 170,
          opacity: isDimmed ? 0.25 : 1,
        }}
      >
        <p className="font-mono text-xs font-bold truncate" style={{ color: borderColor }}>
          {d.code}
        </p>
        <p className="text-xs text-foreground line-clamp-2 mt-0.5">{d.name}</p>
        <Badge variant="secondary" className="text-xs mt-1">{d.credits}cr</Badge>
        {d.hasConflict && (
          <p className="text-xs text-destructive mt-1 font-medium">⚠ missing prereq</p>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} />
    </>
  )
}
