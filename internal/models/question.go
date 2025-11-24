package models

import (
	"time"

	"github.com/google/uuid"
)

// Question represents a question in the game
type Question struct {
	ID             uuid.UUID `json:"id"`
	CategoryID     uuid.UUID `json:"category_id"`
	LanguageCode   string    `json:"lang_code"`
	Text           string    `json:"question_text"`
	BaseQuestionID uuid.UUID `json:"base_question_id"` // Links translations together
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Category represents a question category
type Category struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}



