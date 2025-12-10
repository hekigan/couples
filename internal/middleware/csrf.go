package middleware

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EchoCSRF returns a CSRF protection middleware
// Skips GET requests automatically
func EchoCSRF() echo.MiddlewareFunc {
	config := middleware.CSRFConfig{
		TokenLookup:    "form:csrf,header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookieHTTPOnly: true,
		CookieSecure:   os.Getenv("ENV") == "production",
		CookieSameSite: http.SameSiteLaxMode,
		CookiePath:     "/",
		CookieMaxAge:   86400, // 24 hours
		Skipper: func(c echo.Context) bool {
			// Skip CSRF check for GET, HEAD, OPTIONS requests
			method := c.Request().Method
			return method == "GET" || method == "HEAD" || method == "OPTIONS"
		},
	}

	return middleware.CSRFWithConfig(config)
}
