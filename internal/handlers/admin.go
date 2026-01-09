package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	adminPages "github.com/hekigan/couples/internal/views/pages/admin"
	"github.com/labstack/echo/v4"
)

// AdminDashboardHandler displays admin dashboard
func (h *Handler) AdminDashboardHandler(c echo.Context) error {
	ctx := context.Background()

	// Get admin service from handler
	adminService := h.GetAdminService()
	if adminService == nil {
		log.Printf("⚠️ AdminService not available")
		return echo.NewHTTPError(http.StatusInternalServerError, "Admin service not configured")
	}

	// Get dashboard statistics
	stats, err := adminService.GetDashboardStats(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to fetch dashboard stats: %v", err)
		stats = nil
	}

	// Convert to DashboardStatsData for template
	var statsData *services.DashboardStatsData
	if stats != nil {
		statsData = &services.DashboardStatsData{
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
	}

	data := &TemplateData{
		Title:     "Admin Dashboard",
		User:      GetTemplateUser(c), // Use helper to avoid nil interface gotcha
		Data:      statsData,
		IsAdmin:   true, // Admin pages are protected by middleware
		Env:       os.Getenv("ENV"),
		CSRFToken: GetCSRFToken(c),
	}

	return h.RenderTemplComponent(c, adminPages.DashboardPage(data))
}

// AdminUsersHandler displays user management
func (h *Handler) AdminUsersHandler(c echo.Context) error {
	ctx := context.Background()

	// Parse pagination parameters
	page, perPage := ParsePaginationParams(c)

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

			userType := "registered"
			if user.IsAnonymous {
				userType = "guest"
			}
			if user.IsAdmin {
				userType = "admin"
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
		currentUser := GetSessionUser(c)
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
		Title:     "User Management",
		User:      GetTemplateUser(c), // Use helper to avoid nil interface gotcha
		IsAdmin:   true,
		Data:      usersData,
		Env:       os.Getenv("ENV"),
		CSRFToken: GetCSRFToken(c),
	}
	return h.RenderTemplComponent(c, adminPages.UsersPage(data))
}

// AdminQuestionsHandler displays question management
func (h *Handler) AdminQuestionsHandler(c echo.Context) error {
	ctx := context.Background()

	// Parse pagination parameters
	page, perPage := ParsePaginationParams(c)

	var categoryID *uuid.UUID
	if catIDStr := c.QueryParam("category_id"); catIDStr != "" {
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
	categories, err := h.CategoryService.GetCategories(ctx)
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
				ID:            cat.ID.String(),
				Label:         cat.Label,
				Selected:      selected,
				QuestionCount: count,
			}
		}

		// Create category map for lookups
		categoryMap := make(map[uuid.UUID]string)
		for _, cat := range categories {
			categoryMap[cat.ID] = cat.Label
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
		Title:     "Question Management",
		User:      GetTemplateUser(c), // Use helper to avoid nil interface gotcha
		IsAdmin:   true,
		Data:      questionsData,
		Env:       os.Getenv("ENV"),
		CSRFToken: GetCSRFToken(c),
	}
	return h.RenderTemplComponent(c, adminPages.QuestionsPage(data))
}

// AdminCategoriesHandler displays category management
func (h *Handler) AdminCategoriesHandler(c echo.Context) error {
	ctx := context.Background()

	// Parse pagination parameters
	page, perPage := ParsePaginationParams(c)

	// Calculate offset
	offset := (page - 1) * perPage

	// Fetch categories list for SSR
	categories, err := h.CategoryService.ListCategories(ctx, perPage, offset)
	if err != nil {
		log.Printf("⚠️ Failed to fetch categories: %v", err)
		categories = nil
	}

	totalCount, _ := h.CategoryService.GetCategoryCount(ctx)

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
		Title:     "Category Management",
		User:      GetTemplateUser(c), // Use helper to avoid nil interface gotcha
		IsAdmin:   true,
		Data:      categoriesData,
		Env:       os.Getenv("ENV"),
		CSRFToken: GetCSRFToken(c),
	}
	return h.RenderTemplComponent(c, adminPages.CategoriesPage(data))
}

