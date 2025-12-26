package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	joinroomFragments "github.com/hekigan/couples/internal/views/fragments/joinroom"
	roomFragments "github.com/hekigan/couples/internal/views/fragments/room"
	gamePages "github.com/hekigan/couples/internal/views/pages/game"
	"github.com/labstack/echo/v4"
)

// ListJoinRequestsHandler lists join requests for a room
func (h *Handler) ListJoinRequestsHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	requests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load join requests")
	}

	data := NewTemplateData(c)
	data.Title = "Join Requests"
	data.User = currentUser
	data.Data = requests
	return h.RenderTemplComponent(c, gamePages.JoinRequestsPage(data))
}

// CreateJoinRequestHandler creates a join request
func (h *Handler) CreateJoinRequestHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Parse room ID from form
	roomIDStr := c.FormValue("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	// Get the room to check if it exists
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
	}

	// Check if user is trying to request their own room
	if room.OwnerID == userID {
		return echo.NewHTTPError(http.StatusBadRequest, "You cannot request to join your own room")
	}

	// Check if room is already full
	if room.GuestID != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Room is full")
	}

	// Create the join request
	message := c.FormValue("message")
	request := &models.RoomJoinRequest{
		ID:      uuid.New(),
		RoomID:  roomID,
		UserID:  userID,
		Status:  "pending",
		Message: &message,
	}

	if err := h.RoomService.CreateJoinRequest(ctx, request); err != nil {
		fmt.Printf("ERROR creating join request: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create join request: %v", err))
	}

	// Get user information for SSE broadcast
	user, err := h.UserService.GetUserByID(ctx, userID)
	var username string
	if err == nil && user != nil {
		username = user.Username
	} else {
		username = "Unknown"
	}

	// Render HTML fragment for the join request
	log.Printf("🎨 Rendering join_request template for user %s (%s)", username, userID)
	html, err := h.RenderTemplFragment(c, roomFragments.JoinRequest(&services.JoinRequestData{
		ID:        request.ID.String(),
		RoomID:    roomID.String(),
		Username:  username,
		CreatedAt: request.CreatedAt.Format("3:04 PM"),
	}))
	if err != nil {
		log.Printf("❌ Failed to render join request template: %v (falling back to JSON SSE)", err)
		// Fall back to JSON SSE if template rendering fails
		h.RoomService.BroadcastJoinRequestWithDetails(roomID, request.ID, userID, username, request.CreatedAt)
	} else {
		log.Printf("✅ Successfully rendered join_request HTML fragment (%d bytes)", len(html))
		log.Printf("📤 Broadcasting HTML fragment via SSE to room %s", roomID)
		// Broadcast HTML fragment via SSE
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
			Type:       "join_request",
			Target:     "#join-requests",
			SwapMethod: "beforeend", // Append to the list of join requests
			HTML:       html,
		})
		log.Printf("✅ Join request HTML fragment broadcasted successfully")
	}

	// Update the badge count (count pending requests after adding this one)
	pendingCount := 0
	if allRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID); err == nil {
		for _, req := range allRequests {
			if req.Status == "pending" {
				pendingCount++
			}
		}
	}
	// Broadcast badge update
	if badgeHTML, err := h.RenderTemplFragment(c, roomFragments.BadgeUpdate(pendingCount)); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
			Type:       "badge_update",
			Target:     "#request-count",
			SwapMethod: "innerHTML",
			HTML:       badgeHTML,
		})
	}

	// Return success message for HTMX
	return c.HTML(http.StatusCreated, `<div style="color: #10b981; background: #dcfce7; padding: 0.75rem 1rem; border-radius: 6px;">✅ Join request sent! Waiting for owner approval...</div>`)
}

