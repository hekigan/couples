package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a user notification
type Notification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"` // room_invitation, friend_request, game_start, etc.
	Title     string    `json:"title"`
	Message   string    `json:"message,omitempty"`
	Link      string    `json:"link,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// RoomInvitation represents an invitation to join a room
type RoomInvitation struct {
	ID         uuid.UUID  `json:"id"`
	RoomID     uuid.UUID  `json:"room_id"`
	InviterID  uuid.UUID  `json:"inviter_id"`
	InviteeID  uuid.UUID  `json:"invitee_id"`
	Status     string     `json:"status"` // pending, accepted, declined, cancelled
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// NotificationType constants
const (
	NotificationTypeRoomInvite    = "room_invitation"
	NotificationTypeFriendRequest = "friend_request"
	NotificationTypeGameStart     = "game_start"
	NotificationTypeMessage       = "message"
)

// InvitationStatus constants
const (
	InvitationStatusPending   = "pending"
	InvitationStatusAccepted  = "accepted"
	InvitationStatusDeclined  = "declined"
	InvitationStatusCancelled = "cancelled"
)

