package main

import (
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/labstack/echo/v4"
)

// registerPublicRoutes registers all public-facing routes (auth, home, profile, friends)
func registerPublicRoutes(e *echo.Echo, h *handlers.Handler) {
	// Static files
	e.Static("/static", "./static")

	// Health check (no CSRF needed)
	e.GET("/health", h.HealthHandler)

	// Home route (with CSRF for logout button)
	home := e.Group("")
	home.Use(middleware.EchoCSRF())
	home.GET("/", h.HomeHandler)

	// Login/Signup routes (no auth required + CSRF for forms)
	loginSignup := e.Group("")
	loginSignup.Use(middleware.EchoCSRF())
	loginSignup.GET("/login", h.LoginHandler)
	loginSignup.POST("/login", h.LoginPostHandler)
	loginSignup.GET("/signup", h.SignupHandler)
	loginSignup.POST("/signup", h.SignupPostHandler)

	// Username setup routes (auth required + CSRF)
	setup := e.Group("/setup-username")
	setup.Use(middleware.EchoRequireAuth())
	setup.Use(middleware.EchoCSRF())
	setup.GET("", h.SetupUsernameHandler)
	setup.POST("", h.SetupUsernamePostHandler)

	// Auth routes (no auth required + CSRF for forms)
	auth := e.Group("/auth")
	auth.Use(middleware.EchoCSRF())
	auth.POST("/logout", h.LogoutHandler)
	auth.POST("/anonymous", h.CreateAnonymousHandler)
	auth.GET("/dev-login-admin", h.DevLoginAsAdminHandler) // Development only

	// OAuth routes
	auth.GET("/oauth/google", h.OAuthGoogleHandler)
	auth.GET("/oauth/facebook", h.OAuthFacebookHandler)
	auth.GET("/oauth/github", h.OAuthGithubHandler)
	auth.GET("/oauth/callback", h.OAuthCallbackHandler)
	auth.POST("/oauth/token", h.OAuthTokenHandler)

	// Profile routes (auth required + CSRF)
	profile := e.Group("/profile")
	profile.Use(middleware.EchoRequireAuth())
	profile.Use(middleware.EchoCSRF())
	profile.GET("", h.ProfileHandler)
	profile.POST("/update", h.UpdateProfileHandler)

	// Email invitation acceptance (public route - must be before auth)
	e.GET("/invitations/friend/:token", h.AcceptEmailInvitationHandler)

	// Email test endpoint (development/debugging - no auth required)
	// Usage: GET /api/test-email?to=your@email.com
	e.GET("/api/test-email", h.TestEmailHandler)

	// Friend routes (auth + rate limiting + CSRF)
	friends := e.Group("/friends")
	friends.Use(middleware.EchoRequireAuth())
	friends.Use(middleware.EchoRateLimit())
	friends.Use(middleware.EchoCSRF())
	friends.GET("", h.FriendListHandler)
	friends.GET("/add", h.AddFriendHandler)
	friends.POST("/add", h.AddFriendHandler)
	friends.GET("/add/input-field", h.GetFriendInputFieldHandler) // HTMX endpoint
	friends.POST("/:id/accept", h.AcceptFriendHandler)
	friends.POST("/:id/decline", h.DeclineFriendHandler)
	friends.DELETE("/:id", h.RemoveFriendHandler)
}
