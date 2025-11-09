package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware extracts user info from session
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "couple-card-game-session")
		
		if userID, ok := session.Values["user_id"].(string); ok {
			if id, err := uuid.Parse(userID); err == nil {
				ctx := context.WithValue(r.Context(), UserIDKey, id)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuth ensures user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		if userID == nil {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AnonymousSessionMiddleware handles anonymous user sessions
func AnonymousSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "couple-card-game-session")
		
		// Check if anonymous user flag is set
		if isAnon, ok := session.Values["is_anonymous"].(bool); ok && isAnon {
			if userID, ok := session.Values["user_id"].(string); ok {
				if id, err := uuid.Parse(userID); err == nil {
					ctx := context.WithValue(r.Context(), UserIDKey, id)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}



