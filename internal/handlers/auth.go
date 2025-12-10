package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// LoginHandler displays the login page
func (h *Handler) LoginHandler(c echo.Context) error {
	data := &TemplateData{
		Title: "Login - Couple Card Game",
	}
	return h.RenderTemplate(c, "auth/login.html", data)
}

// LoginPostHandler handles email/password login via HTMX
func (h *Handler) LoginPostHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Validate inputs
	if email == "" || password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}

	// Create auth service
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		log.Printf("Failed to initialize auth service: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication service unavailable",
		})
	}

	// Authenticate with Supabase
	session, err := authService.LoginWithPassword(ctx, email, password)
	if err != nil {
		log.Printf("Login failed for %s: %v", email, err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
		})
	}

	// Create or update user in database (reuse OAuth pattern)
	user, err := authService.CreateOrUpdateUserFromOAuth(ctx, session.User)
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user session",
		})
	}

	// Create session
	sess, _ := middleware.GetSession(c)
	sess.Values["user_id"] = user.ID.String()
	sess.Values["username"] = user.Username
	if user.Email != nil {
		sess.Values["email"] = *user.Email
	}
	sess.Values["is_anonymous"] = false
	sess.Values["is_admin"] = user.IsAdmin
	sess.Values["access_token"] = session.AccessToken
	if session.RefreshToken != "" {
		sess.Values["refresh_token"] = session.RefreshToken
	}

	if err := middleware.SaveSession(c, sess); err != nil {
		log.Printf("Failed to save session: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save session",
		})
	}

	// Return HX-Redirect header for HTMX to handle
	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusOK)
}

// SignupHandler displays the signup page
func (h *Handler) SignupHandler(c echo.Context) error {
	data := &TemplateData{
		Title: "Sign Up - Couple Card Game",
	}
	return h.RenderTemplate(c, "auth/signup.html", data)
}

// SignupPostHandler handles user registration via HTMX
func (h *Handler) SignupPostHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	username := c.FormValue("username")

	// Validate inputs
	if email == "" || password == "" || username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "All fields are required",
		})
	}

	// Validate password match
	if password != passwordConfirm {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Passwords do not match",
		})
	}

	// Validate password strength (minimum 6 characters)
	if len(password) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Password must be at least 6 characters long",
		})
	}

	// Validate username (alphanumeric, 3-50 characters)
	if len(username) < 3 || len(username) > 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username must be between 3 and 50 characters",
		})
	}

	// Create auth service
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		log.Printf("Failed to initialize auth service: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authentication service unavailable",
		})
	}

	// Create user in Supabase Auth
	session, err := authService.SignupWithPassword(ctx, email, password, username)
	if err != nil {
		log.Printf("❌ Signup failed for %s: %v", email, err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Signup failed. Email may already be in use.",
		})
	}

	// Create user in application database
	user, err := authService.CreateOrUpdateUserFromOAuth(ctx, session.User)
	if err != nil {
		log.Printf("Failed to create user in database: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user account",
		})
	}

	// Create session
	sess, _ := middleware.GetSession(c)
	sess.Values["user_id"] = user.ID.String()
	sess.Values["username"] = user.Username
	if user.Email != nil {
		sess.Values["email"] = *user.Email
	}
	sess.Values["is_anonymous"] = false
	sess.Values["is_admin"] = user.IsAdmin
	sess.Values["access_token"] = session.AccessToken
	if session.RefreshToken != "" {
		sess.Values["refresh_token"] = session.RefreshToken
	}

	if err := middleware.SaveSession(c, sess); err != nil {
		log.Printf("Failed to save session: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save session",
		})
	}

	// Return HX-Redirect header for HTMX to handle
	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusOK)
}

// LogoutHandler logs out the user
func (h *Handler) LogoutHandler(c echo.Context) error {
	session, _ := middleware.GetSession(c)
	session.Options.MaxAge = -1
	middleware.SaveSession(c, session)

	// Check if this is an HTMX request
	if c.Request().Header.Get("HX-Request") == "true" {
		// For HTMX, use HX-Redirect header
		c.Response().Header().Set("HX-Redirect", "/")
		return c.NoContent(http.StatusOK)
	}

	// For regular requests, use standard redirect
	return c.Redirect(http.StatusSeeOther, "/")
}

