package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/yourusername/couple-card-game/internal/models"
)

// AnswerService handles answer-related operations
type AnswerService struct {
	client *supabase.Client
}

// NewAnswerService creates a new answer service
func NewAnswerService(client *supabase.Client) *AnswerService {
	return &AnswerService{
		client: client,
	}
}

// CreateAnswer creates a new answer (stub)
func (s *AnswerService) CreateAnswer(ctx context.Context, answer *models.Answer) error {
	// TODO: Implement proper Supabase insert
	return nil
}

// GetAnswersByRoom retrieves all answers for a room (stub)
func (s *AnswerService) GetAnswersByRoom(ctx context.Context, roomID uuid.UUID) ([]models.Answer, error) {
	// TODO: Implement proper Supabase query
	return []models.Answer{}, nil
}
