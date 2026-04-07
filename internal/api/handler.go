package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"
)

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	WorkItems     store.WorkItemStore
	Relationships store.RelationshipStore
	Cycles        store.CycleDetector
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /workitems", h.CreateWorkItem)
	mux.HandleFunc("GET /workitems", h.ListWorkItems)
	mux.HandleFunc("GET /workitems/{id}", h.GetWorkItem)
	mux.HandleFunc("PATCH /workitems/{id}", h.UpdateWorkItem)
	mux.HandleFunc("DELETE /workitems/{id}", h.DeleteWorkItem)
	mux.HandleFunc("POST /workitems/{id}/relationships", h.AddRelationship)
	mux.HandleFunc("DELETE /workitems/{id}/relationships/{rel_id}", h.RemoveRelationship)
	mux.HandleFunc("GET /workitems/{id}/cycles", h.DetectCycles)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

const maxBodySize = 1 << 20 // 1 MB

// --- Create ---

type createRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	ParentID    *string  `json:"parent_id"`
}

func (h *Handler) CreateWorkItem(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	item := &domain.WorkItem{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Tags:        req.Tags,
		ParentID:    req.ParentID,
	}

	if err := h.WorkItems.Create(r.Context(), item); err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// --- Get ---

func (h *Handler) GetWorkItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := h.WorkItems.Get(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// --- Update ---

type updateRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	State       *string  `json:"state"`
	Type        *string  `json:"type"`
	Tags        []string `json:"tags"`
	ParentID    *string  `json:"parent_id"`
	Override    bool     `json:"override"`
}

func (h *Handler) UpdateWorkItem(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	id := r.PathValue("id")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	params := store.UpdateParams{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Tags:        req.Tags,
		ParentID:    req.ParentID,
		Override:    req.Override,
		IsAdmin:     isAdmin(r.Context()),
	}
	if req.State != nil {
		s := domain.State(*req.State)
		params.State = &s
	}

	item, err := h.WorkItems.Update(r.Context(), id, params)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// --- Delete ---

func (h *Handler) DeleteWorkItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.WorkItems.Delete(r.Context(), id); err != nil {
		mapDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- List ---

func (h *Handler) ListWorkItems(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	result, err := h.WorkItems.List(r.Context(), filter)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func parseListFilter(r *http.Request) (store.ListFilter, error) {
	q := r.URL.Query()
	var f store.ListFilter

	if v := q.Get("state"); v != "" {
		s := domain.State(v)
		if !domain.ValidState(s) {
			return f, fmt.Errorf("invalid state %q", v)
		}
		f.State = &s
	}
	if v := q.Get("tags"); v != "" {
		f.Tags = strings.Split(v, ",")
	}
	if v := q.Get("type"); v != "" {
		f.Type = &v
	}
	if v := q.Get("parent_id"); v != "" {
		f.ParentID = &v
	}
	if v := q.Get("relationship_type"); v != "" {
		rt := domain.RelationshipType(v)
		if !domain.ValidRelationshipType(rt) {
			return f, fmt.Errorf("invalid relationship_type %q", v)
		}
		f.RelationshipType = &rt
	}
	if v := q.Get("relationship_target_id"); v != "" {
		f.RelationshipTarget = &v
	}
	if v := q.Get("is_blocked"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return f, fmt.Errorf("invalid is_blocked %q", v)
		}
		f.IsBlocked = &b
	}
	if v := q.Get("is_ready"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return f, fmt.Errorf("invalid is_ready %q", v)
		}
		f.IsReady = &b
	}

	if f.IsBlocked != nil && f.IsReady != nil {
		return f, fmt.Errorf("is_blocked and is_ready cannot be used together")
	}

	f.Page = 1
	if v := q.Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			return f, fmt.Errorf("invalid page %q", v)
		}
		f.Page = p
	}

	f.PageSize = 50
	if v := q.Get("page_size"); v != "" {
		ps, err := strconv.Atoi(v)
		if err != nil || ps < 1 || ps > 200 {
			return f, fmt.Errorf("invalid page_size %q (must be 1-200)", v)
		}
		f.PageSize = ps
	}

	return f, nil
}

// --- Relationships ---

type addRelationshipRequest struct {
	Type     string `json:"type"`
	TargetID string `json:"target_id"`
}

func (h *Handler) AddRelationship(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	ownerID := r.PathValue("id")

	var req addRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	rel := &domain.Relationship{
		Type:     domain.RelationshipType(req.Type),
		TargetID: req.TargetID,
	}

	if err := h.Relationships.Add(r.Context(), ownerID, rel); err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, rel)
}

func (h *Handler) RemoveRelationship(w http.ResponseWriter, r *http.Request) {
	ownerID := r.PathValue("id")
	relID := r.PathValue("rel_id")

	if err := h.Relationships.Remove(r.Context(), ownerID, relID); err != nil {
		mapDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Cycles ---

type cyclesResponse struct {
	Cycles [][]string `json:"cycles"`
}

func (h *Handler) DetectCycles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cycles, err := h.Cycles.DetectCycles(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	if cycles == nil {
		cycles = [][]string{}
	}

	writeJSON(w, http.StatusOK, cyclesResponse{Cycles: cycles})
}
