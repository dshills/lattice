package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/dshills/lattice/internal/domain"
	"github.com/dshills/lattice/internal/store"
)

type projectRoleKey struct{}

// ProjectRoleMiddleware loads the authenticated user's role for project-scoped
// routes and attaches it to the request context. Returns 403 if the user is not
// a member. Skips non-project routes (/auth/, /users/, GET /projects without ID).
func ProjectRoleMiddleware(memberships store.MembershipStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip non-project routes.
		if strings.HasPrefix(r.URL.Path, "/auth/") ||
			strings.HasPrefix(r.URL.Path, "/users/") {
			next.ServeHTTP(w, r)
			return
		}

		// Extract project ID from URL path: /projects/{id}/...
		// PathValue isn't available yet (set by ServeMux during routing),
		// so we parse the path directly.
		projectID := extractProjectID(r.URL.Path)
		if projectID == "" {
			next.ServeHTTP(w, r)
			return
		}

		userID := UserIDFromContext(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
			return
		}

		role, err := memberships.GetRole(r.Context(), projectID, userID)
		if err != nil {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "not a member of this project")
			return
		}

		ctx := context.WithValue(r.Context(), projectRoleKey{}, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractProjectID parses the project ID from paths like /projects/{id}/...
// Returns empty string for /projects (list) or non-project paths.
func extractProjectID(path string) string {
	// Must start with /projects/
	const prefix = "/projects/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := path[len(prefix):]
	if rest == "" {
		return ""
	}
	// Take everything up to the next slash (or end of string).
	if idx := strings.IndexByte(rest, '/'); idx >= 0 {
		return rest[:idx]
	}
	return rest
}

// ProjectRoleFromContext returns the authenticated user's role for the current
// project. Returns an empty string if no role is set.
func ProjectRoleFromContext(ctx context.Context) domain.ProjectRole {
	v, _ := ctx.Value(projectRoleKey{}).(domain.ProjectRole)
	return v
}
