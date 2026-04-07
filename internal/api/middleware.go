package api

import (
	"context"
	"log"
	"mime"
	"net/http"
	"time"
)

type contextKey string

const roleKey contextKey = "role"

// RoleMiddleware extracts the X-Role header and stores isAdmin in context.
func RoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("X-Role")
		isAdmin := role == "admin"
		ctx := context.WithValue(r.Context(), roleKey, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isAdmin returns whether the current request has admin role.
func isAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(roleKey).(bool)
	return v
}

// LoggingMiddleware logs each request with method, path, status, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// JSONContentType ensures request Content-Type is application/json for methods
// that have a body.
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut {
			mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if mt != "application/json" {
				writeError(w, http.StatusBadRequest, "INVALID_INPUT", "Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
