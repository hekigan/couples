package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/yourusername/couple-card-game/internal/middleware"
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

// Template cache for development mode (lazy loading)
var (
	templateCache      = make(map[string]*template.Template)
	templateCacheMutex sync.RWMutex
	isProduction       = os.Getenv("ENV") == "production"
)

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
	var templates *template.Template

	// Production mode: Parse all templates once at startup
	if isProduction {
		log.Println("📦 Production mode: Caching all templates at startup...")
		tmpl := template.New("").Funcs(TemplateFuncMap)

		// Parse layout
		tmpl = template.Must(tmpl.ParseFiles("templates/layout.html"))

		// Parse all page templates
		tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))
		tmpl = template.Must(tmpl.ParseGlob("templates/*/*.html"))

		// Parse all components
		tmpl = template.Must(tmpl.ParseGlob("templates/components/*.html"))

		// Parse all partials
		tmpl = template.Must(tmpl.ParseGlob("templates/partials/*/*.html"))
		tmpl = template.Must(tmpl.ParseGlob("templates/partials/*/*/*.html"))

		templates = tmpl
		log.Println("✅ Templates cached successfully")
	} else {
		log.Println("🔧 Development mode: Using lazy template loading")
		templates = nil // Lazy loading in dev mode
	}

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
		Templates:           templates,
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

// TemplateData represents common data passed to templates
type TemplateData struct {
	Title             string
	User              interface{} // Can be *SessionUser or *models.User
	Error             string
	Success           string
	Data              interface{}
	OwnerUsername     string
	GuestUsername     string
	IsOwner           bool
	IsAdmin           bool // Whether current user is admin
	JoinRequestsCount int  // Number of pending join requests (for badge)
}

// GetSessionUser extracts user data from session without DB call
func GetSessionUser(r *http.Request) *SessionUser {
	session, _ := middleware.Store.Get(r, "couple-card-game-session")

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

// RenderTemplate renders a template with the given data
func (h *Handler) RenderTemplate(w http.ResponseWriter, tmpl string, data *TemplateData) {
	if data.Title == "" {
		data.Title = "Couple Card Game"
	}

	var t *template.Template
	var err error

	// Production mode: Use pre-cached templates
	if h.Templates != nil {
		t = h.Templates
	} else {
		// Development mode: Lazy caching with hot-reload capability
		cacheKey := tmpl

		// Check if template is cached
		templateCacheMutex.RLock()
		cachedTemplate, exists := templateCache[cacheKey]
		templateCacheMutex.RUnlock()

		if exists {
			// Use cached template
			t = cachedTemplate
		} else {
			// Parse and cache on first use
			t, err = template.New("").Funcs(TemplateFuncMap).ParseFiles(
				"templates/layout.html",
				"templates/"+tmpl,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Parse component templates
			t.ParseGlob("templates/components/*.html")

			// Parse partial templates
			t.ParseGlob("templates/partials/*/*.html")
			t.ParseGlob("templates/partials/*/*/*.html")

			// Cache the parsed template
			templateCacheMutex.Lock()
			templateCache[cacheKey] = t
			templateCacheMutex.Unlock()

			log.Printf("🔄 Cached template: %s", tmpl)
		}
	}

	// Execute layout.html which will include the content template
	err = t.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

