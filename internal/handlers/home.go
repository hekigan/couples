package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/labstack/echo/v4"
)

// HomeHandler handles the home page
func (h *Handler) HomeHandler(c echo.Context) error {
	// Get session user data
	sessionUser := GetSessionUser(c)

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
				return c.Redirect(http.StatusSeeOther, "/setup-username")
			}
		}
	}

	data := &TemplateData{
		Title:   "Home - Couple Card Game",
		User:    sessionUser,
		IsAdmin: sessionUser != nil && sessionUser.IsAdmin,
	}

	return h.RenderTemplate(c, "home.html", data)
}

// SetupUsernameHandler shows the username setup page
func (h *Handler) SetupUsernameHandler(c echo.Context) error {
	data := &TemplateData{
		Title: "Setup Username",
	}
	return h.RenderTemplate(c, "setup-username.html", data)
}

// SetupUsernamePostHandler handles username setup submission
func (h *Handler) SetupUsernamePostHandler(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	username := c.FormValue("username")
	if len(username) < 3 || len(username) > 20 {
		return echo.NewHTTPError(http.StatusBadRequest, "Username must be 3-20 characters")
	}

	ctx := context.Background()
	if err := h.UserService.UpdateUsername(ctx, userID, username); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Username already taken or invalid")
	}

	return c.NoContent(http.StatusOK)
}

// HealthHandler handles health check requests
func (h *Handler) HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}
