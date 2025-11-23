package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
)

// AdminAPIHandler handles admin API requests
type AdminAPIHandler struct {
	handler        *handlers.Handler
	adminService   *services.AdminService
	questionService *services.QuestionService
}

// NewAdminAPIHandler creates a new admin API handler
func NewAdminAPIHandler(h *handlers.Handler, adminService *services.AdminService, questionService *services.QuestionService) *AdminAPIHandler {
	return &AdminAPIHandler{
		handler:        h,
		adminService:   adminService,
		questionService: questionService,
	}
}

// ListUsersHandler returns an HTML fragment with the users list
func (ah *AdminAPIHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse pagination parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 25 // Default
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			// Validate against allowed values
			if parsed == 25 || parsed == 50 || parsed == 100 {
				perPage = parsed
			}
		}
	}

	offset := (page - 1) * perPage

	users, err := ah.adminService.ListAllUsers(ctx, perPage, offset)
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		log.Printf("Error listing users: %v", err)
		return
	}

	totalCount, _ := ah.adminService.GetUserCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Build data for template
	userInfos := make([]services.AdminUserInfo, len(users))
	for i, user := range users {
		email := ""
		if user.Email != nil {
			email = *user.Email
		}

		userType := "Registered"
		if user.IsAnonymous {
			userType = "Anonymous"
		}

		userInfos[i] = services.AdminUserInfo{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     email,
			UserType:  userType,
			IsAdmin:   user.IsAdmin,
			CreatedAt: user.CreatedAt.Format("2006-01-02"),
		}
	}

	// Get current user ID from session
	currentUser := handlers.GetSessionUser(r)
	currentUserID := ""
	if currentUser != nil {
		currentUserID = currentUser.ID
	}

	data := services.UsersListData{
		Users:         userInfos,
		TotalCount:    totalCount,
		Page:          page,
		CurrentUserID: currentUserID,
		// Pagination fields
		CurrentPage:     page,
		TotalPages:      totalPages,
		ItemsPerPage:    perPage,
		BaseURL:         "/admin/api/users/list",
		PageURL:         "/admin/users",
		Target:          "#users-list",
		IncludeSelector: "",
		ExtraParams:     "",
		ItemName:        "users",
	}

	html, err := ah.handler.TemplateService.RenderFragment("users_list.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering users list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ToggleUserAdminHandler toggles a user's admin status
func (ah *AdminAPIHandler) ToggleUserAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := ah.adminService.ToggleUserAdmin(ctx, userID); err != nil {
		http.Error(w, "Failed to toggle admin status", http.StatusInternalServerError)
		log.Printf("Error toggling admin status: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// DeleteUserHandler deletes a user
func (ah *AdminAPIHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := ah.adminService.DeleteUser(ctx, userID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		log.Printf("Error deleting user: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

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

	// Parse pagination
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 25 // Default
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			// Validate against allowed values
			if parsed == 25 || parsed == 50 || parsed == 100 {
				perPage = parsed
			}
		}
	}

	// Always filter to English only
	langCode := "en"

	// Calculate offset
	offset := (page - 1) * perPage

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
	categories, _ := ah.questionService.GetCategories(ctx)

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
			ID:           cat.ID.String(),
			Icon:         cat.Icon,
			Label:        cat.Label,
			Selected:     selected,
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
			categoryLabel = fmt.Sprintf("%s %s", cat.Icon, cat.Label)
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

	html, err := ah.handler.TemplateService.RenderFragment("questions_list.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering questions list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetQuestionEditFormHandler returns an HTML form for editing a question
func (ah *AdminAPIHandler) GetQuestionEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	questionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	question, err := ah.questionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	categories, _ := ah.questionService.GetCategories(ctx)

	// Build category options
	categoryOptions := make([]services.AdminCategoryOption, len(categories))
	for i, cat := range categories {
		selected := cat.ID == question.CategoryID
		categoryOptions[i] = services.AdminCategoryOption{
			ID:       cat.ID.String(),
			Icon:     cat.Icon,
			Label:    cat.Label,
			Selected: selected,
		}
	}

	data := services.QuestionEditFormData{
		QuestionID:   question.ID.String(),
		QuestionText: question.Text,
		Categories:   categoryOptions,
		LangEN:       question.LanguageCode == "en",
		LangFR:       question.LanguageCode == "fr",
		LangJA:       question.LanguageCode == "ja",
	}

	html, err := ah.handler.TemplateService.RenderFragment("question_edit_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering question edit form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// UpdateQuestionHandler updates a question
func (ah *AdminAPIHandler) UpdateQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	questionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryID, err := uuid.Parse(r.FormValue("category_id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	question := &models.Question{
		ID:           questionID,
		CategoryID:   categoryID,
		LanguageCode: r.FormValue("lang_code"),
		Text:         r.FormValue("question_text"),
	}

	if err := ah.questionService.UpdateQuestion(ctx, question); err != nil {
		http.Error(w, "Failed to update question", http.StatusInternalServerError)
		log.Printf("Error updating question: %v", err)
		return
	}

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}

// CreateQuestionHandler creates a new question
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

	question := &models.Question{
		CategoryID:   categoryID,
		LanguageCode: r.FormValue("lang_code"),
		Text:         r.FormValue("question_text"),
	}

	if err := ah.questionService.CreateQuestion(ctx, question); err != nil {
		http.Error(w, "Failed to create question", http.StatusInternalServerError)
		log.Printf("Error creating question: %v", err)
		return
	}

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}

// DeleteQuestionHandler deletes a question
func (ah *AdminAPIHandler) DeleteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	questionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
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

// ListCategoriesHandler returns an HTML fragment with the categories list
func (ah *AdminAPIHandler) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 25 // Default
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			// Validate against allowed values
			if parsed == 25 || parsed == 50 || parsed == 100 {
				perPage = parsed
			}
		}
	}

	offset := (page - 1) * perPage

	categories, err := ah.questionService.ListCategories(ctx, perPage, offset)
	if err != nil {
		http.Error(w, "Failed to list categories", http.StatusInternalServerError)
		log.Printf("Error listing categories: %v", err)
		return
	}

	totalCount, _ := ah.questionService.GetCategoryCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Get question counts per category
	counts, _ := ah.questionService.GetQuestionCountsByCategory(ctx, "en")

	// Build data for template
	categoryInfos := make([]services.AdminCategoryInfo, len(categories))
	for i, cat := range categories {
		count := 0
		if c, ok := counts[cat.ID.String()]; ok {
			count = c
		}

		categoryInfos[i] = services.AdminCategoryInfo{
			ID:            cat.ID.String(),
			Icon:          cat.Icon,
			Label:         cat.Label,
			Key:           cat.Key,
			QuestionCount: count,
		}
	}

	data := services.CategoriesListData{
		Categories: categoryInfos,
		// Pagination fields
		TotalCount:      totalCount,
		CurrentPage:     page,
		TotalPages:      totalPages,
		ItemsPerPage:    perPage,
		BaseURL:         "/admin/api/categories/list",
		PageURL:         "/admin/categories",
		Target:          "#categories-list",
		IncludeSelector: "",
		ExtraParams:     "",
		ItemName:        "categories",
	}

	html, err := ah.handler.TemplateService.RenderFragment("categories_list.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering categories list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetCategoryEditFormHandler returns an HTML form for editing a category
func (ah *AdminAPIHandler) GetCategoryEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	categoryID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := ah.questionService.GetCategoryByID(ctx, categoryID)
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	data := services.CategoryEditFormData{
		ID:    category.ID.String(),
		Key:   category.Key,
		Label: category.Label,
		Icon:  category.Icon,
	}

	html, err := ah.handler.TemplateService.RenderFragment("category_edit_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering category edit form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// UpdateCategoryHandler updates a category
func (ah *AdminAPIHandler) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	categoryID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	category := &models.Category{
		ID:    categoryID,
		Key:   r.FormValue("key"),
		Label: r.FormValue("label"),
		Icon:  r.FormValue("icon"),
	}

	if err := ah.questionService.UpdateCategory(ctx, category); err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		log.Printf("Error updating category: %v", err)
		return
	}

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}

// CreateCategoryHandler creates a new category
func (ah *AdminAPIHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	category := &models.Category{
		Key:   r.FormValue("key"),
		Label: r.FormValue("label"),
		Icon:  r.FormValue("icon"),
	}

	if err := ah.questionService.CreateCategory(ctx, category); err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		log.Printf("Error creating category: %v", err)
		return
	}

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}

// DeleteCategoryHandler deletes a category
func (ah *AdminAPIHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	categoryID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	if err := ah.questionService.DeleteCategory(ctx, categoryID); err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		log.Printf("Error deleting category: %v", err)
		return
	}

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}

// ListRoomsHandler returns an HTML fragment with the rooms list
func (ah *AdminAPIHandler) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 25 // Default
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			// Validate against allowed values
			if parsed == 25 || parsed == 50 || parsed == 100 {
				perPage = parsed
			}
		}
	}

	offset := (page - 1) * perPage

	rooms, err := ah.adminService.ListAllRooms(ctx, perPage, offset)
	if err != nil {
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		log.Printf("Error listing rooms: %v", err)
		return
	}

	totalCount, _ := ah.adminService.GetRoomCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Build data for template
	roomInfos := make([]services.AdminRoomInfo, len(rooms))
	for i, room := range rooms {
		owner := "Unknown"
		if room.OwnerUsername != nil {
			owner = *room.OwnerUsername
		}

		guest := "Waiting..."
		if room.GuestUsername != nil {
			guest = *room.GuestUsername
		}

		statusColor := ""
		switch room.Status {
		case "active":
			statusColor = "color: #10b981;"
		case "waiting":
			statusColor = "color: #f59e0b;"
		case "completed":
			statusColor = "color: #6b7280;"
		}

		roomInfos[i] = services.AdminRoomInfo{
			ID:          room.ID.String(),
			ShortID:     room.ID.String()[:8],
			Owner:       owner,
			Guest:       guest,
			Status:      room.Status,
			StatusColor: statusColor,
			CreatedAt:   room.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	data := services.RoomsListData{
		Rooms: roomInfos,
		// Pagination fields
		TotalCount:      totalCount,
		CurrentPage:     page,
		TotalPages:      totalPages,
		ItemsPerPage:    perPage,
		BaseURL:         "/admin/api/rooms/list",
		PageURL:         "/admin/rooms",
		Target:          "#rooms-list",
		IncludeSelector: "",
		ExtraParams:     "",
		ItemName:        "rooms",
	}

	html, err := ah.handler.TemplateService.RenderFragment("rooms_list.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering rooms list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// CloseRoomHandler forces a room to close
func (ah *AdminAPIHandler) CloseRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	if err := ah.adminService.ForceCloseRoom(ctx, roomID); err != nil {
		http.Error(w, "Failed to close room", http.StatusInternalServerError)
		log.Printf("Error closing room: %v", err)
		return
	}

	// Return updated rooms list
	ah.ListRoomsHandler(w, r)
}

// GetDashboardStatsHandler returns dashboard statistics as HTML for HTMX
func (ah *AdminAPIHandler) GetDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	stats, err := ah.adminService.GetDashboardStats(ctx)
	if err != nil {
		http.Error(w, "Failed to get dashboard stats", http.StatusInternalServerError)
		log.Printf("Error getting dashboard stats: %v", err)
		return
	}

	data := services.DashboardStatsData{
		TotalUsers:      stats.TotalUsers,
		AnonymousUsers:  stats.AnonymousUsers,
		RegisteredUsers: stats.RegisteredUsers,
		AdminUsers:      stats.AdminUsers,
		TotalRooms:      stats.TotalRooms,
		ActiveRooms:     stats.ActiveRooms,
		CompletedRooms:  stats.CompletedRooms,
		TotalQuestions:  stats.TotalQuestions,
		TotalCategories: stats.TotalCategories,
	}

	html, err := ah.handler.TemplateService.RenderFragment("dashboard_stats.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering dashboard stats: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// BulkDeleteUsersHandler deletes multiple users at once
func (ah *AdminAPIHandler) BulkDeleteUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Debug: log all form values
	log.Printf("Form values: %+v", r.Form)

	userIDs := r.Form["user_ids[]"]
	log.Printf("Received %d user IDs: %v", len(userIDs), userIDs)

	if len(userIDs) == 0 {
		log.Printf("No users selected for bulk delete")
		http.Error(w, "No users selected", http.StatusBadRequest)
		return
	}

	deletedCount := 0
	for _, idStr := range userIDs {
		userID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid user ID: %s", idStr)
			continue
		}

		if err := ah.adminService.DeleteUser(ctx, userID); err != nil {
			log.Printf("Error deleting user %s: %v", idStr, err)
			continue
		}
		deletedCount++
		log.Printf("Deleted user: %s", idStr)
	}

	log.Printf("✅ Bulk deleted %d users out of %d requested", deletedCount, len(userIDs))

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// BulkDeleteQuestionsHandler deletes multiple questions at once
func (ah *AdminAPIHandler) BulkDeleteQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	questionIDs := r.Form["question_ids[]"]
	if len(questionIDs) == 0 {
		http.Error(w, "No questions selected", http.StatusBadRequest)
		return
	}

	deletedCount := 0
	for _, idStr := range questionIDs {
		questionID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid question ID: %s", idStr)
			continue
		}

		if err := ah.questionService.DeleteQuestion(ctx, questionID); err != nil {
			log.Printf("Error deleting question %s: %v", idStr, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Bulk deleted %d questions", deletedCount)

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}

// BulkDeleteCategoriesHandler deletes multiple categories at once
func (ah *AdminAPIHandler) BulkDeleteCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryIDs := r.Form["category_ids[]"]
	if len(categoryIDs) == 0 {
		http.Error(w, "No categories selected", http.StatusBadRequest)
		return
	}

	deletedCount := 0
	for _, idStr := range categoryIDs {
		categoryID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid category ID: %s", idStr)
			continue
		}

		if err := ah.questionService.DeleteCategory(ctx, categoryID); err != nil {
			log.Printf("Error deleting category %s: %v", idStr, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Bulk deleted %d categories", deletedCount)

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}

// BulkCloseRoomsHandler closes multiple rooms at once
func (ah *AdminAPIHandler) BulkCloseRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	roomIDs := r.Form["room_ids[]"]
	if len(roomIDs) == 0 {
		http.Error(w, "No rooms selected", http.StatusBadRequest)
		return
	}

	closedCount := 0
	for _, idStr := range roomIDs {
		roomID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid room ID: %s", idStr)
			continue
		}

		if err := ah.adminService.ForceCloseRoom(ctx, roomID); err != nil {
			log.Printf("Error closing room %s: %v", idStr, err)
			continue
		}
		closedCount++
	}

	log.Printf("Bulk closed %d rooms", closedCount)

	// Return updated rooms list
	ah.ListRoomsHandler(w, r)
}
