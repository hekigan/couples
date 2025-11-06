package handlers

import (
	"net/http"
)

// AdminDashboardHandler displays admin dashboard
func (h *Handler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Admin Dashboard",
	}
	h.RenderTemplate(w, "admin/dashboard.html", data)
}

// AdminUsersHandler displays user management
func (h *Handler) AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "User Management",
	}
	h.RenderTemplate(w, "admin/users.html", data)
}

// AdminQuestionsHandler displays question management
func (h *Handler) AdminQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Question Management",
	}
	h.RenderTemplate(w, "admin/questions.html", data)
}

// AdminCategoriesHandler displays category management
func (h *Handler) AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Category Management",
	}
	h.RenderTemplate(w, "admin/categories.html", data)
}

// AdminRoomsHandler displays room monitoring
func (h *Handler) AdminRoomsHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Room Monitoring",
	}
	h.RenderTemplate(w, "admin/rooms.html", data)
}

// AdminTranslationsHandler displays translation management
func (h *Handler) AdminTranslationsHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Translation Management",
	}
	h.RenderTemplate(w, "admin/translations.html", data)
}

