package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/yourusername/couple-card-game/internal/models"
)

// FriendService handles friend-related operations
type FriendService struct {
	client *supabase.Client
}

// NewFriendService creates a new friend service
func NewFriendService(client *supabase.Client) *FriendService {
	return &FriendService{
		client: client,
	}
}

// GetFriends retrieves all friends for a user (stub)
func (s *FriendService) GetFriends(ctx context.Context, userID uuid.UUID) ([]models.Friend, error) {
	// TODO: Implement proper Supabase query
	return []models.Friend{}, nil
}

// CreateFriendRequest creates a new friend request (stub)
func (s *FriendService) CreateFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	// TODO: Implement proper Supabase insert
	return nil
}

// AcceptFriendRequest accepts a friend request (stub)
func (s *FriendService) AcceptFriendRequest(ctx context.Context, requestID uuid.UUID) error {
	// TODO: Implement proper Supabase update and friend record creation
	return nil
}
