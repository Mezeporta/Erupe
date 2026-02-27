package api

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

// UserIDFromContext extracts the authenticated user ID from the request context.
// Returns the user ID and true if present, or 0 and false otherwise.
func UserIDFromContext(ctx context.Context) (uint32, bool) {
	uid, ok := ctx.Value(userIDKey).(uint32)
	return uid, ok
}

// AuthMiddleware extracts a Bearer token from the Authorization header,
// validates it, and injects the user ID into the request context.
func (s *APIServer) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		userID, err := s.userIDFromToken(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
