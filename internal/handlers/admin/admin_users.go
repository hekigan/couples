package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// ListUsersHandler returns an HTML fragment with the users list
func (ah *AdminAPIHandler) ListUsersHandler(c echo.Context) error {
	ctx := context.Background()

	// Use helper for pagination
	page, perPage := handlers.ParsePaginationParams(c)
	offset := (page - 1) * perPage

	users, err := ah.adminService.ListAllUsers(ctx, perPage, offset)
	if err != nil {
		log.Printf("Error listing users: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list users")
	}

	totalCount, _ := ah.adminService.GetUserCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Build data for template
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
	currentUser := handlers.GetSessionUser(c)
	currentUserID := ""
	if currentUser != nil {
		currentUserID = currentUser.ID
	}

	data := services.UsersListData{
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

	html, err := ah.handler.TemplateService.RenderFragment("users_list", data)
	if err != nil {
		log.Printf("Error rendering users list: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// ToggleUserAdminHandler toggles a user's admin status
func (ah *AdminAPIHandler) ToggleUserAdminHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	userID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ah.adminService.ToggleUserAdmin(ctx, userID); err != nil {
		log.Printf("Error toggling admin status: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to toggle admin status")
	}

	// Return updated users list
	return ah.ListUsersHandler(c)
}

// DeleteUserHandler deletes a user
func (ah *AdminAPIHandler) DeleteUserHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	userID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ah.adminService.DeleteUser(ctx, userID); err != nil {
		log.Printf("Error deleting user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user")
	}

	// Return updated users list
	return ah.ListUsersHandler(c)
}

// GetUserCreateFormHandler returns an HTML form for creating a user
func (ah *AdminAPIHandler) GetUserCreateFormHandler(c echo.Context) error {
	// Create empty form data for create mode
	data := services.UserFormData{
		ID:          "", // Empty ID indicates create mode
		Username:    "",
		Email:       "",
		IsAdmin:     false,
		IsAnonymous: false,
	}

	html, err := ah.handler.TemplateService.RenderFragment("user_form.html", data)
	if err != nil {
		log.Printf("Error rendering user create form: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// GetUserEditFormHandler returns an HTML form for editing a user
func (ah *AdminAPIHandler) GetUserEditFormHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	userID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := ah.adminService.GetUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	data := services.UserFormData{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       email,
		IsAdmin:     user.IsAdmin,
		IsAnonymous: user.IsAnonymous,
	}

	html, err := ah.handler.TemplateService.RenderFragment("user_form.html", data)
	if err != nil {
		log.Printf("Error rendering user edit form: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// CreateUserHandler creates a new user
func (ah *AdminAPIHandler) CreateUserHandler(c echo.Context) error {
	ctx := context.Background()

	username := c.FormValue("username")
	email := c.FormValue("email")
	isAdmin := c.FormValue("is_admin") == "on"
	isAnonymous := c.FormValue("is_anonymous") == "on"

	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Username is required")
	}

	_, err := ah.adminService.CreateUser(ctx, username, email, isAdmin, isAnonymous)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	// Return updated users list
	return ah.ListUsersHandler(c)
}

// UpdateUserHandler updates a user
func (ah *AdminAPIHandler) UpdateUserHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	userID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	username := c.FormValue("username")
	email := c.FormValue("email")
	isAdmin := c.FormValue("is_admin") == "on"
	isAnonymous := c.FormValue("is_anonymous") == "on"

	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Username is required")
	}

	if err := ah.adminService.UpdateUser(ctx, userID, username, email, isAdmin, isAnonymous); err != nil {
		log.Printf("Error updating user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user")
	}

	// Return updated users list
	return ah.ListUsersHandler(c)
}
