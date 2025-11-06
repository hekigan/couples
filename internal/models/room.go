package models

import (
	"time"

	"github.com/google/uuid"
)

// Room represents a game room
type Room struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	OwnerID         uuid.UUID  `json:"owner_id"`
	GuestID         *uuid.UUID `json:"guest_id"`
	Status          string     `json:"status"`
	Language        string     `json:"language"`
	IsPrivate       bool       `json:"is_private"`
	MaxQuestions    int        `json:"max_questions"`
	CurrentQuestion int        `json:"current_question"`
	CurrentTurn     *uuid.UUID `json:"current_turn"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

