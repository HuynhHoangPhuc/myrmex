package valueobject

import "fmt"

type Role string

const (
	RoleSuperAdmin Role = "super_admin" // unrestricted access
	RoleAdmin      Role = "admin"
	RoleDean       Role = "dean"      // read-all, edit own faculty
	RoleDeptHead   Role = "dept_head" // CRUD teachers/subjects in own dept
	RoleManager    Role = "manager"
	RoleViewer     Role = "viewer"
	RoleStudent    Role = "student"
	RoleTeacher    Role = "teacher" // grade subjects they're assigned to
)

func (r Role) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleAdmin, RoleDean, RoleDeptHead,
		RoleManager, RoleViewer, RoleStudent, RoleTeacher:
		return true
	}
	return false
}

// HasDeptScope returns true for roles that require department-level enforcement.
func (r Role) HasDeptScope() bool {
	return r == RoleDeptHead || r == RoleTeacher
}

// BypassesScope returns true for roles that skip department scope checks.
func (r Role) BypassesScope() bool {
	return r == RoleSuperAdmin || r == RoleAdmin || r == RoleDean || r == "service"
}

func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return r, nil
}
