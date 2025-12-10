package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/labstack/echo/v4"
)

// GetNotificationsHandler returns all notifications for the current user
func (h *Handler) GetNotificationsHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	notifications, err := h.NotificationService.GetUserNotifications(ctx, userID, 50)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load notifications")
	}

	return c.JSON(http.StatusOK, notifications)
}

// GetUnreadCountHandler returns the count of unread notifications
func (h *Handler) GetUnreadCountHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	count, err := h.NotificationService.GetUnreadCount(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load notification count")
	}

	return c.JSON(http.StatusOK, map[string]int{"count": count})
}

// MarkNotificationReadHandler marks a notification as read
func (h *Handler) MarkNotificationReadHandler(c echo.Context) error {
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid notification ID")
	}

	ctx := context.Background()
	if err := h.NotificationService.MarkAsRead(ctx, notificationID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to mark as read")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// MarkAllNotificationsReadHandler marks all notifications as read
func (h *Handler) MarkAllNotificationsReadHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	if err := h.NotificationService.MarkAllAsRead(ctx, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to mark all as read")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// SendRoomInvitationHandler sends a room invitation
func (h *Handler) SendRoomInvitationHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	var req struct {
		RoomID    string `json:"room_id"`
		InviteeID string `json:"invitee_id"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	inviteeID, err := uuid.Parse(req.InviteeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invitee ID")
	}

	// Create the invitation
	invitation := &models.RoomInvitation{
		RoomID:    roomID,
		InviterID: userID,
		InviteeID: inviteeID,
	}

	if err := h.NotificationService.CreateRoomInvitation(ctx, invitation); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to send invitation: %v", err))
	}

	// Get inviter username
	inviter, _ := h.UserService.GetUserByID(ctx, userID)
	inviterName := "Someone"
	if inviter != nil {
		inviterName = inviter.Username
	}

	// Create notification for invitee
	notification := &models.Notification{
		UserID:  inviteeID,
		Type:    models.NotificationTypeRoomInvite,
		Title:   "Room Invitation",
		Message: fmt.Sprintf("%s invited you to join a game room", inviterName),
		Link:    fmt.Sprintf("/game/room/%s", roomID),
		Read:    false,
	}

	if err := h.NotificationService.CreateNotification(ctx, notification); err != nil {
		fmt.Printf("WARNING: Failed to create notification: %v\n", err)
	}

	return c.JSON(http.StatusCreated, map[string]string{"status": "success", "message": "Invitation sent"})
}

// CancelRoomInvitationHandler cancels a room invitation
func (h *Handler) CancelRoomInvitationHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("room_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	inviteeID, err := uuid.Parse(c.Param("invitee_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid invitee ID")
	}

	ctx := context.Background()
	if err := h.NotificationService.CancelRoomInvitation(ctx, roomID, inviteeID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to cancel invitation")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Invitation cancelled"})
}

