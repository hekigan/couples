package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
	roomFragments "github.com/hekigan/couples/internal/views/fragments/room"
	"github.com/labstack/echo/v4"
)

// SetGuestReadyAPIHandler marks the guest as ready
func (h *Handler) SetGuestReadyAPIHandler(c echo.Context) error {
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

	// Verify user is the guest
	if room.GuestID == nil || *room.GuestID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only the guest can mark themselves as ready")
	}

	// Mark guest as ready
	room.GuestReady = true
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("❌ Failed to update guest ready status: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update ready status: "+err.Error())
	}

	log.Printf("✅ Guest %s marked as ready in room %s", userID, roomID)

	// Refetch room to get updated status from database
	room, err = h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("❌ Failed to refetch room: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to refetch room")
	}

	log.Printf("🔍 Room status after refetch: %s (GuestReady: %v)", room.Status, room.GuestReady)

	// Use helper to render HTML fragment for updated button state
	html, err := h.RenderTemplFragment(c, roomFragments.GuestReadyButton(&services.GuestReadyButtonData{
		RoomID:     roomID.String(),
		GuestReady: true,
	}))
	if err != nil {
		log.Printf("Error rendering guest ready button template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render button")
	}

	// Broadcast room status badge update via SSE (so owner sees status change)
	if statusBadgeHTML, err := h.RenderTemplFragment(c, roomFragments.StatusBadge(&services.RoomStatusBadgeData{
		Status: room.Status,
	})); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
			Type:       "room_status_update",
			Target:     "#room-status-badge",
			SwapMethod: "outerHTML",
			HTML:       statusBadgeHTML,
		})
		log.Printf("📡 Broadcasted room status badge update to room %s (status: %s)", roomID, room.Status)
	}

	// Broadcast room_update event to trigger start button refresh
	h.RoomService.BroadcastRoomUpdate(roomID, map[string]interface{}{
		"guest_ready": true,
	})

	log.Printf("📡 Broadcasting guest_ready=true to room %s", roomID)

	// Step 7: Broadcast step transition with role-specific HTML
	if err := h.BroadcastRoleSpecificRoomContainer(c, ctx, roomID); err != nil {
		log.Printf("⚠️ Failed to broadcast room container: %v", err)
	}

	// Return HTML fragment for HTMX
	return c.HTML(http.StatusOK, html)
}

// RoomStatusBadgeAPIHandler returns the room status badge HTML fragment
func (h *Handler) RoomStatusBadgeAPIHandler(c echo.Context) error {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	log.Printf("🔍 RoomStatusBadgeAPIHandler: Room %s status: %s", roomID, room.Status)

	// Use helper to render HTML fragment for status badge
	return h.RenderTemplComponent(c, roomFragments.StatusBadge(&services.RoomStatusBadgeData{
		Status: room.Status,
	}))
}

// GetStartGameButtonHTMLHandler returns start game button as HTML fragment (for HTMX)
func (h *Handler) GetStartGameButtonHTMLHandler(c echo.Context) error {
	// Parse room ID
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "invalid room ID")
	}

	// Use GetRoomWithPlayers to get guest username
	ctx := context.Background()
	room, err := h.RoomService.GetRoomWithPlayers(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "room not found")
	}

	guestUsername := ""
	if room.GuestUsername != nil {
		guestUsername = *room.GuestUsername
	}

	// Use helper to render HTML fragment
	return h.RenderTemplComponent(c, roomFragments.StartGameButton(&services.StartGameButtonData{
		RoomID:        roomID.String(),
		GuestReady:    room.GuestReady,
		GuestUsername: guestUsername,
	}))
}

// GetGuestReadyButtonHTMLHandler returns guest ready button as HTML fragment (for HTMX)
func (h *Handler) GetGuestReadyButtonHTMLHandler(c echo.Context) error {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	// Use helper to render HTML fragment
	return h.RenderTemplComponent(c, roomFragments.GuestReadyButton(&services.GuestReadyButtonData{
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	}))
}
