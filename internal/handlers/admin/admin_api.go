package admin

import (
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/services"
)

// AdminAPIHandler handles admin API requests
// Handler methods are organized across multiple files:
// - admin_users.go: User management (8 handlers)
// - admin_questions.go: Question management (6 handlers)
// - admin_categories.go: Category management (5 handlers)
// - admin_rooms.go: Room management (3 handlers)
// - admin_stats.go: Dashboard statistics (1 handler)
// - admin_bulk.go: Bulk operations (4 handlers)
type AdminAPIHandler struct {
	handler         *handlers.Handler
	adminService    *services.AdminService
	questionService *services.QuestionService
}

// NewAdminAPIHandler creates a new admin API handler
func NewAdminAPIHandler(h *handlers.Handler, adminService *services.AdminService, questionService *services.QuestionService) *AdminAPIHandler {
	return &AdminAPIHandler{
		handler:         h,
		adminService:    adminService,
		questionService: questionService,
	}
}
