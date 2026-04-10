package domain

import "time"

// ProjectRole represents a user's role within a project.
type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RoleMember ProjectRole = "member"
	RoleViewer ProjectRole = "viewer"
)

var validProjectRoles = map[ProjectRole]bool{
	RoleOwner:  true,
	RoleMember: true,
	RoleViewer: true,
}

// ValidProjectRole returns true if the given role is a supported project role.
func ValidProjectRole(r ProjectRole) bool {
	return validProjectRoles[r]
}

// ProjectMembership links a user to a project with a specific role.
type ProjectMembership struct {
	ID        string      `json:"id"`
	ProjectID string      `json:"project_id"`
	UserID    string      `json:"user_id"`
	Role      ProjectRole `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}
