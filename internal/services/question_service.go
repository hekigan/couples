package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

// GetQuestionByID retrieves a question by ID
func (s *QuestionService) GetQuestionByID(ctx context.Context, id uuid.UUID) (*models.Question, error) {
	data, _, err := s.client.From("questions").
		Select("*", "", false).
		Eq("id", id.String()).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("question not found: %w", err)
	}

	var question models.Question
	if err := json.Unmarshal(data, &question); err != nil {
		return nil, fmt.Errorf("failed to parse question: %w", err)
	}

	return &question, nil
}

// GetRandomQuestion gets a random question for a room, filtered by categories and excluding already asked questions
func (s *QuestionService) GetRandomQuestion(ctx context.Context, roomID uuid.UUID, language string, categoryIDs []uuid.UUID) (*models.Question, error) {
	// First, get the list of question IDs already asked in this room
	historyData, _, err := s.client.From("question_history").
		Select("question_id", "", false).
		Eq("room_id", roomID.String()).
		Execute()

	var askedQuestionIDs []string
	if err == nil && len(historyData) > 0 {
		var historyRecords []map[string]interface{}
		if err := json.Unmarshal(historyData, &historyRecords); err == nil {
			for _, record := range historyRecords {
				if qid, ok := record["question_id"].(string); ok {
					askedQuestionIDs = append(askedQuestionIDs, qid)
				}
			}
		}
	}

	// Build query for random question
	query := s.client.From("questions").
		Select("*", "", false).
		Eq("lang_code", language)

	// Filter by categories if provided
	if len(categoryIDs) > 0 {
		categoryIDStrings := make([]string, len(categoryIDs))
		for i, id := range categoryIDs {
			categoryIDStrings[i] = id.String()
		}
		query = query.In("category_id", categoryIDStrings)
	}

	// Exclude already asked questions
	if len(askedQuestionIDs) > 0 {
		// Use string interpolation format for NOT IN clause
		// Format: NOT id IN ('uuid1','uuid2',...)
		query = query.Not("id", "in", "("+strings.Join(askedQuestionIDs, ",")+")")
	}

	// Execute query with limit 1 to get one random question
	// Note: For true randomness, we'd need a custom SQL function, but for now we'll get the first available
	data, _, err := query.Limit(1, "").Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch question: %w", err)
	}

	var questions []models.Question
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse questions: %w", err)
	}

	if len(questions) == 0 {
		// Provide more helpful error message
		if len(categoryIDs) > 0 {
			return nil, fmt.Errorf("no questions available for the selected categories and language '%s'. Please add questions to the database or change category selection", language)
		}
		return nil, fmt.Errorf("no questions available for language '%s'. Please add questions to the database", language)
	}

	return &questions[0], nil
}

// MarkQuestionAsked marks a question as asked in a room to prevent repetition
func (s *QuestionService) MarkQuestionAsked(ctx context.Context, roomID, questionID uuid.UUID) error {
	historyMap := map[string]interface{}{
		"room_id":     roomID.String(),
		"question_id": questionID.String(),
	}

	_, _, err := s.client.From("question_history").Insert(historyMap, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to mark question as asked: %w", err)
	}

	return nil
}

// GetAllQuestions retrieves all questions
func (s *QuestionService) GetAllQuestions(ctx context.Context) ([]models.Question, error) {
	data, _, err := s.client.From("questions").
		Select("*", "", false).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}

	var questions []models.Question
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse questions: %w", err)
	}

	return questions, nil
}

// CreateQuestion creates a new question
func (s *QuestionService) CreateQuestion(ctx context.Context, question *models.Question) error {
	questionMap := map[string]interface{}{
		"category_id":   question.CategoryID.String(),
		"lang_code":     question.LanguageCode,
		"question_text": question.Text,
	}

	_, _, err := s.client.From("questions").Insert(questionMap, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to create question: %w", err)
	}

	return nil
}

// UpdateQuestion updates a question
func (s *QuestionService) UpdateQuestion(ctx context.Context, question *models.Question) error {
	questionMap := map[string]interface{}{
		"category_id":   question.CategoryID.String(),
		"lang_code":     question.LanguageCode,
		"question_text": question.Text,
	}

	_, _, err := s.client.From("questions").
		Update(questionMap, "", "").
		Eq("id", question.ID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}

	return nil
}

// DeleteQuestion deletes a question
func (s *QuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	_, _, err := s.client.From("questions").
		Delete("", "").
		Eq("id", id.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}

	return nil
}

// GetCategories retrieves all categories
func (s *QuestionService) GetCategories(ctx context.Context) ([]models.Category, error) {
	data, _, err := s.client.From("categories").
		Select("*", "", false).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	var categories []models.Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("failed to parse categories: %w", err)
	}

	return categories, nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += "'" + s + "'"
	}
	return result
}
