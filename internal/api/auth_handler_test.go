package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dshills/lattice/internal/api"
	"github.com/dshills/lattice/internal/auth"
	"github.com/dshills/lattice/internal/domain"
)

type mockUserStore struct {
	users map[string]*domain.User // keyed by ID
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{users: make(map[string]*domain.User)}
}

func (m *mockUserStore) Create(_ context.Context, user *domain.User, passwordHash string) error {
	// Check for duplicate email
	for _, u := range m.users {
		if u.Email == user.Email {
			return domain.ErrDuplicateEmail
		}
	}
	user.ID = "test-user-id"
	user.PasswordHash = passwordHash
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	if err := user.Validate(); err != nil {
		return err
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserStore) GetByID(_ context.Context, id string) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserStore) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserStore) UpdateDisplayName(_ context.Context, id, displayName string) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	u.DisplayName = displayName
	return u, nil
}

func (m *mockUserStore) UpdatePassword(_ context.Context, id, hash string) error {
	u, ok := m.users[id]
	if !ok {
		return domain.ErrNotFound
	}
	u.PasswordHash = hash
	return nil
}

func (m *mockUserStore) Delete(_ context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.users, id)
	return nil
}

// --- Auth handler tests ---

func newTokenService() *auth.TokenService {
	return auth.NewTokenService("test-secret-at-least-32-chars!!", 15*time.Minute, 7*24*time.Hour)
}

func setupAuthServer() (*http.ServeMux, *mockUserStore, *auth.TokenService) {
	users := newMockUserStore()
	tokens := newTokenService()
	ah := &api.AuthHandler{Users: users, Tokens: tokens}
	uh := &api.UserHandler{Users: users}

	mux := http.NewServeMux()
	ah.RegisterAuthRoutes(mux)
	uh.RegisterUserRoutes(mux)
	return mux, users, tokens
}

func TestRegister(t *testing.T) {
	mux, _, _ := setupAuthServer()

	body := `{"email":"alice@example.com","display_name":"Alice","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Contains(t, string(resp["access_token"]), ".")
	assert.Contains(t, string(resp["user"]), "alice@example.com")
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mux, _, _ := setupAuthServer()

	body := `{"email":"alice@example.com","display_name":"Alice","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Register again with same email
	req = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_WeakPassword(t *testing.T) {
	mux, _, _ := setupAuthServer()

	body := `{"email":"alice@example.com","display_name":"Alice","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin(t *testing.T) {
	mux, _, _ := setupAuthServer()

	// Register first
	regBody := `{"email":"bob@example.com","display_name":"Bob","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Login
	loginBody := `{"email":"bob@example.com","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Contains(t, string(resp["access_token"]), ".")

	// Check refresh cookie is set
	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "lattice_refresh" {
			found = true
			assert.True(t, c.HttpOnly)
			break
		}
	}
	assert.True(t, found, "refresh cookie should be set")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mux, _, _ := setupAuthServer()

	body := `{"email":"nobody@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefresh(t *testing.T) {
	mux, _, tokens := setupAuthServer()

	// Register to get a user
	regBody := `{"email":"carol@example.com","display_name":"Carol","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Issue a refresh token
	refreshToken, err := tokens.IssueRefreshToken("test-user-id")
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "lattice_refresh", Value: refreshToken})
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp["access_token"])
}

func TestRefresh_MissingCookie(t *testing.T) {
	mux, _, _ := setupAuthServer()

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Auth middleware tests ---

func TestAuthMiddleware_ValidToken(t *testing.T) {
	tokens := newTokenService()
	token, err := tokens.IssueAccessToken("user-789")
	require.NoError(t, err)

	var capturedUserID string
	inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedUserID = api.UserIDFromContext(r.Context())
	})

	handler := api.AuthMiddleware(tokens, inner)
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user-789", capturedUserID)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	tokens := newTokenService()
	inner := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	handler := api.AuthMiddleware(tokens, inner)
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	tokens := newTokenService()
	inner := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	handler := api.AuthMiddleware(tokens, inner)
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_SkipsAuthPaths(t *testing.T) {
	tokens := newTokenService()
	var called bool
	inner := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	handler := api.AuthMiddleware(tokens, inner)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, called, "auth middleware should pass through /auth/ paths")
}

// --- User handler tests ---

func TestGetMe(t *testing.T) {
	mux, users, tokens := setupAuthServer()

	// Create a user in the store
	user := &domain.User{Email: "dave@example.com", DisplayName: "Dave"}
	hash, _ := auth.HashPassword("password123")
	require.NoError(t, users.Create(context.Background(), user, hash))

	token, err := tokens.IssueAccessToken(user.ID)
	require.NoError(t, err)

	// Wrap with auth middleware so UserIDFromContext works
	handler := api.AuthMiddleware(tokens, mux)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.User
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "Dave", resp.DisplayName)
	assert.Empty(t, resp.PasswordHash, "password hash should not be in response")
}

func TestUpdateMe_DisplayName(t *testing.T) {
	mux, users, tokens := setupAuthServer()

	user := &domain.User{Email: "eve@example.com", DisplayName: "Eve"}
	hash, _ := auth.HashPassword("password123")
	require.NoError(t, users.Create(context.Background(), user, hash))

	token, err := tokens.IssueAccessToken(user.ID)
	require.NoError(t, err)

	handler := api.AuthMiddleware(tokens, mux)
	body := `{"display_name":"Eve Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/users/me", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.User
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "Eve Updated", resp.DisplayName)
}