// AdminRoomsHandler displays room monitoring
func (h *Handler) AdminRoomsHandler(c echo.Context) error {
	ctx := context.Background()

	// Parse pagination parameters
	page, perPage := ParsePaginationParams(c)

	// Calculate offset
	offset := (page - 1) * perPage

	// Fetch rooms list for SSR with default filters (all statuses, no search, sort by created_at desc)
	defaultStatuses := []string{"waiting", "ready", "playing", "finished"}
	rooms, err := h.AdminService.ListAllRooms(ctx, perPage, offset, "", defaultStatuses, "created_at", "desc")
	if err != nil {
		log.Printf("⚠️ Failed to fetch rooms: %v", err)
		rooms = nil
	}

	totalCount, _ := h.AdminService.GetFilteredRoomCount(ctx, "", defaultStatuses)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Fetch all categories once for mapping
	allCategories, err := h.CategoryService.GetCategories(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
	}

	// Build category map for quick lookup
	categoryMap := make(map[string]string) // ID -> Label
	for _, cat := range allCategories {
		categoryMap[cat.ID.String()] = cat.Label
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

			// Get category names for this room
			categoryNames := []string{}
			for _, selectedID := range room.SelectedCategories {
				if label, ok := categoryMap[selectedID.String()]; ok {
					categoryNames = append(categoryNames, label)
				}
			}

			roomInfos[i] = services.AdminRoomInfo{
				ID:            room.ID.String(),
				ShortID:       room.ID.String()[:8],
				Name:          room.Name,
				Owner:         owner,
				Guest:         guest,
				Status:        room.Status,
				CreatedAt:     room.CreatedAt.Format("2006-01-02 15:04"),
				CategoryNames: categoryNames,
			}
		}

		roomsData = &services.RoomsListData{
			Rooms: roomInfos,
			// Pagination fields
			TotalCount:      totalCount,
			CurrentPage:     page,
			TotalPages:      totalPages,
			ItemsPerPage:    perPage,
			BaseURL:         "/admin/api/v1/rooms/list",
			PageURL:         "/admin/rooms",
			Target:          "#rooms-list",
			IncludeSelector: "",
			ExtraParams:     "",
			ItemName:        "rooms",
			// Filter/sort state (default values for initial page load)
			SearchTerm:       "",
			SelectedStatuses: defaultStatuses,
			SortBy:           "created_at",
			SortOrder:        "desc",
		}
	}

	data := &TemplateData{
		Title:     "Room Monitoring",
		User:      GetTemplateUser(c), // Use helper to avoid nil interface gotcha
		IsAdmin:   true,
		Data:      roomsData,
		Env:       os.Getenv("ENV"),
		CSRFToken: GetCSRFToken(c),
	}
	return h.RenderTemplComponent(c, adminPages.RoomsPage(data))
}

// AdminTranslationsHandler displays translation management
func (h *Handler) AdminTranslationsHandler(c echo.Context) error {
	data := &TemplateData{
		Title:     "Translation Management",
		User:      GetTemplateUser(c), // Use helper to avoid nil interface gotcha
		IsAdmin:   true,
		Env:       os.Getenv("ENV"),
		CSRFToken: GetCSRFToken(c),
	}
	return h.RenderTemplComponent(c, adminPages.TranslationsPage(data))
}

// AdminRoutesHandler displays all application routes
func (h *Handler) AdminRoutesHandler(c echo.Context) error {
	currentUser := GetTemplateUser(c)

	// Get all routes from Echo instance
	allRoutes := []services.RouteDisplayInfo{}
	stats := services.RouteStats{}

	for _, route := range h.echo.Routes() {
		// Skip HEAD methods (auto-generated by Echo)
		if route.Method == "HEAD" {
			continue
		}

		// Skip internal Echo routes (404 handler, etc.)
		if strings.Contains(route.Name, "echo_route") {
			continue
		}

		// Skip routes with non-standard HTTP methods
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true,
			"PATCH": true, "OPTIONS": true,
		}
		if !validMethods[route.Method] {
			continue
		}

		version := determineVersion(route.Path)
		isHTMX := isHTMXFragment(route.Path)

		displayInfo := services.RouteDisplayInfo{
			Method:      route.Method,
			Path:        route.Path,
			Version:     version,
			IsHTMX:      isHTMX,
			MethodClass: "method-" + strings.ToLower(route.Method),
		}

		allRoutes = append(allRoutes, displayInfo)

		// Count stats
		stats.TotalRoutes++
		switch version {
		case "v1":
			stats.V1Routes++
		case "admin-v1":
			stats.AdminV1Routes++
		case "unversioned":
			stats.UnversionedRoutes++
		}
	}

	// Sort routes by path, then by method
	sort.Slice(allRoutes, func(i, j int) bool {
		if allRoutes[i].Path == allRoutes[j].Path {
			return allRoutes[i].Method < allRoutes[j].Method
		}
		return allRoutes[i].Path < allRoutes[j].Path
	})

	// Group routes by version
	var apiv1Routes, adminAPIv1Routes, unversionedRoutes []services.RouteDisplayInfo
	htmxFragmentCount := 0

	for _, route := range allRoutes {
		if route.IsHTMX {
			htmxFragmentCount++
		}

		switch route.Version {
		case "v1":
			apiv1Routes = append(apiv1Routes, route)
		case "admin-v1":
			adminAPIv1Routes = append(adminAPIv1Routes, route)
		case "unversioned":
			unversionedRoutes = append(unversionedRoutes, route)
		}
	}

	routesData := &services.RoutesPageData{
		Stats:             stats,
		APIv1Routes:       apiv1Routes,
		AdminAPIv1Routes:  adminAPIv1Routes,
		UnversionedRoutes: unversionedRoutes,
		HTMXFragmentCount: htmxFragmentCount,
	}

	data := NewTemplateData(c)
	h.PopulateNotificationCount(c, data)
	data.Title = "Route Registry"
	data.User = currentUser
	data.Data = routesData

	return h.RenderTemplComponent(c, adminPages.RoutesPage(data))
}

// Helper functions (copied from cmd/server/routes_registry.go to avoid import issues)

// determineVersion identifies the route version from its path
func determineVersion(path string) string {
	if len(path) >= 8 && path[:8] == "/api/v1/" {
		return "v1"
	}
	if len(path) >= 15 && path[:15] == "/admin/api/v1/" {
		return "admin-v1"
	}
	return "unversioned"
}

// isHTMXFragment checks if a route path represents an HTMX fragment endpoint
func isHTMXFragment(path string) bool {
	patterns := []string{
		"-button", "-badge", "-indicator", "-card", "-form",
		"-content", "-counter", "/list", "/edit-form", "/new",
	}

	for _, pattern := range patterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}
