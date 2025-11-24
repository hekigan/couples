package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
)

// AdminDashboardHandler displays admin dashboard
func (h *Handler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get admin service from handler
	adminService := h.GetAdminService()
	if adminService == nil {
		log.Printf("⚠️ AdminService not available")
		http.Error(w, "Admin service not configured", http.StatusInternalServerError)
		return
	}

	// Get dashboard statistics
	stats, err := adminService.GetDashboardStats(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to fetch dashboard stats: %v", err)
		stats = nil
	}

	data := &TemplateData{
		Title:   "Admin Dashboard",
		User:    GetSessionUser(r), // Get user from session (no DB call)
		Data:    stats,
		IsAdmin: true, // Admin pages are protected by middleware
	}

	h.RenderTemplate(w, "admin/dashboard.html", data)
}

// AdminUsersHandler displays user management
func (h *Handler) AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25 // Default
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil {
			// Validate against allowed values
			if pp == 25 || pp == 50 || pp == 100 {
				perPage = pp
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Fetch users list for SSR
	users, err := h.AdminService.ListAllUsers(ctx, perPage, offset)
	if err != nil {
		log.Printf("⚠️ Failed to fetch users: %v", err)
		users = nil
	}

	totalCount, _ := h.AdminService.GetUserCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Build users list data
	var usersData *services.UsersListData
	if users != nil {
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
		currentUser := GetSessionUser(r)
		currentUserID := ""
		if currentUser != nil {
			currentUserID = currentUser.ID
		}

		usersData = &services.UsersListData{
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
	}

	data := &TemplateData{
		Title:   "User Management",
		User:    GetSessionUser(r), // Get user from session (no DB call)
		IsAdmin: true,
		Data:    usersData,
	}
	h.RenderTemplate(w, "admin/users.html", data)
}

// AdminQuestionsHandler displays question management
func (h *Handler) AdminQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse query parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25 // Default
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil {
			// Validate against allowed values
			if pp == 25 || pp == 50 || pp == 100 {
				perPage = pp
			}
		}
	}

	var categoryID *uuid.UUID
	if catIDStr := r.URL.Query().Get("category_id"); catIDStr != "" {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			categoryID = &catID
		}
	}

	// Always filter to English only
	langCode := "en"

	// Calculate offset
	offset := (page - 1) * perPage

	// Fetch questions for current page
	questions, err := h.QuestionService.ListQuestions(ctx, perPage, offset, categoryID, &langCode)
	if err != nil {
		log.Printf("⚠️ Failed to fetch questions: %v", err)
		questions = nil
	}

	// Get total count for pagination (English only)
	totalCount, err := h.QuestionService.GetQuestionCountsByCategory(ctx, "en")
	if err != nil {
		log.Printf("⚠️ Failed to get question counts: %v", err)
	}
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

	// Get categories
	categories, err := h.QuestionService.GetCategories(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to get categories: %v", err)
	}

	// Get question counts by category (for dropdown)
	counts, err := h.QuestionService.GetQuestionCountsByCategory(ctx, "en")
	if err != nil {
		log.Printf("⚠️ Failed to get question counts: %v", err)
		counts = make(map[string]int)
	}

	// Build questions list data
	var questionsData *services.QuestionsListData
	if categories != nil {
		// Handle nil questions as empty slice
		if questions == nil {
			questions = []models.Question{}
		}
		// Get translation status for current page questions
		questionIDs := make([]uuid.UUID, len(questions))
		for i, q := range questions {
			questionIDs[i] = q.ID
		}
		translationStatus, _ := h.QuestionService.GetQuestionTranslationStatus(ctx, questionIDs)

		// Calculate total missing translations (across ALL English questions, not just current page)
		allEnglishQuestions, _ := h.QuestionService.ListQuestions(ctx, 10000, 0, nil, &langCode) // Get all
		allIDs := make([]uuid.UUID, len(allEnglishQuestions))
		for i, q := range allEnglishQuestions {
			allIDs[i] = q.ID
		}
		allTranslationStatus, _ := h.QuestionService.GetQuestionTranslationStatus(ctx, allIDs)
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
		categoryMap := make(map[uuid.UUID]string)
		for _, cat := range categories {
			categoryMap[cat.ID] = cat.Icon + " " + cat.Label
		}

		// Build question infos
		questionInfos := make([]services.AdminQuestionInfo, len(questions))
		for i, q := range questions {
			categoryLabel := "Unknown"
			if label, ok := categoryMap[q.CategoryID]; ok {
				categoryLabel = label
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

		questionsData = &services.QuestionsListData{
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
	} else {
		log.Printf("⚠️ Categories is nil, cannot render questions list")
	}

	data := &TemplateData{
		Title:   "Question Management",
		User:    GetSessionUser(r), // Get user from session (no DB call)
		IsAdmin: true,
		Data:    questionsData,
	}
	h.RenderTemplate(w, "admin/questions.html", data)
}

// AdminCategoriesHandler displays category management
func (h *Handler) AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25 // Default
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil {
			// Validate against allowed values
			if pp == 25 || pp == 50 || pp == 100 {
				perPage = pp
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Fetch categories list for SSR
	categories, err := h.QuestionService.ListCategories(ctx, perPage, offset)
	if err != nil {
		log.Printf("⚠️ Failed to fetch categories: %v", err)
		categories = nil
	}

	totalCount, _ := h.QuestionService.GetCategoryCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	counts, _ := h.QuestionService.GetQuestionCountsByCategory(ctx, "en")

	// Build categories list data
	var categoriesData *services.CategoriesListData
	if categories != nil {
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

		categoriesData = &services.CategoriesListData{
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
	}

	data := &TemplateData{
		Title:   "Category Management",
		User:    GetSessionUser(r), // Get user from session (no DB call)
		IsAdmin: true,
		Data:    categoriesData,
	}
	h.RenderTemplate(w, "admin/categories.html", data)
}

// AdminRoomsHandler displays room monitoring
func (h *Handler) AdminRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25 // Default
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil {
			// Validate against allowed values
			if pp == 25 || pp == 50 || pp == 100 {
				perPage = pp
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Fetch rooms list for SSR
	rooms, err := h.AdminService.ListAllRooms(ctx, perPage, offset)
	if err != nil {
		log.Printf("⚠️ Failed to fetch rooms: %v", err)
		rooms = nil
	}

	totalCount, _ := h.AdminService.GetRoomCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Build rooms list data
	var roomsData *services.RoomsListData
	if rooms != nil {
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

		roomsData = &services.RoomsListData{
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
	}

	data := &TemplateData{
		Title:   "Room Monitoring",
		User:    GetSessionUser(r), // Get user from session (no DB call)
		IsAdmin: true,
		Data:    roomsData,
	}
	h.RenderTemplate(w, "admin/rooms.html", data)
}

// AdminTranslationsHandler displays translation management
func (h *Handler) AdminTranslationsHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title:   "Translation Management",
		User:    GetSessionUser(r), // Get user from session (no DB call)
		IsAdmin: true,
	}
	h.RenderTemplate(w, "admin/translations.html", data)
}



