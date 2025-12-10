package middleware

import (
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EchoCORS returns Echo's built-in CORS middleware with custom configuration
func EchoCORS() echo.MiddlewareFunc {
	// Get allowed origins from environment
	allowedOrigins := []string{}
	if originsEnv := os.Getenv("ALLOWED_ORIGINS"); originsEnv != "" {
		allowedOrigins = strings.Split(originsEnv, ",")
	}

	// In development, allow all origins
	if os.Getenv("ENV") == "development" {
		allowedOrigins = []string{"*"}
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
}
