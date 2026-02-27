// Subject module domain types

export type PrerequisiteType = 'hard' | 'soft'

export interface Prerequisite {
  subject_id: string
  prerequisite_id: string
  prerequisite_type: PrerequisiteType
  prerequisite?: SubjectSummary
}

export interface SubjectSummary {
  id: string
  code: string
  name: string
}

export interface Subject {
  id: string
  code: string
  name: string
  credits: number
  description?: string
  department_id: string
  weekly_hours: number
  prerequisites: Prerequisite[]
  created_at: string
  updated_at: string
}

export interface CreateSubjectInput {
  code: string
  name: string
  credits: number
  description?: string
  department_id: string
  weekly_hours: number
}

export interface UpdateSubjectInput extends Partial<CreateSubjectInput> {}

export interface AddPrerequisiteInput {
  prerequisite_id: string
  prerequisite_type: PrerequisiteType
}

export interface SemesterOffering {
  semester_id: string
  subject_id: string
  subject?: SubjectSummary
}

// Full DAG response from GET /api/subjects/dag/full
export interface DAGNode {
  id: string
  code: string
  name: string
  credits: number
  department_id: string
  weekly_hours: number
  is_active: boolean
}

export interface DAGEdge {
  source_id: string  // prerequisite (must be completed first)
  target_id: string  // subject that depends on source
  type: PrerequisiteType
  priority: number
}

export interface FullDAGResponse {
  nodes: DAGNode[]
  edges: DAGEdge[]
}

// Conflict detection response from POST /api/subjects/dag/check-conflicts
export interface MissingPrerequisiteInfo {
  id: string
  name: string
  code: string
  type: PrerequisiteType
}

export interface ConflictDetail {
  subject_id: string
  subject_name: string
  missing: MissingPrerequisiteInfo[]
}

export interface CheckConflictsResponse {
  conflicts: ConflictDetail[]
}
