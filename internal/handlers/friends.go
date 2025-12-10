package handlers

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// FriendsHandler displays the friends list
func (h *Handler) FriendsHandler(c echo.Context) error {
	return h.FriendListHandler(c)
}

// FriendListHandler shows the user's friends list
func (h *Handler) FriendListHandler(c echo.Context) error {
	// Get user from session
	session, _ := middleware.GetSession(c)
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(c.Request().Context(), parsedUserID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	// Get friends
	friends, err := h.FriendService.GetFriends(c.Request().Context(), parsedUserID)
	if err != nil {
		log.Printf("Error getting friends: %v", err)
		friends = []models.FriendWithUserInfo{} // Empty list
	}

	// Get pending friend requests
	pendingRequests, err := h.FriendService.GetPendingRequests(c.Request().Context(), parsedUserID)
	if err != nil {
		log.Printf("Error getting pending requests: %v", err)
		pendingRequests = []models.FriendWithUserInfo{} // Empty list
	}

	data := &TemplateData{
		Title: "Friends",
		User:  currentUser,
		Data: map[string]interface{}{
			"Friends":            friends,
			"PendingInvitations": pendingRequests,
			"CurrentUserID":      userID,
		},
	}

	return h.RenderTemplate(c, "friends/list.html", data)
}

// AddFriendHandler handles sending friend invitations
func (h *Handler) AddFriendHandler(c echo.Context) error {
	// Get user from session
	session, _ := middleware.GetSession(c)
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		return c.Redirect(http.StatusSeeOther, "/auth/login")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(c.Request().Context(), parsedUserID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	if c.Request().Method == "GET" {
		data := &TemplateData{
			Title: "Add Friend",
			User:  currentUser,
			Data: map[string]interface{}{
				"CurrentUserID": userID,
			},
		}
		return h.RenderTemplate(c, "friends/add.html", data)
	}

	// Handle POST request
	friendIdentifier := c.FormValue("friend_identifier")
	if friendIdentifier == "" {
		return h.RenderTemplate(c, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			User:  currentUser,
			Error: "Friend identifier is required",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
	}

	// Parse user IDs
	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid current user ID")
	}

	friendID, err := uuid.Parse(friendIdentifier)
	if err != nil {
		return h.RenderTemplate(c, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			User:  currentUser,
			Error: "Invalid friend ID format. Please use UUID format.",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
	}

	// Create friend request
	err = h.FriendService.CreateFriendRequest(c.Request().Context(), currentUserID, friendID)
	if err != nil {
		return h.RenderTemplate(c, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			User:  currentUser,
			Error: "Failed to send friend request",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
	}

	return h.RenderTemplate(c, "friends/add.html", &TemplateData{
		Title:   "Add Friend",
		User:    currentUser,
		Success: "Friend invitation sent!",
		Data:    map[string]interface{}{"CurrentUserID": userID},
	})
}

// AcceptFriendHandler accepts a friend invitation
func (h *Handler) AcceptFriendHandler(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request ID")
	}

	err = h.FriendService.AcceptFriendRequest(c.Request().Context(), requestID)
	if err != nil {
		log.Printf("Error accepting friend request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to accept friend request")
	}

	return c.Redirect(http.StatusSeeOther, "/friends")
}

// DeclineFriendHandler declines a friend invitation
func (h *Handler) DeclineFriendHandler(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request ID")
	}

	err = h.FriendService.DeclineFriendRequest(c.Request().Context(), requestID)
	if err != nil {
		log.Printf("Error declining friend request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to decline friend request")
	}

	log.Printf("Declined friend request: %s", requestID)
	return c.Redirect(http.StatusSeeOther, "/friends")
}

// SendFriendRequestHandler sends a friend request
func (h *Handler) SendFriendRequestHandler(c echo.Context) error {
	return h.AddFriendHandler(c)
}

// AcceptFriendRequestHandler accepts a friend request
func (h *Handler) AcceptFriendRequestHandler(c echo.Context) error {
	return h.AcceptFriendHandler(c)
}

// RejectFriendRequestHandler rejects a friend request
func (h *Handler) RejectFriendRequestHandler(c echo.Context) error {
	return h.DeclineFriendHandler(c)
}

// RemoveFriendHandler removes a friendship
func (h *Handler) RemoveFriendHandler(c echo.Context) error {
	friendshipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid friendship ID")
	}

	err = h.FriendService.RemoveFriend(c.Request().Context(), friendshipID)
	if err != nil {
		log.Printf("Error removing friend: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to remove friend")
	}

	log.Printf("Removed friendship: %s", friendshipID)

	// Return success JSON for AJAX requests
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Friend removed",
	})
}

// GetFriendsAPIHandler returns friends list as JSON (for room invitations)
func (h *Handler) GetFriendsAPIHandler(c echo.Context) error {
	// Get user from context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get friends
	friends, err := h.FriendService.GetFriends(c.Request().Context(), userID)
	if err != nil {
		log.Printf("Error getting friends: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get friends")
	}

	// Build JSON response
	type FriendJSON struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}

	friendList := make([]FriendJSON, 0, len(friends))
	for _, friend := range friends {
		// Determine which ID is the actual friend (not the current user)
		friendIDStr := friend.FriendID.String()
		if friend.FriendID == userID {
			friendIDStr = friend.UserID.String()
		}
		friendList = append(friendList, FriendJSON{
			ID:       friendIDStr,
			Username: friend.Username,
		})
	}

	return c.JSON(http.StatusOK, friendList)
}

// GetFriendsHTMLHandler returns friends list as HTML fragment (for HTMX)
func (h *Handler) GetFriendsHTMLHandler(c echo.Context) error {
	// Get user from context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get room ID from query parameter
	roomID := c.QueryParam("room_id")

	// Get friends
	friendsList, err := h.FriendService.GetFriends(c.Request().Context(), userID)
	if err != nil {
		log.Printf("Error getting friends: %v", err)
		return c.HTML(http.StatusOK, `<p style="color: #6b7280;">Failed to load friends</p>`)
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

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("friends_list.html", services.FriendsListData{
		Friends: friends,
		RoomID:  roomID,
	})
	if err != nil {
		log.Printf("Error rendering friends list template: %v", err)
		return c.HTML(http.StatusOK, `<p style="color: #6b7280;">Failed to load friends</p>`)
	}

	return c.HTML(http.StatusOK, html)
}
