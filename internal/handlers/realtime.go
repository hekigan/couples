package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/models"
	"github.com/yourusername/couple-card-game/internal/services"
)

// RealtimeHandler handles real-time connections
type RealtimeHandler struct {
	handler         *Handler
	realtimeService *services.RealtimeService
}

// NewRealtimeHandler creates a new realtime handler
func NewRealtimeHandler(h *Handler, realtimeService *services.RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{
		handler:         h,
		realtimeService: realtimeService,
	}
}

// StreamRoomEvents streams room events via SSE
func (h *RealtimeHandler) StreamRoomEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Subscribe to room events
	client := h.realtimeService.Subscribe(roomID, userID)
	defer h.realtimeService.Unsubscribe(client.ID)

	// Stream events
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\",\"room_id\":\"%s\"}\n\n", roomID)
	flusher.Flush()

	// Send initial join requests (if user is room owner)
	ctx := r.Context()
	room, err := h.handler.RoomService.GetRoomByID(ctx, roomID)
	if err == nil && room.OwnerID == userID {
		// Get pending join requests with user info
		requests, err := h.handler.RoomService.GetJoinRequestsWithUserInfo(ctx, roomID)
		if err == nil && len(requests) > 0 {
			// Send each request as a join_request event
			for _, req := range requests {
				data, _ := json.Marshal(req)
				fmt.Fprintf(w, "event: join_request\ndata: %s\n\n", string(data))
			}
			flusher.Flush()
		}
	}

	// Stream events until client disconnects
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-client.Channel:
			if !ok {
				// Channel closed
				return
			}
			// Send event from realtime service
			sseData := services.EventToSSE(event)
			if _, err := fmt.Fprint(w, sseData); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Send keepalive ping every 15 seconds
			if _, err := fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// StreamUserNotifications streams user-specific notifications via SSE
func (h *RealtimeHandler) StreamUserNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Subscribe to user events (use a dummy room ID since we're not in a room)
	dummyRoomID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	client := h.realtimeService.Subscribe(dummyRoomID, userID)
	defer h.realtimeService.Unsubscribe(client.ID)

	// Stream events
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\",\"user_id\":\"%s\"}\n\n", userID)
	flusher.Flush()

	// Stream events until client disconnects
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-client.Channel:
			if !ok {
				// Channel closed
				return
			}
			// Send event from realtime service
			sseData := services.EventToSSE(event)
			if _, err := fmt.Fprint(w, sseData); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Send keepalive ping every 15 seconds
			if _, err := fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// GetRoomPlayers gets current room players
func (h *RealtimeHandler) GetRoomPlayers(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

// RoomStateResponse represents the room state for the frontend
type RoomStateResponse struct {
	ID              uuid.UUID        `json:"id"`
	Status          string           `json:"status"`
	CurrentQuestion int              `json:"current_question"`
	CurrentTurn     *uuid.UUID       `json:"current_turn"` // Note: different field name for frontend compatibility
	MaxQuestions    int              `json:"max_questions"`
	OwnerID         uuid.UUID        `json:"owner_id"`
	GuestID         *uuid.UUID       `json:"guest_id"`
	Language        string           `json:"language"`
	CurrentQuestionData *models.Question `json:"current_question_data,omitempty"` // Include full question data if exists
}

// GetRoomState gets current room state
func (h *RealtimeHandler) GetRoomState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.handler.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Check if user is part of this room
	isOwner := room.OwnerID == userID
	isGuest := room.GuestID != nil && *room.GuestID == userID
	if !isOwner && !isGuest {
		http.Error(w, "You are not a member of this room", http.StatusForbidden)
		return
	}

	// Fetch the current question if it exists
	var currentQuestionData *models.Question
	if room.CurrentQuestionID != nil {
		question, err := h.handler.QuestionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err != nil {
			// Log error but don't fail the request - question might have been deleted
			fmt.Printf("Warning: Failed to fetch current question %s: %v\n", room.CurrentQuestionID, err)
		} else {
			currentQuestionData = question
		}
	}

	// Create response with frontend-compatible field names
	response := RoomStateResponse{
		ID:                  room.ID,
		Status:              room.Status,
		CurrentQuestion:     room.CurrentQuestion,
		CurrentTurn:         room.CurrentTurn, // Maps CurrentTurn to current_turn in JSON
		MaxQuestions:        room.MaxQuestions,
		OwnerID:             room.OwnerID,
		GuestID:             room.GuestID,
		Language:            room.Language,
		CurrentQuestionData: currentQuestionData, // Include question data if exists
	}

	// Return the room state as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

