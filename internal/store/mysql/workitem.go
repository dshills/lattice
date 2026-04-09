package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"

	"github.com/google/uuid"
)

// WorkItemStore implements store.WorkItemStore backed by MySQL.
type WorkItemStore struct {
	db *sql.DB
}

// NewWorkItemStore creates a new MySQL-backed WorkItemStore.
func NewWorkItemStore(db *sql.DB) *WorkItemStore {
	return &WorkItemStore{db: db}
}

// Create inserts a new WorkItem. The store generates id, state, created_at, and
// updated_at — any caller-supplied values for these fields are ignored.
func (s *WorkItemStore) Create(ctx context.Context, item *domain.WorkItem) error {
	item.ID = uuid.New().String()
	item.State = domain.NotDone
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := item.Validate(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Normalize empty parent_id to nil (no parent).
	if item.ParentID != nil && *item.ParentID == "" {
		item.ParentID = nil
	}

	// Validate parent exists, check depth limit, and check for cycles.
	if item.ParentID != nil {
		exists, err := rowExists(ctx, tx, "SELECT 1 FROM work_items WHERE id = ?", *item.ParentID)
		if err != nil {
			return fmt.Errorf("check parent: %w", err)
		}
		if !exists {
			return fmt.Errorf("%w: parent_id %q does not exist", domain.ErrValidation, *item.ParentID)
		}
		depth, err := ancestorDepth(ctx, tx, *item.ParentID)
		if err != nil {
			return fmt.Errorf("check depth: %w", err)
		}
		if depth+1 > domain.MaxHierarchyDepth {
			return fmt.Errorf("%w: hierarchy depth would exceed %d levels", domain.ErrValidation, domain.MaxHierarchyDepth)
		}
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO work_items (id, project_id, title, description, state, type, parent_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.ProjectID, item.Title, item.Description, string(item.State),
		nullableString(item.Type), nullableStringPtr(item.ParentID),
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert work_item: %w", err)
	}

	if err := insertTags(ctx, tx, item.ID, item.Tags); err != nil {
		return err
	}

	return tx.Commit()
}

// Get retrieves a WorkItem by ID, including its tags and relationships.
func (s *WorkItemStore) Get(ctx context.Context, id string) (*domain.WorkItem, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, title, description, state, type, parent_id, created_at, updated_at
		 FROM work_items WHERE id = ?`, id)

	item, err := scanWorkItem(row)
	if err != nil {
		return nil, err
	}

	tags, err := loadTags(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	item.Tags = tags

	rels, err := loadRelationships(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	item.Relationships = rels

	return item, nil
}

// Update applies a partial update to an existing WorkItem. Only non-nil fields
// in params are changed. State transitions are validated under the row lock to
// prevent TOCTOU races. Returns the updated WorkItem.
func (s *WorkItemStore) Update(ctx context.Context, id string, params store.UpdateParams) (*domain.WorkItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Fetch current state under lock.
	row := tx.QueryRowContext(ctx,
		`SELECT id, project_id, title, description, state, type, parent_id, created_at, updated_at
		 FROM work_items WHERE id = ? FOR UPDATE`, id)

	existing, err := scanWorkItem(row)
	if err != nil {
		return nil, err
	}

	// Validate state transition under the lock (prevents TOCTOU race).
	if params.State != nil {
		if err := domain.ValidateTransition(existing.State, *params.State, params.Override); err != nil {
			return nil, err
		}
		existing.State = *params.State
	}

	// Merge fields: only overwrite non-nil caller values.
	if params.Title != nil {
		existing.Title = *params.Title
	}
	if params.Description != nil {
		existing.Description = *params.Description
	}
	if params.Type != nil {
		existing.Type = *params.Type
	}
	if params.ParentID != nil {
		if *params.ParentID == "" {
			existing.ParentID = nil // explicitly unset parent
		} else {
			existing.ParentID = params.ParentID
		}
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	// Validate parent exists and check for cycles if parent changed.
	if params.ParentID != nil && existing.ParentID != nil {
		exists, err := rowExists(ctx, tx, "SELECT 1 FROM work_items WHERE id = ?", *existing.ParentID)
		if err != nil {
			return nil, fmt.Errorf("check parent: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("%w: parent_id %q does not exist", domain.ErrValidation, *existing.ParentID)
		}
		hasCycle, err := hasCycle(ctx, tx, existing.ID, *existing.ParentID)
		if err != nil {
			return nil, fmt.Errorf("check parent cycle: %w", err)
		}
		if hasCycle {
			return nil, fmt.Errorf("%w: setting parent_id would create a cycle", domain.ErrValidation)
		}
		depth, err := ancestorDepth(ctx, tx, *existing.ParentID)
		if err != nil {
			return nil, fmt.Errorf("check depth: %w", err)
		}
		if depth+1 > domain.MaxHierarchyDepth {
			return nil, fmt.Errorf("%w: hierarchy depth would exceed %d levels", domain.ErrValidation, domain.MaxHierarchyDepth)
		}
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE work_items SET title=?, description=?, state=?, type=?, parent_id=?, updated_at=?
		 WHERE id=?`,
		existing.Title, existing.Description, string(existing.State),
		nullableString(existing.Type), nullableStringPtr(existing.ParentID),
		existing.UpdatedAt, existing.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update work_item: %w", err)
	}

	// Replace tags if caller provided any (including empty slice to clear).
	if params.Tags != nil {
		if _, err := tx.ExecContext(ctx, "DELETE FROM work_item_tags WHERE item_id = ?", id); err != nil {
			return nil, fmt.Errorf("delete tags: %w", err)
		}
		if err := insertTags(ctx, tx, id, params.Tags); err != nil {
			return nil, err
		}
	}

	// Load tags and relationships within the transaction for consistency.
	tags, err := loadTags(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("reload tags: %w", err)
	}
	existing.Tags = tags

	rels, err := loadRelationships(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("reload relationships: %w", err)
	}
	existing.Relationships = rels

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update: %w", err)
	}

	return existing, nil
}

