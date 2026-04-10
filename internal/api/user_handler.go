package api

import (
	"encoding/json"
	"net/http"

	"github.com/dshills/lattice/internal/auth"
	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	Users store.UserStore
}

// RegisterUserRoutes registers user routes on the given mux.
func (h *UserHandler) RegisterUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/me", h.GetMe)
	mux.HandleFunc("PATCH /users/me", h.UpdateMe)
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	user, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type updateMeRequest struct {
	DisplayName *string `json:"display_name"`
	Password    *string `json:"password"`
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	if req.Password != nil {
		if err := domain.ValidatePassword(*req.Password); err != nil {
			mapDomainError(w, err)
			return
		}
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
		if err := h.Users.UpdatePassword(r.Context(), userID, hash); err != nil {
			mapDomainError(w, err)
			return
		}
	}

	if req.DisplayName != nil {
		user, err := h.Users.UpdateDisplayName(r.Context(), userID, *req.DisplayName)
		if err != nil {
			mapDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, user)
		return
	}

	// If only password was changed, return the current user.
	user, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}
