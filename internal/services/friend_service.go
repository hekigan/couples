package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/hekigan/couples/internal/models"
)

// FriendService handles friend-related operations
type FriendService struct {
	*BaseService
	client *supabase.Client
}

// NewFriendService creates a new friend service
func NewFriendService(client *supabase.Client) *FriendService {
	return &FriendService{
		BaseService: NewBaseService(client, "FriendService"),
		client:      client,
	}
}

// GetFriends retrieves all accepted friends for a user
func (s *FriendService) GetFriends(ctx context.Context, userID uuid.UUID) ([]models.FriendWithUserInfo, error) {
	// Query friends where user is the sender (user_id)
	var friends1 []models.Friend
	if err := s.BaseService.GetRecords(ctx, "friends", map[string]interface{}{
		"user_id": userID.String(),
		"status":  "accepted",
	}, &friends1); err != nil {
		friends1 = []models.Friend{}
	}

	// Query friends where user is the receiver (friend_id)
	var friends2 []models.Friend
	if err := s.BaseService.GetRecords(ctx, "friends", map[string]interface{}{
		"friend_id": userID.String(),
		"status":    "accepted",
	}, &friends2); err != nil {
		friends2 = []models.Friend{}
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
			Status:    friend.Status,
			CreatedAt: friend.CreatedAt,
		})
	}

	return result, nil
}

// GetPendingRequests retrieves all pending friend requests for a user (requests they received)
func (s *FriendService) GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]models.FriendWithUserInfo, error) {
	var requests []models.Friend
	if err := s.BaseService.GetRecords(ctx, "friends", map[string]interface{}{
		"friend_id": userID.String(),
		"status":    "pending",
	}, &requests); err != nil {
		return nil, err
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
			Status:    req.Status,
			CreatedAt: req.CreatedAt,
		})
	}

	return result, nil
}

// GetSentRequests retrieves all pending friend requests sent by a user
func (s *FriendService) GetSentRequests(ctx context.Context, userID uuid.UUID) ([]models.FriendWithUserInfo, error) {
	var requests []models.Friend
	if err := s.BaseService.GetRecords(ctx, "friends", map[string]interface{}{
		"user_id": userID.String(),
		"status":  "pending",
	}, &requests); err != nil {
		return nil, err
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

	return s.BaseService.InsertRecord(ctx, "friends", friendData)
}

// AcceptFriendRequest accepts a friend request
func (s *FriendService) AcceptFriendRequest(ctx context.Context, requestID uuid.UUID) error {
	data := map[string]interface{}{
		"status": "accepted",
	}

	return s.BaseService.UpdateRecord(ctx, "friends", requestID, data)
}

// DeclineFriendRequest declines a friend request
func (s *FriendService) DeclineFriendRequest(ctx context.Context, requestID uuid.UUID) error {
	data := map[string]interface{}{
		"status": "declined",
	}

	return s.BaseService.UpdateRecord(ctx, "friends", requestID, data)
}

// RemoveFriend removes a friendship (deletes the record)
func (s *FriendService) RemoveFriend(ctx context.Context, friendshipID uuid.UUID) error {
	return s.BaseService.DeleteRecord(ctx, "friends", friendshipID)
}

// GetFriendshipByID retrieves a specific friendship/request by ID
func (s *FriendService) GetFriendshipByID(ctx context.Context, friendshipID uuid.UUID) (*models.Friend, error) {
	var friend models.Friend
	if err := s.BaseService.GetSingleRecord(ctx, "friends", friendshipID, &friend); err != nil {
		return nil, err
	}
	return &friend, nil
}

// SearchUsersByUsername searches for users by username (for adding friends)
func (s *FriendService) SearchUsersByUsername(ctx context.Context, query string) ([]models.User, error) {
	// Custom query - uses Ilike and Limit which are not supported by BaseService
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
	// Custom query - only selecting specific fields, not all (*)
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
	// Custom query - complex logic checking both directions
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
