package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
	"github.com/hekigan/couples/internal/models"
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

// CategoryCount represents a category with its question count
type CategoryCount struct {
	CategoryID string `json:"category_id"`
	Count      int    `json:"count"`
}

// GetQuestionCountsByCategory returns the number of questions per category for a given language
func (s *QuestionService) GetQuestionCountsByCategory(ctx context.Context, language string) (map[string]int, error) {
	// Query questions grouped by category for the given language
	data, _, err := s.client.From("questions").
		Select("category_id", "", false).
		Eq("lang_code", language).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch question counts: %w", err)
	}

	// Parse the response - we get all questions with their category_id
	var questions []struct {
		CategoryID string `json:"category_id"`
	}
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse questions: %w", err)
	}

	// Count questions per category
	counts := make(map[string]int)
	for _, q := range questions {
		counts[q.CategoryID]++
	}

	return counts, nil
}

// GetQuestionTranslationStatus returns the number of translations (0-3) for each question ID
// This checks how many language versions exist for each question (en, fr, ja)
// Note: Questions are matched by text content since there's no base_question_id in the schema
func (s *QuestionService) GetQuestionTranslationStatus(ctx context.Context, questionIDs []uuid.UUID) (map[string]int, error) {
	if len(questionIDs) == 0 {
		return make(map[string]int), nil
	}

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(questionIDs))
	for i, id := range questionIDs {
		idStrings[i] = id.String()
	}

	// Step 1: Fetch the English questions to get their texts
	data, _, err := s.client.From("questions").
		Select("id, question_text", "", false).
		In("id", idStrings).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}

	var englishQuestions []struct {
		ID   string `json:"id"`
		Text string `json:"question_text"`
	}
	if err := json.Unmarshal(data, &englishQuestions); err != nil {
		return nil, fmt.Errorf("failed to parse questions: %w", err)
	}

	// Step 2: Build a map of ID → Text and collect all texts
	idToText := make(map[string]string)
	texts := make([]string, 0, len(englishQuestions))
	for _, q := range englishQuestions {
		idToText[q.ID] = q.Text
		texts = append(texts, q.Text)
	}

	// Step 3: Query all questions with these texts across all languages
	data, _, err = s.client.From("questions").
		Select("question_text, lang_code", "", false).
		In("question_text", texts).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch translations: %w", err)
	}

	var allTranslations []struct {
		Text     string `json:"question_text"`
		LangCode string `json:"lang_code"`
	}
	if err := json.Unmarshal(data, &allTranslations); err != nil {
		return nil, fmt.Errorf("failed to parse translations: %w", err)
	}

	// Step 4: Count distinct languages per question text
	textToLanguages := make(map[string]map[string]bool)
	for _, t := range allTranslations {
		if textToLanguages[t.Text] == nil {
			textToLanguages[t.Text] = make(map[string]bool)
		}
		textToLanguages[t.Text][t.LangCode] = true
	}

	// Step 5: Map back to question IDs
	counts := make(map[string]int)
	for id, text := range idToText {
		if langs, ok := textToLanguages[text]; ok {
			counts[id] = len(langs)
		} else {
			counts[id] = 1 // At least the English version exists
		}
	}

	return counts, nil
}

// CountQuestionsForCategories counts total questions available for selected categories and language
func (s *QuestionService) CountQuestionsForCategories(ctx context.Context, language string, categoryIDs []uuid.UUID) (int, error) {
	query := s.client.From("questions").
		Select("id", "exact", false).
		Eq("lang_code", language)

	// Filter by categories if provided
	if len(categoryIDs) > 0 {
		categoryIDStrings := make([]string, len(categoryIDs))
		for i, id := range categoryIDs {
			categoryIDStrings[i] = id.String()
		}
		query = query.In("category_id", categoryIDStrings)
	}

	data, _, err := query.Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to count questions: %w", err)
	}

	// Parse the data array length
	var questions []map[string]interface{}
	if err := json.Unmarshal(data, &questions); err != nil {
		return 0, fmt.Errorf("failed to parse questions: %w", err)
	}

	return len(questions), nil
}

// CreateCategory creates a new category
func (s *QuestionService) CreateCategory(ctx context.Context, category *models.Category) error {
	categoryMap := map[string]interface{}{
		"key":   category.Key,
		"label": category.Label,
		"icon":  category.Icon,
	}

	_, _, err := s.client.From("categories").Insert(categoryMap, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// UpdateCategory updates a category
func (s *QuestionService) UpdateCategory(ctx context.Context, category *models.Category) error {
	categoryMap := map[string]interface{}{
		"key":   category.Key,
		"label": category.Label,
		"icon":  category.Icon,
	}

	_, _, err := s.client.From("categories").
		Update(categoryMap, "", "").
		Eq("id", category.ID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// DeleteCategory deletes a category
func (s *QuestionService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	_, _, err := s.client.From("categories").
		Delete("", "").
		Eq("id", id.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// GetCategoryByID retrieves a category by ID
func (s *QuestionService) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	data, _, err := s.client.From("categories").
		Select("*", "", false).
		Eq("id", id.String()).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	var category models.Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, fmt.Errorf("failed to parse category: %w", err)
	}

	return &category, nil
}

// ListQuestions retrieves questions with pagination and optional filtering
func (s *QuestionService) ListQuestions(ctx context.Context, limit, offset int, categoryID *uuid.UUID, langCode *string) ([]models.Question, error) {
	query := s.client.From("questions").
		Select("*", "", false).
		Order("created_at", &postgrest.OrderOpts{Ascending: false})

	if categoryID != nil {
		query = query.Eq("category_id", categoryID.String())
	}

	if langCode != nil {
		query = query.Eq("lang_code", *langCode)
	}

	if limit > 0 {
		query = query.Range(offset, offset+limit-1, "")
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list questions: %w", err)
	}

	var questions []models.Question
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse questions: %w", err)
	}

	return questions, nil
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
