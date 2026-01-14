package main

import (
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// registerAPIv1Routes registers all versioned API routes under /api/v1
// This provides a stable API contract for future compatibility
func registerAPIv1Routes(e *echo.Echo, h *handlers.Handler, realtimeService *services.RealtimeService) {
	// API v1 routes (auth + rate limiting + CSRF)
	apiv1 := e.Group("/api/v1")
	apiv1.Use(middleware.EchoRequireAuth())
	apiv1.Use(middleware.EchoRateLimit())
	apiv1.Use(middleware.EchoCSRF())

	// ============================================================================
	// Room Management API
	// ============================================================================
	rooms := apiv1.Group("/rooms")

	// Room CRUD operations
	rooms.DELETE("/:id", h.DeleteRoomAPIHandler)                    // Delete a room
	rooms.POST("/:id/leave", h.LeaveRoomHandler)                    // Leave a room

	// Game state management
	rooms.POST("/:id/start", h.StartGameAPIHandler)                 // Start game
	rooms.POST("/:id/guest-ready", h.SetGuestReadyAPIHandler)       // Mark guest as ready
	rooms.POST("/:id/typing", h.PlayerTypingAPIHandler)             // Typing indicator
	rooms.POST("/:id/draw", h.DrawQuestionAPIHandler)               // Draw next question
	rooms.POST("/:id/answer", h.SubmitAnswerAPIHandler)             // Submit answer
	rooms.POST("/:id/finish", h.FinishGameAPIHandler)               // Finish game
	rooms.POST("/:id/next-question", h.NextQuestionHTMLHandler)     // Get next question (HTMX)

	// Room UI fragments (HTMX)
	rooms.GET("/:id/start-button", h.GetStartGameButtonHTMLHandler)    // Start button state
	rooms.GET("/:id/ready-button", h.GetGuestReadyButtonHTMLHandler)   // Ready button state
	rooms.GET("/:id/status-badge", h.RoomStatusBadgeAPIHandler)        // Status badge
	rooms.GET("/:id/turn-indicator", h.GetTurnIndicatorHandler)        // Turn indicator
	rooms.GET("/:id/question-card", h.GetQuestionCardHandler)          // Question card
	rooms.GET("/:id/game-forms", h.GetGameFormsHandler)                // Game forms
	rooms.GET("/:id/game-content", h.GetGameContentHandler)            // Game content
	rooms.GET("/:id/progress-counter", h.GetProgressCounterHandler)    // Progress counter

	// Room categories
	rooms.GET("/:id/categories", h.GetRoomCategoriesHTMLHandler)       // Get room categories
	rooms.POST("/:id/categories", h.UpdateCategoriesAPIHandler)        // Update categories
	rooms.POST("/:id/categories/toggle", h.ToggleCategoryAPIHandler)   // Toggle category

	// Join requests management
	rooms.GET("/:id/join-requests", h.ListJoinRequestsHandler)               // List join requests
	rooms.GET("/:id/join-requests-json", h.GetJoinRequestsJSONHandler)       // Get as JSON
	rooms.GET("/:id/join-requests-count", h.GetJoinRequestsCountHandler)     // Get count
	rooms.GET("/:id/my-join-request", h.CheckMyJoinRequestHandler)           // Check my request
	rooms.POST("/:id/cancel-my-request", h.CancelMyJoinRequestHTMLHandler)   // Cancel my request

	// ============================================================================
	// Categories API
	// ============================================================================
	categories := apiv1.Group("/categories")
	categories.GET("", h.GetCategoriesAPIHandler)                    // List all categories

	// ============================================================================
	// Friends API
	// ============================================================================
	friends := apiv1.Group("/friends")
	friends.GET("/list", h.GetFriendsAPIHandler)                     // Get friends (JSON)
	friends.GET("/list-html", h.GetFriendsHTMLHandler)               // Get friends (HTML)

	// ============================================================================
	// Join Requests API
	// ============================================================================
	joinRequests := apiv1.Group("/join-requests")
	joinRequests.POST("", h.CreateJoinRequestHandler)                // Create join request
	joinRequests.GET("/my-requests", h.GetMyJoinRequestsHTMLHandler) // My join requests
	joinRequests.GET("/my-accepted", h.GetMyAcceptedRequestsHandler) // My accepted requests
	joinRequests.POST("/:request_id/accept", h.AcceptJoinRequestHandler) // Accept request
	joinRequests.POST("/:request_id/reject", h.RejectJoinRequestHandler) // Reject request

	// ============================================================================
	// Invitations API
	// ============================================================================
	invitations := apiv1.Group("/invitations")
	invitations.POST("", h.SendRoomInvitationHandler)                        // Send invitation
	invitations.DELETE("/:room_id/:invitee_id", h.CancelRoomInvitationHandler) // Cancel invitation

	// ============================================================================
	// Notifications API
	// ============================================================================
	notifications := apiv1.Group("/notifications")
	notifications.GET("", h.GetNotificationsHandler)                 // List notifications
	notifications.GET("/unread-count", h.GetUnreadCountHandler)      // Unread count
	notifications.POST("/:id/read", h.MarkNotificationReadHandler)   // Mark as read
	notifications.POST("/read-all", h.MarkAllNotificationsReadHandler) // Mark all read

	// ============================================================================
	// Server-Sent Events (SSE) - Real-time updates (NO rate limiting, NO CSRF)
	// ============================================================================
	apiSSE := e.Group("/api/v1/stream")
	apiSSE.Use(middleware.EchoRequireAuth())
	realtimeHandler := handlers.NewRealtimeHandler(h, realtimeService)

	// Room event streams
	apiSSE.GET("/rooms/:id/events", realtimeHandler.StreamRoomEvents)           // Room events
	apiSSE.GET("/rooms/:id/players", realtimeHandler.GetRoomPlayers)            // Player updates
	apiSSE.GET("/rooms/:id/state", realtimeHandler.GetRoomState)                // Room state

	// User event streams
	apiSSE.GET("/user/events", realtimeHandler.StreamUserNotifications)         // User notifications
	apiSSE.GET("/notifications", h.NotificationStreamHandler)                   // Notification stream
}
