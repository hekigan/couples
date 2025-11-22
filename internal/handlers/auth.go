package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/services"
)

// LoginHandler displays the login page
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Login - Couple Card Game",
	}
	h.RenderTemplate(w, "auth/login.html", data)
}

// LogoutHandler logs out the user
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "couple-card-game-session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DevLoginAsAdminHandler is a development-only endpoint to login as the seeded admin user
func (h *Handler) DevLoginAsAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// This is the seeded admin user ID from sql/seed.sql
	adminUserID := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	adminUUID, err := uuid.Parse(adminUserID)
	if err != nil {
		http.Error(w, "Invalid admin user ID", http.StatusInternalServerError)
		return
	}

	// Get the admin user from database
	adminUser, err := h.UserService.GetUserByID(ctx, adminUUID)
	if err != nil {
		http.Error(w, "Admin user not found. Please run sql/seed.sql first.", http.StatusNotFound)
		return
	}

	// Verify the user is actually an admin
	if !adminUser.IsAdmin {
		http.Error(w, "User is not an admin", http.StatusForbidden)
		return
	}

	// Create session with both user auth and admin auth flags
	session, _ := middleware.Store.Get(r, "couple-card-game-session")
	session.Values["user_id"] = adminUser.ID.String()
	session.Values["username"] = adminUser.Username
	if adminUser.Email != nil {
		session.Values["email"] = *adminUser.Email
	}
	session.Values["is_anonymous"] = false
	session.Values["is_admin"] = true
	session.Values["admin_authenticated"] = true // Required by AdminPasswordGate

	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Dev login successful for admin user: %s", adminUser.Username)

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// CreateAnonymousHandler creates an anonymous user
func (h *Handler) CreateAnonymousHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	log.Printf("Creating anonymous user...")
	
	user, err := h.UserService.CreateAnonymousUser(ctx)
	if err != nil {
		log.Printf("ERROR creating anonymous user: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create anonymous user: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Anonymous user created successfully: %s", user.ID.String())

	// Save to session
	session, _ := middleware.Store.Get(r, "couple-card-game-session")
	session.Values["user_id"] = user.ID.String()
	session.Values["username"] = user.Username
	session.Values["email"] = "" // Anonymous users don't have email
	session.Values["is_anonymous"] = true
	if err := session.Save(r, w); err != nil {
		log.Printf("ERROR saving session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// OAuthGoogleHandler initiates Google OAuth flow
func (h *Handler) OAuthGoogleHandler(w http.ResponseWriter, r *http.Request) {
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		http.Error(w, "Failed to initialize auth service", http.StatusInternalServerError)
		return
	}

	authURL, err := authService.GetOAuthURL(services.ProviderGoogle)
	if err != nil {
		http.Error(w, "Failed to generate OAuth URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// OAuthFacebookHandler initiates Facebook OAuth flow
func (h *Handler) OAuthFacebookHandler(w http.ResponseWriter, r *http.Request) {
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		http.Error(w, "Failed to initialize auth service", http.StatusInternalServerError)
		return
	}

	authURL, err := authService.GetOAuthURL(services.ProviderFacebook)
	if err != nil {
		http.Error(w, "Failed to generate OAuth URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// OAuthGithubHandler initiates GitHub OAuth flow
func (h *Handler) OAuthGithubHandler(w http.ResponseWriter, r *http.Request) {
	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		http.Error(w, "Failed to initialize auth service", http.StatusInternalServerError)
		return
	}

	authURL, err := authService.GetOAuthURL(services.ProviderGithub)
	if err != nil {
		http.Error(w, "Failed to generate OAuth URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// OAuthCallbackHandler handles OAuth callbacks from Supabase
func (h *Handler) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Check if tokens are in query params (PKCE flow)
	accessToken := r.URL.Query().Get("access_token")
	refreshToken := r.URL.Query().Get("refresh_token")

	if accessToken != "" {
		// Process tokens directly
		h.processOAuthTokens(w, r, accessToken, refreshToken)
		return
	}

	// No tokens in query, render callback page to extract from fragment (implicit flow)
	data := &TemplateData{
		Title: "Completing Login...",
	}

	h.RenderTemplate(w, "auth/oauth-callback.html", data)
}

// OAuthTokenHandler receives tokens from the client-side script
func (h *Handler) OAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	var tokenData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tokenData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.processOAuthTokens(w, r, tokenData.AccessToken, tokenData.RefreshToken)
}

// processOAuthTokens is a helper to process OAuth tokens and create session
func (h *Handler) processOAuthTokens(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string) {
	ctx := r.Context()

	authService, err := services.NewAuthService(h.UserService.GetSupabaseClient())
	if err != nil {
		log.Printf("Failed to initialize auth service: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Get user info from access token
	oauthUser, err := authService.GetUserFromAccessToken(ctx, accessToken)
	if err != nil {
		log.Printf("Failed to get user from token: %v", err)
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Create or update user in our database
	user, err := authService.CreateOrUpdateUserFromOAuth(ctx, oauthUser)
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	// Save to session
	session, _ := middleware.Store.Get(r, "couple-card-game-session")
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

	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	log.Printf("OAuth login successful for user: %s (%s)", user.ID, *user.Email)

	// Return success for JSON requests, redirect for browser
	if r.Header.Get("Content-Type") == "application/json" {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

