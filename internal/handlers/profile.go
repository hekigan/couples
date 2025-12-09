package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
)

// ProfileHandler displays the user's profile
func (h *Handler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	user, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	data := &TemplateData{
		Title: "My Profile",
		User:  user,
		Data:  user,
	}
	h.RenderTemplate(w, "profile.html", data)
}

// UpdateProfileHandler updates the user's profile
func (h *Handler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	user, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	// Update user fields from form
	if username := r.FormValue("username"); username != "" {
		user.Username = username
	}

	if err := h.UserService.UpdateUser(ctx, user); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

