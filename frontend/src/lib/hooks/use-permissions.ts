import { authStore } from '@/lib/stores/auth-store'
import type { UserRole } from '@/lib/api/types'

// Admin-level roles that can manage the system
const ADMIN_ROLES: UserRole[] = ['super_admin', 'admin']

// Roles that can view/manage their own department's resources
const DEPT_MANAGER_ROLES: UserRole[] = ['super_admin', 'admin', 'dean', 'dept_head']

// Roles that can grade students
const GRADER_ROLES: UserRole[] = ['super_admin', 'admin', 'teacher']

export interface Permissions {
  role: UserRole | null
  departmentId: string | undefined
  // Role checks
  isSuperAdmin: boolean
  isAdmin: boolean       // super_admin | admin
  isDean: boolean
  isDeptHead: boolean
  isTeacher: boolean
  isStudent: boolean
  // Capability checks
  canManageUsers: boolean        // admin, super_admin
  canManageDept: boolean         // admin, super_admin, dept_head (own dept)
  canViewAllDepts: boolean       // admin, super_admin, dean
  canGrade: boolean              // admin, super_admin, teacher
  canDeleteResources: boolean    // admin, super_admin
}

// Returns RBAC permission flags for the current user.
// Read from the stored user object (populated on login/refresh).
export function usePermissions(): Permissions {
  const user = authStore.getUser()
  const role = (user?.role ?? null) as UserRole | null
  const departmentId = user?.department_id

  const hasRole = (...roles: UserRole[]) => role !== null && roles.includes(role)

  return {
    role,
    departmentId,
    isSuperAdmin: hasRole('super_admin'),
    isAdmin: hasRole('super_admin', 'admin'),
    isDean: hasRole('dean'),
    isDeptHead: hasRole('dept_head'),
    isTeacher: hasRole('teacher'),
    isStudent: hasRole('student'),
    canManageUsers: hasRole(...ADMIN_ROLES),
    canManageDept: hasRole(...DEPT_MANAGER_ROLES),
    canViewAllDepts: hasRole('super_admin', 'admin', 'dean'),
    canGrade: hasRole(...GRADER_ROLES),
    canDeleteResources: hasRole(...ADMIN_ROLES),
  }
}
