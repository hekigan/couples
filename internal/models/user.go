package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID  `json:"id"`
	Email              *string    `json:"email,omitempty"`
	Username           string     `json:"username"`
	AvatarURL          *string    `json:"avatar_url,omitempty"`
	IsAdmin            bool       `json:"is_admin"`
	IsAnonymous        bool       `json:"is_anonymous"`
	LanguagePreference string     `json:"language_preference,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	LastSeenAt         *time.Time `json:"last_seen_at,omitempty"`
}