// AcceptJoinRequestHandler accepts a join request
func (h *Handler) AcceptJoinRequestHandler(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("request_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request ID")
	}

	ctx := context.Background()

	// Get the join request to find the user ID
	joinRequest, err := h.RoomService.GetJoinRequestByID(ctx, requestID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Join request not found")
	}

	// Accept the request
	if err := h.RoomService.AcceptJoinRequest(ctx, requestID); err != nil {
		fmt.Printf("ERROR: AcceptJoinRequest failed: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to accept join request: %v", err))
	}

	// Get the guest username
	guestUser, err := h.UserService.GetUserByID(ctx, joinRequest.UserID)
	var guestUsername string
	if err == nil && guestUser != nil {
		guestUsername = guestUser.Username
	}

	// Render HTML fragment for request accepted (shown to room owner)
	html, err := h.RenderTemplFragment(c, roomFragments.RequestAccepted(&services.RequestAcceptedData{
		GuestUsername: guestUsername,
	}))
	if err != nil {
		log.Printf("⚠️ Failed to render request_accepted template: %v (falling back to JSON SSE)", err)
		// Fall back to JSON SSE
		h.RoomService.BroadcastRequestAccepted(joinRequest.RoomID, guestUsername)
	} else {
		// Broadcast HTML fragment to room owner
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
			Type:       "request_accepted",
			Target:     "#guest-info",
			SwapMethod: "innerHTML", // Replace guest info section
			HTML:       html,
		})
	}

	// Notify the guest user that their request was accepted (use JSON for now, will enhance later)
	h.RoomService.BroadcastRequestAcceptedToGuest(joinRequest.UserID, joinRequest.RoomID)

	// Update the badge count (count pending requests after accepting)
	pendingCount := 0
	if allRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, joinRequest.RoomID); err == nil {
		for _, req := range allRequests {
			if req.Status == "pending" {
				pendingCount++
			}
		}
	}
	// Broadcast badge update
	if badgeHTML, err := h.RenderTemplFragment(c, roomFragments.BadgeUpdate(pendingCount)); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
			Type:       "badge_update",
			Target:     "#request-count",
			SwapMethod: "innerHTML",
			HTML:       badgeHTML,
		})
	}

	// Broadcast room status badge update (room is now "ready")
	// Refetch the room to get the updated status
	room, err := h.RoomService.GetRoomByID(ctx, joinRequest.RoomID)
	if err != nil {
		log.Printf("⚠️ Failed to refetch room for status badge update: %v", err)
	} else {
		log.Printf("🔍 Room status after accepting join request: %s", room.Status)
		if statusBadgeHTML, err := h.RenderTemplFragment(c, roomFragments.StatusBadge(&services.RoomStatusBadgeData{
			Status: room.Status,
		})); err == nil {
			h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
				Type:       "room_status_update",
				Target:     "#room-status-badge",
				SwapMethod: "outerHTML",
				HTML:       statusBadgeHTML,
			})
			log.Printf("📡 Broadcasted room status badge update (status: %s) after accepting join request", room.Status)
		}

		// Broadcast step transition: Re-render room container to show step 2
		// Fetch room with player info for complete template data
		roomWithPlayers, err := h.RoomService.GetRoomWithPlayers(ctx, joinRequest.RoomID)
		if err == nil {
			// Get current user for template rendering
			userID, _ := middleware.GetUserID(c)
			currentUser, err := h.FetchCurrentUser(c)
			if err == nil {
				// Determine usernames
				ownerUsername := ""
				if roomWithPlayers.OwnerUsername != nil {
					ownerUsername = *roomWithPlayers.OwnerUsername
				}
				guestUsername := ""
				if roomWithPlayers.GuestUsername != nil {
					guestUsername = *roomWithPlayers.GuestUsername
				}
				isOwner := userID == roomWithPlayers.OwnerID

				// Render fragments
				categoriesHTML, friendsHTML, actionButtonHTML, _ := h.RenderRoomFragments(c, &roomWithPlayers.Room, joinRequest.RoomID, userID, isOwner)
				joinRequestsHTML, joinRequestsCount, _ := h.renderJoinRequests(c, ctx, joinRequest.RoomID)

				// Build template data
				data := NewTemplateData(c)
				data.Title = "Room - " + roomWithPlayers.Name
				data.User = currentUser
				data.Data = map[string]interface{}{
					"room": &roomWithPlayers.Room,
				}
				data.OwnerUsername = ownerUsername
				data.GuestUsername = guestUsername
				data.IsOwner = isOwner
				data.JoinRequestsCount = joinRequestsCount
				data.CategoriesGridHTML = categoriesHTML
				data.FriendsListHTML = friendsHTML
				data.ActionButtonHTML = actionButtonHTML
				data.JoinRequestsHTML = joinRequestsHTML

				// Render just the room container (without SSE wrapper)
				roomHTML, err := h.RenderTemplFragment(c, gamePages.RoomContainer(data))
				if err == nil {
					// Broadcast the re-rendered room container
					h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
						Type:       "step_transition",
						Target:     ".room-container",
						SwapMethod: "outerHTML",
						HTML:       roomHTML,
					})
					log.Printf("📡 Broadcasted step_transition event (step 1 → 2) - swapped room container")
				}
			}
		}
	}

	// Return empty HTML - HTMX will remove the element via outerHTML swap
	return c.HTML(http.StatusOK, "")
}

