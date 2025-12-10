package middleware

import (
	"github.com/labstack/echo/v4"
)

// EchoSecurityHeaders adds security headers to all responses
func EchoSecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent MIME type sniffing
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent the page from being displayed in an iframe
			c.Response().Header().Set("X-Frame-Options", "DENY")

			// Enable browser's XSS protection
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Control information sent in the Referer header
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			return next(c)
		}
	}
}
