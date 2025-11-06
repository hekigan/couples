package services

import (
	"context"
	"encoding/json"
	"fmt"

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

// GetFriends retrieves all accepted friends for a user
func (s *FriendService) GetFriends(ctx context.Context, userID uuid.UUID) ([]models.FriendWithUserInfo, error) {
	// Query friends where user is the sender (user_id)
	data1, _, err1 := s.client.From("friends").
		Select("*", "", false).
		Eq("user_id", userID.String()).
		Eq("status", "accepted").
		Execute()

	// Query friends where user is the receiver (friend_id)
	data2, _, err2 := s.client.From("friends").
		Select("*", "", false).
		Eq("friend_id", userID.String()).
		Eq("status", "accepted").
		Execute()

	var friends1, friends2 []models.Friend
	if err1 == nil {
		json.Unmarshal(data1, &friends1)
	}
	if err2 == nil {
		json.Unmarshal(data2, &friends2)
	}

	// Combine and enrich with user info
	var result []models.FriendWithUserInfo

	// Process friends where current user is user_id
	for _, friend := range friends1 {
		userInfo, err := s.getUserInfo(ctx, friend.FriendID)
		if err != nil {
			continue
		}
		result = append(result, models.FriendWithUserInfo{
			ID:        friend.ID,
			UserID:    friend.UserID,
			FriendID:  friend.FriendID,
			Username:  userInfo.Username,
			Name:      userInfo.Name,
			Status:    friend.Status,
			CreatedAt: friend.CreatedAt,
		})
	}

	// Process friends where current user is friend_id
	for _, friend := range friends2 {
		userInfo, err := s.getUserInfo(ctx, friend.UserID)
		if err != nil {
			continue
		}
		result = append(result, models.FriendWithUserInfo{
			ID:        friend.ID,
			UserID:    friend.UserID,
			FriendID:  friend.FriendID,
			Username:  userInfo.Username,
			Name:      userInfo.Name,
			Status:    friend.Status,
			CreatedAt: friend.CreatedAt,
		})
	}

	return result, nil
}

// GetPendingRequests retrieves all pending friend requests for a user (requests they received)
func (s *FriendService) GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]models.FriendWithUserInfo, error) {
	data, _, err := s.client.From("friends").
		Select("*", "", false).
		Eq("friend_id", userID.String()).
		Eq("status", "pending").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending requests: %w", err)
	}

	var requests []models.Friend
	if err := json.Unmarshal(data, &requests); err != nil {
		return nil, fmt.Errorf("failed to parse requests: %w", err)
	}

	// Enrich with sender info
	var result []models.FriendWithUserInfo
	for _, req := range requests {
		userInfo, err := s.getUserInfo(ctx, req.UserID)
		if err != nil {
			continue
		}
		result = append(result, models.FriendWithUserInfo{
			ID:        req.ID,
			UserID:    req.UserID,
			FriendID:  req.FriendID,
			Username:  userInfo.Username,
			Name:      userInfo.Name,
			Status:    req.Status,
			CreatedAt: req.CreatedAt,
		})
	}

	return result, nil
}

// GetSentRequests retrieves all pending friend requests sent by a user
func (s *FriendService) GetSentRequests(ctx context.Context, userID uuid.UUID) ([]models.FriendWithUserInfo, error) {
	data, _, err := s.client.From("friends").
		Select("*", "", false).
		Eq("user_id", userID.String()).
		Eq("status", "pending").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch sent requests: %w", err)
	}

	var requests []models.Friend
	if err := json.Unmarshal(data, &requests); err != nil {
		return nil, fmt.Errorf("failed to parse requests: %w", err)
	}

	// Enrich with receiver info
	var result []models.FriendWithUserInfo
	for _, req := range requests {
		userInfo, err := s.getUserInfo(ctx, req.FriendID)
		if err != nil {
			continue
		}
		result = append(result, models.FriendWithUserInfo{
			ID:        req.ID,
			UserID:    req.UserID,
			FriendID:  req.FriendID,
			Username:  userInfo.Username,
			Name:      userInfo.Name,
			Status:    req.Status,
			CreatedAt: req.CreatedAt,
		})
	}

	return result, nil
}

