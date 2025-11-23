package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go"
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
	supabaseKey := os.Getenv("SUPABASE_KEY")
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")

	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/oauth/callback"
	}

	// Create GoTrue client for auth operations
	authClient := gotrue.New(supabaseURL, supabaseKey)

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
	ID     string
	Email  string
	Name   string
	Avatar string
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
		if name, ok := resp.UserMetadata["full_name"].(string); ok {
			oauthUser.Name = name
		} else if name, ok := resp.UserMetadata["name"].(string); ok {
			oauthUser.Name = name
		}

		if avatar, ok := resp.UserMetadata["avatar_url"].(string); ok {
			oauthUser.Avatar = avatar
		}
	}

	if oauthUser.Name == "" {
		oauthUser.Name = resp.Email
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
		// Update user if necessary (e.g., display name, avatar)
		updateData := map[string]interface{}{}
		if user.DisplayName == nil || *user.DisplayName != oauthUser.Name {
			updateData["display_name"] = oauthUser.Name
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
		DisplayName: &oauthUser.Name,
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

