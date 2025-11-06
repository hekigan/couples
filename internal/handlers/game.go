package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/models"
)

// DeleteRoomAPIHandler handles room deletion via API
func (h *Handler) DeleteRoomAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room to check ownership
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Only the owner can delete the room
	if room.OwnerID != userID {
		http.Error(w, "You are not the owner of this room", http.StatusForbidden)
		return
	}

	// Delete the room
	if err := h.RoomService.DeleteRoom(ctx, roomID); err != nil {
		http.Error(w, "Failed to delete room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Room deleted successfully"}`))
}

// RoomWithUsername is a room enriched with the other player's username
type RoomWithUsername struct {
	*models.Room
	OtherPlayerUsername string
}

// ListRoomsHandler lists all rooms for the current user
func (h *Handler) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	rooms, err := h.RoomService.GetRoomsByUserID(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load rooms", http.StatusInternalServerError)
		return
	}

	// Enrich rooms with the other player's username
	enrichedRooms := make([]RoomWithUsername, 0, len(rooms))
	for _, room := range rooms {
		enrichedRoom := RoomWithUsername{
			Room: &room,
		}
		
		// Determine who the "other player" is
		var otherPlayerID uuid.UUID
		if room.OwnerID == userID {
			// Current user is owner, so other player is guest
			if room.GuestID != nil {
				otherPlayerID = *room.GuestID
			}
		} else if room.GuestID != nil && *room.GuestID == userID {
			// Current user is guest, so other player is owner
			otherPlayerID = room.OwnerID
		}
		
		// Fetch the other player's username
		if otherPlayerID != uuid.Nil {
			otherPlayer, err := h.UserService.GetUserByID(ctx, otherPlayerID)
			if err == nil && otherPlayer != nil {
				enrichedRoom.OtherPlayerUsername = otherPlayer.Username
			}
		}
		
		enrichedRooms = append(enrichedRooms, enrichedRoom)
	}

	data := &TemplateData{
		Title: "My Rooms",
		Data:  enrichedRooms,
	}
	h.RenderTemplate(w, "game/rooms.html", data)
}

// CreateRoomHandler handles room creation
func (h *Handler) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := &TemplateData{
			Title: "Create Room",
		}
		h.RenderTemplate(w, "game/create-room.html", data)
		return
	}

	// POST - Create room
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

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
	if r.Method == "GET" {
		data := &TemplateData{
			Title: "Join Room",
		}
		h.RenderTemplate(w, "game/join-room.html", data)
		return
	}

	// POST - Join room logic
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	
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

// RoomHandler displays the room lobby
func (h *Handler) RoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, _ := uuid.Parse(vars["id"])

	ctx := context.Background()
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get current user ID from context
	currentUserID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get owner username
	owner, err := h.UserService.GetUserByID(ctx, room.OwnerID)
	var ownerUsername string
	if err == nil && owner != nil {
		ownerUsername = owner.Username
	}

	// Get guest username if present
	var guestUsername string
	if room.GuestID != nil {
		guest, err := h.UserService.GetUserByID(ctx, *room.GuestID)
		if err == nil && guest != nil {
			guestUsername = guest.Username
		}
	}

	// Check if current user is the owner
	isOwner := currentUserID == room.OwnerID

	data := &TemplateData{
		Title:          "Room - " + room.Name,
		Data:           room,
		OwnerUsername:  ownerUsername,
		GuestUsername:  guestUsername,
		IsOwner:        isOwner,
	}
	h.RenderTemplate(w, "game/room.html", data)
}

// PlayHandler handles the game play screen
func (h *Handler) PlayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, _ := uuid.Parse(vars["id"])

	ctx := context.Background()
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	data := &TemplateData{
		Title: "Play - " + room.Name,
		Data:  room,
	}
	h.RenderTemplate(w, "game/play.html", data)
}

// GameFinishedHandler shows game results
func (h *Handler) GameFinishedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, _ := uuid.Parse(vars["id"])

	ctx := context.Background()
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	data := &TemplateData{
		Title: "Game Finished",
		Data:  room,
	}
	h.RenderTemplate(w, "game/finished.html", data)
}

// DeleteRoomHandler deletes a room
func (h *Handler) DeleteRoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, _ := uuid.Parse(vars["id"])

	ctx := context.Background()
	if err := h.RoomService.DeleteRoom(ctx, roomID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?success=Room+deleted", http.StatusSeeOther)
}

// API Handlers (stubs)
func (h *Handler) StartGameAPIHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func (h *Handler) DrawQuestionAPIHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func (h *Handler) SubmitAnswerAPIHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func (h *Handler) FinishGameAPIHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

