package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/models"
	"github.com/supabase-community/supabase-go"
)

// FriendService handles friend-related operations
type FriendService struct {
	*BaseService
	client              *supabase.Client
	notificationService *NotificationService
	realtimeService     *RealtimeService
}

// NewFriendService creates a new friend service
func NewFriendService(
	client *supabase.Client,
	notificationService *NotificationService,
	realtimeService *RealtimeService,
) *FriendService {
	return &FriendService{
		BaseService:         NewBaseService(client, "FriendService"),
		client:              client,
		notificationService: notificationService,
		realtimeService:     realtimeService,
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
// Returns (autoAccepted, error) - autoAccepted is true if a reverse pending request was auto-accepted
func (s *FriendService) CreateFriendRequest(ctx context.Context, senderID, receiverID uuid.UUID) (bool, error) {
	// Check if request already exists (in either direction)
	existing, err := s.checkExistingFriendship(ctx, senderID, receiverID)
	if err == nil && existing != nil {
		if existing.Status == "accepted" {
			return false, fmt.Errorf("you are already friends with this user")
		}
		if existing.Status == "pending" {
			// If the other user already sent us a request, auto-accept it (mutual interest!)
			if existing.UserID == receiverID && existing.FriendID == senderID {
				data := map[string]interface{}{"status": "accepted"}
				if err := s.BaseService.UpdateRecord(ctx, "friends", existing.ID, data); err != nil {
					return false, fmt.Errorf("failed to accept existing request: %w", err)
				}
				// Notify the original sender that their request was accepted
				senderInfo, _ := s.getUserInfo(ctx, senderID)
				if senderInfo != nil && s.notificationService != nil {
					notification := &models.Notification{
						UserID:  receiverID, // Notify original sender
						Type:    models.NotificationTypeFriendAccepted,
						Title:   "Friend Request Accepted",
						Message: fmt.Sprintf("🎉 %s accepted your friend request!", senderInfo.Username),
						Link:    "/friends",
						Read:    false,
					}
					go func() {
						s.notificationService.CreateAndBroadcastNotification(context.Background(), notification, s.realtimeService)
					}()
				}
				return true, nil // Success - friendship established via auto-accept
			}
			// Same direction pending request already exists
			return false, fmt.Errorf("a friend request already exists")
		}
		// Any other status (shouldn't happen, but handle gracefully)
		return false, fmt.Errorf("unexpected friendship status: %s", existing.Status)
	}

	// Create new friend request (no existing record)
	friendData := map[string]interface{}{
		"id":        uuid.New().String(),
		"user_id":   senderID.String(),
		"friend_id": receiverID.String(),
		"status":    "pending",
	}

	if err := s.BaseService.InsertRecord(ctx, "friends", friendData); err != nil {
		return false, err
	}

	// Get sender info for notification
	senderInfo, err := s.getUserInfo(ctx, senderID)
	if err != nil {
		// Don't fail the request if notification fails
		return false, nil
	}

	// Create notification synchronously to ensure it's in the database
	notification := &models.Notification{
		UserID:  receiverID,
		Type:    models.NotificationTypeFriendRequest,
		Title:   "New Friend Request",
		Message: fmt.Sprintf("%s sent you a friend request", senderInfo.Username),
		Link:    "/friends",
		Read:    false,
	}

	// Create notification synchronously (must succeed for consistency)
	if err := s.notificationService.CreateAndBroadcastNotification(ctx, notification, s.realtimeService); err != nil {
		fmt.Printf("⚠️ Failed to create notification: %v\n", err)
		// Don't fail the friend request, but log the error
	}

	return false, nil // Normal invitation sent (not auto-accepted)
}

// AcceptFriendRequest accepts a friend request and returns the Friend record
func (s *FriendService) AcceptFriendRequest(
	ctx context.Context,
	requestID uuid.UUID,
	accepterID uuid.UUID,
) (*models.Friend, error) {
	// Get friend request details BEFORE updating
	friendRequest, err := s.GetFriendshipByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Verify accepter is the receiver (security check)
	if friendRequest.FriendID != accepterID {
		return nil, fmt.Errorf("you are not authorized to accept this request")
	}

	// Update status to accepted
	data := map[string]interface{}{
		"status": "accepted",
	}

	if err := s.BaseService.UpdateRecord(ctx, "friends", requestID, data); err != nil {
		return nil, err
	}

	// Get accepter info for notification message
	accepterInfo, err := s.getUserInfo(ctx, accepterID)
	if err != nil {
		// Don't fail if notification fails
		return friendRequest, nil
	}

	// Create notification for sender (async with background context)
	notification := &models.Notification{
		UserID:  friendRequest.UserID, // Notify the sender
		Type:    models.NotificationTypeFriendAccepted,
		Title:   "Friend Request Accepted",
		Message: fmt.Sprintf("🎉 %s accepted your friend request!", accepterInfo.Username),
		Link:    "/friends",
		Read:    false,
	}

	go func() {
		if err := s.notificationService.CreateAndBroadcastNotification(context.Background(), notification, s.realtimeService); err != nil {
			fmt.Printf("⚠️ Failed to create acceptance notification: %v\n", err)
		}
	}()

	// Mark friend_request notification as read, THEN broadcast badge update (in sequence)
	go func() {
		// First mark notification as read
		if err := s.notificationService.MarkFriendRequestNotificationAsRead(context.Background(), accepterID, friendRequest.UserID); err != nil {
			fmt.Printf("⚠️ Failed to mark notification as read: %v\n", err)
		}
		// Then broadcast updated badge count (after notification is marked read)
		count, _ := s.notificationService.GetNotificationBadgeCount(context.Background(), accepterID)
		s.realtimeService.BroadcastToUser(accepterID, RealtimeEvent{
			Type: "badge_update",
			Data: map[string]interface{}{"count": count},
		})
	}()

	return friendRequest, nil
}

// DeclineFriendRequest declines a friend request by deleting the record
func (s *FriendService) DeclineFriendRequest(
	ctx context.Context,
	requestID uuid.UUID,
	declinerID uuid.UUID,
) error {
	// Get friend request details for verification
	friendRequest, err := s.GetFriendshipByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Verify decliner is the receiver (security check)
	if friendRequest.FriendID != declinerID {
		return fmt.Errorf("you are not authorized to decline this request")
	}

	// Delete the friend request record (allows sender to invite again in the future)
	if err := s.BaseService.DeleteRecord(ctx, "friends", requestID); err != nil {
		return err
	}

	// Mark friend_request notification as read, THEN broadcast badge update (in sequence)
	go func() {
		// First mark notification as read
		if err := s.notificationService.MarkFriendRequestNotificationAsRead(context.Background(), declinerID, friendRequest.UserID); err != nil {
			fmt.Printf("⚠️ Failed to mark notification as read: %v\n", err)
		}
		// Then broadcast updated badge count (after notification is marked read)
		count, _ := s.notificationService.GetNotificationBadgeCount(context.Background(), declinerID)
		s.realtimeService.BroadcastToUser(declinerID, RealtimeEvent{
			Type: "badge_update",
			Data: map[string]interface{}{"count": count},
		})
	}()

	// NO notification to sender (silent decline per requirement)

	return nil
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
		Select("id,username", "", false).
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

// GetUserByEmail fetches user by email
func (s *FriendService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var users []models.User
	if err := s.BaseService.GetRecords(ctx, "users", map[string]interface{}{
		"email": email,
	}, &users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

// CreateFriendRequestByEmail creates friend request via email
// Security: Always returns success to prevent email enumeration
func (s *FriendService) CreateFriendRequestByEmail(
	ctx context.Context,
	senderID uuid.UUID,
	senderUsername string,
	recipientEmail string,
	emailService *EmailService,
	notificationService *NotificationService,
) error {
	// 1. Check if user exists with this email
	user, err := s.GetUserByEmail(ctx, recipientEmail)

	// 2a. User exists → create friend request + send email + create notification
	if err == nil && user != nil {
		// Check for existing friendship
		existing, _ := s.checkExistingFriendship(ctx, senderID, user.ID)
		if existing != nil {
			// Already friends or pending - silently succeed
			return nil
		}

		// Create friend request (this now handles notification creation and broadcasting)
		if _, err := s.CreateFriendRequest(ctx, senderID, user.ID); err != nil {
			return err
		}

		// Send email notification (async) - use background context to avoid cancellation
		go emailService.SendFriendInvitationToExistingUser(context.Background(), recipientEmail, senderUsername)

		return nil
	}

	// 2b. User doesn't exist → create email invitation + send join email
	token := generateSecureToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	invitation := map[string]interface{}{
		"id":              uuid.New().String(),
		"sender_id":       senderID.String(),
		"recipient_email": recipientEmail,
		"token":           token,
		"status":          "pending",
		"expires_at":      expiresAt,
	}

	if err := s.BaseService.InsertRecord(ctx, "friend_email_invitations", invitation); err != nil {
		// Duplicate invitation - silently succeed (security)
		return nil
	}

	// Send join invitation email (async) - use background context to avoid cancellation
	go emailService.SendFriendInvitationToNewUser(context.Background(), recipientEmail, senderUsername, token)

	return nil
}

// AcceptEmailInvitation accepts email invitation and creates friendship
func (s *FriendService) AcceptEmailInvitation(ctx context.Context, token string, newUserID uuid.UUID) error {
	// Find invitation by token
	var invitations []models.FriendEmailInvitation
	if err := s.BaseService.GetRecords(ctx, "friend_email_invitations", map[string]interface{}{
		"token":  token,
		"status": "pending",
	}, &invitations); err != nil || len(invitations) == 0 {
		return fmt.Errorf("invalid or expired invitation")
	}

	invitation := invitations[0]

	// Check expiration
	if time.Now().After(invitation.ExpiresAt) {
		// Mark as expired
		s.BaseService.UpdateRecord(ctx, "friend_email_invitations", invitation.ID, map[string]interface{}{
			"status": "expired",
		})
		return fmt.Errorf("invitation has expired")
	}

	// Create friendship
	if _, err := s.CreateFriendRequest(ctx, invitation.SenderID, newUserID); err != nil {
		return err
	}

	// Mark invitation as accepted
	now := time.Now()
	s.BaseService.UpdateRecord(ctx, "friend_email_invitations", invitation.ID, map[string]interface{}{
		"status":      "accepted",
		"accepted_at": now,
	})

	return nil
}

// generateSecureToken generates cryptographically secure token
func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
