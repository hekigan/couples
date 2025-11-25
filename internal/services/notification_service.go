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
