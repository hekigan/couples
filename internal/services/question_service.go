package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/yourusername/couple-card-game/internal/models"
)

// QuestionService handles question-related operations
type QuestionService struct {
	client *supabase.Client
}

// NewQuestionService creates a new question service
func NewQuestionService(client *supabase.Client) *QuestionService {
	return &QuestionService{
		client: client,
	}
}

// GetQuestionByID retrieves a question by ID (stub)
func (s *QuestionService) GetQuestionByID(ctx context.Context, id uuid.UUID) (*models.Question, error) {
	return &models.Question{ID: id, Text: "Sample question?"}, nil
}

// GetRandomQuestion gets a random question for a room (stub)
func (s *QuestionService) GetRandomQuestion(ctx context.Context, roomID uuid.UUID, language string) (*models.Question, error) {
	return &models.Question{ID: uuid.New(), Text: "What is your favorite memory together?"}, nil
}

// GetAllQuestions retrieves all questions (stub)
func (s *QuestionService) GetAllQuestions(ctx context.Context) ([]models.Question, error) {
	return []models.Question{}, nil
}

// CreateQuestion creates a new question (stub)
func (s *QuestionService) CreateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}

// UpdateQuestion updates a question (stub)
func (s *QuestionService) UpdateQuestion(ctx context.Context, question *models.Question) error {
	return nil
}

// DeleteQuestion deletes a question (stub)
func (s *QuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	return nil
}

// GetCategories retrieves all categories (stub)
func (s *QuestionService) GetCategories(ctx context.Context) ([]models.Category, error) {
	return []models.Category{}, nil
}
