package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dshills/lattice/internal/api"
	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/graph"
	"github.com/dshills/lattice/internal/store"
	mysqlstore "github.com/dshills/lattice/internal/store/mysql"
)

func integrationServer(t *testing.T) *httptest.Server {
	t.Helper()
	dsn := os.Getenv("LATTICE_TEST_DSN")
	if dsn == "" {
		t.Skip("LATTICE_TEST_DSN not set; skipping integration test")
	}

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Ping())

	require.NoError(t, mysqlstore.MigrateUp(context.Background(), db, "../../migrations"))

	// Clean tables (order matters for FK constraints).
	for _, table := range []string{"work_item_relationships", "work_item_tags", "work_items"} {
		_, err := db.Exec("DELETE FROM " + table)
		require.NoError(t, err)
	}
	// Delete non-default projects; keep the default project for backward compat.
	_, err = db.Exec("DELETE FROM projects WHERE id != ?", domain.DefaultProjectID)
	require.NoError(t, err)

	h := &api.Handler{
		Projects:      mysqlstore.NewProjectStore(db),
		WorkItems:     mysqlstore.NewWorkItemStore(db),
		Relationships: mysqlstore.NewRelationshipStore(db),
		Cycles:        graph.NewCycleDetector(db),
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	handler := api.RoleMiddleware(api.JSONContentType(mux))

	return httptest.NewServer(handler)
}

func postJSON(t *testing.T, url string, body any, headers ...string) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func patchJSON(t *testing.T, url string, body any, headers ...string) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func getJSON(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	return resp
}

func deleteReq(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	var v T
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&v))
	return v
}

const integrationProjectPrefix = "/projects/" + domain.DefaultProjectID

// --- Integration Tests ---

func TestIntegration_FullLifecycle(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// Create.
	resp := postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{
		"title": "Lifecycle Item", "description": "test", "tags": []string{"v1"},
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	item := decodeJSON[domain.WorkItem](t, resp)
	assert.Equal(t, "Lifecycle Item", item.Title)
	assert.Equal(t, domain.NotDone, item.State)
	assert.NotEmpty(t, item.ID)

	// Get.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	got := decodeJSON[domain.WorkItem](t, resp)
	assert.Equal(t, item.ID, got.ID)
	assert.Equal(t, []string{"v1"}, got.Tags)

	// Update state to InProgress.
	state := "InProgress"
	resp = patchJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID, map[string]any{"state": state})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	updated := decodeJSON[domain.WorkItem](t, resp)
	assert.Equal(t, domain.InProgress, updated.State)

	// List.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?page=1&page_size=10")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	list := decodeJSON[store.ListResult](t, resp)
	assert.Equal(t, 1, list.Total)

	// Delete.
	resp = deleteReq(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify gone.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestIntegration_StateTransitions(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// Create.
	resp := postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "State Test"})
	item := decodeJSON[domain.WorkItem](t, resp)

	// Skip forward (NotDone -> Completed) should fail.
	resp = patchJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID, map[string]any{"state": "Completed"})
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	_ = resp.Body.Close()

	// Forward to InProgress.
	resp = patchJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID, map[string]any{"state": "InProgress"})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Backward without override should fail.
	resp = patchJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID, map[string]any{"state": "NotDone"})
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	_ = resp.Body.Close()

	// Backward with override but no admin should fail.
	resp = patchJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID, map[string]any{"state": "NotDone", "override": true})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	_ = resp.Body.Close()

	// Backward with override + admin should succeed.
	resp = patchJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+item.ID,
		map[string]any{"state": "NotDone", "override": true},
		"X-Role", "admin")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	result := decodeJSON[domain.WorkItem](t, resp)
	assert.Equal(t, domain.NotDone, result.State)
}

func TestIntegration_RelationshipCRUD(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// Create two items.
	resp := postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Source"})
	source := decodeJSON[domain.WorkItem](t, resp)
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Target"})
	target := decodeJSON[domain.WorkItem](t, resp)

	// Add relationship.
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+source.ID+"/relationships", map[string]any{
		"type": "depends_on", "target_id": target.ID,
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	rel := decodeJSON[domain.Relationship](t, resp)
	assert.NotEmpty(t, rel.ID)

	// Verify via Get.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+source.ID)
	got := decodeJSON[domain.WorkItem](t, resp)
	require.Len(t, got.Relationships, 1)
	assert.Equal(t, target.ID, got.Relationships[0].TargetID)

	// Remove relationship.
	resp = deleteReq(t, srv.URL+integrationProjectPrefix+"/workitems/"+source.ID+"/relationships/"+rel.ID)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify removed.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+source.ID)
	got = decodeJSON[domain.WorkItem](t, resp)
	assert.Empty(t, got.Relationships)
}

