package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/services"
)

// ListRoomsHandler returns an HTML fragment with the rooms list
func (ah *AdminAPIHandler) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper for pagination
	page, perPage, offset := handlers.ParsePaginationParams(r)

	rooms, err := ah.adminService.ListAllRooms(ctx, perPage, offset)
	if err != nil {
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		log.Printf("Error listing rooms: %v", err)
		return
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

	html, err := ah.handler.TemplateService.RenderFragment("rooms_list", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering rooms list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// CloseRoomHandler forces a room to close
func (ah *AdminAPIHandler) CloseRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	roomID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ah.adminService.ForceCloseRoom(ctx, roomID); err != nil {
		http.Error(w, "Failed to close room", http.StatusInternalServerError)
		log.Printf("Error closing room: %v", err)
		return
	}

	// Return updated rooms list
	ah.ListRoomsHandler(w, r)
}

// GetRoomDetailsHandler returns room details for viewing (read-only)
func (ah *AdminAPIHandler) GetRoomDetailsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	roomID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	room, err := ah.handler.RoomService.GetRoomWithPlayers(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
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

	html, err := ah.handler.TemplateService.RenderFragment("room_details.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering room details: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
