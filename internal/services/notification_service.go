package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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
// This counts unread notifications (friend requests create notifications in the table)
func (s *NotificationService) GetNotificationBadgeCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	return s.GetUnreadCount(ctx, userID)
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
