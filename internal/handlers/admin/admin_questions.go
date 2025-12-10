package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	adminFragments "github.com/hekigan/couples/internal/views/fragments/admin"
	"github.com/labstack/echo/v4"
)

// ListQuestionsHandler returns an HTML fragment with the questions list
func (ah *AdminAPIHandler) ListQuestionsHandler(c echo.Context) error {
	ctx := context.Background()

	// Parse filters
	var categoryID *uuid.UUID
	if catID := c.QueryParam("category_id"); catID != "" {
		if parsed, err := uuid.Parse(catID); err == nil {
			categoryID = &parsed
		}
	}

	// Use helper for pagination
	page, perPage := handlers.ParsePaginationParams(c)
	offset := (page - 1) * perPage

	// Always filter to English only
	langCode := "en"

	// Fetch questions
	questions, err := ah.questionService.ListQuestions(ctx, perPage, offset, categoryID, &langCode)
	if err != nil {
		log.Printf("Error listing questions: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list questions")
	}

	// Get total count for pagination (English only)
	totalCount, _ := ah.questionService.GetQuestionCountsByCategory(ctx, "en")
	total := 0
	for _, count := range totalCount {
		total += count
	}
	if categoryID != nil {
		// Filter count by category
		if count, ok := totalCount[categoryID.String()]; ok {
			total = count
		} else {
			total = 0
		}
	}

	// Calculate pagination
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Get categories for dropdown
	categories, _ := ah.categoryService.GetCategories(ctx)

	// Get question counts by category (for dropdown)
	counts, err := ah.questionService.GetQuestionCountsByCategory(ctx, "en")
	if err != nil {
		log.Printf("⚠️ Failed to get question counts: %v", err)
		counts = make(map[string]int)
	}

	// Get translation status for current page questions
	questionIDs := make([]uuid.UUID, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}
	translationStatus, _ := ah.questionService.GetQuestionTranslationStatus(ctx, questionIDs)

	// Calculate total missing translations (across ALL English questions)
	allEnglishQuestions, _ := ah.questionService.ListQuestions(ctx, 10000, 0, nil, &langCode)
	allIDs := make([]uuid.UUID, len(allEnglishQuestions))
	for i, q := range allEnglishQuestions {
		allIDs[i] = q.ID
	}
	allTranslationStatus, _ := ah.questionService.GetQuestionTranslationStatus(ctx, allIDs)
	missingTranslationsCount := 0
	for _, count := range allTranslationStatus {
		if count < 3 {
			missingTranslationsCount += (3 - count)
		}
	}

	// Build category options
	categoryOptions := make([]services.AdminCategoryOption, len(categories))
	for i, cat := range categories {
		count := 0
		if c, ok := counts[cat.ID.String()]; ok {
			count = c
		}
		selected := categoryID != nil && *categoryID == cat.ID
		categoryOptions[i] = services.AdminCategoryOption{
			ID:            cat.ID.String(),
			Label:         cat.Label,
			Selected:      selected,
			QuestionCount: count,
		}
	}

	// Create category map for lookups
	categoryMap := make(map[uuid.UUID]models.Category)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}

	// Build question infos
	questionInfos := make([]services.AdminQuestionInfo, len(questions))
	for i, q := range questions {
		categoryLabel := "Unknown"
		if cat, ok := categoryMap[q.CategoryID]; ok {
			categoryLabel = cat.Label
		}

		tCount := 1 // Default to 1 (at least English exists)
		if count, ok := translationStatus[q.ID.String()]; ok {
			tCount = count
		}

		questionInfos[i] = services.AdminQuestionInfo{
			ID:               q.ID.String(),
			Text:             q.Text,
			CategoryLabel:    categoryLabel,
			LanguageCode:     q.LanguageCode,
			TranslationCount: tCount,
		}
	}

	selectedCategoryID := ""
	extraParams := ""
	if categoryID != nil {
		selectedCategoryID = categoryID.String()
		extraParams = "&category_id=" + selectedCategoryID
	}

	data := services.QuestionsListData{
		Questions:                questionInfos,
		Categories:               categoryOptions,
		SelectedCategoryID:       selectedCategoryID,
		TotalCount:               total,
		TotalPages:               totalPages,
		MissingTranslationsCount: missingTranslationsCount,
		// Pagination template fields
		BaseURL:         "/admin/api/questions/list",
		PageURL:         "/admin/questions",
		Target:          "#questions-list",
		IncludeSelector: "[name='category_id']",
		ExtraParams:     extraParams,
		ItemName:        "questions",
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.QuestionsList(&data))
	if err != nil {
		log.Printf("Error rendering questions list: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// GetQuestionCreateFormHandler returns an HTML form for creating a new question
func (ah *AdminAPIHandler) GetQuestionCreateFormHandler(c echo.Context) error {
	ctx := context.Background()

	categories, err := ah.categoryService.GetCategories(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch categories")
	}

	// Build category options (none selected)
	categoryOptions := make([]services.AdminCategoryOption, len(categories))
	for i, cat := range categories {
		categoryOptions[i] = services.AdminCategoryOption{
			ID:       cat.ID.String(),
			Label:    cat.Label,
			Selected: false,
		}
	}

	// Create empty form data for new question
	data := services.QuestionFormData{
		QuestionID:     "", // Empty indicates create mode
		BaseQuestionID: "",
		QuestionText:   "",
		TranslationFR:  "",
		TranslationJA:  "",
		Categories:     categoryOptions,
		LangEN:         true, // Default to English
		LangFR:         false,
		LangJA:         false,
		SelectedLang:   "en",
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.QuestionForm(&data))
	if err != nil {
		log.Printf("Error rendering question create form: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// GetQuestionEditFormHandler returns an HTML form for editing a question
func (ah *AdminAPIHandler) GetQuestionEditFormHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	questionID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	question, err := ah.questionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Question not found")
	}

	// Fetch all translations for this question
	translations, err := ah.questionService.GetQuestionTranslations(ctx, question.BaseQuestionID)
	if err != nil {
		log.Printf("Error fetching translations: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch translations")
	}

	categories, _ := ah.categoryService.GetCategories(ctx)

	// Build category options
	categoryOptions := make([]services.AdminCategoryOption, len(categories))
	for i, cat := range categories {
		selected := cat.ID == question.CategoryID
		categoryOptions[i] = services.AdminCategoryOption{
			ID:       cat.ID.String(),
			Label:    cat.Label,
			Selected: selected,
		}
	}

	// Prepare translation texts
	var translationFR, translationJA string
	if translations.French != nil {
		translationFR = translations.French.Text
	}
	if translations.Japanese != nil {
		translationJA = translations.Japanese.Text
	}

	// Determine which text to show in the main question field based on selected language
	questionText := question.Text
	if translations.English != nil && question.LanguageCode != "en" {
		questionText = translations.English.Text // Always show English in main field for reference
	}

	data := services.QuestionFormData{
		QuestionID:     question.ID.String(),
		BaseQuestionID: question.BaseQuestionID.String(),
		QuestionText:   questionText,
		TranslationFR:  translationFR,
		TranslationJA:  translationJA,
		Categories:     categoryOptions,
		LangEN:         question.LanguageCode == "en",
		LangFR:         question.LanguageCode == "fr",
		LangJA:         question.LanguageCode == "ja",
		SelectedLang:   question.LanguageCode,
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.QuestionForm(&data))
	if err != nil {
		log.Printf("Error rendering question edit form: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// UpdateQuestionHandler updates a question or its translation
func (ah *AdminAPIHandler) UpdateQuestionHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	questionID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Get the current question to access base_question_id
	currentQuestion, err := ah.questionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Question not found"})
	}

	categoryID, err := uuid.Parse(c.FormValue("category_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
	}

	langCode := c.FormValue("lang_code")
	questionText := c.FormValue("question_text")
	translationText := c.FormValue("question_text_translation")

	// Fetch all translations to find the correct question to update
	translations, err := ah.questionService.GetQuestionTranslations(ctx, currentQuestion.BaseQuestionID)
	if err != nil {
		log.Printf("Error fetching translations: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch translations"})
	}

	// Determine which question to update based on selected language
	var targetQuestion *models.Question
	textToUpdate := translationText

	switch langCode {
	case "fr":
		targetQuestion = translations.French
	case "ja":
		targetQuestion = translations.Japanese
	default:
		targetQuestion = translations.English
		textToUpdate = questionText
	}

	if targetQuestion != nil {
		// Update existing question
		question := &models.Question{
			ID:             targetQuestion.ID,
			CategoryID:     categoryID,
			LanguageCode:   langCode,
			Text:           textToUpdate,
			BaseQuestionID: currentQuestion.BaseQuestionID,
		}

		if err := ah.questionService.UpdateQuestion(ctx, question); err != nil {
			log.Printf("Error updating question: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update question"})
		}
	} else {
		// Create new translation (only for FR/JA, English should always exist)
		if langCode == "en" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "English question not found"})
		}

		newTranslation := &models.Question{
			ID:             uuid.New(),
			CategoryID:     categoryID,
			LanguageCode:   langCode,
			Text:           textToUpdate,
			BaseQuestionID: currentQuestion.BaseQuestionID,
		}

		if err := ah.questionService.CreateQuestion(ctx, newTranslation); err != nil {
			log.Printf("Error creating translation: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create translation"})
		}
	}

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{"success": "Question updated successfully"})
}

