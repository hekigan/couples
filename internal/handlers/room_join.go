package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/models"
)

// ListJoinRequestsHandler lists join requests for a room
func (h *Handler) ListJoinRequestsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	requests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		http.Error(w, "Failed to load join requests", http.StatusInternalServerError)
		return
	}

	data := &TemplateData{
		Title: "Join Requests",
		Data:  requests,
	}
	h.RenderTemplate(w, "game/join-requests.html", data)
}

// CreateJoinRequestHandler creates a join request
func (h *Handler) CreateJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Parse room ID from form
	roomIDStr := r.FormValue("room_id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Get the room to check if it exists
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Check if user is trying to request their own room
	if room.OwnerID == userID {
		http.Error(w, "You cannot request to join your own room", http.StatusBadRequest)
		return
	}

	// Check if room is already full
	if room.GuestID != nil {
		http.Error(w, "Room is full", http.StatusBadRequest)
		return
	}

	// Create the join request
	message := r.FormValue("message")
	request := &models.RoomJoinRequest{
		ID:      uuid.New(),
		RoomID:  roomID,
		UserID:  userID,
		Status:  "pending",
		Message: &message,
	}

	if err := h.RoomService.CreateJoinRequest(ctx, request); err != nil {
		fmt.Printf("ERROR creating join request: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to create join request: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success","message":"Join request sent"}`))
}

// AcceptJoinRequestHandler accepts a join request
func (h *Handler) AcceptJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["request_id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.RoomService.AcceptJoinRequest(ctx, requestID); err != nil {
		http.Error(w, "Failed to accept join request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success","message":"Join request accepted"}`))
}

// RejectJoinRequestHandler rejects a join request
func (h *Handler) RejectJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["request_id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.RoomService.RejectJoinRequest(ctx, requestID); err != nil {
		http.Error(w, "Failed to reject join request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success","message":"Join request rejected"}`))
}

// GetJoinRequestsCountHandler gets count of pending requests
func (h *Handler) GetJoinRequestsCountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	requests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		http.Error(w, "Failed to load join requests", http.StatusInternalServerError)
		return
	}

	// Count pending requests
	count := 0
	for _, req := range requests {
		if req.Status == "pending" {
			count++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"count":%d}`, count)))
}

// GetJoinRequestsJSONHandler returns join requests with user information as JSON
func (h *Handler) GetJoinRequestsJSONHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	requests, err := h.RoomService.GetJoinRequestsWithUserInfo(ctx, roomID)
	if err != nil {
		http.Error(w, "Failed to load join requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// CheckMyJoinRequestHandler checks the status of the current user's join request for a room
func (h *Handler) CheckMyJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Check if user is already a guest in the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err == nil && room.GuestID != nil && *room.GuestID == userID {
		// User is now a guest - request was accepted!
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"accepted","message":"Request accepted"}`))
		return
	}

	// Check ALL join requests (not just pending) to see status
	allRequests, err := h.RoomService.GetAllJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"pending","message":"Waiting for approval"}`))
		return
	}

	// Find this user's request
	for _, req := range allRequests {
		if req.UserID == userID {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"status":"%s","message":"Request %s"}`, req.Status, req.Status)))
			return
		}
	}

	// No request found
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"not_found","message":"Request not found"}`))
}

// CancelMyJoinRequestHandler cancels the current user's join request
func (h *Handler) CancelMyJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Find and delete the user's request
	if err := h.RoomService.CancelJoinRequest(ctx, roomID, userID); err != nil {
		http.Error(w, "Failed to cancel request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success","message":"Request cancelled"}`))
}

// GetMyAcceptedRequestsHandler returns all accepted join requests for the current user
func (h *Handler) GetMyAcceptedRequestsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get all accepted requests for this user
	requests, err := h.RoomService.GetAcceptedRequestsByUser(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load accepted requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

