package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/postgrest-go"
	supabase "github.com/supabase-community/supabase-go"

	"github.com/hekigan/couples/internal/models"
)

type NotificationService struct {
	*BaseService
	client *supabase.Client
}

func NewNotificationService(client *supabase.Client) *NotificationService {
	return &NotificationService{
		BaseService: NewBaseService(client, "NotificationService"),
		client:      client,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, notification *models.Notification) error {
	notification.ID = uuid.New()
	notification.CreatedAt = time.Now()

	data := map[string]interface{}{
		"id":         notification.ID.String(),
		"user_id":    notification.UserID.String(),
		"type":       notification.Type,
		"title":      notification.Title,
		"message":    notification.Message,
		"link":       notification.Link,
		"read":       notification.Read,
		"created_at": notification.CreatedAt,
	}

	return s.BaseService.InsertRecord(ctx, "notifications", data)
}

// GetUserNotifications gets all notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	filters := map[string]interface{}{
		"user_id": userID.String(),
	}

	var notifications []models.Notification
	if err := s.BaseService.GetRecords(ctx, "notifications", filters, &notifications); err != nil {
		return nil, err
	}

	// Apply limit if specified
	if limit > 0 && len(notifications) > limit {
		notifications = notifications[:limit]
	}

	return notifications, nil
}

// GetNotificationsForDisplay returns notifications for display in the header dropdown
// This includes actual notifications + synthetic notifications for pending friend requests
func (s *NotificationService) GetNotificationsForDisplay(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	// Get actual notifications (excluding friend_request type to avoid duplicates)
	data, _, err := s.client.From("notifications").
		Select("*", "", false).
		Eq("user_id", userID.String()).
		Neq("type", models.NotificationTypeFriendRequest).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Execute()

	if err != nil {
		return nil, err
	}

	var notifications []models.Notification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return nil, err
	}

	// Get pending friend requests and create synthetic notifications
	friendRequests, err := s.getPendingFriendRequestsForDisplay(ctx, userID)
	if err == nil {
		// Merge friend request notifications with other notifications
		notifications = append(notifications, friendRequests...)
	}

	// Sort by created_at descending (most recent first)
	// Use a simple bubble sort since the list should be small
	for i := 0; i < len(notifications)-1; i++ {
		for j := i + 1; j < len(notifications); j++ {
			if notifications[j].CreatedAt.After(notifications[i].CreatedAt) {
				notifications[i], notifications[j] = notifications[j], notifications[i]
			}
		}
	}

	// Apply limit if specified
	if limit > 0 && len(notifications) > limit {
		notifications = notifications[:limit]
	}

	return notifications, nil
}

// getPendingFriendRequestsForDisplay creates synthetic notifications for pending friend requests
func (s *NotificationService) getPendingFriendRequestsForDisplay(ctx context.Context, userID uuid.UUID) ([]models.Notification, error) {
	// Query pending friend requests where user is the receiver
	filters := map[string]interface{}{
		"friend_id": userID.String(),
		"status":    "pending",
	}

	var friendRequests []models.Friend
	if err := s.BaseService.GetRecords(ctx, "friends", filters, &friendRequests); err != nil {
		return nil, err
	}

	// Create synthetic notifications for each friend request
	notifications := make([]models.Notification, 0, len(friendRequests))
	for _, req := range friendRequests {
		// Get sender info
		senderInfo, err := s.getUserInfo(ctx, req.UserID)
		if err != nil {
			continue // Skip if we can't get user info
		}

		notifications = append(notifications, models.Notification{
			ID:        req.ID, // Use friend request ID
			UserID:    userID,
			Type:      models.NotificationTypeFriendRequest,
			Title:     "New Friend Request",
			Message:   fmt.Sprintf("%s sent you a friend request", senderInfo.Username),
			Link:      "/friends",
			Read:      false, // Pending requests are always unread
			CreatedAt: req.CreatedAt,
		})
	}

	return notifications, nil
}

