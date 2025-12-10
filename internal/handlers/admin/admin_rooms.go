package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
	adminFragments "github.com/hekigan/couples/internal/views/fragments/admin"
)

// ListRoomsHandler returns an HTML fragment with the rooms list
func (ah *AdminAPIHandler) ListRoomsHandler(c echo.Context) error {
	ctx := context.Background()

	// Use helper for pagination
	page, perPage := handlers.ParsePaginationParams(c)
	offset := (page - 1) * perPage

	rooms, err := ah.adminService.ListAllRooms(ctx, perPage, offset)
	if err != nil {
		log.Printf("Error listing rooms: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list rooms")
	}

	totalCount, _ := ah.adminService.GetRoomCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Build data for template
	roomInfos := make([]services.AdminRoomInfo, len(rooms))
	for i, room := range rooms {
		owner := "Unknown"
		if room.OwnerUsername != nil {
			owner = *room.OwnerUsername
		}

		guest := "Waiting..."
		if room.GuestUsername != nil {
			guest = *room.GuestUsername
		}

		roomInfos[i] = services.AdminRoomInfo{
			ID:        room.ID.String(),
			ShortID:   room.ID.String()[:8],
			Owner:     owner,
			Guest:     guest,
			Status:    room.Status,
			CreatedAt: room.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	data := services.RoomsListData{
		Rooms: roomInfos,
		// Pagination fields
		TotalCount:      totalCount,
		CurrentPage:     page,
		TotalPages:      totalPages,
		ItemsPerPage:    perPage,
		BaseURL:         "/admin/api/rooms/list",
		PageURL:         "/admin/rooms",
		Target:          "#rooms-list",
		IncludeSelector: "",
		ExtraParams:     "",
		ItemName:        "rooms",
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.RoomsList(&data))
	if err != nil {
		log.Printf("Error rendering rooms list: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// CloseRoomHandler forces a room to close
func (ah *AdminAPIHandler) CloseRoomHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	roomID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ah.adminService.ForceCloseRoom(ctx, roomID); err != nil {
		log.Printf("Error closing room: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to close room")
	}

	// Return updated rooms list
	return ah.ListRoomsHandler(c)
}

// GetRoomDetailsHandler returns room details for viewing (read-only)
func (ah *AdminAPIHandler) GetRoomDetailsHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	roomID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	room, err := ah.handler.RoomService.GetRoomWithPlayers(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
	}

	// TODO: Get category count - for now default to 0
	categoryCount := 0

	ownerUsername := ""
	if room.OwnerUsername != nil {
		ownerUsername = *room.OwnerUsername
	}

	ownerEmail := ""
	if room.OwnerEmail != nil {
		ownerEmail = *room.OwnerEmail
	}

	guestUsername := ""
	if room.GuestUsername != nil {
		guestUsername = *room.GuestUsername
	}

	guestEmail := ""
	if room.GuestEmail != nil {
		guestEmail = *room.GuestEmail
	}

	data := services.RoomDetailsData{
		ID:            room.ID.String(),
		ShortID:       room.ID.String()[:8],
		Name:          room.Name,
		Status:        room.Status,
		Language:      room.Language,
		MaxQuestions:  room.MaxQuestions,
		CreatedAt:     room.CreatedAt.Format("2006-01-02 15:04:05"),
		OwnerUsername: ownerUsername,
		OwnerEmail:    ownerEmail,
		GuestUsername: guestUsername,
		GuestEmail:    guestEmail,
		CategoryCount: categoryCount,
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.RoomDetails(&data))
	if err != nil {
		log.Printf("Error rendering room details: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}
