package models

import (
	"time"

	"github.com/google/uuid"
)

// Room represents a game room
type Room struct {
	ID                 uuid.UUID   `json:"id"`
	Name               string      `json:"name"`
	OwnerID            uuid.UUID   `json:"owner_id"`
	GuestID            *uuid.UUID  `json:"guest_id"`
	Status             string      `json:"status"` // waiting, ready, playing, paused, finished
	Language           string      `json:"language"`
	IsPrivate          bool        `json:"is_private"`
	GuestReady         bool        `json:"guest_ready"`
	MaxQuestions       int         `json:"max_questions"`
	CurrentQuestion    int         `json:"current_question"`
	CurrentQuestionID  *uuid.UUID  `json:"current_question_id"`
	CurrentTurn        *uuid.UUID  `json:"current_player_id"`
	SelectedCategories []uuid.UUID `json:"selected_categories"`
	PausedAt           *time.Time  `json:"paused_at,omitempty"`
	DisconnectedUser   *uuid.UUID  `json:"disconnected_user,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}