// CreateQuestionHandler creates a new question with optional translations
func (ah *AdminAPIHandler) CreateQuestionHandler(c echo.Context) error {
	ctx := context.Background()

	categoryID, err := uuid.Parse(c.FormValue("category_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
	}

	// Get form values
	questionTextEN := c.FormValue("question_text_en")
	questionTextFR := c.FormValue("question_text_fr")
	questionTextJA := c.FormValue("question_text_ja")

	if questionTextEN == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "English question text is required")
	}

	// Generate UUID for the base English question
	baseQuestionID := uuid.New()

	// Create English question (base question)
	englishQuestion := &models.Question{
		ID:             baseQuestionID,
		CategoryID:     categoryID,
		LanguageCode:   "en",
		Text:           questionTextEN,
		BaseQuestionID: baseQuestionID, // Self-reference for base question
	}

	if err := ah.questionService.CreateQuestion(ctx, englishQuestion); err != nil {
		log.Printf("Error creating English question: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create English question")
	}

	// Create French translation if provided
	if questionTextFR != "" {
		frenchQuestion := &models.Question{
			ID:             uuid.New(),
			CategoryID:     categoryID,
			LanguageCode:   "fr",
			Text:           questionTextFR,
			BaseQuestionID: baseQuestionID, // Reference to English question
		}

		if err := ah.questionService.CreateQuestion(ctx, frenchQuestion); err != nil {
			log.Printf("Error creating French translation: %v", err)
			// Continue even if French creation fails
		}
	}

	// Create Japanese translation if provided
	if questionTextJA != "" {
		japaneseQuestion := &models.Question{
			ID:             uuid.New(),
			CategoryID:     categoryID,
			LanguageCode:   "ja",
			Text:           questionTextJA,
			BaseQuestionID: baseQuestionID, // Reference to English question
		}

		if err := ah.questionService.CreateQuestion(ctx, japaneseQuestion); err != nil {
			log.Printf("Error creating Japanese translation: %v", err)
			// Continue even if Japanese creation fails
		}
	}

	// Return updated questions list
	return ah.ListQuestionsHandler(c)
}

// DeleteQuestionHandler deletes a question
func (ah *AdminAPIHandler) DeleteQuestionHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	questionID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ah.questionService.DeleteQuestion(ctx, questionID); err != nil {
		log.Printf("Error deleting question: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete question")
	}

	// Return updated questions list
	return ah.ListQuestionsHandler(c)
}
