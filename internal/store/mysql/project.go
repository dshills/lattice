package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// ProjectStore implements store.ProjectStore backed by MySQL.
type ProjectStore struct {
	db *sql.DB
}

// NewProjectStore creates a new MySQL-backed ProjectStore.
func NewProjectStore(db *sql.DB) *ProjectStore {
	return &ProjectStore{db: db}
}

// Create inserts a new Project. The store generates id, created_at, and updated_at.
func (s *ProjectStore) Create(ctx context.Context, project *domain.Project) error {
	project.ID = uuid.New().String()
	now := time.Now().UTC()
	project.CreatedAt = now
	project.UpdatedAt = now

	if err := project.Validate(); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO projects (id, name, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		project.ID, project.Name, project.Description, project.CreatedAt, project.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("%w: project name %q already exists", domain.ErrConflict, project.Name)
		}
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

// Get retrieves a Project by ID.
func (s *ProjectStore) Get(ctx context.Context, id string) (*domain.Project, error) {
	var p domain.Project
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &p, nil
}

// Update applies a partial update to an existing Project using SELECT FOR UPDATE.
func (s *ProjectStore) Update(ctx context.Context, id string, params store.ProjectUpdateParams) (*domain.Project, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var p domain.Project
	err = tx.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ? FOR UPDATE`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get project for update: %w", err)
	}

	if params.Name != nil {
		p.Name = *params.Name
	}
	if params.Description != nil {
		p.Description = *params.Description
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}

	p.UpdatedAt = time.Now().UTC()
	_, err = tx.ExecContext(ctx,
		`UPDATE projects SET name=?, description=?, updated_at=? WHERE id=?`,
		p.Name, p.Description, p.UpdatedAt, p.ID,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, fmt.Errorf("%w: project name %q already exists", domain.ErrConflict, p.Name)
		}
		return nil, fmt.Errorf("update project: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update: %w", err)
	}
	return &p, nil
}

// Delete removes a Project. Fails with ErrConflict if work items exist in the project.
func (s *ProjectStore) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		if isFKDeleteError(err) {
			return fmt.Errorf("%w: project contains work items", domain.ErrConflict)
		}
		return fmt.Errorf("delete project: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List returns all projects with their work item counts, ordered by name.
func (s *ProjectStore) List(ctx context.Context) ([]store.ProjectWithCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.description, p.created_at, p.updated_at,
		        COUNT(w.id) AS item_count
		 FROM projects p
		 LEFT JOIN work_items w ON w.project_id = p.id
		 GROUP BY p.id
		 ORDER BY p.name`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var projects []store.ProjectWithCount
	for rows.Next() {
		var pc store.ProjectWithCount
		if err := rows.Scan(&pc.ID, &pc.Name, &pc.Description,
			&pc.CreatedAt, &pc.UpdatedAt, &pc.ItemCount); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	if projects == nil {
		projects = []store.ProjectWithCount{}
	}
	return projects, nil
}

// isFKDeleteError checks if a MySQL error is a FK constraint violation on delete (error 1451).
func isFKDeleteError(err error) bool {
	var mysqlErr *gomysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1451
	}
	return false
}
