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
