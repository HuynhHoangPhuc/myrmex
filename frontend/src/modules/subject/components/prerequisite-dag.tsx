// Interactive prerequisite DAG using React Flow + Dagre layout.
// Replaces the old flex-layout PrerequisiteGraph with zoom/pan/minimap support.

import { useMemo, useCallback, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  BackgroundVariant,
  type Node,
  type Edge,
  type NodeMouseHandler,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { LoadingSpinner } from '@/components/shared/loading-spinner'
import { useFullDAG } from '../hooks/use-subjects'
import { DagSubjectNode } from './dag-subject-node'
import { layoutDAG } from '../utils/dag-layout'
import { getDeptColor } from '../utils/dept-color'

const nodeTypes = { subject: DagSubjectNode }

interface PrerequisiteDAGProps {
  focusSubjectId?: string
  /** Set of subject IDs that have missing hard prerequisites (conflict highlighting) */
  conflicts?: Set<string>
}

// Build a map: nodeId → Set of all transitive ancestor IDs (prerequisites).
// An "ancestor" is any node that must be completed before this one.
function buildAncestorMap(edges: Edge[]): Map<string, Set<string>> {
  // Build reverse adjacency: targetId → [sourceIds]
  const reverseAdj = new Map<string, string[]>()
  for (const e of edges) {
    const targets = reverseAdj.get(e.target) ?? []
    targets.push(e.source)
    reverseAdj.set(e.target, targets)
  }

  const cache = new Map<string, Set<string>>()

  function getAncestors(nodeId: string): Set<string> {
    if (cache.has(nodeId)) return cache.get(nodeId)!
    const ancestors = new Set<string>()
    for (const src of reverseAdj.get(nodeId) ?? []) {
      ancestors.add(src)
      for (const a of getAncestors(src)) ancestors.add(a)
    }
    cache.set(nodeId, ancestors)
    return ancestors
  }

  const allIds = new Set([
    ...edges.map((e) => e.source),
    ...edges.map((e) => e.target),
  ])
  for (const id of allIds) getAncestors(id)

  return cache
}

export function PrerequisiteDAG({ focusSubjectId, conflicts }: PrerequisiteDAGProps) {
  const { data: dag, isLoading } = useFullDAG()
  const navigate = useNavigate()
  const [hoveredId, setHoveredId] = useState<string | null>(null)

  // Convert raw DAG edges to React Flow edges (before layout, for ancestor map).
  const rawEdges = useMemo((): Edge[] => {
    if (!dag) return []
    return dag.edges.map((e) => ({
      id: `${e.source_id}-${e.target_id}`,
      source: e.source_id,
      target: e.target_id,
      animated: e.type === 'hard',
      style: e.type === 'soft' ? { strokeDasharray: '5 5' } : undefined,
    }))
  }, [dag])

  const ancestorMap = useMemo(() => buildAncestorMap(rawEdges), [rawEdges])

  // Determine which node IDs are visible in focus mode.
  const visibleIds = useMemo((): Set<string> | null => {
    if (!focusSubjectId || !dag) return null
    const ancestors = ancestorMap.get(focusSubjectId) ?? new Set()
    return new Set([focusSubjectId, ...ancestors])
  }, [focusSubjectId, dag, ancestorMap])

  // Determine highlighted set when hovering (hovered node + all its ancestors).
  const highlightedIds = useMemo((): Set<string> | null => {
    if (!hoveredId) return null
    const ancestors = ancestorMap.get(hoveredId) ?? new Set()
    return new Set([hoveredId, ...ancestors])
  }, [hoveredId, ancestorMap])

  const { nodes, edges } = useMemo((): { nodes: Node[]; edges: Edge[] } => {
    if (!dag) return { nodes: [], edges: [] }

    // Filter nodes for focus mode.
    const filteredNodes = visibleIds
      ? dag.nodes.filter((n) => visibleIds.has(n.id))
      : dag.nodes

    // Filter edges to only show edges between visible nodes.
    const visibleNodeIds = new Set(filteredNodes.map((n) => n.id))
    const filteredEdges = rawEdges.filter(
      (e) => visibleNodeIds.has(e.source) && visibleNodeIds.has(e.target),
    )

    const rfNodes: Node[] = filteredNodes.map((n) => ({
      id: n.id,
      type: 'subject',
      position: { x: 0, y: 0 }, // overwritten by layoutDAG
      data: {
        code: n.code,
        name: n.name,
        credits: n.credits,
        departmentId: n.department_id,
        hasConflict: conflicts?.has(n.id) ?? false,
        // Dim nodes not in the hover ancestor chain.
        highlighted: highlightedIds ? highlightedIds.has(n.id) : undefined,
      },
    }))

    return layoutDAG(rfNodes, filteredEdges)
  }, [dag, visibleIds, rawEdges, conflicts, highlightedIds])

  const onNodeClick: NodeMouseHandler = useCallback(
    (_event, node) => {
      void navigate({ to: '/subjects/$id', params: { id: node.id } })
    },
    [navigate],
  )

  if (isLoading) return <LoadingSpinner />

  if (nodes.length === 0) {
    return <p className="text-sm text-muted-foreground p-4">No subjects found.</p>
  }

  return (
    <div style={{ height: focusSubjectId ? 420 : 620 }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodeClick={onNodeClick}
        onNodeMouseEnter={(_event, node) => setHoveredId(node.id)}
        onNodeMouseLeave={() => setHoveredId(null)}
        fitView
        minZoom={0.2}
        maxZoom={2}
      >
        <Controls />
        <MiniMap nodeColor={(n) => getDeptColor((n.data as { departmentId?: string }).departmentId ?? '')} />
        <Background variant={BackgroundVariant.Dots} />
      </ReactFlow>
      {/* Edge type legend */}
      <div className="flex gap-4 px-3 py-2 text-xs text-muted-foreground border-t bg-muted/30">
        <span className="flex items-center gap-1.5">
          <span className="inline-block w-6 border-t-2 border-foreground" />
          Hard (required)
        </span>
        <span className="flex items-center gap-1.5">
          <span className="inline-block w-6 border-t-2 border-dashed border-foreground/50" />
          Soft (recommended)
        </span>
      </div>
    </div>
  )
}