// RejectJoinRequestHandler rejects a join request
func (h *Handler) RejectJoinRequestHandler(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("request_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request ID")
	}

	ctx := context.Background()

	// Get the join request to find the user ID and room ID before deleting
	joinRequest, err := h.RoomService.GetJoinRequestByID(ctx, requestID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Join request not found")
	}

	// Reject the request
	if err := h.RoomService.RejectJoinRequest(ctx, requestID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reject join request")
	}

	// Notify the guest user that their request was rejected
	h.RoomService.BroadcastRequestRejectedToGuest(joinRequest.UserID, joinRequest.RoomID)

	// Update the badge count (count pending requests after rejecting)
	pendingCount := 0
	if allRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, joinRequest.RoomID); err == nil {
		for _, req := range allRequests {
			if req.Status == "pending" {
				pendingCount++
			}
		}
	}
	// Broadcast badge update
	if badgeHTML, err := h.RenderTemplFragment(c, roomFragments.BadgeUpdate(pendingCount)); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
			Type:       "badge_update",
			Target:     "#request-count",
			SwapMethod: "innerHTML",
			HTML:       badgeHTML,
		})
	}

	// Return empty HTML - HTMX will remove the element via outerHTML swap
	return c.HTML(http.StatusOK, "")
}

// GetJoinRequestsCountHandler gets count of pending requests
func (h *Handler) GetJoinRequestsCountHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	requests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load join requests")
	}

	// Count pending requests
	count := 0
	for _, req := range requests {
		if req.Status == "pending" {
			count++
		}
	}

	return c.JSON(http.StatusOK, map[string]int{"count": count})
}

// GetJoinRequestsJSONHandler returns join requests with user information as JSON
func (h *Handler) GetJoinRequestsJSONHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	requests, err := h.RoomService.GetJoinRequestsWithUserInfo(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load join requests")
	}

	return c.JSON(http.StatusOK, requests)
}

// CheckMyJoinRequestHandler checks the status of the current user's join request for a room
func (h *Handler) CheckMyJoinRequestHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Check if user is already a guest in the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err == nil && room.GuestID != nil && *room.GuestID == userID {
		// User is now a guest - request was accepted!
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "accepted",
			"message": "Request accepted",
		})
	}

	// Check ALL join requests (not just pending) to see status
	allRequests, err := h.RoomService.GetAllJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "pending",
			"message": "Waiting for approval",
		})
	}

	// Find this user's request
	for _, req := range allRequests {
		if req.UserID == userID {
			return c.JSON(http.StatusOK, map[string]string{
				"status":  req.Status,
				"message": fmt.Sprintf("Request %s", req.Status),
			})
		}
	}

	// No request found
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "not_found",
		"message": "Request not found",
	})
}

// CancelMyJoinRequestHandler cancels the current user's join request
func (h *Handler) CancelMyJoinRequestHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Find and delete the user's request
	if err := h.RoomService.CancelJoinRequest(ctx, roomID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to cancel request")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Request cancelled",
	})
}

// GetMyAcceptedRequestsHandler returns all accepted join requests for the current user
func (h *Handler) GetMyAcceptedRequestsHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get all accepted requests for this user
	requests, err := h.RoomService.GetAcceptedRequestsByUser(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load accepted requests")
	}

	return c.JSON(http.StatusOK, requests)
}

// GetMyJoinRequestsHTMLHandler returns HTML fragment for all user's join requests
func (h *Handler) GetMyJoinRequestsHTMLHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get all join requests for this user (pending, accepted, rejected)
	allRequests, err := h.RoomService.GetJoinRequestsByUser(ctx, userID)
	if err != nil {
		log.Printf("Error fetching user's join requests: %v", err)
		return c.HTML(http.StatusOK, `<!-- No pending requests -->`)
	}

	// Convert to template data format
	var requestsData []services.JoinRequestData
	for _, req := range allRequests {
		message := ""
		if req.Message != nil {
			message = *req.Message
		}
		requestsData = append(requestsData, services.JoinRequestData{
			RoomID:  req.RoomID.String(),
			UserID:  req.UserID.String(),
			Message: message,
			Status:  req.Status,
		})
	}

	// Render HTML fragment
	html, err := h.RenderTemplFragment(c, joinroomFragments.PendingRequestsList(requestsData))
	if err != nil {
		log.Printf("Error rendering pending_requests_list template: %v", err)
		return c.HTML(http.StatusOK, `<p style="color: #6b7280;">Failed to load pending requests</p>`)
	}

	return c.HTML(http.StatusOK, html)
}

// CancelMyJoinRequestHTMLHandler cancels request and returns empty (for HTMX swap removal)
func (h *Handler) CancelMyJoinRequestHTMLHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Find and delete the user's request
	if err := h.RoomService.CancelJoinRequest(ctx, roomID, userID); err != nil {
		log.Printf("Error cancelling join request: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to cancel request")
	}

	// Return empty response - HTMX will remove the element
	return c.HTML(http.StatusOK, "")
}
