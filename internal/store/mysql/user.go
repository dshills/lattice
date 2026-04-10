package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dshills/lattice/internal/domain"
	"github.com/google/uuid"
)

// UserStore implements store.UserStore backed by MySQL.
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new MySQL-backed UserStore.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// Create inserts a new User. The passwordHash should be pre-hashed with bcrypt.
func (s *UserStore) Create(ctx context.Context, user *domain.User, passwordHash string) error {
	user.ID = uuid.New().String()
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.PasswordHash = passwordHash

	if err := user.Validate(); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, display_name, password_hash, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrDuplicateEmail
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// GetByID retrieves a User by ID.
func (s *UserStore) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, display_name, password_hash, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

// GetByEmail retrieves a User by email address.
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, display_name, password_hash, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// UpdateDisplayName updates only the user's display name.
func (s *UserStore) UpdateDisplayName(ctx context.Context, id, displayName string) (*domain.User, error) {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE users SET display_name = ?, updated_at = ? WHERE id = ?`,
		displayName, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update display name: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return nil, domain.ErrNotFound
	}
	return s.GetByID(ctx, id)
}

// UpdatePassword updates the user's password hash.
func (s *UserStore) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		passwordHash, now, id,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
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

// Delete removes a User by ID.
func (s *UserStore) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
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
