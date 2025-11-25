package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
)

// DeleteRoomAPIHandler handles room deletion via API
func (h *Handler) DeleteRoomAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify ownership
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Only the owner can delete the room
	if room.OwnerID != userID {
		http.Error(w, "You are not the owner of this room", http.StatusForbidden)
		return
	}

	ctx := context.Background()
	if err := h.RoomService.DeleteRoom(ctx, roomID); err != nil {
		http.Error(w, "Failed to delete room", http.StatusInternalServerError)
		return
	}

	// Return empty response for HTMX to replace with nothing (removes the element)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

// LeaveRoomHandler handles a guest leaving a room
func (h *Handler) LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	log.Printf("DEBUG LeaveRoom: Room before update - ID: %s, OwnerID: %s, GuestID: %v, Status: %s",
		room.ID, room.OwnerID, room.GuestID, room.Status)
	log.Printf("DEBUG LeaveRoom: User requesting leave: %s", userID)

	// Check if user is the guest
	if room.GuestID == nil || *room.GuestID != userID {
		log.Printf("ERROR LeaveRoom: User %s is not the guest of room %s (GuestID: %v)",
			userID, roomID, room.GuestID)
		http.Error(w, "You are not a guest in this room", http.StatusForbidden)
		return
	}

	// Remove the guest from the room by setting GuestID to nil
	room.GuestID = nil
	room.Status = "waiting" // Reset status to waiting

	log.Printf("DEBUG LeaveRoom: Updating room to remove guest - GuestID will be set to nil")

	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("ERROR LeaveRoom: Failed to update room: %v", err)
		http.Error(w, "Failed to leave room", http.StatusInternalServerError)
		return
	}

	log.Printf("DEBUG LeaveRoom: Successfully updated room %s, guest removed", roomID)

	// Return empty response for HTMX to replace with nothing (removes the element)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

// CreateRoomHandler handles room creation
func (h *Handler) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		data := &TemplateData{
			Title: "Create Room",
			User:  currentUser,
		}
		h.RenderTemplate(w, "game/create-room.html", data)
		return
	}

	// POST - Create room

	room := &models.Room{
		ID:       uuid.New(),
		Name:     r.FormValue("name"),
		OwnerID:  userID,
		Status:   "waiting",
		Language: "en",
	}

	if err := h.RoomService.CreateRoom(ctx, room); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/game/room/"+room.ID.String(), http.StatusSeeOther)
}

// JoinRoomHandler handles joining a room
func (h *Handler) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		data := &TemplateData{
			Title: "Join Room",
			User:  currentUser,
		}
		h.RenderTemplate(w, "game/join-room.html", data)
		return
	}

	// POST - Join room logic

	// Parse room ID from form
	roomIDStr := r.FormValue("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Check if room is already full
	if room.GuestID != nil {
		http.Error(w, "Room is full", http.StatusBadRequest)
		return
	}

	// Check if user is trying to join their own room
	if room.OwnerID == userID {
		http.Error(w, "You cannot join your own room", http.StatusBadRequest)
		return
	}

	// Join the room by setting the guest_id
	room.GuestID = &userID
	room.Status = "ready" // Room is ready when both players are present

	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	// Redirect to the room
	http.Redirect(w, r, "/game/room/"+room.ID.String(), http.StatusSeeOther)
}

// DeleteRoomHandler deletes a room (full page handler)
func (h *Handler) DeleteRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to extract room ID
	roomID, err := ExtractIDFromRouteVars(r, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.RoomService.DeleteRoom(ctx, roomID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?success=Room+deleted", http.StatusSeeOther)
}
