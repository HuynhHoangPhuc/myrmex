package valueobject

import "fmt"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleViewer  Role = "viewer"
	RoleStudent Role = "student"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleManager, RoleViewer, RoleStudent:
		return true
	}
	return false
}

func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return r, nil
}
