package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// EchoRateLimit returns a rate limiter middleware
// Limits requests to 20 per second with a burst of 50
func EchoRateLimit() echo.MiddlewareFunc {
	config := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:  rate.Limit(20), // 20 requests per second
				Burst: 50,              // Allow burst of 50 requests
				// ExpiresIn omitted - uses default
			},
		),
	}

	return middleware.RateLimiterWithConfig(config)
}
