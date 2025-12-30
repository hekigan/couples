package models

import (
	"time"

	"github.com/google/uuid"
)

// Friend represents a friendship between two users (or a pending request)
type Friend struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FriendID  uuid.UUID `json:"friend_id"`
	Status    string    `json:"status"` // "pending", "accepted", "declined"
	CreatedAt time.Time `json:"created_at"`
}

// FriendWithUserInfo represents a friend with user details
type FriendWithUserInfo struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FriendID  uuid.UUID `json:"friend_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// FriendEmailInvitation represents an email-based friend invitation
type FriendEmailInvitation struct {
	ID             uuid.UUID  `json:"id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	RecipientEmail string     `json:"recipient_email"`
	Token          string     `json:"token"`
	Status         string     `json:"status"` // "pending", "accepted", "expired", "cancelled"
	ExpiresAt      time.Time  `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

