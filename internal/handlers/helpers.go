package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hekigan/couples/internal/models"
)

// GetRoomFromRequest parses room ID from URL and fetches the room
// Eliminates 18 duplications in game.go
func (h *Handler) GetRoomFromRequest(r *http.Request) (*models.Room, uuid.UUID, error) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
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
// Eliminates 5 duplications in game.go
func (h *Handler) VerifyRoomParticipant(room *models.Room, userID uuid.UUID) error {
	isOwner := room.OwnerID == userID
	isGuest := room.GuestID != nil && *room.GuestID == userID

	if !isOwner && !isGuest {
		return fmt.Errorf("user is not a participant in this room")
	}

	return nil
}

// RenderHTMLFragment renders an HTML fragment using TemplateService and writes it to the response
// Eliminates 13+ duplications in admin_api.go and game.go
func (h *Handler) RenderHTMLFragment(w http.ResponseWriter, templateName string, data interface{}) error {
	html, err := h.TemplateService.RenderFragment(templateName, data)
	if err != nil {
		log.Printf("Error rendering fragment %s: %v", templateName, err)
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
	return nil
}

// RenderHTMLFragmentOrFallback renders an HTML fragment with a fallback on error
// Enhanced version of RenderHTMLFragment with graceful degradation
func (h *Handler) RenderHTMLFragmentOrFallback(w http.ResponseWriter, templateName string, data interface{}, fallback string) error {
	html, err := h.TemplateService.RenderFragment(templateName, data)
	if err != nil {
		log.Printf("Error rendering fragment %s: %v", templateName, err)
		// Write fallback HTML
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fallback))
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
	return nil
}

// FetchCurrentUser gets the current user from context with proper error handling
// Eliminates 6 duplications in game.go
func (h *Handler) FetchCurrentUser(r *http.Request) (*models.User, error) {
	ctx := context.Background()

	// Get user ID from session
	sessionUser := GetSessionUser(r)
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

// ExtractIDFromRoute extracts and validates a UUID from route variables
// Eliminates 12 duplications in admin_api.go
func ExtractIDFromRoute(vars map[string]string, key string) (uuid.UUID, error) {
	idStr, ok := vars[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("missing %s in route", key)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s: %w", key, err)
	}

	return id, nil
}

// ExtractIDFromRouteVars is a convenience wrapper that extracts mux vars from request
// and parses the UUID in one call
func ExtractIDFromRouteVars(r *http.Request, key string) (uuid.UUID, error) {
	vars := mux.Vars(r)
	return ExtractIDFromRoute(vars, key)
}

// ParsePaginationParams extracts pagination parameters from the request
// Eliminates 4 duplications in admin_api.go
// Returns: page (1-indexed), perPage (validated), offset (0-indexed)
func ParsePaginationParams(r *http.Request) (page, perPage, offset int) {
	// Parse page parameter
	page = 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	// Parse per_page parameter with validation
	perPage = 25 // Default
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			// Validate against allowed values
			if parsed == 25 || parsed == 50 || parsed == 100 {
				perPage = parsed
			}
		}
	}

	// Calculate offset
	offset = (page - 1) * perPage

	return page, perPage, offset
}

// WriteJSONResponse writes a JSON response with proper headers
// Standardizes JSON response writing across handlers
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteJSONError writes a JSON error response
// Convenience wrapper for error responses
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) error {
	return WriteJSONResponse(w, statusCode, map[string]string{
		"error": message,
	})
}

// WriteJSONSuccess writes a JSON success response
// Convenience wrapper for success responses
func WriteJSONSuccess(w http.ResponseWriter, message string) error {
	return WriteJSONResponse(w, http.StatusOK, map[string]string{
		"success": message,
	})
}
