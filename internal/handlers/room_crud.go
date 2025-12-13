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
	roomFragments "github.com/hekigan/couples/internal/views/fragments/room"
	gamePages "github.com/hekigan/couples/internal/views/pages/game"
	"github.com/labstack/echo/v4"
)

// DeleteRoomAPIHandler handles room deletion via API
func (h *Handler) DeleteRoomAPIHandler(c echo.Context) error {
	// Use helper to get room and verify ownership
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Only the owner can delete the room
	if room.OwnerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not the owner of this room")
	}

	ctx := context.Background()
	if err := h.RoomService.DeleteRoom(ctx, roomID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete room")
	}

	// Return empty response for HTMX to replace with nothing (removes the element)
	return c.HTML(http.StatusOK, "")
}

// LeaveRoomHandler handles a guest leaving a room
func (h *Handler) LeaveRoomHandler(c echo.Context) error {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	log.Printf("DEBUG LeaveRoom: Room before update - ID: %s, OwnerID: %s, GuestID: %v, Status: %s",
		room.ID, room.OwnerID, room.GuestID, room.Status)
	log.Printf("DEBUG LeaveRoom: User requesting leave: %s", userID)

	// Check if user is the guest
	if room.GuestID == nil || *room.GuestID != userID {
		log.Printf("ERROR LeaveRoom: User %s is not the guest of room %s (GuestID: %v)",
			userID, roomID, room.GuestID)
		return echo.NewHTTPError(http.StatusForbidden, "You are not a guest in this room")
	}

	// Remove the guest from the room by setting GuestID to nil
	room.GuestID = nil
	room.Status = "waiting" // Reset status to waiting

	log.Printf("DEBUG LeaveRoom: Updating room to remove guest - GuestID will be set to nil")

	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("ERROR LeaveRoom: Failed to update room: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to leave room")
	}

	log.Printf("DEBUG LeaveRoom: Successfully updated room %s, guest removed", roomID)

	// Return empty response for HTMX to replace with nothing (removes the element)
	return c.HTML(http.StatusOK, "")
}

// CreateRoomHandler handles room creation
func (h *Handler) CreateRoomHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	if c.Request().Method == "GET" {
		data := NewTemplateData(c)
		data.Title = "Create Room"
		data.User = currentUser
		return h.RenderTemplComponent(c, gamePages.CreateRoomPage(data))
	}

	// POST - Create room

	// Parse is_private checkbox (checkbox is "on" when checked, empty when unchecked)
	isPrivate := c.FormValue("is_private") == "on"

	room := &models.Room{
		ID:        uuid.New(),
		Name:      c.FormValue("name"),
		OwnerID:   userID,
		Status:    "waiting",
		Language:  "en",
		IsPrivate: isPrivate,
	}

	if err := h.RoomService.CreateRoom(ctx, room); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/game/room/"+room.ID.String())
}

// JoinRoomHandler handles joining a room
func (h *Handler) JoinRoomHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	if c.Request().Method == "GET" {
		data := NewTemplateData(c)
		data.Title = "Join Room"
		data.User = currentUser
		return h.RenderTemplComponent(c, gamePages.JoinRoomPage(data))
	}

	// POST - Join room logic

	// Parse room ID from form
	roomIDStr := c.FormValue("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
	}

	// Check if room is already full
	if room.GuestID != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Room is full")
	}

	// Check if user is trying to join their own room
	if room.OwnerID == userID {
		return echo.NewHTTPError(http.StatusBadRequest, "You cannot join your own room")
	}

	// If room is private, create a join request instead of joining directly
	if room.IsPrivate {
		// Check if user already has a pending join request
		existingRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID)
		if err == nil {
			for _, req := range existingRequests {
				if req.UserID == userID && req.Status == "pending" {
					// Redirect back with message that request already exists
					return c.Redirect(http.StatusSeeOther, "/game/rooms?info=Join+request+already+pending")
				}
			}
		}

		// Create the join request
		message := c.FormValue("message") // Optional message field
		request := &models.RoomJoinRequest{
			ID:      uuid.New(),
			RoomID:  roomID,
			UserID:  userID,
			Status:  "pending",
			Message: &message,
		}

		if err := h.RoomService.CreateJoinRequest(ctx, request); err != nil {
			log.Printf("ERROR creating join request: %v\n", err)
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
			UserID:    userID.String(),
			Username:  username,
			Status:    request.Status,
			CreatedAt: request.CreatedAt.Format("Jan 2, 15:04"),
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

		// Redirect to rooms list with success message
		return c.Redirect(http.StatusSeeOther, "/game/rooms?success=Join+request+sent")
	}

	// For public rooms, join directly
	// Join the room by setting the guest_id
	room.GuestID = &userID
	room.Status = "ready" // Room is ready when both players are present

	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to join room")
	}

	// Redirect to the room
	return c.Redirect(http.StatusSeeOther, "/game/room/"+room.ID.String())
}

// DeleteRoomHandler deletes a room (full page handler)
func (h *Handler) DeleteRoomHandler(c echo.Context) error {
	// Use helper to extract room ID
	roomID, err := ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := context.Background()
	if err := h.RoomService.DeleteRoom(ctx, roomID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/?success=Room+deleted")
}
