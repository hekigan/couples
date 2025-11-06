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

	// Get user information for SSE broadcast
	user, err := h.UserService.GetUserByID(ctx, userID)
	var username string
	if err == nil && user != nil {
		username = user.Username
	} else {
		username = "Unknown"
	}

	// Broadcast to SSE that a new join request was created with full details
	h.RoomService.BroadcastJoinRequestWithDetails(roomID, request.ID, userID, username, request.CreatedAt)

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
	
	// Get the join request to find the user ID
	joinRequest, err := h.RoomService.GetJoinRequestByID(ctx, requestID)
	if err != nil {
		http.Error(w, "Join request not found", http.StatusNotFound)
		return
	}
	
	// Accept the request
	if err := h.RoomService.AcceptJoinRequest(ctx, requestID); err != nil {
		fmt.Printf("ERROR: AcceptJoinRequest failed: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to accept join request: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Get the guest username
	guestUser, err := h.UserService.GetUserByID(ctx, joinRequest.UserID)
	var guestUsername string
	if err == nil && guestUser != nil {
		guestUsername = guestUser.Username
	}

	// Broadcast to room that request was accepted
	h.RoomService.BroadcastRequestAccepted(joinRequest.RoomID, guestUsername)

	// Notify the guest user that their request was accepted
	h.RoomService.BroadcastRequestAcceptedToGuest(joinRequest.UserID, joinRequest.RoomID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Join request accepted",
		"guest_username": guestUsername,
	})
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

	// Get the join request to find the user ID and room ID before deleting
	joinRequest, err := h.RoomService.GetJoinRequestByID(ctx, requestID)
	if err != nil {
		http.Error(w, "Join request not found", http.StatusNotFound)
		return
	}

	// Reject the request
	if err := h.RoomService.RejectJoinRequest(ctx, requestID); err != nil {
		http.Error(w, "Failed to reject join request", http.StatusInternalServerError)
		return
	}

	// Notify the guest user that their request was rejected
	h.RoomService.BroadcastRequestRejectedToGuest(joinRequest.UserID, joinRequest.RoomID)

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

