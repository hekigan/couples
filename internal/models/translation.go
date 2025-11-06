package models

import (
	"time"

	"github.com/google/uuid"
)

// Translation represents a translation entry
type Translation struct {
	ID               uuid.UUID `json:"id"`
	LanguageCode     string    `json:"language_code"`
	TranslationKey   string    `json:"translation_key"`
	TranslationValue string    `json:"translation_value"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