func TestIntegration_CascadeDelete(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// Create parent + child.
	resp := postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Parent"})
	parent := decodeJSON[domain.WorkItem](t, resp)
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Child", "parent_id": parent.ID})
	child := decodeJSON[domain.WorkItem](t, resp)

	// Add relationship from parent to another item.
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Related"})
	related := decodeJSON[domain.WorkItem](t, resp)
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+parent.ID+"/relationships", map[string]any{
		"type": "blocks", "target_id": related.ID,
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	_ = resp.Body.Close()

	// Delete parent.
	resp = deleteReq(t, srv.URL+integrationProjectPrefix+"/workitems/"+parent.ID)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Child should have nil parent_id.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+child.ID)
	childGot := decodeJSON[domain.WorkItem](t, resp)
	assert.Nil(t, childGot.ParentID)
}

func TestIntegration_CycleDetection(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// Create A -> B -> A cycle.
	resp := postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "A"})
	a := decodeJSON[domain.WorkItem](t, resp)
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "B"})
	b := decodeJSON[domain.WorkItem](t, resp)

	postJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+a.ID+"/relationships", map[string]any{
		"type": "depends_on", "target_id": b.ID,
	})
	postJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+b.ID+"/relationships", map[string]any{
		"type": "depends_on", "target_id": a.ID,
	})

	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+a.ID+"/cycles")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var cycleResp struct {
		Cycles [][]string `json:"cycles"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&cycleResp))
	_ = resp.Body.Close()
	assert.Len(t, cycleResp.Cycles, 1)
}

func TestIntegration_ListFilters(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// Create items with different types and tags.
	postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Bug1", "type": "bug", "tags": []string{"urgent"}})
	postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Feature1", "type": "feature", "tags": []string{"urgent", "backend"}})
	postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "Bug2", "type": "bug"})

	// Filter by type.
	resp := getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?type=bug")
	list := decodeJSON[store.ListResult](t, resp)
	assert.Equal(t, 2, list.Total)

	// Filter by tags (AND).
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?tags=urgent,backend")
	list = decodeJSON[store.ListResult](t, resp)
	assert.Equal(t, 1, list.Total)

	// Filter by state.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?state=NotDone")
	list = decodeJSON[store.ListResult](t, resp)
	assert.Equal(t, 3, list.Total)
}

func TestIntegration_Pagination(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	for i := range 5 {
		postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": fmt.Sprintf("Item %d", i)})
	}

	resp := getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?page=1&page_size=2")
	list := decodeJSON[store.ListResult](t, resp)
	assert.Equal(t, 5, list.Total)
	assert.Len(t, list.Items, 2)
	assert.Equal(t, 1, list.Page)
	assert.Equal(t, 2, list.PageSize)

	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?page=3&page_size=2")
	list = decodeJSON[store.ListResult](t, resp)
	assert.Len(t, list.Items, 1) // last page
}

func TestIntegration_ErrorResponses(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// 404.
	resp := getJSON(t, srv.URL+integrationProjectPrefix+"/workitems/nonexistent")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	_ = resp.Body.Close()
	assert.Equal(t, "NOT_FOUND", errResp.Error.Code)

	// 400 — missing title.
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"description": "no title"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	// 400 — invalid page.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?page=-1")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	// 400 — invalid page_size.
	resp = getJSON(t, srv.URL+integrationProjectPrefix+"/workitems?page_size=999")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestIntegration_HierarchyParentCycle(t *testing.T) {
	srv := integrationServer(t)
	defer srv.Close()

	// Create A -> B hierarchy.
	resp := postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "A"})
	a := decodeJSON[domain.WorkItem](t, resp)
	resp = postJSON(t, srv.URL+integrationProjectPrefix+"/workitems", map[string]any{"title": "B", "parent_id": a.ID})
	b := decodeJSON[domain.WorkItem](t, resp)

	// Try to set A's parent to B (would create cycle).
	resp = patchJSON(t, srv.URL+integrationProjectPrefix+"/workitems/"+a.ID, map[string]any{"parent_id": b.ID})
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	_ = resp.Body.Close()
}
