package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// RealtimeHandler handles real-time connections
type RealtimeHandler struct {
	handler         *Handler
	realtimeService *services.RealtimeService
}

// NewRealtimeHandler creates a new realtime handler
func NewRealtimeHandler(h *Handler, realtimeService *services.RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{
		handler:         h,
		realtimeService: realtimeService,
	}
}

// formatSSEData formats data for Server-Sent Events protocol
// SSE requires multi-line data to have each line prefixed with "data: "
// Per SSE spec, ALL lines (including empty ones) must be prefixed with "data: "
func formatSSEData(eventType, data string) string {
	var b strings.Builder
	b.WriteString("event: ")
	b.WriteString(eventType)
	b.WriteString("\n")

	// Split data on newlines and prefix EVERY line with "data: "
	// This includes empty lines - they become "data: \n" which is valid SSE
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		b.WriteString("data: ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n") // Final newline to terminate the event

	return b.String()
}

// StreamRoomEvents streams room events via SSE
func (h *RealtimeHandler) StreamRoomEvents(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	// Access underlying writer for SSE
	w := c.Response().Writer

	// Subscribe to room events
	client := h.realtimeService.Subscribe(roomID, userID)
	defer h.realtimeService.Unsubscribe(client.ID)

	// Stream events
	flusher, ok := w.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Streaming not supported")
	}

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\",\"room_id\":\"%s\"}\n\n", roomID)
	flusher.Flush()

	// NOTE: Initial join requests are now rendered server-side in the template
	// We only broadcast NEW join requests via SSE to avoid duplicates on reconnect

	// Stream events until client disconnects
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-client.Channel:
			if !ok {
				// Channel closed
				return nil
			}
			// Check if this is an HTML fragment (string data) or JSON data
			if htmlStr, isString := event.Data.(string); isString {
				// HTML fragment - use proper SSE format for multi-line data
				// HTMX expects raw HTML in the data field
				fmt.Fprint(w, formatSSEData(event.Type, htmlStr))
			} else {
				// JSON data - use the normal EventToSSE function
				sseData := services.EventToSSE(event)
				if _, err := fmt.Fprint(w, sseData); err != nil {
					return nil
				}
			}
			flusher.Flush()
		case <-c.Request().Context().Done():
			return nil
		case <-ticker.C:
			// Send keepalive ping every 15 seconds
			if _, err := fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339)); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

// StreamUserNotifications streams user-specific notifications via SSE
func (h *RealtimeHandler) StreamUserNotifications(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	// Access underlying writer for SSE
	w := c.Response().Writer

	// Subscribe to user events (use a dummy room ID since we're not in a room)
	dummyRoomID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	client := h.realtimeService.Subscribe(dummyRoomID, userID)
	defer h.realtimeService.Unsubscribe(client.ID)

	// Stream events
	flusher, ok := w.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Streaming not supported")
	}

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\",\"user_id\":\"%s\"}\n\n", userID)
	flusher.Flush()

	// Stream events until client disconnects
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-client.Channel:
			if !ok {
				// Channel closed
				return nil
			}
			// Check if this is an HTML fragment (string data) or JSON data
			if htmlStr, isString := event.Data.(string); isString {
				// HTML fragment - use proper SSE format for multi-line data
				// HTMX expects raw HTML in the data field
				fmt.Fprint(w, formatSSEData(event.Type, htmlStr))
			} else {
				// JSON data - use the normal EventToSSE function
				sseData := services.EventToSSE(event)
				if _, err := fmt.Fprint(w, sseData); err != nil {
					return nil
				}
			}
			flusher.Flush()
		case <-c.Request().Context().Done():
			return nil
		case <-ticker.C:
			// Send keepalive ping every 15 seconds
			if _, err := fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339)); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

// GetRoomPlayers gets current room players
func (h *RealtimeHandler) GetRoomPlayers(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "Not yet implemented")
}

// RoomStateResponse represents the room state for the frontend
type RoomStateResponse struct {
	ID                  uuid.UUID        `json:"id"`
	Status              string           `json:"status"`
	CurrentQuestion     int              `json:"current_question"`
	CurrentTurn         *uuid.UUID       `json:"current_turn"` // Note: different field name for frontend compatibility
	MaxQuestions        int              `json:"max_questions"`
	OwnerID             uuid.UUID        `json:"owner_id"`
	GuestID             *uuid.UUID       `json:"guest_id"`
	Language            string           `json:"language"`
	CurrentQuestionData *models.Question `json:"current_question_data,omitempty"` // Include full question data if exists
}

// GetRoomState gets current room state
func (h *RealtimeHandler) GetRoomState(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := c.Request().Context()

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get the room
	room, err := h.handler.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
	}

	// Check if user is part of this room
	isOwner := room.OwnerID == userID
	isGuest := room.GuestID != nil && *room.GuestID == userID
	if !isOwner && !isGuest {
		return echo.NewHTTPError(http.StatusForbidden, "You are not a member of this room")
	}

	// Fetch the current question if it exists
	var currentQuestionData *models.Question
	if room.CurrentQuestionID != nil {
		question, err := h.handler.QuestionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err != nil {
			// Log error but don't fail the request - question might have been deleted
			fmt.Printf("Warning: Failed to fetch current question %s: %v\n", room.CurrentQuestionID, err)
		} else {
			currentQuestionData = question
		}
	}

	// Create response with frontend-compatible field names
	response := RoomStateResponse{
		ID:                  room.ID,
		Status:              room.Status,
		CurrentQuestion:     room.CurrentQuestion,
		CurrentTurn:         room.CurrentTurn, // Maps CurrentTurn to current_turn in JSON
		MaxQuestions:        room.MaxQuestions,
		OwnerID:             room.OwnerID,
		GuestID:             room.GuestID,
		Language:            room.Language,
		CurrentQuestionData: currentQuestionData, // Include question data if exists
	}

	return c.JSON(http.StatusOK, response)
}
