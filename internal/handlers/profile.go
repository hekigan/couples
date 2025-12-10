package handlers

import (
	"context"
	"net/http"

	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// ProfileHandler displays the user's profile
func (h *Handler) ProfileHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	user, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load profile")
	}

	data := NewTemplateData(c)
	data.Title = "My Profile"
	data.User = user
	data.Data = user
	return h.RenderTemplComponent(c, pages.ProfilePage(data))
}

// UpdateProfileHandler updates the user's profile
func (h *Handler) UpdateProfileHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	user, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load profile")
	}

	// Update user fields from form
	if username := c.FormValue("username"); username != "" {
		user.Username = username
	}

	if err := h.UserService.UpdateUser(ctx, user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}

	return c.Redirect(http.StatusSeeOther, "/profile")
}
