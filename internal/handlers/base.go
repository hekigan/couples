package handlers

import (
	"html/template"
	"net/http"

	"github.com/yourusername/couple-card-game/internal/services"
)

// Handler holds all service dependencies
type Handler struct {
	UserService         *services.UserService
	RoomService         *services.RoomService
	GameService         *services.GameService
	QuestionService     *services.QuestionService
	AnswerService       *services.AnswerService
	FriendService       *services.FriendService
	I18nService         *services.I18nService
	NotificationService *services.NotificationService
	TemplateService     *services.TemplateService // For SSE HTML fragments
	AdminService        *services.AdminService    // For admin operations
	Templates           *template.Template
}

// Template function map for custom functions
var TemplateFuncMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
}

// NewHandler creates a new handler with all dependencies
func NewHandler(
	userService *services.UserService,
	roomService *services.RoomService,
	gameService *services.GameService,
	questionService *services.QuestionService,
	answerService *services.AnswerService,
	friendService *services.FriendService,
	i18nService *services.I18nService,
	notificationService *services.NotificationService,
	templateService *services.TemplateService,
	adminService *services.AdminService,
) *Handler {
	return &Handler{
		UserService:         userService,
		RoomService:         roomService,
		GameService:         gameService,
		QuestionService:     questionService,
		AnswerService:       answerService,
		FriendService:       friendService,
		I18nService:         i18nService,
		NotificationService: notificationService,
		TemplateService:     templateService,
		AdminService:        adminService,
		Templates:           nil, // We'll parse templates on demand
	}
}

// GetAdminService returns the admin service
func (h *Handler) GetAdminService() *services.AdminService {
	return h.AdminService
}

// TemplateData represents common data passed to templates
type TemplateData struct {
	Title             string
	User              interface{}
	Error             string
	Success           string
	Data              interface{}
	OwnerUsername     string
	GuestUsername     string
	IsOwner           bool
	IsAdmin           bool // Whether current user is admin
	JoinRequestsCount int  // Number of pending join requests (for badge)
}

// RenderTemplate renders a template with the given data
func (h *Handler) RenderTemplate(w http.ResponseWriter, tmpl string, data *TemplateData) {
	if data.Title == "" {
		data.Title = "Couple Card Game"
	}

	// Note: IsAdmin flag should be set by the caller from request context
	// This is typically done in each handler using the session data

	// Parse layout.html and the specific content template together
	t, err := template.New("").Funcs(TemplateFuncMap).ParseFiles(
		"templates/layout.html",
		"templates/"+tmpl,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Also parse component templates if they exist
	t.ParseGlob("templates/components/*.html")
	
	// Also parse partial templates (for reusable fragments)
	// Note: ParseGlob doesn't support ** so we parse each subdirectory
	t.ParseGlob("templates/partials/*/*.html")
	t.ParseGlob("templates/partials/*/*/*.html")

	// Execute layout.html which will include the content template
	err = t.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

