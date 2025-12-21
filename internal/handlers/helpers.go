package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	friendsFragments "github.com/hekigan/couples/internal/views/fragments/friends"
	roomFragments "github.com/hekigan/couples/internal/views/fragments/room"
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

// RenderRoomFragments fetches data and renders all fragments for the room page
// This eliminates the need for hx-trigger="load" by rendering fragments server-side
func (h *Handler) RenderRoomFragments(c echo.Context, room *models.Room, roomID uuid.UUID, userID uuid.UUID, isOwner bool) (categoriesHTML, friendsHTML, actionButtonHTML string, err error) {
	ctx := context.Background()

	// 1. Render categories grid (always shown)
	categoriesHTML, err = h.renderCategoriesGrid(c, ctx, room, roomID, isOwner)
	if err != nil {
		log.Printf("⚠️ Failed to render categories grid: %v", err)
		categoriesHTML = `<p style="color: #6b7280;">Failed to load categories</p>`
	}

	// 2. Render friends list (owner only, when waiting for guest)
	if isOwner && room.GuestID == nil {
		friendsHTML, err = h.renderFriendsList(c, ctx, userID, roomID)
		if err != nil {
			log.Printf("⚠️ Failed to render friends list: %v", err)
			friendsHTML = `<p style="color: #6b7280;">Failed to load friends</p>`
		}
	}

	// 3. Render action button (based on role and room status)
	if room.Status == "ready" {
		actionButtonHTML, err = h.renderActionButton(c, ctx, room, roomID, isOwner)
		if err != nil {
			log.Printf("⚠️ Failed to render action button: %v", err)
			actionButtonHTML = `<p style="color: #6b7280;">Failed to load button</p>`
		}
	}

	return categoriesHTML, friendsHTML, actionButtonHTML, nil
}

// renderCategoriesGrid fetches and renders the categories grid fragment
func (h *Handler) renderCategoriesGrid(c echo.Context, ctx context.Context, room *models.Room, roomID uuid.UUID, isOwner bool) (string, error) {
	// Import needed at top of file
	// roomFragments "github.com/hekigan/couples/internal/views/fragments/room"
	// "github.com/hekigan/couples/internal/services"

	// Get all categories
	allCategories, err := h.CategoryService.GetCategories(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Get question counts per category for the room's language
	questionCounts, err := h.QuestionService.GetQuestionCountsByCategory(ctx, room.Language)
	if err != nil {
		log.Printf("⚠️ Failed to fetch question counts: %v", err)
		questionCounts = make(map[string]int) // Continue without counts
	}

	// Build CategoryInfo array with selection state and question counts
	categoryInfos := make([]services.CategoryInfo, 0, len(allCategories))
	for _, cat := range allCategories {
		isSelected := false
		// Check if this category is in room's selected categories
		for _, selectedID := range room.SelectedCategories {
			if selectedID == cat.ID {
				isSelected = true
				break
			}
		}

		categoryInfos = append(categoryInfos, services.CategoryInfo{
			ID:            cat.ID.String(),
			Key:           cat.Key,
			Label:         cat.Label,
			IsSelected:    isSelected,
			QuestionCount: questionCounts[cat.ID.String()],
		})
	}

	// Import needed: roomFragments "github.com/hekigan/couples/internal/views/fragments/room"
	return h.RenderTemplFragment(c, roomFragments.CategoriesGrid(&services.CategoriesGridData{
		Categories: categoryInfos,
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
		IsOwner:    isOwner,
	}))
}

// renderFriendsList fetches and renders the friends list fragment
func (h *Handler) renderFriendsList(c echo.Context, ctx context.Context, userID uuid.UUID, roomID uuid.UUID) (string, error) {
	// Import needed: friendsFragments "github.com/hekigan/couples/internal/views/fragments/friends"

	// Get friends
	friendsList, err := h.FriendService.GetFriends(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch friends: %w", err)
	}

	// Build FriendsListData
	friends := make([]services.FriendInfo, 0, len(friendsList))
	for _, friend := range friendsList {
		// Determine which ID is the actual friend (not the current user)
		friendIDStr := friend.FriendID.String()
		if friend.FriendID == userID {
			friendIDStr = friend.UserID.String()
		}
		friends = append(friends, services.FriendInfo{
			ID:       friendIDStr,
			Username: friend.Username,
		})
	}

	// Import needed: friendsFragments "github.com/hekigan/couples/internal/views/fragments/friends"
	return h.RenderTemplFragment(c, friendsFragments.FriendsList(&services.FriendsListData{
		Friends: friends,
		RoomID:  roomID.String(),
	}))
}

// renderActionButton renders the appropriate action button (start or ready)
func (h *Handler) renderActionButton(c echo.Context, ctx context.Context, room *models.Room, roomID uuid.UUID, isOwner bool) (string, error) {
	// Import needed: roomFragments "github.com/hekigan/couples/internal/views/fragments/room"

	if isOwner {
		// Get guest username if there's a guest
		guestUsername := ""
		if room.GuestID != nil {
			if guest, err := h.UserService.GetUserByID(ctx, *room.GuestID); err == nil && guest != nil {
				guestUsername = guest.Username
			}
		}
		// Render start game button
		return h.RenderTemplFragment(c, roomFragments.StartGameButton(&services.StartGameButtonData{
			RoomID:        roomID.String(),
			GuestReady:    room.GuestReady,
			GuestUsername: guestUsername,
		}))
	} else {
		// Render guest ready button
		return h.RenderTemplFragment(c, roomFragments.GuestReadyButton(&services.GuestReadyButtonData{
			RoomID:     roomID.String(),
			GuestReady: room.GuestReady,
		}))
	}
}

// renderJoinRequests fetches and renders pending join requests as HTML
// Uses database view for optimal performance (single query with usernames)
func (h *Handler) renderJoinRequests(c echo.Context, ctx context.Context, roomID uuid.UUID) (html string, count int, err error) {
	// Use database view - includes username, already filtered for pending
	joinRequests, err := h.RoomService.GetJoinRequestsWithUserInfo(ctx, roomID)
	if err != nil {
		return "", 0, err
	}

	count = len(joinRequests)
	if count == 0 {
		return "", 0, nil
	}

	var htmlParts []string
	for _, req := range joinRequests {
		fragment, err := h.RenderTemplFragment(c, roomFragments.JoinRequest(&services.JoinRequestData{
			ID:        req.ID.String(),
			RoomID:    req.RoomID.String(),
			UserID:    req.UserID.String(),
			Username:  req.Username,
			Message:   "", // Database view doesn't include message field
			Status:    req.Status,
			CreatedAt: req.CreatedAt.Format("Jan 2, 15:04"),
		}))
		if err == nil {
			htmlParts = append(htmlParts, fragment)
		} else {
			log.Printf("⚠️ Failed to render join request %s: %v", req.ID, err)
		}
	}

	return strings.Join(htmlParts, "\n"), count, nil
}