// getUserInfo fetches basic user information
func (s *NotificationService) getUserInfo(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.BaseService.GetSingleRecord(ctx, "users", userID, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUnreadCount gets count of unread notifications
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	filters := map[string]interface{}{
		"user_id": userID.String(),
		"read":    "false",
	}

	return s.BaseService.CountRecords(ctx, "notifications", filters)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	data := map[string]interface{}{
		"read": true,
	}

	return s.BaseService.UpdateRecord(ctx, "notifications", notificationID, data)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	filters := map[string]interface{}{
		"user_id": userID.String(),
		"read":    "false",
	}

	data := map[string]interface{}{
		"read": true,
	}

	return s.BaseService.UpdateRecordsWithFilter(ctx, "notifications", filters, data)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID) error {
	return s.BaseService.DeleteRecord(ctx, "notifications", notificationID)
}

// CreateRoomInvitation creates a room invitation
func (s *NotificationService) CreateRoomInvitation(ctx context.Context, invitation *models.RoomInvitation) error {
	invitation.ID = uuid.New()
	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()
	invitation.Status = models.InvitationStatusPending

	data := map[string]interface{}{
		"id":         invitation.ID.String(),
		"room_id":    invitation.RoomID.String(),
		"inviter_id": invitation.InviterID.String(),
		"invitee_id": invitation.InviteeID.String(),
		"status":     invitation.Status,
		"created_at": invitation.CreatedAt,
		"updated_at": invitation.UpdatedAt,
	}

	return s.BaseService.InsertRecord(ctx, "room_invitations", data)
}

// GetRoomInvitation gets a specific room invitation
func (s *NotificationService) GetRoomInvitation(ctx context.Context, roomID, inviteeID uuid.UUID) (*models.RoomInvitation, error) {
	// Custom query with multiple filters - BaseService doesn't have a GetSingleRecordWithFilters method
	data, _, err := s.client.From("room_invitations").
		Select("*", "", false).
		Eq("room_id", roomID.String()).
		Eq("invitee_id", inviteeID.String()).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}

	var invitation models.RoomInvitation
	if err := json.Unmarshal(data, &invitation); err != nil {
		return nil, fmt.Errorf("failed to parse invitation: %w", err)
	}

	return &invitation, nil
}

// CancelRoomInvitation cancels a room invitation
func (s *NotificationService) CancelRoomInvitation(ctx context.Context, roomID, inviteeID uuid.UUID) error {
	filters := map[string]interface{}{
		"room_id":    roomID.String(),
		"invitee_id": inviteeID.String(),
	}

	return s.BaseService.DeleteRecordsWithFilter(ctx, "room_invitations", filters)
}

// CreateAndBroadcastNotification creates a notification in DB and broadcasts it via SSE
// This is the hybrid approach: persistence + instant delivery to online users
func (s *NotificationService) CreateAndBroadcastNotification(
	ctx context.Context,
	notification *models.Notification,
	realtimeService *RealtimeService,
) error {
	// Create notification record in DB for persistence
	if err := s.CreateNotification(ctx, notification); err != nil {
		return err
	}

	// Broadcast via SSE for instant delivery (non-blocking)
	go realtimeService.BroadcastToUser(notification.UserID, RealtimeEvent{
		Type: "notification",
		Data: notification,
	})

	return nil
}

// GetNotificationBadgeCount returns the count for the notification badge
// This combines unread notifications + pending friend requests
func (s *NotificationService) GetNotificationBadgeCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	return s.GetUnreadCountWithPendingFriends(ctx, userID)
}

// GetUnreadCountWithPendingFriends returns combined count of unread notifications + pending friend requests
// This ensures the badge shows all actionable items even if notification records weren't created
func (s *NotificationService) GetUnreadCountWithPendingFriends(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	// Count unread notifications EXCLUDING friend_request type (to avoid double-counting)
	// We'll count friend requests directly from the friends table
	nonFriendNotificationCount, err := s.getUnreadCountExcludingType(ctx, userID, models.NotificationTypeFriendRequest)
	if err != nil {
		return 0, err
	}

	// Count pending friend requests where user is the RECEIVER (friend_id)
	// This is the source of truth for friend requests
	pendingFriendsCount, err := s.countPendingFriendRequests(ctx, userID)
	if err != nil {
		// If friend count fails, still return notification count
		return nonFriendNotificationCount, nil
	}

	return nonFriendNotificationCount + pendingFriendsCount, nil
}

// getUnreadCountExcludingType counts unread notifications excluding a specific type
func (s *NotificationService) getUnreadCountExcludingType(
	ctx context.Context,
	userID uuid.UUID,
	excludeType string,
) (int, error) {
	// Query unread notifications excluding the specified type
	data, _, err := s.client.From("notifications").
		Select("id", "count", false).
		Eq("user_id", userID.String()).
		Eq("read", "false").
		Neq("type", excludeType).
		Execute()

	if err != nil {
		return 0, err
	}

	var notifications []models.Notification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return 0, err
	}

	return len(notifications), nil
}

// countPendingFriendRequests counts pending friend requests where user is the receiver
// This is separate to avoid double-counting if notifications exist
func (s *NotificationService) countPendingFriendRequests(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	filters := map[string]interface{}{
		"friend_id": userID.String(),
		"status":    "pending",
	}

	return s.BaseService.CountRecords(ctx, "friends", filters)
}

// MarkFriendRequestNotificationAsRead marks friend_request notification as read
// Called when user accepts/declines a friend request
func (s *NotificationService) MarkFriendRequestNotificationAsRead(
	ctx context.Context,
	userID uuid.UUID,
	senderID uuid.UUID,
) error {
	// Find unread friend_request notifications for this user
	data, _, err := s.client.From("notifications").
		Select("*", "", false).
		Eq("user_id", userID.String()).
		Eq("type", models.NotificationTypeFriendRequest).
		Eq("read", "false").
		Execute()

	if err != nil {
		return err
	}

	var notifications []models.Notification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return err
	}

	// Mark all matching notifications as read
	for _, notif := range notifications {
		s.MarkAsRead(ctx, notif.ID)
	}

	return nil
}
