package models

import (
	"time"

	"github.com/google/uuid"
)

// RoomJoinRequest represents a request to join a room
type RoomJoinRequest struct {
	ID        uuid.UUID  `json:"id"`
	RoomID    uuid.UUID  `json:"room_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Status    string     `json:"status"` // pending, accepted, rejected
	Message   *string    `json:"message,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}



