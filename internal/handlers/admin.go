package handlers

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourusername/couple-card-game/internal/middleware"
)

// AdminDashboardHandler displays admin dashboard
func (h *Handler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		currentUser = nil
	}

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
		User:    currentUser,
		Data:    stats,
		IsAdmin: true, // Admin pages are protected by middleware
	}
	h.RenderTemplate(w, "admin/dashboard.html", data)
}

// AdminUsersHandler displays user management
func (h *Handler) AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		currentUser = nil
	}

	// Fetch users list for SSR
	limit := 50
	offset := 0
	users, err := h.AdminService.ListAllUsers(ctx, limit, offset)
	if err != nil {
		log.Printf("⚠️ Failed to fetch users: %v", err)
		users = nil
	}

	totalCount, _ := h.AdminService.GetUserCount(ctx)

	// Build users list HTML
	var usersListHTML string
	if users != nil {
		userInfos := make([]interface{}, len(users))
		for i, user := range users {
			email := ""
			if user.Email != nil {
				email = *user.Email
			}

			userType := "Registered"
			if user.IsAnonymous {
				userType = "Anonymous"
			}

			userInfos[i] = map[string]interface{}{
				"ID":        user.ID.String(),
				"Username":  user.Username,
				"Email":     email,
				"UserType":  userType,
				"IsAdmin":   user.IsAdmin,
				"CreatedAt": user.CreatedAt.Format("2006-01-02"),
			}
		}

		usersData := map[string]interface{}{
			"Users":      userInfos,
			"TotalCount": totalCount,
			"Page":       1,
		}

		usersListHTML, _ = h.TemplateService.RenderFragment("users_list.html", usersData)
	}

	data := &TemplateData{
		Title:   "User Management",
		User:    currentUser,
		IsAdmin: true,
		Data:    template.HTML(usersListHTML),
	}
	h.RenderTemplate(w, "admin/users.html", data)
}

// AdminQuestionsHandler displays question management
func (h *Handler) AdminQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		currentUser = nil
	}

	// Fetch questions list for SSR
	limit := 50
	offset := 0
	questions, err := h.QuestionService.ListQuestions(ctx, limit, offset, nil, nil)
	if err != nil {
		log.Printf("⚠️ Failed to fetch questions: %v", err)
		questions = nil
	}

	categories, _ := h.QuestionService.GetCategories(ctx)

	// Build questions list HTML
	var questionsListHTML string
	if questions != nil && categories != nil {
		// Build category options
		categoryOptions := make([]interface{}, len(categories))
		for i, cat := range categories {
			categoryOptions[i] = map[string]interface{}{
				"ID":       cat.ID.String(),
				"Icon":     cat.Icon,
				"Label":    cat.Label,
				"Selected": false,
			}
		}

		// Create category map for lookups
		categoryMap := make(map[uuid.UUID]string)
		for _, cat := range categories {
			categoryMap[cat.ID] = cat.Icon + " " + cat.Label
		}

		// Build question infos
		questionInfos := make([]interface{}, len(questions))
		for i, q := range questions {
			categoryLabel := "Unknown"
			if label, ok := categoryMap[q.CategoryID]; ok {
				categoryLabel = label
			}

			questionInfos[i] = map[string]interface{}{
				"ID":            q.ID.String(),
				"Text":          q.Text,
				"CategoryLabel": categoryLabel,
				"LanguageCode":  q.LanguageCode,
			}
		}

		questionsData := map[string]interface{}{
			"Questions":          questionInfos,
			"Categories":         categoryOptions,
			"SelectedLanguage":   "",
			"LanguageEnSelected": false,
			"LanguageFrSelected": false,
			"LanguageJaSelected": false,
		}

		questionsListHTML, _ = h.TemplateService.RenderFragment("questions_list.html", questionsData)
	}

	data := &TemplateData{
		Title:   "Question Management",
		User:    currentUser,
		IsAdmin: true,
		Data:    template.HTML(questionsListHTML),
	}
	h.RenderTemplate(w, "admin/questions.html", data)
}

// AdminCategoriesHandler displays category management
func (h *Handler) AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		currentUser = nil
	}

	// Fetch categories list for SSR
	categories, err := h.QuestionService.GetCategories(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to fetch categories: %v", err)
		categories = nil
	}

	counts, _ := h.QuestionService.GetQuestionCountsByCategory(ctx, "en")

	// Build categories list HTML
	var categoriesListHTML string
	if categories != nil {
		categoryInfos := make([]interface{}, len(categories))
		for i, cat := range categories {
			count := 0
			if c, ok := counts[cat.ID.String()]; ok {
				count = c
			}

			categoryInfos[i] = map[string]interface{}{
				"ID":            cat.ID.String(),
				"Icon":          cat.Icon,
				"Label":         cat.Label,
				"Key":           cat.Key,
				"QuestionCount": count,
			}
		}

		categoriesData := map[string]interface{}{
			"Categories": categoryInfos,
		}

		categoriesListHTML, _ = h.TemplateService.RenderFragment("categories_list.html", categoriesData)
	}

	data := &TemplateData{
		Title:   "Category Management",
		User:    currentUser,
		IsAdmin: true,
		Data:    template.HTML(categoriesListHTML),
	}
	h.RenderTemplate(w, "admin/categories.html", data)
}

// AdminRoomsHandler displays room monitoring
func (h *Handler) AdminRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		currentUser = nil
	}

	// Fetch rooms list for SSR
	limit := 50
	offset := 0
	rooms, err := h.AdminService.ListAllRooms(ctx, limit, offset)
	if err != nil {
		log.Printf("⚠️ Failed to fetch rooms: %v", err)
		rooms = nil
	}

	// Build rooms list HTML
	var roomsListHTML string
	if rooms != nil {
		roomInfos := make([]interface{}, len(rooms))
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

			roomInfos[i] = map[string]interface{}{
				"ID":          room.ID.String(),
				"ShortID":     room.ID.String()[:8],
				"Owner":       owner,
				"Guest":       guest,
				"Status":      room.Status,
				"StatusColor": statusColor,
				"CreatedAt":   room.CreatedAt.Format("2006-01-02 15:04"),
			}
		}

		roomsData := map[string]interface{}{
			"Rooms": roomInfos,
		}

		roomsListHTML, _ = h.TemplateService.RenderFragment("rooms_list.html", roomsData)
	}

	data := &TemplateData{
		Title:   "Room Monitoring",
		User:    currentUser,
		IsAdmin: true,
		Data:    template.HTML(roomsListHTML),
	}
	h.RenderTemplate(w, "admin/rooms.html", data)
}

// AdminTranslationsHandler displays translation management
func (h *Handler) AdminTranslationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		currentUser = nil
	}

	data := &TemplateData{
		Title:   "Translation Management",
		User:    currentUser,
		IsAdmin: true,
	}
	h.RenderTemplate(w, "admin/translations.html", data)
}



