package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dshills/lattice/internal/api"
	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"
)

// --- Mock stores ---

type mockWorkItemStore struct {
	items map[string]*domain.WorkItem
}

func newMockWorkItemStore() *mockWorkItemStore {
	return &mockWorkItemStore{items: make(map[string]*domain.WorkItem)}
}

func (m *mockWorkItemStore) Create(_ context.Context, item *domain.WorkItem) error {
	item.ID = "test-id-1"
	item.State = domain.NotDone
	if err := item.Validate(); err != nil {
		return err
	}
	m.items[item.ID] = item
	return nil
}

func (m *mockWorkItemStore) Get(_ context.Context, id string) (*domain.WorkItem, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (m *mockWorkItemStore) Update(_ context.Context, id string, params store.UpdateParams) (*domain.WorkItem, error) {
	existing, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if params.State != nil {
		if err := domain.ValidateTransition(existing.State, *params.State, params.Override, params.IsAdmin); err != nil {
			return nil, err
		}
		existing.State = *params.State
	}
	if params.Title != nil {
		existing.Title = *params.Title
	}
	if params.Description != nil {
		existing.Description = *params.Description
	}
	if params.Type != nil {
		existing.Type = *params.Type
	}
	result := *existing
	return &result, nil
}

func (m *mockWorkItemStore) Delete(_ context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *mockWorkItemStore) List(_ context.Context, filter store.ListFilter) (*store.ListResult, error) {
	items := make([]domain.WorkItem, 0)
	for _, item := range m.items {
		if filter.State != nil && item.State != *filter.State {
			continue
		}
		items = append(items, *item)
	}
	return &store.ListResult{
		Items:    items,
		Total:    len(items),
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (m *mockWorkItemStore) AncestorDepth(_ context.Context, _ string) (int, error) {
	return 1, nil
}

func (m *mockWorkItemStore) HasCycle(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

type mockRelationshipStore struct {
	rels map[string]*domain.Relationship
}

func newMockRelationshipStore() *mockRelationshipStore {
	return &mockRelationshipStore{rels: make(map[string]*domain.Relationship)}
}

func (m *mockRelationshipStore) Add(_ context.Context, _ string, rel *domain.Relationship) error {
	if !domain.ValidRelationshipType(rel.Type) {
		return domain.ErrInvalidInput
	}
	rel.ID = "rel-id-1"
	m.rels[rel.ID] = rel
	return nil
}

func (m *mockRelationshipStore) Remove(_ context.Context, _, relID string) error {
	if _, ok := m.rels[relID]; !ok {
		return domain.ErrNotFound
	}
	delete(m.rels, relID)
	return nil
}

func (m *mockRelationshipStore) ListByTarget(_ context.Context, _ string) ([]domain.Relationship, error) {
	return nil, nil
}

type mockCycleDetector struct{}

func (m *mockCycleDetector) DetectCycles(_ context.Context, _ string) ([][]string, error) {
	return nil, nil
}

func newHandler() *api.Handler {
	return &api.Handler{
		WorkItems:     newMockWorkItemStore(),
		Relationships: newMockRelationshipStore(),
		Cycles:        &mockCycleDetector{},
	}
}

func setupServer(h *api.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// --- Tests ---

func TestCreateWorkItem(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	body := `{"title":"Test Item","description":"A test","tags":["alpha"]}`
	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	assert.Equal(t, "Test Item", item.Title)
	assert.Equal(t, domain.NotDone, item.State)
	assert.NotEmpty(t, item.ID)
}

func TestCreateWorkItemInvalidJSON(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateWorkItemMissingTitle(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	body := `{"description":"no title"}`
	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetWorkItem(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	// Create first.
	body := `{"title":"Get Me"}`
	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Get.
	req = httptest.NewRequest(http.MethodGet, "/workitems/test-id-1", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	assert.Equal(t, "Get Me", item.Title)
}

func TestGetWorkItemNotFound(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodGet, "/workitems/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWorkItem(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	// Create.
	createBody := `{"title":"Before"}`
	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Update.
	updateBody := `{"title":"After","state":"InProgress"}`
	req = httptest.NewRequest(http.MethodPatch, "/workitems/test-id-1", bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	assert.Equal(t, "After", item.Title)
	assert.Equal(t, domain.InProgress, item.State)
}

func TestUpdateWorkItemNotFound(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	body := `{"title":"x"}`
	req := httptest.NewRequest(http.MethodPatch, "/workitems/nonexistent", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWorkItemInvalidTransition(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	// Create (state=NotDone).
	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(`{"title":"t"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Try to skip to Completed.
	body := `{"state":"Completed"}`
	req = httptest.NewRequest(http.MethodPatch, "/workitems/test-id-1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUpdateWorkItemBackwardWithAdminOverride(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	// Create and advance to InProgress.
	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(`{"title":"t"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	req = httptest.NewRequest(http.MethodPatch, "/workitems/test-id-1", bytes.NewBufferString(`{"state":"InProgress"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Backward transition without admin should fail.
	body := `{"state":"NotDone","override":true}`
	req = httptest.NewRequest(http.MethodPatch, "/workitems/test-id-1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// With admin role should succeed.
	req = httptest.NewRequest(http.MethodPatch, "/workitems/test-id-1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", "admin")
	w = httptest.NewRecorder()
	// Need to apply middleware for role extraction.
	handler := api.RoleMiddleware(mux)
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteWorkItem(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	// Create.
	req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(`{"title":"Delete Me"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Delete.
	req = httptest.NewRequest(http.MethodDelete, "/workitems/test-id-1", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify gone.
	req = httptest.NewRequest(http.MethodGet, "/workitems/test-id-1", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteWorkItemNotFound(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodDelete, "/workitems/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListWorkItems(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	// Create two items.
	for _, title := range []string{"One", "Two"} {
		body := `{"title":"` + title + `"}`
		req := httptest.NewRequest(http.MethodPost, "/workitems", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/workitems?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result store.ListResult
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, 1, result.Page)
}

func TestListWorkItemsInvalidPage(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodGet, "/workitems?page=-1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListWorkItemsInvalidPageSize(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodGet, "/workitems?page_size=999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddRelationship(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	body := `{"type":"depends_on","target_id":"some-target"}`
	req := httptest.NewRequest(http.MethodPost, "/workitems/owner-id/relationships", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var rel domain.Relationship
	require.NoError(t, json.NewDecoder(w.Body).Decode(&rel))
	assert.Equal(t, domain.DependsOn, rel.Type)
	assert.NotEmpty(t, rel.ID)
}

func TestAddRelationshipInvalidType(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	body := `{"type":"invalid","target_id":"some-target"}`
	req := httptest.NewRequest(http.MethodPost, "/workitems/owner-id/relationships", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemoveRelationship(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	// Add first.
	body := `{"type":"blocks","target_id":"t"}`
	req := httptest.NewRequest(http.MethodPost, "/workitems/owner/relationships", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Remove.
	req = httptest.NewRequest(http.MethodDelete, "/workitems/owner/relationships/rel-id-1", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRemoveRelationshipNotFound(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodDelete, "/workitems/owner/relationships/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDetectCycles(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodGet, "/workitems/some-id/cycles", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Cycles [][]string `json:"cycles"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Empty(t, resp.Cycles)
}

func TestErrorResponseFormat(t *testing.T) {
	h := newHandler()
	mux := setupServer(h)

	req := httptest.NewRequest(http.MethodGet, "/workitems/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	assert.NotEmpty(t, resp.Error.Message)
}