// CreateFriendRequest creates a new friend request
func (s *FriendService) CreateFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) error {
	// Check if request already exists (in either direction)
	existing, err := s.checkExistingFriendship(ctx, senderID, receiverID)
	if err == nil && existing != nil {
		if existing.Status == "accepted" {
			return fmt.Errorf("you are already friends with this user")
		}
		return fmt.Errorf("a friend request already exists")
	}

	// Create new friend request
	friendData := map[string]interface{}{
		"id":        uuid.New().String(),
		"user_id":   senderID.String(),
		"friend_id": receiverID.String(),
		"status":    "pending",
	}

	_, _, err = s.client.From("friends").Insert(friendData, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to create friend request: %w", err)
	}

	return nil
}

// AcceptFriendRequest accepts a friend request
func (s *FriendService) AcceptFriendRequest(ctx context.Context, requestID uuid.UUID) error {
	// Update status to accepted
	data := map[string]interface{}{
		"status": "accepted",
	}

	_, _, err := s.client.From("friends").
		Update(data, "", "").
		Eq("id", requestID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to accept friend request: %w", err)
	}

	return nil
}

// DeclineFriendRequest declines a friend request
func (s *FriendService) DeclineFriendRequest(ctx context.Context, requestID uuid.UUID) error {
	// Update status to declined
	data := map[string]interface{}{
		"status": "declined",
	}

	_, _, err := s.client.From("friends").
		Update(data, "", "").
		Eq("id", requestID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to decline friend request: %w", err)
	}

	return nil
}

// RemoveFriend removes a friendship (deletes the record)
func (s *FriendService) RemoveFriend(ctx context.Context, friendshipID uuid.UUID) error {
	_, _, err := s.client.From("friends").
		Delete("", "").
		Eq("id", friendshipID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	return nil
}

// GetFriendshipByID retrieves a specific friendship/request by ID
func (s *FriendService) GetFriendshipByID(ctx context.Context, friendshipID uuid.UUID) (*models.Friend, error) {
	data, _, err := s.client.From("friends").
		Select("*", "", false).
		Eq("id", friendshipID.String()).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("friendship not found: %w", err)
	}

	var friend models.Friend
	if err := json.Unmarshal(data, &friend); err != nil {
		return nil, fmt.Errorf("failed to parse friendship: %w", err)
	}

	return &friend, nil
}

// SearchUsersByUsername searches for users by username (for adding friends)
func (s *FriendService) SearchUsersByUsername(ctx context.Context, query string) ([]models.User, error) {
	// Use ilike for case-insensitive search
	data, _, err := s.client.From("users").
		Select("id,username,name,is_anonymous", "", false).
		Ilike("username", fmt.Sprintf("%%%s%%", query)).
		Limit(10, "").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse users: %w", err)
	}

	return users, nil
}

// Helper function to get user info
func (s *FriendService) getUserInfo(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	data, _, err := s.client.From("users").
		Select("id,username,name", "", false).
		Eq("id", userID.String()).
		Single().
		Execute()

	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Helper function to check if friendship already exists
func (s *FriendService) checkExistingFriendship(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Friend, error) {
	// Check direction 1: userID1 -> userID2
	data1, _, _ := s.client.From("friends").
		Select("*", "", false).
		Eq("user_id", userID1.String()).
		Eq("friend_id", userID2.String()).
		Single().
		Execute()

	if data1 != nil && len(data1) > 0 {
		var friend models.Friend
		if err := json.Unmarshal(data1, &friend); err == nil {
			return &friend, nil
		}
	}

	// Check direction 2: userID2 -> userID1
	data2, _, _ := s.client.From("friends").
		Select("*", "", false).
		Eq("user_id", userID2.String()).
		Eq("friend_id", userID1.String()).
		Single().
		Execute()

	if data2 != nil && len(data2) > 0 {
		var friend models.Friend
		if err := json.Unmarshal(data2, &friend); err == nil {
			return &friend, nil
		}
	}

	return nil, fmt.Errorf("not found")
}
