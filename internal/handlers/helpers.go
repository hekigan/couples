package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/labstack/echo/v4"
)

// GetRoomFromRequest parses room ID from URL and fetches the room
// Updated to work with echo.Context
func (h *Handler) GetRoomFromRequest(c echo.Context) (*models.Room, uuid.UUID, error) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("invalid room ID: %w", err)
	}

	ctx := context.Background()
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, roomID, fmt.Errorf("room not found: %w", err)
	}

	return room, roomID, nil
}

// VerifyRoomParticipant checks if the user is a participant (owner or guest) in the room
// No changes needed - works with uuid.UUID
func (h *Handler) VerifyRoomParticipant(room *models.Room, userID uuid.UUID) error {
	isOwner := room.OwnerID == userID
	isGuest := room.GuestID != nil && *room.GuestID == userID

	if !isOwner && !isGuest {
		return fmt.Errorf("user is not a participant in this room")
	}

	return nil
}

// RenderHTMLFragment is deprecated. Use h.RenderTemplFragment() with templ components instead.
// This function remains for backward compatibility but should not be used in new code.
func (h *Handler) RenderHTMLFragment(c echo.Context, templateName string, data interface{}) error {
	log.Printf("⚠️ DEPRECATED: RenderHTMLFragment called with template %s. Use RenderTemplFragment with templ components instead.", templateName)
	return echo.NewHTTPError(http.StatusInternalServerError, "RenderHTMLFragment is deprecated")
}

// RenderHTMLFragmentOrFallback is deprecated. Use h.RenderTemplFragment() with templ components instead.
// This function remains for backward compatibility but should not be used in new code.
func (h *Handler) RenderHTMLFragmentOrFallback(c echo.Context, templateName string, data interface{}, fallback string) error {
	log.Printf("⚠️ DEPRECATED: RenderHTMLFragmentOrFallback called with template %s. Use RenderTemplFragment with templ components instead.", templateName)
	return c.HTML(http.StatusOK, fallback)
}

// FetchCurrentUser gets the current user from context with proper error handling
// Updated to work with echo.Context
func (h *Handler) FetchCurrentUser(c echo.Context) (*models.User, error) {
	ctx := context.Background()

	// Get user ID from session
	sessionUser := GetSessionUser(c)
	if sessionUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	userID, err := uuid.Parse(sessionUser.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Fetch full user from database
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		return nil, fmt.Errorf("failed to load user information: %w", err)
	}

	return currentUser, nil
}

// ExtractIDFromParam extracts and validates a UUID from URL parameter
// Replaces ExtractIDFromRoute for Echo
func ExtractIDFromParam(c echo.Context, key string) (uuid.UUID, error) {
	idStr := c.Param(key)
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("missing %s in route", key)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s: %w", key, err)
	}

	return id, nil
}

// ParsePaginationParams extracts pagination parameters from the request
// Updated to work with echo.Context
// ParsePaginationParams extracts and validates pagination parameters from request
// Returns: page (1-indexed), perPage (validated)
// Callers should calculate offset as: offset = (page - 1) * perPage
func ParsePaginationParams(c echo.Context) (page, perPage int) {
	// Parse page parameter
	page = 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	// Parse per_page parameter with validation
	perPage = 25 // Default
	if pp := c.QueryParam("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			// Validate against allowed values
			if parsed == 25 || parsed == 50 || parsed == 100 {
				perPage = parsed
			}
		}
	}

	return page, perPage
}

// GetUserIDFromContext retrieves user ID from Echo context
// Helper function for cleaner code
func GetUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, fmt.Errorf("user not authenticated")
	}
	return userID, nil
}
