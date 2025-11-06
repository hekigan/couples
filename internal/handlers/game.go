package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/models"
)

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
	userID := r.Context().Value("user_id").(uuid.UUID)

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
	http.Error(w, "Join room not yet implemented", http.StatusNotImplemented)
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

	data := &TemplateData{
		Title: "Room - " + room.Name,
		Data:  room,
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

