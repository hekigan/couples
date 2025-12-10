package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// EchoRequireAdmin enforces that a user must be an admin
// Returns 404 (not 403) if unauthorized for security by obscurity
func EchoRequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// First check if user is authenticated
			_, ok := GetUserID(c)
			if !ok {
				return echo.NewHTTPError(http.StatusNotFound, "Page not found")
			}

			// Check if user is admin
			isAdmin, ok := GetIsAdmin(c)
			if !ok || !isAdmin {
				return echo.NewHTTPError(http.StatusNotFound, "Page not found")
			}

			return next(c)
		}
	}
}
