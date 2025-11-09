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
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}



