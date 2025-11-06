package models

import (
	"time"

	"github.com/google/uuid"
)

// Answer represents a player's answer to a question
type Answer struct {
	ID         uuid.UUID `json:"id"`
	RoomID     uuid.UUID `json:"room_id"`
	QuestionID uuid.UUID `json:"question_id"`
	UserID     uuid.UUID `json:"user_id"`
	AnswerText string    `json:"answer_text"`
	ActionType string    `json:"action_type"` // "answered" or "passed"
	CreatedAt  time.Time `json:"created_at"`
}


