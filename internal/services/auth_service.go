package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
	"github.com/hekigan/couples/internal/models"
)

// AuthService handles authentication operations with Supabase Auth
type AuthService struct {
	supabase    *supabase.Client
	authClient  gotrue.Client
	redirectURL string
}

// NewAuthService creates a new authentication service
func NewAuthService(client *supabase.Client) (*AuthService, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	// Use SUPABASE_ANON_KEY for auth operations (signup, login)
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")

	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/oauth/callback"
	}

	// Extract project reference from URL (e.g., "fsymkxnnmwyhafcdnpzx" from "https://fsymkxnnmwyhafcdnpzx.supabase.co")
	projectRef := strings.TrimPrefix(supabaseURL, "https://")
	projectRef = strings.TrimPrefix(projectRef, "http://")
	projectRef = strings.Split(projectRef, ".")[0]

	// Create GoTrue client for auth operations
	// gotrue.New expects just the project reference, not the full URL
	authClient := gotrue.New(projectRef, supabaseKey)

	return &AuthService{
		supabase:    client,
		authClient:  authClient,
		redirectURL: redirectURL,
	}, nil
}

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	ProviderGoogle   OAuthProvider = "google"
	ProviderFacebook OAuthProvider = "facebook"
	ProviderGithub   OAuthProvider = "github"
)

// OAuthUser represents a user obtained from an OAuth provider
type OAuthUser struct {
	ID       string
	Email    string
	Username string
	Avatar   string
}

// OAuthSession represents an OAuth session
type OAuthSession struct {
	AccessToken  string
	RefreshToken string
	User         *OAuthUser
	ExpiresIn    int
}

// GetOAuthURL generates the URL to initiate the OAuth flow
func (s *AuthService) GetOAuthURL(provider OAuthProvider) (string, error) {
	// For gotrue-go, we need to construct the OAuth URL manually
	baseURL := os.Getenv("SUPABASE_URL")
	providerURL := fmt.Sprintf("%s/auth/v1/authorize?provider=%s&redirect_to=%s", 
		baseURL, string(provider), s.redirectURL)
	return providerURL, nil
}

// GetUserFromAccessToken retrieves user information from an access token
func (s *AuthService) GetUserFromAccessToken(ctx context.Context, accessToken string) (*OAuthUser, error) {
	client := s.authClient.WithToken(accessToken)
	resp, err := client.GetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	oauthUser := &OAuthUser{
		ID:    resp.ID.String(),
		Email: resp.Email,
	}

	if resp.UserMetadata != nil {
		// Map Supabase display_name/username to Username field
		if username, ok := resp.UserMetadata["username"].(string); ok {
			oauthUser.Username = username
		} else if displayName, ok := resp.UserMetadata["display_name"].(string); ok {
			oauthUser.Username = displayName
		} else if name, ok := resp.UserMetadata["full_name"].(string); ok {
			oauthUser.Username = name
		} else if name, ok := resp.UserMetadata["name"].(string); ok {
			oauthUser.Username = name
		}

		if avatar, ok := resp.UserMetadata["avatar_url"].(string); ok {
			oauthUser.Avatar = avatar
		}
	}

	// Fallback to email prefix if no username found
	if oauthUser.Username == "" {
		oauthUser.Username = strings.Split(resp.Email, "@")[0]
	}

	return oauthUser, nil
}

// CreateOrUpdateUserFromOAuth creates or updates a user in our database from OAuth data
func (s *AuthService) CreateOrUpdateUserFromOAuth(ctx context.Context, oauthUser *OAuthUser) (*models.User, error) {
	userID, err := uuid.Parse(oauthUser.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var existingUsers []models.User
	query := s.supabase.From("users").
		Select("*", "exact", false).
		Eq("id", userID.String())

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if err := json.Unmarshal(data, &existingUsers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	if len(existingUsers) > 0 {
		user := &existingUsers[0]
		// Update user if necessary (e.g., username, avatar)
		updateData := map[string]interface{}{}
		if user.Username != oauthUser.Username {
			updateData["username"] = oauthUser.Username
		}
		if user.Email == nil || *user.Email != oauthUser.Email {
			updateData["email"] = oauthUser.Email
		}
		if oauthUser.Avatar != "" && (user.AvatarURL == nil || *user.AvatarURL != oauthUser.Avatar) {
			updateData["avatar_url"] = oauthUser.Avatar
		}

		if len(updateData) > 0 {
			updatedData, _, err := s.supabase.From("users").
				Update(updateData, "", "").
				Eq("id", userID.String()).
				Execute()
			if err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
			var updatedUsers []models.User
			if err := json.Unmarshal(updatedData, &updatedUsers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal updated user: %w", err)
			}
			if len(updatedUsers) > 0 {
				return &updatedUsers[0], nil
			}
		}
		return user, nil
	}

	// User does not exist, create new user
	newUser := &models.User{
		ID:          userID,
		Email:       &oauthUser.Email,
		Username:    oauthUser.Username, // Username from Supabase metadata
		IsAnonymous: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if oauthUser.Avatar != "" {
		newUser.AvatarURL = &oauthUser.Avatar
	}

	insertData, _, err := s.supabase.From("users").
		Insert(newUser, false, "", "", "").
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	var createdUsers []models.User
	if err := json.Unmarshal(insertData, &createdUsers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created user: %w", err)
	}
	if len(createdUsers) == 0 {
		return nil, fmt.Errorf("failed to create user: no data returned")
	}

	return &createdUsers[0], nil
}

// RefreshAccessToken refreshes an expired access token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*OAuthSession, error) {
	resp, err := s.authClient.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	user, err := s.GetUserFromAccessToken(ctx, resp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user after refresh: %w", err)
	}

	return &OAuthSession{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User:         user,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

// SignOut signs out a user
func (s *AuthService) SignOut(ctx context.Context, accessToken string) error {
	client := s.authClient.WithToken(accessToken)
	return client.Logout()
}

// LoginWithPassword authenticates a user with email/password
func (s *AuthService) LoginWithPassword(ctx context.Context, email, password string) (*OAuthSession, error) {
	// Use gotrue-go SignInWithEmailPassword convenience method
	resp, err := s.authClient.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	user, err := s.GetUserFromAccessToken(ctx, resp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &OAuthSession{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User:         user,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

// SignupWithPassword creates a new user account with email/password
func (s *AuthService) SignupWithPassword(ctx context.Context, email, password, username string) (*OAuthSession, error) {
	// Create user in Supabase Auth
	resp, err := s.authClient.Signup(types.SignupRequest{
		Email:    email,
		Password: password,
		Data: map[string]interface{}{
			"username":     username,
			"display_name": username,
		},
	})
	if err != nil {
		fmt.Printf("🔴 Supabase Auth signup error: %v\n", err)
		fmt.Printf("🔴 Error type: %T\n", err)
		return nil, fmt.Errorf("signup failed: %w", err)
	}

	// Get user details from access token
	user, err := s.GetUserFromAccessToken(ctx, resp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &OAuthSession{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User:         user,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