// DevLoginAsAdminHandler is a development-only endpoint to login as the seeded admin user
func (h *Handler) DevLoginAsAdminHandler(c echo.Context) error {
	ctx := context.Background()

	// This is the seeded admin user ID from sql/seed.sql
	adminUserID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	adminUUID, err := uuid.Parse(adminUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid admin user ID")
	}

	// Get the admin user from database
	adminUser, err := h.UserService.GetUserByID(ctx, adminUUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Admin user not found. Please run sql/seed.sql first.")
	}

	// Verify the user is actually an admin
	if !adminUser.IsAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "User is not an admin")
	}

	// Create session with both user auth and admin auth flags
	session, _ := middleware.GetSession(c)
	session.Values["user_id"] = adminUser.ID.String()
	session.Values["username"] = adminUser.Username
	if adminUser.Email != nil {
		session.Values["email"] = *adminUser.Email
	}
	session.Values["is_anonymous"] = false
	session.Values["is_admin"] = true

	if err := middleware.SaveSession(c, session); err != nil {
		log.Printf("Failed to save session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	log.Printf("✅ Dev login successful for admin user: %s", adminUser.Username)

	// Redirect to admin dashboard
	return c.Redirect(http.StatusSeeOther, "/admin")
}

// CreateAnonymousHandler creates an anonymous user
func (h *Handler) CreateAnonymousHandler(c echo.Context) error {
	ctx := context.Background()

	log.Printf("Creating anonymous user...")

	user, err := h.UserService.CreateAnonymousUser(ctx)
	if err != nil {
		log.Printf("ERROR creating anonymous user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create anonymous user: %v", err))
	}

	log.Printf("Anonymous user created successfully: %s", user.ID.String())

	// Save to session
	session, _ := middleware.GetSession(c)
	session.Values["user_id"] = user.ID.String()
	session.Values["username"] = user.Username
	session.Values["email"] = "" // Anonymous users don't have email
	session.Values["is_anonymous"] = true
	if err := middleware.SaveSession(c, session); err != nil {
		log.Printf("ERROR saving session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// OAuthGoogleHandler initiates Google OAuth flow
func (h *Handler) OAuthGoogleHandler(c echo.Context) error {
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initialize auth service")
	}

	authURL, err := authService.GetOAuthURL(services.ProviderGoogle)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate OAuth URL")
	}

	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// OAuthFacebookHandler initiates Facebook OAuth flow
func (h *Handler) OAuthFacebookHandler(c echo.Context) error {
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initialize auth service")
	}

	authURL, err := authService.GetOAuthURL(services.ProviderFacebook)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate OAuth URL")
	}

	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// OAuthGithubHandler initiates GitHub OAuth flow
func (h *Handler) OAuthGithubHandler(c echo.Context) error {
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initialize auth service")
	}

	authURL, err := authService.GetOAuthURL(services.ProviderGithub)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate OAuth URL")
	}

	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// OAuthCallbackHandler handles OAuth callbacks from Supabase
func (h *Handler) OAuthCallbackHandler(c echo.Context) error {
	// Check if tokens are in query params (PKCE flow)
	accessToken := c.QueryParam("access_token")
	refreshToken := c.QueryParam("refresh_token")

	if accessToken != "" {
		// Process tokens directly
		return h.processOAuthTokens(c, accessToken, refreshToken)
	}

	// No tokens in query, render callback page to extract from fragment (implicit flow)
	data := &TemplateData{
		Title: "Completing Login...",
	}

	return h.RenderTemplate(c, "auth/oauth-callback.html", data)
}

// OAuthTokenHandler receives tokens from the client-side script
func (h *Handler) OAuthTokenHandler(c echo.Context) error {
	var tokenData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Bind(&tokenData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	return h.processOAuthTokens(c, tokenData.AccessToken, tokenData.RefreshToken)
}

// processOAuthTokens is a helper to process OAuth tokens and create session
func (h *Handler) processOAuthTokens(c echo.Context, accessToken, refreshToken string) error {
	ctx := c.Request().Context()

	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		log.Printf("Failed to initialize auth service: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication failed")
	}

	// Get user info from access token
	oauthUser, err := authService.GetUserFromAccessToken(ctx, accessToken)
	if err != nil {
		log.Printf("Failed to get user from token: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication failed")
	}

	// Create or update user in our database
	user, err := authService.CreateOrUpdateUserFromOAuth(ctx, oauthUser)
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save user")
	}

	// Save to session
	session, _ := middleware.GetSession(c)
	session.Values["user_id"] = user.ID.String()
	session.Values["username"] = user.Username
	if user.Email != nil {
		session.Values["email"] = *user.Email
	}
	session.Values["is_anonymous"] = false
	session.Values["is_admin"] = user.IsAdmin
	session.Values["access_token"] = accessToken
	if refreshToken != "" {
		session.Values["refresh_token"] = refreshToken
	}

	if err := middleware.SaveSession(c, session); err != nil {
		log.Printf("Failed to save session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	log.Printf("OAuth login successful for user: %s (%s)", user.ID, *user.Email)

	// Return success for JSON requests, redirect for browser
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.NoContent(http.StatusOK)
	}

	return c.Redirect(http.StatusSeeOther, "/")
}
