package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/dshills/lattice/internal/auth"
)

type userIDKey struct{}

// AuthMiddleware validates the Bearer token and stores the user ID in context.
// Requests to /auth/ paths are passed through without authentication.
func AuthMiddleware(tokens *auth.TokenService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/auth/") {
			next.ServeHTTP(w, r)
			return
		}

		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := tokens.ValidateToken(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext returns the authenticated user's ID from the request context.
// Returns an empty string if no user ID is set (e.g., on public routes).
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey{}).(string)
	return v
}

// TestSetUserID injects a user ID into the context. Exported for testing only.
func TestSetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}
