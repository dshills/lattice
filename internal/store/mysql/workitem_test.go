package mysql_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"
	mysqlstore "github.com/dshills/lattice/internal/store/mysql"
)

// testDB opens a connection using LATTICE_TEST_DSN and runs migrations.
// Skips if the env var is not set.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("LATTICE_TEST_DSN")
	if dsn == "" {
		t.Skip("LATTICE_TEST_DSN not set; skipping integration test")
	}
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, db.Ping())

	// Run migrations.
	ctx := context.Background()
	require.NoError(t, mysqlstore.MigrateUp(ctx, db, "../../../migrations"))

	// Clean tables before each test.
	for _, table := range []string{"work_item_relationships", "work_item_tags", "work_items"} {
		_, err := db.ExecContext(ctx, "DELETE FROM "+table)
		require.NoError(t, err)
	}

	return db
}

func newItem(title string) *domain.WorkItem {
	return &domain.WorkItem{
		ProjectID:   domain.DefaultProjectID,
		Title:       title,
		Description: "test description",
	}
}

const testProjectID = domain.DefaultProjectID

func TestCreate(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Create Test")
	item.Tags = []string{"alpha", "beta"}
	item.Type = "task"

	err := s.Create(ctx, item)
	require.NoError(t, err)

	assert.NotEmpty(t, item.ID)
	assert.Equal(t, domain.NotDone, item.State)
	assert.False(t, item.CreatedAt.IsZero())
	assert.False(t, item.UpdatedAt.IsZero())
}

