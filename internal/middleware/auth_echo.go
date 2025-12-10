package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EchoAuth extracts the user ID from the session and adds it to the Echo context
// This middleware does NOT enforce authentication - it only extracts the user if present
func EchoAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := GetSession(c)
			if err != nil {
				return next(c)
			}

			// Try to extract user_id from session
			if userID, ok := session.Values["user_id"].(string); ok && userID != "" {
				if id, err := uuid.Parse(userID); err == nil {
					SetUserID(c, id)
				}
			}

			// Extract is_admin flag if present
			if isAdmin, ok := session.Values["is_admin"].(bool); ok {
				SetIsAdmin(c, isAdmin)
			}

			return next(c)
		}
	}
}

// EchoRequireAuth enforces that a user must be authenticated
// If not authenticated, redirects to /auth/login
func EchoRequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, ok := GetUserID(c)
			if !ok {
				// Check if this is an HTMX request
				if c.Request().Header.Get("HX-Request") == "true" {
					c.Response().Header().Set("HX-Redirect", "/auth/login")
					return c.NoContent(http.StatusOK)
				}

				// Regular redirect
				return c.Redirect(http.StatusSeeOther, "/auth/login")
			}

			return next(c)
		}
	}
}

// EchoAnonymousSession handles anonymous user sessions
// If a user is not authenticated but has an anonymous session, it extracts the user_id
func EchoAnonymousSession() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If user is already authenticated, skip
			if _, ok := GetUserID(c); ok {
				return next(c)
			}

			// Check for anonymous session
			session, err := GetSession(c)
			if err != nil {
				return next(c)
			}

			// Check if this is an anonymous user
			isAnonymous, ok := session.Values["is_anonymous"].(bool)
			if ok && isAnonymous {
				// Extract anonymous user_id
				if userID, ok := session.Values["user_id"].(string); ok && userID != "" {
					if id, err := uuid.Parse(userID); err == nil {
						SetUserID(c, id)
					}
				}
			}

			return next(c)
		}
	}
}
