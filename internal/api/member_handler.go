package api

import (
	"encoding/json"
	"net/http"

	"github.com/dshills/lattice/internal/domain"
)

// ListMembers returns all members of a project. Any role can view.
func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	projectID := h.projectID(r)
	members, err := h.Memberships.ListByProject(r.Context(), projectID)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"members": members})
}

type addMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// AddMember adds a user to the project by email. Owner only.
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	if err := requireOwner(r.Context()); err != nil {
		mapDomainError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	role := domain.ProjectRole(req.Role)
	if !domain.ValidProjectRole(role) {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid role")
		return
	}

	user, err := h.Users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	membership := &domain.ProjectMembership{
		ProjectID: h.projectID(r),
		UserID:    user.ID,
		Role:      role,
	}
	if err := h.Memberships.Add(r.Context(), membership); err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, membership)
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

// UpdateMemberRole changes a member's role. Owner only.
func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	if err := requireOwner(r.Context()); err != nil {
		mapDomainError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	role := domain.ProjectRole(req.Role)
	if !domain.ValidProjectRole(role) {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid role")
		return
	}

	projectID := h.projectID(r)
	userID := r.PathValue("user_id")
	if err := h.Memberships.UpdateRole(r.Context(), projectID, userID, role); err != nil {
		mapDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember removes a member from the project. Owner only.
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	if err := requireOwner(r.Context()); err != nil {
		mapDomainError(w, err)
		return
	}

	projectID := h.projectID(r)
	userID := r.PathValue("user_id")
	if err := h.Memberships.Remove(r.Context(), projectID, userID); err != nil {
		mapDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