// Delete atomically removes a WorkItem and cascades: removes relationships
// (both directions), nulls children's parent_id, removes tags, then deletes
// the item itself.
func (s *WorkItemStore) Delete(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Verify item exists.
	exists, err := rowExists(ctx, tx, "SELECT 1 FROM work_items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("check existence: %w", err)
	}
	if !exists {
		return domain.ErrNotFound
	}

	// Null out children's parent_id (not covered by ON DELETE CASCADE).
	if _, err := tx.ExecContext(ctx, "UPDATE work_items SET parent_id = NULL WHERE parent_id = ?", id); err != nil {
		return fmt.Errorf("null children parent_id: %w", err)
	}

	// Delete item — tags and relationships are removed by ON DELETE CASCADE.
	if _, err := tx.ExecContext(ctx, "DELETE FROM work_items WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete work_item: %w", err)
	}

	return tx.Commit()
}

// List retrieves a filtered, paginated list of WorkItems.
func (s *WorkItemStore) List(ctx context.Context, filter store.ListFilter) (*store.ListResult, error) {
	page := max(filter.Page, 1)
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 50
	}

	where, args := buildWhereClause(filter)

	// Count total matching items.
	countQuery := "SELECT COUNT(DISTINCT w.id) FROM work_items w" + buildJoins(filter) + where
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count work_items: %w", err)
	}

	// Fetch page of IDs.
	offset := (page - 1) * pageSize
	idQuery := "SELECT DISTINCT w.id, w.created_at FROM work_items w" + buildJoins(filter) + where +
		" ORDER BY w.created_at DESC LIMIT ? OFFSET ?"
	idArgs := append(args, pageSize, offset)

	rows, err := s.db.QueryContext(ctx, idQuery, idArgs...)
	if err != nil {
		return nil, fmt.Errorf("list work_items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		var createdAt time.Time
		if err := rows.Scan(&id, &createdAt); err != nil {
			return nil, fmt.Errorf("scan work_item id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate work_item ids: %w", err)
	}

	if len(ids) == 0 {
		return &store.ListResult{
			Items:    []domain.WorkItem{},
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	// Batch-load items, tags, and relationships in 3 queries total.
	items, err := s.batchLoadItems(ctx, ids)
	if err != nil {
		return nil, err
	}

	return &store.ListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// AncestorDepth uses a recursive CTE to walk the parent_id chain from the
// given parentID and returns the depth (number of ancestors including the
// starting node). Used for the 100-level hierarchy limit.
func (s *WorkItemStore) AncestorDepth(ctx context.Context, parentID string) (int, error) {
	return ancestorDepth(ctx, s.db, parentID)
}

// HasCycle checks whether setting childID's parent to parentID would create a
// cycle. Uses a recursive CTE to walk from parentID up; returns true if childID
// is encountered in the chain.
func (s *WorkItemStore) HasCycle(ctx context.Context, childID, parentID string) (bool, error) {
	return hasCycle(ctx, s.db, childID, parentID)
}

func ancestorDepth(ctx context.Context, q querier, parentID string) (int, error) {
	const query = `
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id, 1 AS depth
			FROM work_items WHERE id = ?
			UNION ALL
			SELECT w.id, w.parent_id, a.depth + 1
			FROM work_items w
			JOIN ancestors a ON w.id = a.parent_id
		)
		SELECT COALESCE(MAX(depth), 0) FROM ancestors`

	var depth int
	if err := q.QueryRowContext(ctx, query, parentID).Scan(&depth); err != nil {
		return 0, fmt.Errorf("ancestor depth: %w", err)
	}
	return depth, nil
}

func hasCycle(ctx context.Context, q querier, childID, parentID string) (bool, error) {
	const query = `
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id
			FROM work_items WHERE id = ?
			UNION ALL
			SELECT w.id, w.parent_id
			FROM work_items w
			JOIN ancestors a ON w.id = a.parent_id
		)
		SELECT COUNT(*) FROM ancestors WHERE id = ?`

	var count int
	if err := q.QueryRowContext(ctx, query, parentID, childID).Scan(&count); err != nil {
		return false, fmt.Errorf("has cycle: %w", err)
	}
	return count > 0, nil
}

// --- helpers ---

// querier is satisfied by both *sql.DB and *sql.Tx.
type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// scanWorkItemFrom scans a WorkItem from any scanner (Row or Rows).
func scanWorkItemFrom(s scanner) (*domain.WorkItem, error) {
	var item domain.WorkItem
	var state string
	var typ sql.NullString
	var parentID sql.NullString

	err := s.Scan(&item.ID, &item.ProjectID, &item.Title, &item.Description, &state,
		&typ, &parentID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan work_item: %w", err)
	}

	item.State = domain.State(state)
	if typ.Valid {
		item.Type = typ.String
	}
	if parentID.Valid {
		item.ParentID = &parentID.String
	}
	return &item, nil
}

func scanWorkItem(row *sql.Row) (*domain.WorkItem, error) {
	return scanWorkItemFrom(row)
}

// batchLoadItems loads full WorkItem structs for a slice of IDs using 3 queries
// total (items, tags, relationships) instead of 3N queries.
func (s *WorkItemStore) batchLoadItems(ctx context.Context, ids []string) ([]domain.WorkItem, error) {
	placeholders := make([]string, len(ids))
	idArgs := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		idArgs[i] = id
	}
	inClause := strings.Join(placeholders, ", ")

	// 1. Load items.
	itemQuery := `SELECT id, project_id, title, description, state, type, parent_id, created_at, updated_at
		FROM work_items WHERE id IN (` + inClause + `)`
	rows, err := s.db.QueryContext(ctx, itemQuery, idArgs...)
	if err != nil {
		return nil, fmt.Errorf("batch load items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	itemMap := make(map[string]*domain.WorkItem, len(ids))
	for rows.Next() {
		item, err := scanWorkItemFrom(rows)
		if err != nil {
			return nil, err
		}
		item.Tags = []string{}
		item.Relationships = []domain.Relationship{}
		itemMap[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate batch items: %w", err)
	}

	// 2. Bulk load tags.
	tagQuery := `SELECT item_id, tag FROM work_item_tags WHERE item_id IN (` + inClause + `) ORDER BY tag`
	tagRows, err := s.db.QueryContext(ctx, tagQuery, idArgs...)
	if err != nil {
		return nil, fmt.Errorf("batch load tags: %w", err)
	}
	defer func() { _ = tagRows.Close() }()
	for tagRows.Next() {
		var itemID, tag string
		if err := tagRows.Scan(&itemID, &tag); err != nil {
			return nil, fmt.Errorf("scan batch tag: %w", err)
		}
		if item, ok := itemMap[itemID]; ok {
			item.Tags = append(item.Tags, tag)
		}
	}
	if err := tagRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate batch tags: %w", err)
	}

	// 3. Bulk load relationships.
	relQuery := `SELECT source_id, id, type, target_id FROM work_item_relationships WHERE source_id IN (` + inClause + `)`
	relRows, err := s.db.QueryContext(ctx, relQuery, idArgs...)
	if err != nil {
		return nil, fmt.Errorf("batch load relationships: %w", err)
	}
	defer func() { _ = relRows.Close() }()
	for relRows.Next() {
		var sourceID string
		var r domain.Relationship
		var relType string
		if err := relRows.Scan(&sourceID, &r.ID, &relType, &r.TargetID); err != nil {
			return nil, fmt.Errorf("scan batch relationship: %w", err)
		}
		r.Type = domain.RelationshipType(relType)
		if item, ok := itemMap[sourceID]; ok {
			item.Relationships = append(item.Relationships, r)
		}
	}
	if err := relRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate batch relationships: %w", err)
	}

	// Assemble in the caller's original ID order to preserve sorting.
	result := make([]domain.WorkItem, 0, len(ids))
	for _, id := range ids {
		if item, ok := itemMap[id]; ok {
			result = append(result, *item)
		}
	}
	return result, nil
}

func loadTags(ctx context.Context, q querier, itemID string) ([]string, error) {
	rows, err := q.QueryContext(ctx,
		"SELECT tag FROM work_item_tags WHERE item_id = ? ORDER BY tag", itemID)
	if err != nil {
		return nil, fmt.Errorf("load tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func loadRelationships(ctx context.Context, q querier, sourceID string) ([]domain.Relationship, error) {
	rows, err := q.QueryContext(ctx,
		"SELECT id, type, target_id FROM work_item_relationships WHERE source_id = ?", sourceID)
	if err != nil {
		return nil, fmt.Errorf("load relationships: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rels := make([]domain.Relationship, 0)
	for rows.Next() {
		var r domain.Relationship
		var relType string
		if err := rows.Scan(&r.ID, &relType, &r.TargetID); err != nil {
			return nil, fmt.Errorf("scan relationship: %w", err)
		}
		r.Type = domain.RelationshipType(relType)
		rels = append(rels, r)
	}
	return rels, rows.Err()
}

func insertTags(ctx context.Context, tx *sql.Tx, itemID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	// Deduplicate tags.
	seen := make(map[string]bool, len(tags))
	deduped := make([]string, 0, len(tags))
	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			deduped = append(deduped, tag)
		}
	}
	query := "INSERT INTO work_item_tags (item_id, tag) VALUES "
	vals := make([]string, 0, len(deduped))
	args := make([]any, 0, len(deduped)*2)
	for _, tag := range deduped {
		vals = append(vals, "(?, ?)")
		args = append(args, itemID, tag)
	}
	query += strings.Join(vals, ", ")
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("insert tags: %w", err)
	}
	return nil
}

func rowExists(ctx context.Context, tx *sql.Tx, query string, args ...any) (bool, error) {
	var one int
	err := tx.QueryRowContext(ctx, query, args...).Scan(&one)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableStringPtr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func buildJoins(filter store.ListFilter) string {
	var joins string
	// Tags use EXISTS subqueries, so no JOIN needed.
	if filter.RelationshipType != nil || filter.RelationshipTarget != nil {
		joins += " JOIN work_item_relationships r ON r.source_id = w.id"
	}
	return joins
}

func buildWhereClause(filter store.ListFilter) (string, []any) {
	var conditions []string
	var args []any

	if filter.State != nil {
		conditions = append(conditions, "w.state = ?")
		args = append(args, string(*filter.State))
	}
	if filter.Type != nil {
		if *filter.Type == "" {
			conditions = append(conditions, "w.type IS NULL")
		} else {
			conditions = append(conditions, "w.type = ?")
			args = append(args, *filter.Type)
		}
	}
	if filter.ParentID != nil {
		conditions = append(conditions, "w.parent_id = ?")
		args = append(args, *filter.ParentID)
	}
	for _, tag := range filter.Tags {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM work_item_tags t2 WHERE t2.item_id = w.id AND t2.tag = ?)")
		args = append(args, tag)
	}
	if filter.RelationshipType != nil {
		conditions = append(conditions, "r.type = ?")
		args = append(args, string(*filter.RelationshipType))
	}
	if filter.RelationshipTarget != nil {
		conditions = append(conditions, "r.target_id = ?")
		args = append(args, *filter.RelationshipTarget)
	}
	blockedSubquery := `EXISTS (
				SELECT 1 FROM work_item_relationships dep
				JOIN work_items tgt ON tgt.id = dep.target_id
				WHERE dep.source_id = w.id AND dep.type = ? AND tgt.state != ?
			)`
	notBlockedSubquery := `NOT EXISTS (
				SELECT 1 FROM work_item_relationships dep
				JOIN work_items tgt ON tgt.id = dep.target_id
				WHERE dep.source_id = w.id AND dep.type = ? AND tgt.state != ?
			)`
	if filter.IsBlocked != nil && *filter.IsBlocked {
		conditions = append(conditions, blockedSubquery)
		args = append(args, string(domain.DependsOn), string(domain.Completed))
	}
	if filter.IsBlocked != nil && !*filter.IsBlocked {
		conditions = append(conditions, notBlockedSubquery)
		args = append(args, string(domain.DependsOn), string(domain.Completed))
	}
	if filter.IsReady != nil && *filter.IsReady {
		conditions = append(conditions, notBlockedSubquery)
		args = append(args, string(domain.DependsOn), string(domain.Completed))
	}
	if filter.IsReady != nil && !*filter.IsReady {
		conditions = append(conditions, blockedSubquery)
		args = append(args, string(domain.DependsOn), string(domain.Completed))
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
