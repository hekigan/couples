package admin

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

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

	// Parse query parameters
	search := strings.TrimSpace(c.QueryParam("search"))
	statuses := c.QueryParams()["status"]
	sortBy := c.QueryParam("sort")
	sortOrder := c.QueryParam("order")

	// Default to all statuses if none selected
	if len(statuses) == 0 {
		statuses = []string{"waiting", "ready", "playing", "finished"}
	}

	// Default sort to created_at DESC
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Call service with filters and sorting
	rooms, err := ah.adminService.ListAllRooms(ctx, perPage, offset, search, statuses, sortBy, sortOrder)
	if err != nil {
		log.Printf("Error listing rooms: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list rooms")
	}

	// Get filtered count instead of total count
	totalCount, _ := ah.adminService.GetFilteredRoomCount(ctx, search, statuses)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Fetch all categories once for mapping
	allCategories, err := ah.handler.CategoryService.GetCategories(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
	}

	// Build category map for quick lookup
	categoryMap := make(map[string]string) // ID -> Label
	for _, cat := range allCategories {
		categoryMap[cat.ID.String()] = cat.Label
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

		// Get category names for this room
		categoryNames := []string{}
		for _, selectedID := range room.SelectedCategories {
			if label, ok := categoryMap[selectedID.String()]; ok {
				categoryNames = append(categoryNames, label)
			}
		}

		roomInfos[i] = services.AdminRoomInfo{
			ID:            room.ID.String(),
			ShortID:       room.ID.String()[:8],
			Name:          room.Name,
			Owner:         owner,
			Guest:         guest,
			Status:        room.Status,
			CreatedAt:     room.CreatedAt.Format("2006-01-02 15:04"),
			CategoryNames: categoryNames,
		}
	}

	// Build extra params for pagination (preserve filters and sorting)
	extraParams := url.Values{}
	if search != "" {
		extraParams.Set("search", search)
	}
	for _, status := range statuses {
		extraParams.Add("status", status)
	}
	if sortBy != "" {
		extraParams.Set("sort", sortBy)
	}
	if sortOrder != "" {
		extraParams.Set("order", sortOrder)
	}
	extraParamsString := ""
	if len(extraParams) > 0 {
		extraParamsString = "&" + extraParams.Encode()
	}

	data := services.RoomsListData{
		Rooms: roomInfos,
		// Pagination fields
		TotalCount:      totalCount,
		CurrentPage:     page,
		TotalPages:      totalPages,
		ItemsPerPage:    perPage,
		BaseURL:         "/admin/api/v1/rooms/list",
		PageURL:         "/admin/rooms",
		Target:          "#rooms-list",
		IncludeSelector: "",
		ExtraParams:     extraParamsString,
		ItemName:        "rooms",
		// Filter/sort state
		SearchTerm:       search,
		SelectedStatuses: statuses,
		SortBy:           sortBy,
		SortOrder:        sortOrder,
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

// DeleteRoomHandler deletes a room
func (ah *AdminAPIHandler) DeleteRoomHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	roomID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ah.handler.RoomService.DeleteRoom(ctx, roomID); err != nil {
		log.Printf("Error deleting room: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete room")
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

	// Get selected category names
	categoryNames := []string{}
	categoryCount := len(room.SelectedCategories)

	if categoryCount > 0 {
		// Fetch all categories
		allCategories, err := ah.handler.CategoryService.GetCategories(ctx)
		if err != nil {
			log.Printf("Error fetching categories: %v", err)
		} else {
			// Build a map for quick lookup
			categoryMap := make(map[string]string) // ID -> Label
			for _, cat := range allCategories {
				categoryMap[cat.ID.String()] = cat.Label
			}

			// Get labels for selected categories
			for _, selectedID := range room.SelectedCategories {
				if label, ok := categoryMap[selectedID.String()]; ok {
					categoryNames = append(categoryNames, label)
				}
			}
		}
	}

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
		CategoryNames: categoryNames,
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.RoomDetails(&data))
	if err != nil {
		log.Printf("Error rendering room details: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}
