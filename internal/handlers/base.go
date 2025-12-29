package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
	"github.com/hekigan/couples/internal/viewmodels"
	"github.com/labstack/echo/v4"
)

// Handler holds all service dependencies
type Handler struct {
	UserService         *services.UserService
	RoomService         *services.RoomService
	GameService         *services.GameService
	QuestionService     *services.QuestionService
	CategoryService     *services.CategoryService
	AnswerService       *services.AnswerService
	FriendService       *services.FriendService
	I18nService         *services.I18nService
	NotificationService *services.NotificationService
	AdminService        *services.AdminService // For admin operations
	echo                *echo.Echo             // Echo instance for route introspection
}

// NewHandler creates a new handler with all dependencies
func NewHandler(
	userService *services.UserService,
	roomService *services.RoomService,
	gameService *services.GameService,
	questionService *services.QuestionService,
	categoryService *services.CategoryService,
	answerService *services.AnswerService,
	friendService *services.FriendService,
	i18nService *services.I18nService,
	notificationService *services.NotificationService,
	adminService *services.AdminService,
	e *echo.Echo,
) *Handler {
	return &Handler{
		UserService:         userService,
		RoomService:         roomService,
		GameService:         gameService,
		QuestionService:     questionService,
		CategoryService:     categoryService,
		AnswerService:       answerService,
		FriendService:       friendService,
		I18nService:         i18nService,
		NotificationService: notificationService,
		AdminService:        adminService,
		echo:                e,
	}
}

// GetAdminService returns the admin service
func (h *Handler) GetAdminService() *services.AdminService {
	return h.AdminService
}

// SessionUser represents user data from session (no DB call needed)
type SessionUser struct {
	ID          string
	Username    string
	Email       string
	IsAnonymous bool
	IsAdmin     bool
}

// TemplateData is an alias for viewmodels.TemplateData to maintain backward compatibility
type TemplateData = viewmodels.TemplateData

// GetSessionUser extracts user data from session without DB call
// Updated to work with echo.Context
func GetSessionUser(c echo.Context) *SessionUser {
	session, err := middleware.GetSession(c)
	if err != nil {
		return nil
	}

	userID, _ := session.Values["user_id"].(string)
	if userID == "" {
		return nil
	}

	username, _ := session.Values["username"].(string)
	email, _ := session.Values["email"].(string)
	isAnonymous, _ := session.Values["is_anonymous"].(bool)
	isAdmin, _ := session.Values["is_admin"].(bool)

	return &SessionUser{
		ID:          userID,
		Username:    username,
		Email:       email,
		IsAnonymous: isAnonymous,
		IsAdmin:     isAdmin,
	}
}

// GetTemplateUser returns the session user as interface{} for templates.
// This handles Go's nil interface gotcha: a nil *SessionUser assigned to
// interface{} is NOT nil. This function returns a true nil when there's no user.
func GetTemplateUser(c echo.Context) interface{} {
	user := GetSessionUser(c)
	if user == nil {
		return nil
	}
	return user
}

// GetCSRFToken extracts the CSRF token from Echo context
func GetCSRFToken(c echo.Context) string {
	// Echo CSRF middleware uses "csrf" as the default context key
	token, ok := c.Get("csrf").(string)
	if !ok {
		return ""
	}
	return token
}

// NewTemplateData creates a base TemplateData with common fields populated.
// This ensures User, IsAdmin, CSRFToken, and Env are always set correctly.
// Handlers can then add their specific fields (Title, Data, Error, Success, etc.)
func NewTemplateData(c echo.Context) *TemplateData {
	sessionUser := GetSessionUser(c)

	// Handle Go's nil interface gotcha
	var user interface{}
	if sessionUser != nil {
		user = sessionUser
	}

	return &TemplateData{
		User:      user,
		IsAdmin:   sessionUser != nil && sessionUser.IsAdmin,
		CSRFToken: GetCSRFToken(c),
		Env:       os.Getenv("ENV"),
	}
}

// RenderTemplComponent renders a templ component to the response
// This is the new method for templ-based templates
func (h *Handler) RenderTemplComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=UTF-8")
	c.Response().WriteHeader(http.StatusOK)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// RenderTemplFragment renders a templ component and returns HTML string (for SSE)
// This is used for SSE HTML fragment broadcasting
func (h *Handler) RenderTemplFragment(c echo.Context, component templ.Component) (string, error) {
	var buf bytes.Buffer
	err := component.Render(c.Request().Context(), &buf)
	if err != nil {
		return "", fmt.Errorf("failed to render templ component: %w", err)
	}
	return buf.String(), nil
}
