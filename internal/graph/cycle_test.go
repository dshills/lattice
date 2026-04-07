package graph_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/graph"
	mysqlstore "github.com/dshills/lattice/internal/store/mysql"
)

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

	ctx := context.Background()
	require.NoError(t, mysqlstore.MigrateUp(ctx, db, "../../../migrations"))

	for _, table := range []string{"work_item_relationships", "work_item_tags", "work_items"} {
		_, err := db.ExecContext(ctx, "DELETE FROM "+table)
		require.NoError(t, err)
	}
	return db
}

func createItem(t *testing.T, db *sql.DB, title string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO work_items (id, title, description, state, created_at, updated_at)
		 VALUES (?, ?, '', 'NotDone', NOW(3), NOW(3))`, id, title)
	require.NoError(t, err)
	return id
}

func addEdge(t *testing.T, db *sql.DB, sourceID, targetID string, relType domain.RelationshipType) {
	t.Helper()
	id := uuid.New().String()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO work_item_relationships (id, source_id, target_id, type)
		 VALUES (?, ?, ?, ?)`, id, sourceID, targetID, string(relType))
	require.NoError(t, err)
}

func TestDetectCycles_NoCycles(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	a := createItem(t, db, "A")
	b := createItem(t, db, "B")
	addEdge(t, db, a, b, domain.DependsOn)

	detector := graph.NewCycleDetector(db)
	cycles, err := detector.DetectCycles(ctx, a)
	require.NoError(t, err)
	assert.Empty(t, cycles)
}

func TestDetectCycles_Simple2Node(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	a := createItem(t, db, "A")
	b := createItem(t, db, "B")
	addEdge(t, db, a, b, domain.DependsOn)
	addEdge(t, db, b, a, domain.DependsOn)

	detector := graph.NewCycleDetector(db)
	cycles, err := detector.DetectCycles(ctx, a)
	require.NoError(t, err)
	assert.Len(t, cycles, 1)
	assert.Len(t, cycles[0], 2) // [A, B]
}

func TestDetectCycles_3NodeCycle(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	a := createItem(t, db, "A")
	b := createItem(t, db, "B")
	c := createItem(t, db, "C")
	addEdge(t, db, a, b, domain.DependsOn)
	addEdge(t, db, b, c, domain.Blocks)
	addEdge(t, db, c, a, domain.DependsOn)

	detector := graph.NewCycleDetector(db)
	cycles, err := detector.DetectCycles(ctx, a)
	require.NoError(t, err)
	assert.Len(t, cycles, 1)
	assert.Len(t, cycles[0], 3) // [A, B, C]
}

func TestDetectCycles_MixedEdgeTypes(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	a := createItem(t, db, "A")
	b := createItem(t, db, "B")
	addEdge(t, db, a, b, domain.Blocks)
	addEdge(t, db, b, a, domain.DependsOn)

	detector := graph.NewCycleDetector(db)
	cycles, err := detector.DetectCycles(ctx, a)
	require.NoError(t, err)
	assert.Len(t, cycles, 1)
}

func TestDetectCycles_MultiplesCycles(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	a := createItem(t, db, "A")
	b := createItem(t, db, "B")
	c := createItem(t, db, "C")
	// Cycle 1: A -> B -> A
	addEdge(t, db, a, b, domain.DependsOn)
	addEdge(t, db, b, a, domain.DependsOn)
	// Cycle 2: A -> C -> A
	addEdge(t, db, a, c, domain.DependsOn)
	addEdge(t, db, c, a, domain.DependsOn)

	detector := graph.NewCycleDetector(db)
	cycles, err := detector.DetectCycles(ctx, a)
	require.NoError(t, err)
	assert.Len(t, cycles, 2)
}

func TestDetectCycles_IgnoresRelatesToAndDuplicateOf(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	a := createItem(t, db, "A")
	b := createItem(t, db, "B")
	addEdge(t, db, a, b, domain.RelatesTo)
	addEdge(t, db, b, a, domain.DuplicateOf)

	detector := graph.NewCycleDetector(db)
	cycles, err := detector.DetectCycles(ctx, a)
	require.NoError(t, err)
	assert.Empty(t, cycles)
}

func TestDetectCycles_NoEdges(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	a := createItem(t, db, "A")

	detector := graph.NewCycleDetector(db)
	cycles, err := detector.DetectCycles(ctx, a)
	require.NoError(t, err)
	assert.Empty(t, cycles)
}
