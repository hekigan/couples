package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/services"
)

// ListUsersHandler returns an HTML fragment with the users list
func (ah *AdminAPIHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper for pagination
	page, perPage, offset := handlers.ParsePaginationParams(r)

	users, err := ah.adminService.ListAllUsers(ctx, perPage, offset)
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		log.Printf("Error listing users: %v", err)
		return
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
	currentUser := handlers.GetSessionUser(r)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering users list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ToggleUserAdminHandler toggles a user's admin status
func (ah *AdminAPIHandler) ToggleUserAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	userID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ah.adminService.ToggleUserAdmin(ctx, userID); err != nil {
		http.Error(w, "Failed to toggle admin status", http.StatusInternalServerError)
		log.Printf("Error toggling admin status: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// DeleteUserHandler deletes a user
func (ah *AdminAPIHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	userID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ah.adminService.DeleteUser(ctx, userID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		log.Printf("Error deleting user: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// GetUserCreateFormHandler returns an HTML form for creating a user
func (ah *AdminAPIHandler) GetUserCreateFormHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering user create form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetUserEditFormHandler returns an HTML form for editing a user
func (ah *AdminAPIHandler) GetUserEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	userID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := ah.adminService.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering user edit form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// CreateUserHandler creates a new user
func (ah *AdminAPIHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	isAdmin := r.FormValue("is_admin") == "on"
	isAnonymous := r.FormValue("is_anonymous") == "on"

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	_, err := ah.adminService.CreateUser(ctx, username, email, isAdmin, isAnonymous)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		log.Printf("Error creating user: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// UpdateUserHandler updates a user
func (ah *AdminAPIHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	userID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	isAdmin := r.FormValue("is_admin") == "on"
	isAnonymous := r.FormValue("is_anonymous") == "on"

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if err := ah.adminService.UpdateUser(ctx, userID, username, email, isAdmin, isAnonymous); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		log.Printf("Error updating user: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}
