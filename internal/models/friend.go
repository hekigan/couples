package models

import (
	"time"

	"github.com/google/uuid"
)

// Friend represents a friendship between two users
type Friend struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FriendID  uuid.UUID `json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FriendRequest represents a pending friend request
type FriendRequest struct {
	ID         uuid.UUID `json:"id"`
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

