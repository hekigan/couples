package main

import (
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/handlers/admin"
	csvapi "github.com/hekigan/couples/internal/handlers/admin/api"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// registerAdminAPIv1Routes registers versioned admin API routes under /admin/api/v1
// This provides a stable admin API contract for future compatibility
func registerAdminAPIv1Routes(e *echo.Echo, h *handlers.Handler, adminService *services.AdminService, questionService *services.QuestionService, categoryService *services.CategoryService) {
	// Admin routes (admin auth required + CSRF)
	adminRouter := e.Group("/admin")
	adminRouter.Use(middleware.EchoRequireAdmin())
	adminRouter.Use(middleware.EchoCSRF())

	// Admin page routes (unversioned - these are UI pages)
	adminRouter.GET("", h.AdminDashboardHandler)
	adminRouter.GET("/users", h.AdminUsersHandler)
	adminRouter.GET("/questions", h.AdminQuestionsHandler)
	adminRouter.GET("/categories", h.AdminCategoriesHandler)
	adminRouter.GET("/rooms", h.AdminRoomsHandler)
	adminRouter.GET("/translations", h.AdminTranslationsHandler)
	adminRouter.GET("/routes", h.AdminRoutesHandler)

	// Admin API v1 routes
	adminAPIv1 := adminRouter.Group("/api/v1")
	adminAPIHandler := admin.NewAdminAPIHandler(h, adminService, questionService, categoryService)

	// ============================================================================
	// Users Management API
	// ============================================================================
	users := adminAPIv1.Group("/users")
	users.GET("/list", adminAPIHandler.ListUsersHandler)                  // List users (paginated)
	users.GET("/new", adminAPIHandler.GetUserCreateFormHandler)           // Create form
	users.POST("", adminAPIHandler.CreateUserHandler)                     // Create user
	users.GET("/:id/edit-form", adminAPIHandler.GetUserEditFormHandler)   // Edit form
	users.PUT("/:id", adminAPIHandler.UpdateUserHandler)                  // Update user
	users.POST("/:id/toggle-admin", adminAPIHandler.ToggleUserAdminHandler) // Toggle admin
	users.DELETE("/:id", adminAPIHandler.DeleteUserHandler)               // Delete user
	users.POST("/bulk-delete", adminAPIHandler.BulkDeleteUsersHandler)    // Bulk delete

	// ============================================================================
	// Questions Management API
	// ============================================================================
	questions := adminAPIv1.Group("/questions")
	questions.GET("/list", adminAPIHandler.ListQuestionsHandler)                // List questions
	questions.GET("/new", adminAPIHandler.GetQuestionCreateFormHandler)         // Create form
	questions.POST("", adminAPIHandler.CreateQuestionHandler)                   // Create question
	questions.GET("/:id/edit-form", adminAPIHandler.GetQuestionEditFormHandler) // Edit form
	questions.PUT("/:id", adminAPIHandler.UpdateQuestionHandler)                // Update question
	questions.DELETE("/:id", adminAPIHandler.DeleteQuestionHandler)             // Delete question
	questions.POST("/bulk-delete", adminAPIHandler.BulkDeleteQuestionsHandler)  // Bulk delete

	// ============================================================================
	// Categories Management API
	// ============================================================================
	categories := adminAPIv1.Group("/categories")
	categories.GET("/list", adminAPIHandler.ListCategoriesHandler)                // List categories
	categories.GET("/new", adminAPIHandler.GetCategoryCreateFormHandler)          // Create form
	categories.POST("", adminAPIHandler.CreateCategoryHandler)                    // Create category
	categories.GET("/:id/edit-form", adminAPIHandler.GetCategoryEditFormHandler)  // Edit form
	categories.PUT("/:id", adminAPIHandler.UpdateCategoryHandler)                 // Update category
	categories.DELETE("/:id", adminAPIHandler.DeleteCategoryHandler)              // Delete category
	categories.POST("/bulk-delete", adminAPIHandler.BulkDeleteCategoriesHandler)  // Bulk delete

	// ============================================================================
	// Rooms Management API
	// ============================================================================
	rooms := adminAPIv1.Group("/rooms")
	rooms.GET("/list", adminAPIHandler.ListRoomsHandler)              // List rooms
	rooms.GET("/:id/details", adminAPIHandler.GetRoomDetailsHandler)  // Room details
	rooms.POST("/:id/close", adminAPIHandler.CloseRoomHandler)        // Close room
	rooms.DELETE("/:id", adminAPIHandler.DeleteRoomHandler)           // Delete room
	rooms.POST("/bulk-close", adminAPIHandler.BulkCloseRoomsHandler)  // Bulk close

	// ============================================================================
	// Dashboard & Statistics API
	// ============================================================================
	dashboard := adminAPIv1.Group("/dashboard")
	dashboard.GET("/stats", adminAPIHandler.GetDashboardStatsHandler) // Dashboard stats

	// ============================================================================
	// CSV Import/Export API
	// ============================================================================
	csv := adminAPIv1.Group("/csv")
	csvHandler := csvapi.NewCSVHandler(questionService, categoryService)
	csv.GET("/questions/export", csvHandler.ExportQuestionsCSV)       // Export questions CSV
	csv.POST("/questions/import", csvHandler.ImportQuestionsCSV)      // Import questions CSV
	csv.GET("/questions/template", csvHandler.GetImportTemplate)      // CSV template
	csv.GET("/categories/export", csvHandler.ExportCategoriesCSV)     // Export categories CSV

	// ============================================================================
	// Translations Management API
	// ============================================================================
	translations := adminAPIv1.Group("/translations")
	translationHandler := admin.NewTranslationHandler(h)
	translations.GET("/languages", translationHandler.ListLanguagesHandler)              // List languages
	translations.GET("/:lang_code", translationHandler.GetTranslationsHandler)           // Get translations
	translations.PUT("/:lang_code", translationHandler.UpdateTranslationHandler)         // Update translation
	translations.POST("", translationHandler.CreateTranslationHandler)                   // Create translation
	translations.DELETE("/:lang_code", translationHandler.DeleteTranslationHandler)      // Delete translation
	translations.GET("/:lang_code/export", translationHandler.ExportTranslationsHandler) // Export translations
	translations.POST("/:lang_code/import", translationHandler.ImportTranslationsHandler) // Import translations
	translations.GET("/validate", translationHandler.ValidateMissingKeysHandler)         // Validate keys
	translations.POST("/language/add", translationHandler.AddLanguageHandler)            // Add language
}
