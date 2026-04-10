package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dshills/lattice/internal/domain"
	"github.com/google/uuid"
)

// MembershipStore implements store.MembershipStore backed by MySQL.
type MembershipStore struct {
	db *sql.DB
}

// NewMembershipStore creates a new MySQL-backed MembershipStore.
func NewMembershipStore(db *sql.DB) *MembershipStore {
	return &MembershipStore{db: db}
}

// Add inserts a new project membership.
func (s *MembershipStore) Add(ctx context.Context, membership *domain.ProjectMembership) error {
	membership.ID = uuid.New().String()
	membership.CreatedAt = time.Now().UTC()

	if !domain.ValidProjectRole(membership.Role) {
		return fmt.Errorf("%w: invalid role %q", domain.ErrInvalidInput, membership.Role)
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO project_memberships (id, project_id, user_id, role, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		membership.ID, membership.ProjectID, membership.UserID, membership.Role, membership.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("%w: user is already a member of this project", domain.ErrConflict)
		}
		if isForeignKeyError(err) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("insert membership: %w", err)
	}
	return nil
}

// Remove deletes a project membership. Prevents removing the last owner.
func (s *MembershipStore) Remove(ctx context.Context, projectID, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Check if the user is an owner
	var role string
	err = tx.QueryRowContext(ctx,
		`SELECT role FROM project_memberships WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrNotFound
		}
		return fmt.Errorf("get membership: %w", err)
	}

	// If removing an owner, ensure they're not the last one
	if domain.ProjectRole(role) == domain.RoleOwner {
		var ownerCount int
		err = tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM project_memberships WHERE project_id = ? AND role = 'owner'`,
			projectID,
		).Scan(&ownerCount)
		if err != nil {
			return fmt.Errorf("count owners: %w", err)
		}
		if ownerCount <= 1 {
			return fmt.Errorf("%w: cannot remove the last owner", domain.ErrForbidden)
		}
	}

	_, err = tx.ExecContext(ctx,
		`DELETE FROM project_memberships WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete membership: %w", err)
	}

	return tx.Commit()
}

// UpdateRole changes a user's role in a project. Prevents demoting the last owner.
func (s *MembershipStore) UpdateRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error {
	if !domain.ValidProjectRole(role) {
		return fmt.Errorf("%w: invalid role %q", domain.ErrInvalidInput, role)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Get current role
	var currentRole string
	err = tx.QueryRowContext(ctx,
		`SELECT role FROM project_memberships WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	).Scan(&currentRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrNotFound
		}
		return fmt.Errorf("get membership: %w", err)
	}

	// If demoting from owner, ensure they're not the last one
	if domain.ProjectRole(currentRole) == domain.RoleOwner && role != domain.RoleOwner {
		var ownerCount int
		err = tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM project_memberships WHERE project_id = ? AND role = 'owner'`,
			projectID,
		).Scan(&ownerCount)
		if err != nil {
			return fmt.Errorf("count owners: %w", err)
		}
		if ownerCount <= 1 {
			return fmt.Errorf("%w: cannot demote the last owner", domain.ErrForbidden)
		}
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE project_memberships SET role = ? WHERE project_id = ? AND user_id = ?`,
		role, projectID, userID,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	return tx.Commit()
}

// ListByProject returns all memberships for a project, ordered by role then email.
func (s *MembershipStore) ListByProject(ctx context.Context, projectID string) ([]domain.ProjectMembership, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.created_at
		 FROM project_memberships pm
		 WHERE pm.project_id = ?
		 ORDER BY FIELD(pm.role, 'owner', 'member', 'viewer'), pm.created_at`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var memberships []domain.ProjectMembership
	for rows.Next() {
		var m domain.ProjectMembership
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan membership: %w", err)
		}
		memberships = append(memberships, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memberships: %w", err)
	}
	if memberships == nil {
		memberships = []domain.ProjectMembership{}
	}
	return memberships, nil
}

// GetRole returns the user's role for a project, or ErrNotFound if not a member.
func (s *MembershipStore) GetRole(ctx context.Context, projectID, userID string) (domain.ProjectRole, error) {
	var role domain.ProjectRole
	err := s.db.QueryRowContext(ctx,
		`SELECT role FROM project_memberships WHERE project_id = ? AND user_id = ?`,
		projectID, userID,
	).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("get role: %w", err)
	}
	return role, nil
}
