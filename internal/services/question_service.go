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
	*BaseService
	client *supabase.Client
}

// NewQuestionService creates a new question service
func NewQuestionService(client *supabase.Client) *QuestionService {
	return &QuestionService{
		BaseService: NewBaseService(client, "QuestionService"),
		client:      client,
	}
}

// QuestionTranslations holds all language versions of a question
type QuestionTranslations struct {
	BaseQuestionID uuid.UUID
	English        *models.Question
	French         *models.Question
	Japanese       *models.Question
}

// GetQuestionByID retrieves a question by ID
func (s *QuestionService) GetQuestionByID(ctx context.Context, id uuid.UUID) (*models.Question, error) {
	var question models.Question
	if err := s.BaseService.GetSingleRecord(ctx, "questions", id, &question); err != nil {
		return nil, err
	}
	return &question, nil
}

// GetQuestionTranslations retrieves all translations for a question by its base_question_id
func (s *QuestionService) GetQuestionTranslations(ctx context.Context, baseQuestionID uuid.UUID) (*QuestionTranslations, error) {
	var questions []models.Question
	if err := s.BaseService.GetRecords(ctx, "questions", map[string]interface{}{
		"base_question_id": baseQuestionID.String(),
	}, &questions); err != nil {
		return nil, fmt.Errorf("failed to fetch translations: %w", err)
	}

	translations := &QuestionTranslations{
		BaseQuestionID: baseQuestionID,
	}

	// Organize by language code
	for i := range questions {
		switch questions[i].LanguageCode {
		case "en":
			translations.English = &questions[i]
		case "fr":
			translations.French = &questions[i]
		case "ja":
			translations.Japanese = &questions[i]
		}
	}

	return translations, nil
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

	return s.BaseService.InsertRecord(ctx, "question_history", historyMap)
}

// GetAllQuestions retrieves all questions
func (s *QuestionService) GetAllQuestions(ctx context.Context) ([]models.Question, error) {
	var questions []models.Question
	if err := s.BaseService.GetRecords(ctx, "questions", map[string]interface{}{}, &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

// CreateQuestion creates a new question
func (s *QuestionService) CreateQuestion(ctx context.Context, question *models.Question) error {
	questionMap := map[string]interface{}{
		"category_id":      question.CategoryID.String(),
		"lang_code":        question.LanguageCode,
		"question_text":    question.Text,
		"base_question_id": question.BaseQuestionID.String(),
	}

	// Include ID if provided (for base questions that self-reference)
	if question.ID != uuid.Nil {
		questionMap["id"] = question.ID.String()
	}

	return s.BaseService.InsertRecord(ctx, "questions", questionMap)
}

// UpdateQuestion updates a question
func (s *QuestionService) UpdateQuestion(ctx context.Context, question *models.Question) error {
	questionMap := map[string]interface{}{
		"category_id":      question.CategoryID.String(),
		"lang_code":        question.LanguageCode,
		"question_text":    question.Text,
		"base_question_id": question.BaseQuestionID.String(),
	}

	return s.BaseService.UpdateRecord(ctx, "questions", question.ID, questionMap)
}

// DeleteQuestion deletes a question
func (s *QuestionService) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	return s.BaseService.DeleteRecord(ctx, "questions", id)
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
// Uses base_question_id to link translations together
func (s *QuestionService) GetQuestionTranslationStatus(ctx context.Context, questionIDs []uuid.UUID) (map[string]int, error) {
	if len(questionIDs) == 0 {
		return make(map[string]int), nil
	}

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(questionIDs))
	for i, id := range questionIDs {
		idStrings[i] = id.String()
	}

	// Step 1: Fetch the questions to get their base_question_ids
	data, _, err := s.client.From("questions").
		Select("id, base_question_id", "", false).
		In("id", idStrings).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}

	var questions []struct {
		ID             string `json:"id"`
		BaseQuestionID string `json:"base_question_id"`
	}
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse questions: %w", err)
	}

	// Step 2: Build map of ID → BaseQuestionID and collect all base question IDs
	idToBaseID := make(map[string]string)
	baseIDSet := make(map[string]bool)
	for _, q := range questions {
		idToBaseID[q.ID] = q.BaseQuestionID
		baseIDSet[q.BaseQuestionID] = true
	}

	// Convert base IDs to slice for query
	baseIDStrings := make([]string, 0, len(baseIDSet))
	for baseID := range baseIDSet {
		baseIDStrings = append(baseIDStrings, baseID)
	}

	// Step 3: Query all questions with these base_question_ids to count translations
	data, _, err = s.client.From("questions").
		Select("base_question_id, lang_code", "", false).
		In("base_question_id", baseIDStrings).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch translations: %w", err)
	}

	var allTranslations []struct {
		BaseQuestionID string `json:"base_question_id"`
		LangCode       string `json:"lang_code"`
	}
	if err := json.Unmarshal(data, &allTranslations); err != nil {
		return nil, fmt.Errorf("failed to parse translations: %w", err)
	}

	// Step 4: Count distinct languages per base_question_id
	baseIDToLanguages := make(map[string]map[string]bool)
	for _, t := range allTranslations {
		if baseIDToLanguages[t.BaseQuestionID] == nil {
			baseIDToLanguages[t.BaseQuestionID] = make(map[string]bool)
		}
		baseIDToLanguages[t.BaseQuestionID][t.LangCode] = true
	}

	// Step 5: Map back to original question IDs
	counts := make(map[string]int)
	for id, baseID := range idToBaseID {
		if langs, ok := baseIDToLanguages[baseID]; ok {
			counts[id] = len(langs)
		} else {
			counts[id] = 1 // At least one version exists
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

// ListQuestions retrieves questions with pagination and optional filtering
func (s *QuestionService) ListQuestions(ctx context.Context, limit, offset int, categoryID *uuid.UUID, langCode *string) ([]models.Question, error) {
	// Custom query - uses Order and Range with optional filters, not supported by BaseService
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
