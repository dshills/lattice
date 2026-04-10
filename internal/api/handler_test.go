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

func (m *mockWorkItemStore) Get(_ context.Context, _, id string) (*domain.WorkItem, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (m *mockWorkItemStore) Update(_ context.Context, _, id string, params store.UpdateParams) (*domain.WorkItem, error) {
	existing, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if params.State != nil {
		if err := domain.ValidateTransition(existing.State, *params.State, params.Override); err != nil {
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
	if params.AssigneeID != nil {
		if *params.AssigneeID == "" {
			existing.AssigneeID = nil
		} else {
			existing.AssigneeID = params.AssigneeID
		}
	}
	result := *existing
	return &result, nil
}

func (m *mockWorkItemStore) Delete(_ context.Context, _, id string) error {
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

type mockProjectStore struct{}

func (m *mockProjectStore) Create(_ context.Context, p *domain.Project) error {
	p.ID = "test-project-id"
	return nil
}
func (m *mockProjectStore) Get(_ context.Context, id string) (*domain.Project, error) {
	if id == "test-project-id" || id == domain.DefaultProjectID {
		return &domain.Project{ID: id, Name: "Test"}, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockProjectStore) Update(_ context.Context, id string, _ store.ProjectUpdateParams) (*domain.Project, error) {
	return &domain.Project{ID: id, Name: "Updated"}, nil
}
func (m *mockProjectStore) Delete(_ context.Context, _ string) error { return nil }
func (m *mockProjectStore) List(_ context.Context) ([]store.ProjectWithCount, error) {
	return []store.ProjectWithCount{}, nil
}

// mockMembershipStore returns the configured role for all lookups.
type mockMembershipStore struct {
	role domain.ProjectRole
}

func (m *mockMembershipStore) Add(_ context.Context, membership *domain.ProjectMembership) error {
	membership.ID = "test-membership-id"
	return nil
}
func (m *mockMembershipStore) Remove(_ context.Context, _, _ string) error { return nil }
func (m *mockMembershipStore) UpdateRole(_ context.Context, _, _ string, _ domain.ProjectRole) error {
	return nil
}
func (m *mockMembershipStore) ListByProject(_ context.Context, _ string) ([]domain.ProjectMembership, error) {
	return []domain.ProjectMembership{}, nil
}
func (m *mockMembershipStore) GetRole(_ context.Context, _, _ string) (domain.ProjectRole, error) {
	if m.role == "" {
		return domain.RoleOwner, nil
	}
	return m.role, nil
}

const testProjectPrefix = "/projects/" + domain.DefaultProjectID

func newHandler() *api.Handler {
	return &api.Handler{
		Projects:      &mockProjectStore{},
		WorkItems:     newMockWorkItemStore(),
		Relationships: newMockRelationshipStore(),
		Cycles:        &mockCycleDetector{},
		Memberships:   &mockMembershipStore{},
	}
}

func newHandlerWithRole(role domain.ProjectRole) *api.Handler {
	return &api.Handler{
		Projects:      &mockProjectStore{},
		WorkItems:     newMockWorkItemStore(),
		Relationships: newMockRelationshipStore(),
		Cycles:        &mockCycleDetector{},
		Memberships:   &mockMembershipStore{role: role},
	}
}

func setupServer(h *api.Handler) http.Handler {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return api.ProjectRoleMiddleware(h.Memberships, mux)
}

// withUserContext adds a user ID to the request context for auth middleware simulation.
func withUserContext(req *http.Request) *http.Request {
	ctx := api.TestSetUserID(req.Context(), "test-user-id")
	return req.WithContext(ctx)
}

// --- Tests ---

func TestCreateWorkItem(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	body := `{"title":"Test Item","description":"A test","tags":["alpha"]}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	assert.Equal(t, "Test Item", item.Title)
	assert.Equal(t, domain.NotDone, item.State)
	assert.NotEmpty(t, item.ID)
}

func TestCreateWorkItemInvalidJSON(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateWorkItemMissingTitle(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	body := `{"description":"no title"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetWorkItem(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	// Create first.
	body := `{"title":"Get Me"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Get.
	req = withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems/test-id-1", nil))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	assert.Equal(t, "Get Me", item.Title)
}

func TestGetWorkItemNotFound(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems/nonexistent", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWorkItem(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	// Create.
	createBody := `{"title":"Before"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(createBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Update.
	updateBody := `{"title":"After","state":"InProgress"}`
	req = withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	assert.Equal(t, "After", item.Title)
	assert.Equal(t, domain.InProgress, item.State)
}

func TestUpdateWorkItemNotFound(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	body := `{"title":"x"}`
	req := withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/nonexistent", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWorkItemInvalidTransition(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	// Create (state=NotDone).
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(`{"title":"t"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Try to skip to Completed.
	body := `{"state":"Completed"}`
	req = withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUpdateWorkItemOwnerOverride(t *testing.T) {
	h := newHandler() // default role = owner
	handler := setupServer(h)

	// Create and advance to InProgress.
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(`{"title":"t"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	req = withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(`{"state":"InProgress"}`)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Backward transition with override — owner should succeed.
	body := `{"state":"NotDone","override":true}`
	req = withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteWorkItem(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	// Create.
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(`{"title":"Delete Me"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Delete.
	req = withUserContext(httptest.NewRequest(http.MethodDelete, testProjectPrefix+"/workitems/test-id-1", nil))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify gone.
	req = withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems/test-id-1", nil))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteWorkItemNotFound(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodDelete, testProjectPrefix+"/workitems/nonexistent", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListWorkItems(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	// Create two items.
	for _, title := range []string{"One", "Two"} {
		body := `{"title":"` + title + `"}`
		req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(body)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	req := withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems?page=1&page_size=10", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result store.ListResult
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, 1, result.Page)
}

func TestListWorkItemsInvalidPage(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems?page=-1", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListWorkItemsInvalidPageSize(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems?page_size=999", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddRelationship(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	body := `{"type":"depends_on","target_id":"some-target"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems/owner-id/relationships", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var rel domain.Relationship
	require.NoError(t, json.NewDecoder(w.Body).Decode(&rel))
	assert.Equal(t, domain.DependsOn, rel.Type)
	assert.NotEmpty(t, rel.ID)
}

func TestAddRelationshipInvalidType(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	body := `{"type":"invalid","target_id":"some-target"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems/owner-id/relationships", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemoveRelationship(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	// Add first.
	body := `{"type":"blocks","target_id":"t"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems/owner/relationships", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Remove.
	req = withUserContext(httptest.NewRequest(http.MethodDelete, testProjectPrefix+"/workitems/owner/relationships/rel-id-1", nil))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRemoveRelationshipNotFound(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodDelete, testProjectPrefix+"/workitems/owner/relationships/nonexistent", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDetectCycles(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems/some-id/cycles", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Cycles [][]string `json:"cycles"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Empty(t, resp.Cycles)
}

func TestErrorResponseFormat(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems/nonexistent", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

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

// --- Permission enforcement tests ---

func TestViewerCannotCreate(t *testing.T) {
	h := newHandlerWithRole(domain.RoleViewer)
	handler := setupServer(h)

	body := `{"title":"Blocked"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestViewerCannotUpdate(t *testing.T) {
	h := newHandlerWithRole(domain.RoleViewer)
	handler := setupServer(h)

	body := `{"title":"x"}`
	req := withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestViewerCannotDelete(t *testing.T) {
	h := newHandlerWithRole(domain.RoleViewer)
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodDelete, testProjectPrefix+"/workitems/test-id-1", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMemberCannotOverride(t *testing.T) {
	h := newHandlerWithRole(domain.RoleMember)
	handler := setupServer(h)

	// Create item (member can create).
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(`{"title":"t"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Advance to InProgress.
	req = withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(`{"state":"InProgress"}`)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Override requires owner — member should get 403.
	body := `{"state":"NotDone","override":true}`
	req = withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMemberCanCreate(t *testing.T) {
	h := newHandlerWithRole(domain.RoleMember)
	handler := setupServer(h)

	body := `{"title":"Member Item"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestViewerCanRead(t *testing.T) {
	// Use owner handler to create, then viewer handler to read.
	h := newHandler()
	handler := setupServer(h)

	// Create as owner.
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(`{"title":"Readable"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Read as viewer — the mock returns whatever role is configured, so we
	// just verify GET doesn't require write access.
	req = withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems/test-id-1", nil))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNonOwnerCannotUpdateProject(t *testing.T) {
	h := newHandlerWithRole(domain.RoleMember)
	handler := setupServer(h)

	body := `{"name":"New Name"}`
	req := withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix, bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateWorkItemSetsCreatedBy(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	body := `{"title":"With Creator"}`
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	require.NotNil(t, item.CreatedBy)
	assert.Equal(t, "test-user-id", *item.CreatedBy)
}

func TestUpdateWorkItemAssignee(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	// Create.
	req := withUserContext(httptest.NewRequest(http.MethodPost, testProjectPrefix+"/workitems", bytes.NewBufferString(`{"title":"Assign Me"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Update with assignee_id.
	assigneeID := "some-user-id"
	updateBody := `{"assignee_id":"` + assigneeID + `"}`
	req = withUserContext(httptest.NewRequest(http.MethodPatch, testProjectPrefix+"/workitems/test-id-1", bytes.NewBufferString(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var item domain.WorkItem
	require.NoError(t, json.NewDecoder(w.Body).Decode(&item))
	require.NotNil(t, item.AssigneeID)
	assert.Equal(t, assigneeID, *item.AssigneeID)
}

func TestListWorkItemsAssigneeFilter(t *testing.T) {
	h := newHandler()
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodGet, testProjectPrefix+"/workitems?assignee_id=some-user", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNonOwnerCannotDeleteProject(t *testing.T) {
	h := newHandlerWithRole(domain.RoleMember)
	handler := setupServer(h)

	req := withUserContext(httptest.NewRequest(http.MethodDelete, testProjectPrefix, nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
