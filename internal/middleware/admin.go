package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequireAdmin ensures user has admin privileges (checks user's is_admin flag)
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by AuthMiddleware)
		userIDVal := r.Context().Value(UserIDKey)
		if userIDVal == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		_, ok := userIDVal.(uuid.UUID)
		if !ok {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Get session to check admin flag
		session, err := Store.Get(r, "couple-card-game-session")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Check if user is admin (stored in session from login)
		isAdmin, ok := session.Values["is_admin"].(bool)
		if !ok || !isAdmin {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// User is admin, set in context for handlers to use if needed
		ctx := context.WithValue(r.Context(), "is_admin", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}



