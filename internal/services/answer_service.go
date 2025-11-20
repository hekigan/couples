package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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

// CreateAnswer creates a new answer
func (s *AnswerService) CreateAnswer(ctx context.Context, answer *models.Answer) error {
	// Validate action type
	if answer.ActionType != "answered" && answer.ActionType != "passed" {
		return fmt.Errorf("invalid action type: must be 'answered' or 'passed'")
	}

	answerMap := map[string]interface{}{
		"room_id":     answer.RoomID.String(),
		"question_id": answer.QuestionID.String(),
		"user_id":     answer.UserID.String(),
		"answer_text": answer.AnswerText,
		"action_type": answer.ActionType,
	}

	log.Printf("📝 Creating answer: room=%s, question=%s, user=%s, action=%s",
		answer.RoomID, answer.QuestionID, answer.UserID, answer.ActionType)

	data, _, err := s.client.From("answers").Insert(answerMap, false, "", "", "").Execute()
	if err != nil {
		log.Printf("❌ Failed to create answer: %v", err)
		return fmt.Errorf("failed to create answer: %w", err)
	}

	log.Printf("✅ Answer created successfully: %s", string(data))
	return nil
}

// GetAnswersByRoom retrieves all answers for a room, ordered by creation time
func (s *AnswerService) GetAnswersByRoom(ctx context.Context, roomID uuid.UUID) ([]models.Answer, error) {
	data, _, err := s.client.From("answers").
		Select("*", "", false).
		Eq("room_id", roomID.String()).
		Order("created_at", nil).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch answers: %w", err)
	}

	var answers []models.Answer
	if err := json.Unmarshal(data, &answers); err != nil {
		return nil, fmt.Errorf("failed to parse answers: %w", err)
	}

	return answers, nil
}

// GetAnswerByID retrieves a specific answer by ID
func (s *AnswerService) GetAnswerByID(ctx context.Context, id uuid.UUID) (*models.Answer, error) {
	data, _, err := s.client.From("answers").
		Select("*", "", false).
		Eq("id", id.String()).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("answer not found: %w", err)
	}

	var answer models.Answer
	if err := json.Unmarshal(data, &answer); err != nil {
		return nil, fmt.Errorf("failed to parse answer: %w", err)
	}

	return &answer, nil
}

// GetAnswersByQuestion retrieves all answers for a specific question
func (s *AnswerService) GetAnswersByQuestion(ctx context.Context, questionID uuid.UUID) ([]models.Answer, error) {
	data, _, err := s.client.From("answers").
		Select("*", "", false).
		Eq("question_id", questionID.String()).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch answers: %w", err)
	}

	var answers []models.Answer
	if err := json.Unmarshal(data, &answers); err != nil {
		return nil, fmt.Errorf("failed to parse answers: %w", err)
	}

	return answers, nil
}

// GetLastAnswerForQuestion retrieves the most recent answer for a specific question in a room
func (s *AnswerService) GetLastAnswerForQuestion(ctx context.Context, roomID, questionID uuid.UUID) (*models.Answer, error) {
	// Get the answer for this question in this room
	log.Printf("📥 GetLastAnswerForQuestion: room=%s, question=%s", roomID, questionID)

	data, _, err := s.client.From("answers").
		Select("*", "", false).
		Eq("room_id", roomID.String()).
		Eq("question_id", questionID.String()).
		Execute()

	if err != nil {
		log.Printf("❌ GetLastAnswerForQuestion query error: %v", err)
		return nil, fmt.Errorf("failed to fetch answer: %w", err)
	}

	log.Printf("📥 GetLastAnswerForQuestion response: %s", string(data))

	var answers []models.Answer
	if err := json.Unmarshal(data, &answers); err != nil {
		log.Printf("❌ GetLastAnswerForQuestion parse error: %v", err)
		return nil, fmt.Errorf("failed to parse answer: %w", err)
	}

	if len(answers) == 0 {
		log.Printf("⚠️ GetLastAnswerForQuestion: no answers found")
		return nil, nil // No answer found
	}

	log.Printf("✅ GetLastAnswerForQuestion: found %d answers, returning first", len(answers))
	return &answers[0], nil
}
