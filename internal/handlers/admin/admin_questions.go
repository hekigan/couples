package admin

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
)

// ListQuestionsHandler returns an HTML fragment with the questions list  
func (ah *AdminAPIHandler) ListQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse filters
	var categoryID *uuid.UUID
	if catID := r.URL.Query().Get("category_id"); catID != "" {
		if parsed, err := uuid.Parse(catID); err == nil {
			categoryID = &parsed
		}
	}

	// Use helper for pagination
	page, perPage, offset := handlers.ParsePaginationParams(r)

	// Always filter to English only
	langCode := "en"

	// Fetch questions
	questions, err := ah.questionService.ListQuestions(ctx, perPage, offset, categoryID, &langCode)
	if err != nil {
		http.Error(w, "Failed to list questions", http.StatusInternalServerError)
		log.Printf("Error listing questions: %v", err)
		return
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
		CurrentPage:              page,
		TotalPages:               totalPages,
		ItemsPerPage:             perPage,
		MissingTranslationsCount: missingTranslationsCount,
		// Pagination template fields
		BaseURL:         "/admin/api/questions/list",
		PageURL:         "/admin/questions",
		Target:          "#questions-list",
		IncludeSelector: "[name='category_id']",
		ExtraParams:     extraParams,
		ItemName:        "questions",
	}

	html, err := ah.handler.TemplateService.RenderFragment("questions_list", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering questions list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetQuestionCreateFormHandler returns an HTML form for creating a new question
func (ah *AdminAPIHandler) GetQuestionCreateFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	categories, err := ah.categoryService.GetCategories(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		log.Printf("Error fetching categories: %v", err)
		return
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
	data := services.QuestionEditFormData{
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
		Page:           0, // Not needed for create
	}

	html, err := ah.handler.TemplateService.RenderFragment("question_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering question create form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetQuestionEditFormHandler returns an HTML form for editing a question
func (ah *AdminAPIHandler) GetQuestionEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	questionID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question, err := ah.questionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Fetch all translations for this question
	translations, err := ah.questionService.GetQuestionTranslations(ctx, question.BaseQuestionID)
	if err != nil {
		http.Error(w, "Failed to fetch translations", http.StatusInternalServerError)
		log.Printf("Error fetching translations: %v", err)
		return
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

	data := services.QuestionEditFormData{
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

	html, err := ah.handler.TemplateService.RenderFragment("question_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering question edit form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// UpdateQuestionHandler updates a question or its translation
func (ah *AdminAPIHandler) UpdateQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	questionID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
		return
	}

	// Get the current question to access base_question_id
	currentQuestion, err := ah.questionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Question not found"})
		return
	}

	categoryID, err := uuid.Parse(r.FormValue("category_id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid category ID"})
		return
	}

	langCode := r.FormValue("lang_code")
	questionText := r.FormValue("question_text")
	translationText := r.FormValue("question_text_translation")

	// Fetch all translations to find the correct question to update
	translations, err := ah.questionService.GetQuestionTranslations(ctx, currentQuestion.BaseQuestionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch translations"})
		log.Printf("Error fetching translations: %v", err)
		return
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update question"})
			log.Printf("Error updating question: %v", err)
			return
		}
	} else {
		// Create new translation (only for FR/JA, English should always exist)
		if langCode == "en" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "English question not found"})
			return
		}

		newTranslation := &models.Question{
			ID:             uuid.New(),
			CategoryID:     categoryID,
			LanguageCode:   langCode,
			Text:           textToUpdate,
			BaseQuestionID: currentQuestion.BaseQuestionID,
		}

		if err := ah.questionService.CreateQuestion(ctx, newTranslation); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create translation"})
			log.Printf("Error creating translation: %v", err)
			return
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"success": "Question updated successfully"})
}

// CreateQuestionHandler creates a new question with optional translations
func (ah *AdminAPIHandler) CreateQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryID, err := uuid.Parse(r.FormValue("category_id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Get form values
	questionTextEN := r.FormValue("question_text_en")
	questionTextFR := r.FormValue("question_text_fr")
	questionTextJA := r.FormValue("question_text_ja")

	if questionTextEN == "" {
		http.Error(w, "English question text is required", http.StatusBadRequest)
		return
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
		http.Error(w, "Failed to create English question", http.StatusInternalServerError)
		log.Printf("Error creating English question: %v", err)
		return
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
	ah.ListQuestionsHandler(w, r)
}

// DeleteQuestionHandler deletes a question
func (ah *AdminAPIHandler) DeleteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	questionID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ah.questionService.DeleteQuestion(ctx, questionID); err != nil {
		http.Error(w, "Failed to delete question", http.StatusInternalServerError)
		log.Printf("Error deleting question: %v", err)
		return
	}

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}
