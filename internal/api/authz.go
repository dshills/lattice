package api

import (
	"context"

	"github.com/dshills/lattice/internal/domain"
)

// requireWriteAccess checks that the user has at least member role (member or owner).
func requireWriteAccess(ctx context.Context) error {
	role := ProjectRoleFromContext(ctx)
	if role == domain.RoleOwner || role == domain.RoleMember {
		return nil
	}
	return domain.ErrForbidden
}

// requireOwner checks that the user has the owner role.
func requireOwner(ctx context.Context) error {
	if ProjectRoleFromContext(ctx) == domain.RoleOwner {
		return nil
	}
	return domain.ErrForbidden
}
