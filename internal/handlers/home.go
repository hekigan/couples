package handlers

import (
	"net/http"

	"github.com/yourusername/couple-card-game/internal/middleware"
)

// HomeHandler handles the home page
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context if logged in
	var user interface{}
	if userID := r.Context().Value(middleware.UserIDKey); userID != nil {
		// User is logged in
		user = userID
	}

	data := &TemplateData{
		Title: "Home - Couple Card Game",
		User:  user,
	}

	h.RenderTemplate(w, "home.html", data)
}

// HealthHandler handles health check requests
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

