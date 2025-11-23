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
	client *supabase.Client
}

func NewNotificationService(client *supabase.Client) *NotificationService {
	return &NotificationService{
		client: client,
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

	_, _, err := s.client.From("notifications").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetUserNotifications gets all notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	query := s.client.From("notifications").
		Select("*", "", false).
		Eq("user_id", userID.String())

	// Note: Order and Limit might not be available in this client version
	// We'll sort on the client side if needed

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	var notifications []models.Notification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return nil, fmt.Errorf("failed to parse notifications: %w", err)
	}

	// Apply limit if specified
	if limit > 0 && len(notifications) > limit {
		notifications = notifications[:limit]
	}

	return notifications, nil
}

// GetUnreadCount gets count of unread notifications
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	data, _, err := s.client.From("notifications").
		Select("id", "", false).
		Eq("user_id", userID.String()).
		Eq("read", "false").
		Execute()

	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	// Count from data
	var notifications []models.Notification
	if err := json.Unmarshal(data, &notifications); err == nil {
		return len(notifications), nil
	}

	return 0, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	data := map[string]interface{}{
		"read": true,
	}

	_, _, err := s.client.From("notifications").
		Update(data, "", "").
		Eq("id", notificationID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	data := map[string]interface{}{
		"read": true,
	}

	_, _, err := s.client.From("notifications").
		Update(data, "", "").
		Eq("user_id", userID.String()).
		Eq("read", "false").
		Execute()

	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID) error {
	_, _, err := s.client.From("notifications").
		Delete("", "").
		Eq("id", notificationID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
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

	_, _, err := s.client.From("room_invitations").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to create room invitation: %w", err)
	}

	return nil
}

// GetRoomInvitation gets a specific room invitation
func (s *NotificationService) GetRoomInvitation(ctx context.Context, roomID, inviteeID uuid.UUID) (*models.RoomInvitation, error) {
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
	_, _, err := s.client.From("room_invitations").
		Delete("", "").
		Eq("room_id", roomID.String()).
		Eq("invitee_id", inviteeID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return nil
}

