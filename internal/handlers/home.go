package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
)

// HomeHandler handles the home page
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get session user data
	sessionUser := GetSessionUser(r)

	// Check if user has a username set
	if sessionUser != nil {
		ctx := context.Background()
		userID, _ := uuid.Parse(sessionUser.ID)
		userObj, err := h.UserService.GetUserByID(ctx, userID)
		if err == nil && userObj != nil {
			// Check if username is temporary (starts with "guest_" or "user_")
			isTemporary := len(userObj.Username) >= 6 && (
				userObj.Username[:6] == "guest_" ||
				userObj.Username[:5] == "user_")

			// If username is empty or temporary, redirect to setup
			if userObj.Username == "" || isTemporary {
				http.Redirect(w, r, "/setup-username", http.StatusSeeOther)
				return
			}
		}
	}

	data := &TemplateData{
		Title:   "Home - Couple Card Game",
		User:    sessionUser,
		IsAdmin: sessionUser != nil && sessionUser.IsAdmin,
	}

	h.RenderTemplate(w, "home.html", data)
}

// SetupUsernameHandler shows the username setup page
func (h *Handler) SetupUsernameHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Setup Username",
	}
	h.RenderTemplate(w, "setup-username.html", data)
}

// SetupUsernamePostHandler handles username setup submission
func (h *Handler) SetupUsernamePostHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	username := r.FormValue("username")
	if len(username) < 3 || len(username) > 20 {
		http.Error(w, "Username must be 3-20 characters", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.UserService.UpdateUsername(ctx, userID.(uuid.UUID), username); err != nil {
		http.Error(w, "Username already taken or invalid", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HealthHandler handles health check requests
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

