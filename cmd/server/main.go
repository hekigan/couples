package main

import (
	"log"
	"os"

	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/rendering"
	"github.com/hekigan/couples/internal/server"
	"github.com/hekigan/couples/internal/services"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize session store
	middleware.InitSessionStore()

	// Initialize Supabase client
	supabaseClient, err := services.NewSupabaseClient()
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Initialize render service for SSE HTML fragments (templ-based)
	renderService := rendering.NewTemplService()

	// Initialize services
	realtimeService := services.NewRealtimeService()
	userService := services.NewUserService(supabaseClient)
	roomService := services.NewRoomService(supabaseClient, realtimeService)
	questionService := services.NewQuestionService(supabaseClient)
	categoryService := services.NewCategoryService(supabaseClient)
	answerService := services.NewAnswerService(supabaseClient)
	gameService := services.NewGameService(supabaseClient, roomService, questionService, categoryService, answerService, realtimeService, renderService)
	notificationService := services.NewNotificationService(supabaseClient)
	friendService := services.NewFriendService(supabaseClient, notificationService, realtimeService)
	i18nService := services.NewI18nService(supabaseClient, "./static/i18n")
	adminService := services.NewAdminService(supabaseClient)

	// Initialize email service (optional - graceful degradation if not configured)
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	emailFrom := os.Getenv("EMAIL_FROM")
	appBaseURL := os.Getenv("APP_BASE_URL")
	var emailService *services.EmailService
	if resendAPIKey != "" && emailFrom != "" && appBaseURL != "" {
		emailService = services.NewEmailService(resendAPIKey, emailFrom, appBaseURL)
		log.Println("✅ Email service initialized")
	} else {
		log.Println("⚠️  Email service not configured - email invitations will be disabled")
		emailService = nil
	}

	// Show development mode hot-reload instructions
	if os.Getenv("ENV") == "development" {
		log.Println("")
		log.Println("\033[1;33m╔═══════════════════════════════════════════════════════════════════════════╗\033[0m")
		log.Println("\033[1;33m║\033[0m \033[1;36m⚠️  IMPORTANT: For full hot-reload experience, run these commands:        \033[0m \033[1;33m║\033[0m")
		log.Println("\033[1;33m║\033[0m                                                                           \033[1;33m║\033[0m")
		log.Println("\033[1;33m║\033[0m   \033[1;32m📂 Terminal 2:\033[0m  \033[1;97mmake sass-watch\033[0m  \033[90m(auto-compile CSS)\033[0m                     \033[1;33m║\033[0m")
		log.Println("\033[1;33m║\033[0m   \033[1;32m📂 Terminal 3:\033[0m  \033[1;97mmake js-watch\033[0m    \033[90m(auto-compile JS bundles)\033[0m              \033[1;33m║\033[0m")
		log.Println("\033[1;33m╚═══════════════════════════════════════════════════════════════════════════╝\033[0m")
		log.Println("")
	}

	// Log service initialization
	log.Println("Services initialized successfully")

	// Create Echo instance (before handler to enable route introspection)
	e := echo.New()
	e.HideBanner = true

	// Set custom validator and error handler
	e.Validator = server.NewValidator()
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Create handler with all services (including Echo for route introspection)
	h := handlers.NewHandler(
		userService,
		roomService,
		gameService,
		questionService,
		categoryService,
		answerService,
		friendService,
		i18nService,
		notificationService,
		adminService,
		emailService,
		e,
	)

	// Global middleware (order matters!)
	// Error-only logger (only logs status >= 400, excludes known browser auto-requests)
	e.Use(echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogError:    true,
		LogLatency:  true,
		LogRemoteIP: true,
		HandleError: true,
		Skipper: func(c echo.Context) bool {
			// Skip logging for known browser auto-requests that commonly 404
			path := c.Request().URL.Path
			return path == "/.well-known/appspecific/com.chrome.devtools.json" ||
				path == "/favicon.ico"
		},
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
			// Only log errors (status >= 400)
			if v.Status >= 400 {
				if v.Error != nil {
					log.Printf("ERROR: %s %s | Status: %d | IP: %s | Latency: %v | Error: %v",
						v.Method, v.URI, v.Status, v.RemoteIP, v.Latency, v.Error)
				} else {
					log.Printf("ERROR: %s %s | Status: %d | IP: %s | Latency: %v",
						v.Method, v.URI, v.Status, v.RemoteIP, v.Latency)
				}
			}
			return nil
		},
	}))
	e.Use(echoMiddleware.Recover()) // Echo's built-in recovery
	e.Use(echoMiddleware.Gzip())    // Echo's built-in gzip
	e.Use(middleware.EchoSecurityHeaders())
	e.Use(middleware.EchoCORS())
	e.Use(middleware.EchoAuth())
	e.Use(middleware.EchoI18n())
	e.Use(middleware.EchoAnonymousSession())

	// ============================================================================
	// Route Registration
	// ============================================================================

	// Public routes (unversioned - UI pages)
	registerPublicRoutes(e, h)
	registerGameRoutes(e, h)

	// API v1 routes - Current version
	registerAPIv1Routes(e, h, realtimeService)

	// Admin API v1 routes - Current version
	registerAdminAPIv1Routes(e, h, adminService, questionService, categoryService)

	// ============================================================================
	// Development Mode - Route Registry
	// ============================================================================
	if os.Getenv("ENV") == "development" {
		// Print route statistics
		PrintRouteStats(e)

		// Uncomment to see full route inventory
		// PrintRouteRegistry(e)

		log.Println("\033[1;36m💡 API Versioning:\033[0m")
		log.Println("   • Public Pages:  / (unversioned)")
		log.Println("   • API:           /api/v1/*")
		log.Println("   • Admin Pages:   /admin/* (unversioned)")
		log.Println("   • Admin API:     /admin/api/v1/*")
		log.Println("")
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server with graceful shutdown
	server.StartWithGracefulShutdown(e, port)
}
