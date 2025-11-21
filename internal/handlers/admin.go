package handlers

import (
	"context"
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

	data := &TemplateData{
		Title:   "User Management",
		User:    currentUser,
		IsAdmin: true,
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

	data := &TemplateData{
		Title:   "Question Management",
		User:    currentUser,
		IsAdmin: true,
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

	data := &TemplateData{
		Title:   "Category Management",
		User:    currentUser,
		IsAdmin: true,
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

	data := &TemplateData{
		Title:   "Room Monitoring",
		User:    currentUser,
		IsAdmin: true,
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



