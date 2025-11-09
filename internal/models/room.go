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

// RoomWithPlayers extends Room with player username information from the database view
// This eliminates N+1 queries when fetching room details with player info
type RoomWithPlayers struct {
	Room
	OwnerUsername         *string `json:"owner_username"`
	OwnerEmail            *string `json:"owner_email"`
	OwnerName             *string `json:"owner_name"`
	GuestUsername         *string `json:"guest_username"`
	GuestEmail            *string `json:"guest_email"`
	GuestName             *string `json:"guest_name"`
	CurrentPlayerUsername *string `json:"current_player_username"`
}

// ActiveGame represents an active game with player and question information
// Fetched from the active_games database view
// This eliminates multiple queries for game state (room + owner + guest + question + category)
type ActiveGame struct {
	// Room fields
	ID                 uuid.UUID   `json:"id"`
	Name               string      `json:"name"`
	Status             string      `json:"status"`
	Language           string      `json:"language"`
	MaxQuestions       int         `json:"max_questions"`
	CurrentQuestion    int         `json:"current_question"`
	CurrentQuestionID  *uuid.UUID  `json:"current_question_id"`
	CurrentPlayerID    *uuid.UUID  `json:"current_player_id"`
	SelectedCategories []uuid.UUID `json:"selected_categories"`
	PausedAt           *time.Time  `json:"paused_at,omitempty"`
	DisconnectedUser   *uuid.UUID  `json:"disconnected_user,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`

	// Player information
	OwnerID               uuid.UUID `json:"owner_id"`
	OwnerUsername         *string   `json:"owner_username"`
	GuestID               *uuid.UUID `json:"guest_id"`
	GuestUsername         *string   `json:"guest_username"`
	CurrentPlayerUsername *string   `json:"current_player_username"`

	// Current question details
	QuestionID            *uuid.UUID `json:"question_id"`
	CurrentQuestionText   *string    `json:"current_question_text"`
	CurrentQuestionLang   *string    `json:"current_question_lang"`
	QuestionCategoryID    *uuid.UUID `json:"question_category_id"`
	QuestionCategoryKey   *string    `json:"question_category_key"`
	QuestionCategoryLabel *string    `json:"question_category_label"`
}



