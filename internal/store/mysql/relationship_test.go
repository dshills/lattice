package mysql_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dshills/lattice/internal/domain"
	mysqlstore "github.com/dshills/lattice/internal/store/mysql"
)

func TestRelationshipAdd(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	source := newItem("Source")
	require.NoError(t, ws.Create(ctx, source))
	target := newItem("Target")
	require.NoError(t, ws.Create(ctx, target))

	rel := &domain.Relationship{
		Type:     domain.DependsOn,
		TargetID: target.ID,
	}
	err := rs.Add(ctx, source.ID, rel)
	require.NoError(t, err)
	assert.NotEmpty(t, rel.ID)

	// Verify via Get.
	got, err := ws.Get(ctx, testProjectID, source.ID)
	require.NoError(t, err)
	require.Len(t, got.Relationships, 1)
	assert.Equal(t, domain.DependsOn, got.Relationships[0].Type)
	assert.Equal(t, target.ID, got.Relationships[0].TargetID)
}

func TestRelationshipAddDuplicate(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	source := newItem("Source")
	require.NoError(t, ws.Create(ctx, source))
	target := newItem("Target")
	require.NoError(t, ws.Create(ctx, target))

	rel := &domain.Relationship{Type: domain.Blocks, TargetID: target.ID}
	require.NoError(t, rs.Add(ctx, source.ID, rel))

	// Same source+target+type should fail.
	rel2 := &domain.Relationship{Type: domain.Blocks, TargetID: target.ID}
	err := rs.Add(ctx, source.ID, rel2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestRelationshipAddMissingSource(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	target := newItem("Target")
	require.NoError(t, ws.Create(ctx, target))

	rel := &domain.Relationship{Type: domain.DependsOn, TargetID: target.ID}
	err := rs.Add(ctx, "nonexistent", rel)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRelationshipAddMissingTarget(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	source := newItem("Source")
	require.NoError(t, ws.Create(ctx, source))

	rel := &domain.Relationship{Type: domain.DependsOn, TargetID: "nonexistent"}
	err := rs.Add(ctx, source.ID, rel)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestRelationshipAddInvalidType(t *testing.T) {
	db := testDB(t)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	rel := &domain.Relationship{Type: "invalid_type", TargetID: "some-id"}
	err := rs.Add(ctx, "some-owner", rel)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestRelationshipRemove(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	source := newItem("Source")
	require.NoError(t, ws.Create(ctx, source))
	target := newItem("Target")
	require.NoError(t, ws.Create(ctx, target))

	rel := &domain.Relationship{Type: domain.RelatesTo, TargetID: target.ID}
	require.NoError(t, rs.Add(ctx, source.ID, rel))

	err := rs.Remove(ctx, source.ID, rel.ID)
	require.NoError(t, err)

	// Verify removed.
	got, err := ws.Get(ctx, testProjectID, source.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Relationships)
}

func TestRelationshipRemoveNotFound(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	source := newItem("Source")
	require.NoError(t, ws.Create(ctx, source))

	err := rs.Remove(ctx, source.ID, "nonexistent-rel")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRelationshipRemoveWrongOwner(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	a := newItem("A")
	require.NoError(t, ws.Create(ctx, a))
	b := newItem("B")
	require.NoError(t, ws.Create(ctx, b))

	rel := &domain.Relationship{Type: domain.DependsOn, TargetID: b.ID}
	require.NoError(t, rs.Add(ctx, a.ID, rel))

	// Try to remove using wrong owner.
	err := rs.Remove(ctx, b.ID, rel.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRelationshipListByTarget(t *testing.T) {
	db := testDB(t)
	ws := mysqlstore.NewWorkItemStore(db)
	rs := mysqlstore.NewRelationshipStore(db)
	ctx := context.Background()

	target := newItem("Target")
	require.NoError(t, ws.Create(ctx, target))

	a := newItem("A")
	require.NoError(t, ws.Create(ctx, a))
	b := newItem("B")
	require.NoError(t, ws.Create(ctx, b))

	require.NoError(t, rs.Add(ctx, a.ID, &domain.Relationship{Type: domain.DependsOn, TargetID: target.ID}))
	require.NoError(t, rs.Add(ctx, b.ID, &domain.Relationship{Type: domain.Blocks, TargetID: target.ID}))

	rels, err := rs.ListByTarget(ctx, target.ID)
	require.NoError(t, err)
	assert.Len(t, rels, 2)
}