func TestCreateValidatesParent(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Orphan")
	badParent := "nonexistent-id"
	item.ParentID = &badParent

	err := s.Create(ctx, item)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestGet(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Get Test")
	item.Tags = []string{"gamma"}
	require.NoError(t, s.Create(ctx, item))

	got, err := s.Get(ctx, testProjectID, item.ID)
	require.NoError(t, err)

	assert.Equal(t, item.ID, got.ID)
	assert.Equal(t, "Get Test", got.Title)
	assert.Equal(t, []string{"gamma"}, got.Tags)
	assert.Equal(t, domain.NotDone, got.State)
}

func TestGetNotFound(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	_, err := s.Get(ctx, testProjectID, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUpdate(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Before Update")
	item.Tags = []string{"old"}
	require.NoError(t, s.Create(ctx, item))

	title := "After Update"
	state := domain.InProgress
	got, err := s.Update(ctx, testProjectID, item.ID, store.UpdateParams{
		Title: &title,
		State: &state,
		Tags:  []string{"new1", "new2"},
	})
	require.NoError(t, err)

	assert.Equal(t, "After Update", got.Title)
	assert.Equal(t, domain.InProgress, got.State)
	assert.Equal(t, []string{"new1", "new2"}, got.Tags)
	assert.True(t, got.UpdatedAt.After(item.UpdatedAt) || got.UpdatedAt.Equal(item.UpdatedAt))
}

func TestUpdateNotFound(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	title := "x"
	_, err := s.Update(ctx, testProjectID, "nonexistent", store.UpdateParams{Title: &title})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDelete(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Delete Me")
	item.Tags = []string{"doomed"}
	require.NoError(t, s.Create(ctx, item))

	err := s.Delete(ctx, testProjectID, item.ID)
	require.NoError(t, err)

	_, err = s.Get(ctx, testProjectID, item.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeleteNotFound(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	err := s.Delete(ctx, testProjectID, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeleteNullsChildren(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	parent := newItem("Parent")
	require.NoError(t, s.Create(ctx, parent))

	child := newItem("Child")
	child.ParentID = &parent.ID
	require.NoError(t, s.Create(ctx, child))

	require.NoError(t, s.Delete(ctx, testProjectID, parent.ID))

	got, err := s.Get(ctx, testProjectID, child.ID)
	require.NoError(t, err)
	assert.Nil(t, got.ParentID)
}

func TestListBasic(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, s.Create(ctx, newItem("Item")))
	}

	result, err := s.List(ctx, store.ListFilter{ProjectID: testProjectID, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Items, 3)
}

func TestListFilterByState(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Progressing")
	require.NoError(t, s.Create(ctx, item))

	state := domain.InProgress
	_, err := s.Update(ctx, testProjectID, item.ID, store.UpdateParams{State: &state})
	require.NoError(t, err)

	require.NoError(t, s.Create(ctx, newItem("Still NotDone")))

	filterState := domain.InProgress
	result, err := s.List(ctx, store.ListFilter{ProjectID: testProjectID, State: &filterState, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, domain.InProgress, result.Items[0].State)
}

func TestListFilterByType(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Typed")
	item.Type = "bug"
	require.NoError(t, s.Create(ctx, item))

	item2 := newItem("Other")
	item2.Type = "feature"
	require.NoError(t, s.Create(ctx, item2))

	typ := "bug"
	result, err := s.List(ctx, store.ListFilter{ProjectID: testProjectID, Type: &typ, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}

func TestListFilterByTags(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Tagged")
	item.Tags = []string{"urgent", "backend"}
	require.NoError(t, s.Create(ctx, item))

	item2 := newItem("Other Tags")
	item2.Tags = []string{"urgent", "frontend"}
	require.NoError(t, s.Create(ctx, item2))

	// AND logic: both tags must match
	result, err := s.List(ctx, store.ListFilter{ProjectID: testProjectID, Tags: []string{"urgent", "backend"}, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}

func TestListPagination(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, s.Create(ctx, newItem("Page Item")))
	}

	result, err := s.List(ctx, store.ListFilter{ProjectID: testProjectID, Page: 1, PageSize: 2})
	require.NoError(t, err)
	assert.Equal(t, 5, result.Total)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.PageSize)

	result2, err := s.List(ctx, store.ListFilter{ProjectID: testProjectID, Page: 3, PageSize: 2})
	require.NoError(t, err)
	assert.Len(t, result2.Items, 1) // last page
}

func TestAncestorDepth(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	// Create a chain: root -> mid -> leaf
	root := newItem("Root")
	require.NoError(t, s.Create(ctx, root))

	mid := newItem("Mid")
	mid.ParentID = &root.ID
	require.NoError(t, s.Create(ctx, mid))

	leaf := newItem("Leaf")
	leaf.ParentID = &mid.ID
	require.NoError(t, s.Create(ctx, leaf))

	depth, err := s.AncestorDepth(ctx, leaf.ID)
	require.NoError(t, err)
	// leaf itself counts, then mid, then root = 3 nodes walked
	// But AncestorDepth starts from leaf.parent_id in real usage.
	// Here we pass leaf.ID, so: leaf(1) -> mid(2) -> root(3) = 3
	assert.Equal(t, 3, depth)

	depth, err = s.AncestorDepth(ctx, mid.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, depth) // mid -> root

	depth, err = s.AncestorDepth(ctx, root.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, depth) // root only
}

func TestHasCycle(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	// root -> mid -> leaf
	root := newItem("Root")
	require.NoError(t, s.Create(ctx, root))

	mid := newItem("Mid")
	mid.ParentID = &root.ID
	require.NoError(t, s.Create(ctx, mid))

	leaf := newItem("Leaf")
	leaf.ParentID = &mid.ID
	require.NoError(t, s.Create(ctx, leaf))

	// Setting root's parent to leaf would create a cycle.
	hasCycle, err := s.HasCycle(ctx, root.ID, leaf.ID)
	require.NoError(t, err)
	assert.True(t, hasCycle)

	// Setting root's parent to some unrelated item should not cycle.
	other := newItem("Other")
	require.NoError(t, s.Create(ctx, other))

	hasCycle, err = s.HasCycle(ctx, root.ID, other.ID)
	require.NoError(t, err)
	assert.False(t, hasCycle)
}

func TestListFilterByParent(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	parent := newItem("Parent")
	require.NoError(t, s.Create(ctx, parent))

	child1 := newItem("Child1")
	child1.ParentID = &parent.ID
	require.NoError(t, s.Create(ctx, child1))

	child2 := newItem("Child2")
	child2.ParentID = &parent.ID
	require.NoError(t, s.Create(ctx, child2))

	require.NoError(t, s.Create(ctx, newItem("Orphan")))

	result, err := s.List(ctx, store.ListFilter{ProjectID: testProjectID, ParentID: &parent.ID, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
}

func TestUpdateClearsTags(t *testing.T) {
	db := testDB(t)
	s := mysqlstore.NewWorkItemStore(db)
	ctx := context.Background()

	item := newItem("Tagged")
	item.Tags = []string{"a", "b"}
	require.NoError(t, s.Create(ctx, item))

	_, err := s.Update(ctx, testProjectID, item.ID, store.UpdateParams{Tags: []string{}})
	require.NoError(t, err)

	got, err := s.Get(ctx, testProjectID, item.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Tags)
}
