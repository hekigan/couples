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
		Templates:           nil, // We'll parse templates on demand
	}
}

// TemplateData represents common data passed to templates
type TemplateData struct {
	Title          string
	User           interface{}
	Error          string
	Success        string
	Data           interface{}
	OwnerUsername  string
	GuestUsername  string
	IsOwner        bool
}

// RenderTemplate renders a template with the given data
func (h *Handler) RenderTemplate(w http.ResponseWriter, tmpl string, data *TemplateData) {
	if data.Title == "" {
		data.Title = "Couple Card Game"
	}

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

	// Execute layout.html which will include the content template
	err = t.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

