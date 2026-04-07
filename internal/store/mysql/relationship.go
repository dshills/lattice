package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	gomysql "github.com/go-sql-driver/mysql"

	"github.com/dshills/lattice/internal/domain"
	"github.com/google/uuid"
)

// RelationshipStore implements store.RelationshipStore backed by MySQL.
type RelationshipStore struct {
	db *sql.DB
}

// NewRelationshipStore creates a new MySQL-backed RelationshipStore.
func NewRelationshipStore(db *sql.DB) *RelationshipStore {
	return &RelationshipStore{db: db}
}

// Add creates a new relationship from ownerID to rel.TargetID.
// Generates a UUID for the relationship ID. Validates that both source and
// target WorkItems exist and that the relationship type is valid.
// Returns ErrValidation on duplicate (source_id, target_id, type).
func (s *RelationshipStore) Add(ctx context.Context, ownerID string, rel *domain.Relationship) error {
	if !domain.ValidRelationshipType(rel.Type) {
		return fmt.Errorf("%w: invalid relationship type %q", domain.ErrInvalidInput, rel.Type)
	}
	if rel.TargetID == "" {
		return fmt.Errorf("%w: target_id is required", domain.ErrInvalidInput)
	}

	rel.ID = uuid.New().String()

	// Rely on FK constraints for referential integrity. Map MySQL errors
	// to domain errors for the caller.
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO work_item_relationships (id, source_id, target_id, type)
		 VALUES (?, ?, ?, ?)`,
		rel.ID, ownerID, rel.TargetID, string(rel.Type),
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("%w: relationship already exists", domain.ErrValidation)
		}
		if isForeignKeyError(err) {
			// Determine which FK failed: source or target.
			if containsFKRef(err, "fk_rel_source") {
				return fmt.Errorf("%w: source work item %q", domain.ErrNotFound, ownerID)
			}
			return fmt.Errorf("%w: target work item %q does not exist", domain.ErrValidation, rel.TargetID)
		}
		return fmt.Errorf("insert relationship: %w", err)
	}
	return nil
}

// Remove deletes a relationship that belongs to the specified WorkItem.
// Returns ErrNotFound if the relationship does not exist or does not belong
// to the specified owner.
func (s *RelationshipStore) Remove(ctx context.Context, ownerID, relID string) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM work_item_relationships WHERE id = ? AND source_id = ?",
		relID, ownerID)
	if err != nil {
		return fmt.Errorf("delete relationship: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: relationship %q for work item %q", domain.ErrNotFound, relID, ownerID)
	}
	return nil
}

// ListByTarget returns all relationships where the given ID is the target.
// Supports reverse lookup (who depends on / blocks this item).
// In the returned Relationships, TargetID contains the source_id (the item
// that owns the outgoing edge), since from the queried target's perspective
// the source is the "other end" of the relationship.
func (s *RelationshipStore) ListByTarget(ctx context.Context, targetID string) ([]domain.Relationship, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, type, source_id FROM work_item_relationships WHERE target_id = ?",
		targetID)
	if err != nil {
		return nil, fmt.Errorf("list by target: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rels := make([]domain.Relationship, 0)
	for rows.Next() {
		var r domain.Relationship
		var relType string
		var sourceID string
		if err := rows.Scan(&r.ID, &relType, &sourceID); err != nil {
			return nil, fmt.Errorf("scan relationship: %w", err)
		}
		r.Type = domain.RelationshipType(relType)
		r.TargetID = sourceID // reverse: source is the "other end"
		rels = append(rels, r)
	}
	return rels, rows.Err()
}

// isDuplicateKeyError checks if a MySQL error is a duplicate key violation (error 1062).
func isDuplicateKeyError(err error) bool {
	var mysqlErr *gomysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

// isForeignKeyError checks if a MySQL error is a foreign key constraint violation (error 1452).
func isForeignKeyError(err error) bool {
	var mysqlErr *gomysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1452
	}
	return false
}

// containsFKRef checks if the error message references a specific foreign key name.
func containsFKRef(err error, fkName string) bool {
	return strings.Contains(err.Error(), fkName)
}
