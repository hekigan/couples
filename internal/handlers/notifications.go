package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
)

// GetNotificationsHandler returns all notifications for the current user
func (h *Handler) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	notifications, err := h.NotificationService.GetUserNotifications(ctx, userID, 50)
	if err != nil {
		http.Error(w, "Failed to load notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// GetUnreadCountHandler returns the count of unread notifications
func (h *Handler) GetUnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	count, err := h.NotificationService.GetUnreadCount(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load notification count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// MarkNotificationReadHandler marks a notification as read
func (h *Handler) MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.NotificationService.MarkAsRead(ctx, notificationID); err != nil {
		http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success"}`))
}

// MarkAllNotificationsReadHandler marks all notifications as read
func (h *Handler) MarkAllNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	if err := h.NotificationService.MarkAllAsRead(ctx, userID); err != nil {
		http.Error(w, "Failed to mark all as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success"}`))
}

// SendRoomInvitationHandler sends a room invitation
func (h *Handler) SendRoomInvitationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var req struct {
		RoomID    string `json:"room_id"`
		InviteeID string `json:"invitee_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	inviteeID, err := uuid.Parse(req.InviteeID)
	if err != nil {
		http.Error(w, "Invalid invitee ID", http.StatusBadRequest)
		return
	}

	// Create the invitation
	invitation := &models.RoomInvitation{
		RoomID:    roomID,
		InviterID: userID,
		InviteeID: inviteeID,
	}

	if err := h.NotificationService.CreateRoomInvitation(ctx, invitation); err != nil {
		http.Error(w, fmt.Sprintf("Failed to send invitation: %v", err), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Invitation sent"})
}

// CancelRoomInvitationHandler cancels a room invitation
func (h *Handler) CancelRoomInvitationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["room_id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	inviteeID, err := uuid.Parse(vars["invitee_id"])
	if err != nil {
		http.Error(w, "Invalid invitee ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.NotificationService.CancelRoomInvitation(ctx, roomID, inviteeID); err != nil {
		http.Error(w, "Failed to cancel invitation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success","message":"Invitation cancelled"}`))
}

