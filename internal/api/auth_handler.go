package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dshills/lattice/internal/auth"
	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"
)

const refreshCookieName = "lattice_refresh"

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	Users  store.UserStore
	Tokens *auth.TokenService
}

// RegisterAuthRoutes registers auth routes on the given mux.
func (h *AuthHandler) RegisterAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)
}

type registerRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type authResponse struct {
	User        *domain.User `json:"user"`
	AccessToken string       `json:"access_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	if err := domain.ValidatePassword(req.Password); err != nil {
		mapDomainError(w, err)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	user := &domain.User{
		Email:       req.Email,
		DisplayName: req.DisplayName,
	}
	if err := h.Users.Create(r.Context(), user, hash); err != nil {
		mapDomainError(w, err)
		return
	}

	h.issueTokens(w, user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid JSON body")
		return
	}

	user, err := h.Users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	h.issueTokens(w, user)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing refresh token")
		return
	}

	userID, err := h.Tokens.ValidateToken(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid refresh token")
		return
	}

	accessToken, err := h.Tokens.IssueAccessToken(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"access_token": accessToken})
}

func (h *AuthHandler) issueTokens(w http.ResponseWriter, user *domain.User) {
	accessToken, err := h.Tokens.IssueAccessToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	refreshToken, err := h.Tokens.IssueRefreshToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
	})

	writeJSON(w, http.StatusOK, authResponse{
		User:        user,
		AccessToken: accessToken,
	})
}
