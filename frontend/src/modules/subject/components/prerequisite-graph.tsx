import * as React from 'react'
import { Link } from '@tanstack/react-router'
import { Badge } from '@/components/ui/badge'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useAllSubjects } from '../hooks/use-subjects'
import type { Subject, Prerequisite } from '../types'

// Topological sort to assign layer depth (BFS from roots)
function assignLayers(subjects: Subject[]): Map<string, number> {
  const inDegree = new Map<string, number>()

  subjects.forEach((s) => inDegree.set(s.id, 0))
  subjects.forEach((s) => {
    s.prerequisites?.forEach(() => {
      inDegree.set(s.id, (inDegree.get(s.id) ?? 0) + 1)
    })
  })

  const layers = new Map<string, number>()
  const queue: string[] = []

  inDegree.forEach((deg, id) => {
    if (deg === 0) queue.push(id)
  })

  let layer = 0
  let currentCount = queue.length
  let nextCount = 0

  while (queue.length > 0) {
    const id = queue.shift()!
    layers.set(id, layer)
    currentCount--

    // Find subjects that have this as prereq
    subjects.forEach((s) => {
      if (s.prerequisites?.some((p) => p.prerequisite_id === id)) {
        const deg = (inDegree.get(s.id) ?? 0) - 1
        inDegree.set(s.id, deg)
        if (deg === 0) {
          queue.push(s.id)
          nextCount++
        }
      }
    })

    if (currentCount === 0) {
      layer++
      currentCount = nextCount
      nextCount = 0
    }
  }

  // Assign remaining (cycles) to last layer
  subjects.forEach((s) => {
    if (!layers.has(s.id)) layers.set(s.id, layer)
  })

  return layers
}

interface SubjectNodeProps {
  subject: Subject
  prerequisites: Prerequisite[]
}

function SubjectNode({ subject, prerequisites }: SubjectNodeProps) {
  const hardCount = prerequisites.filter((p) => p.prerequisite_type === 'hard').length
  const softCount = prerequisites.filter((p) => p.prerequisite_type === 'soft').length

  return (
    <Link
      to="/subjects/$id"
      params={{ id: subject.id }}
      className="block rounded-lg border bg-card p-3 shadow-sm hover:border-primary hover:shadow-md transition-all w-44"
    >
      <p className="font-mono text-xs font-bold text-primary">{subject.code}</p>
      <p className="mt-0.5 text-xs text-foreground line-clamp-2">{subject.name}</p>
      <div className="mt-2 flex gap-1 flex-wrap">
        <Badge variant="secondary" className="text-xs">{subject.credits}cr</Badge>
        {hardCount > 0 && <Badge variant="destructive" className="text-xs">{hardCount} hard</Badge>}
        {softCount > 0 && <Badge variant="outline" className="text-xs">{softCount} soft</Badge>}
      </div>
    </Link>
  )
}

interface PrerequisiteGraphProps {
  /** When provided, highlight only this subject and its direct prereqs */
  focusSubjectId?: string
}

// DAG visualization using layered flex layout — simple, no canvas needed
export function PrerequisiteGraph({ focusSubjectId }: PrerequisiteGraphProps) {
  const { data: subjects = [], isLoading } = useAllSubjects()

  const visibleSubjects = React.useMemo(() => {
    if (!focusSubjectId) return subjects
    const focus = subjects.find((s) => s.id === focusSubjectId)
    if (!focus) return subjects
    const prereqIds = new Set(focus.prerequisites?.map((p) => p.prerequisite_id) ?? [])
    prereqIds.add(focusSubjectId)
    return subjects.filter((s) => prereqIds.has(s.id))
  }, [subjects, focusSubjectId])

  const layers = React.useMemo(() => assignLayers(visibleSubjects), [visibleSubjects])

  // Group subjects by layer
  const layerGroups = React.useMemo(() => {
    const groups = new Map<number, Subject[]>()
    visibleSubjects.forEach((s) => {
      const layer = layers.get(s.id) ?? 0
      const group = groups.get(layer) ?? []
      group.push(s)
      groups.set(layer, group)
    })
    return Array.from(groups.entries()).sort((a, b) => a[0] - b[0])
  }, [visibleSubjects, layers])

  if (isLoading) return <LoadingSpinner />

  if (visibleSubjects.length === 0) {
    return <p className="text-sm text-muted-foreground">No subjects found.</p>
  }

  return (
    <div className="overflow-x-auto">
      <div className="flex gap-8 p-4 min-w-max">
        {layerGroups.map(([layer, layerSubjects]) => (
          <div key={layer} className="flex flex-col gap-3">
            <p className="text-xs text-muted-foreground text-center">Layer {layer + 1}</p>
            {layerSubjects.map((s) => (
              <SubjectNode
                key={s.id}
                subject={s}
                prerequisites={s.prerequisites ?? []}
              />
            ))}
          </div>
        ))}
      </div>
      {/* Legend */}
      <div className="flex gap-4 px-4 pb-2 text-xs text-muted-foreground">
        <span className="flex items-center gap-1">
          <Badge variant="destructive" className="text-xs">hard</Badge>
          = required prerequisite
        </span>
        <span className="flex items-center gap-1">
          <Badge variant="outline" className="text-xs">soft</Badge>
          = recommended
        </span>
      </div>
    </div>
  )
}
