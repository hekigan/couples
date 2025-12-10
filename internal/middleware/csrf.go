package middleware

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EchoCSRF returns a CSRF protection middleware
// Echo's CSRF middleware automatically validates only on POST/PUT/DELETE/PATCH
// and generates tokens on all requests (including GET)
func EchoCSRF() echo.MiddlewareFunc {
	config := middleware.CSRFConfig{
		TokenLookup:    "form:csrf,header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookieHTTPOnly: true,
		CookieSecure:   os.Getenv("ENV") == "production",
		CookieSameSite: http.SameSiteLaxMode,
		CookiePath:     "/",
		CookieMaxAge:   86400, // 24 hours
		// No Skipper - let Echo's CSRF middleware handle all requests
		// It will generate tokens on GET and validate on POST/DELETE/etc
	}

	return middleware.CSRFWithConfig(config)
}
