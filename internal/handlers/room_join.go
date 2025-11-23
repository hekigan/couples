package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
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
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

	requests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		http.Error(w, "Failed to load join requests", http.StatusInternalServerError)
		return
	}

	data := &TemplateData{
		Title: "Join Requests",
		User:  currentUser,
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

	// Render HTML fragment for the join request
	log.Printf("🎨 Rendering join_request template for user %s (%s)", username, userID)
	html, err := h.TemplateService.RenderFragment("join_request.html", services.JoinRequestData{
		ID:        request.ID.String(),
		RoomID:    roomID.String(),
		Username:  username,
		CreatedAt: request.CreatedAt.Format("3:04 PM"),
	})
	if err != nil {
		log.Printf("❌ Failed to render join request template: %v (falling back to JSON SSE)", err)
		// Fall back to JSON SSE if template rendering fails
		h.RoomService.BroadcastJoinRequestWithDetails(roomID, request.ID, userID, username, request.CreatedAt)
	} else {
		log.Printf("✅ Successfully rendered join_request HTML fragment (%d bytes)", len(html))
		log.Printf("📤 Broadcasting HTML fragment via SSE to room %s", roomID)
		// Broadcast HTML fragment via SSE
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
			Type:       "join_request",
			Target:     "#join-requests",
			SwapMethod: "beforeend", // Append to the list of join requests
			HTML:       html,
		})
		log.Printf("✅ Join request HTML fragment broadcasted successfully")
	}

	// Update the badge count (count pending requests after adding this one)
	pendingCount := 0
	if allRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID); err == nil {
		for _, req := range allRequests {
			if req.Status == "pending" {
				pendingCount++
			}
		}
	}
	// Broadcast badge update
	if badgeHTML, err := h.TemplateService.RenderFragment("badge_update.html", services.BadgeUpdateData{Count: pendingCount}); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
			Type:       "badge_update",
			Target:     "#request-count",
			SwapMethod: "innerHTML",
			HTML:       badgeHTML,
		})
	}

	// Return success message for HTMX
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, `<div style="color: #10b981; background: #dcfce7; padding: 0.75rem 1rem; border-radius: 6px;">✅ Join request sent! Waiting for owner approval...</div>`)
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

	// Render HTML fragment for request accepted (shown to room owner)
	html, err := h.TemplateService.RenderFragment("request_accepted.html", services.RequestAcceptedData{
		GuestUsername: guestUsername,
	})
	if err != nil {
		log.Printf("⚠️ Failed to render request_accepted template: %v (falling back to JSON SSE)", err)
		// Fall back to JSON SSE
		h.RoomService.BroadcastRequestAccepted(joinRequest.RoomID, guestUsername)
	} else {
		// Broadcast HTML fragment to room owner
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
			Type:       "request_accepted",
			Target:     "#guest-info",
			SwapMethod: "innerHTML", // Replace guest info section
			HTML:       html,
		})
	}

	// Notify the guest user that their request was accepted (use JSON for now, will enhance later)
	h.RoomService.BroadcastRequestAcceptedToGuest(joinRequest.UserID, joinRequest.RoomID)

	// Update the badge count (count pending requests after accepting)
	pendingCount := 0
	if allRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, joinRequest.RoomID); err == nil {
		for _, req := range allRequests {
			if req.Status == "pending" {
				pendingCount++
			}
		}
	}
	// Broadcast badge update
	if badgeHTML, err := h.TemplateService.RenderFragment("badge_update.html", services.BadgeUpdateData{Count: pendingCount}); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
			Type:       "badge_update",
			Target:     "#request-count",
			SwapMethod: "innerHTML",
			HTML:       badgeHTML,
		})
	}

	// Broadcast room status badge update (room is now "ready")
	// Refetch the room to get the updated status
	room, err := h.RoomService.GetRoomByID(ctx, joinRequest.RoomID)
	if err != nil {
		log.Printf("⚠️ Failed to refetch room for status badge update: %v", err)
	} else {
		log.Printf("🔍 Room status after accepting join request: %s", room.Status)
		if statusBadgeHTML, err := h.TemplateService.RenderFragment("status_badge.html", services.RoomStatusBadgeData{
			Status: room.Status,
		}); err == nil {
			h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
				Type:       "room_status_update",
				Target:     "#room-status-badge",
				SwapMethod: "outerHTML",
				HTML:       statusBadgeHTML,
			})
			log.Printf("📡 Broadcasted room status badge update (status: %s) after accepting join request", room.Status)
		}
	}

	// Return empty HTML - HTMX will remove the element via outerHTML swap
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
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

	// Update the badge count (count pending requests after rejecting)
	pendingCount := 0
	if allRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, joinRequest.RoomID); err == nil {
		for _, req := range allRequests {
			if req.Status == "pending" {
				pendingCount++
			}
		}
	}
	// Broadcast badge update
	if badgeHTML, err := h.TemplateService.RenderFragment("badge_update.html", services.BadgeUpdateData{Count: pendingCount}); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(joinRequest.RoomID, services.HTMLFragmentEvent{
			Type:       "badge_update",
			Target:     "#request-count",
			SwapMethod: "innerHTML",
			HTML:       badgeHTML,
		})
	}

	// Return empty HTML - HTMX will remove the element via outerHTML swap
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
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

// GetMyJoinRequestsHTMLHandler returns HTML fragment for all user's join requests
func (h *Handler) GetMyJoinRequestsHTMLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get all join requests for this user (pending, accepted, rejected)
	allRequests, err := h.RoomService.GetJoinRequestsByUser(ctx, userID)
	if err != nil {
		log.Printf("Error fetching user's join requests: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!-- No pending requests -->`))
		return
	}

	// Convert to template data format
	var requestsData []services.JoinRequestData
	for _, req := range allRequests {
		message := ""
		if req.Message != nil {
			message = *req.Message
		}
		requestsData = append(requestsData, services.JoinRequestData{
			RoomID:  req.RoomID.String(),
			UserID:  req.UserID.String(),
			Message: message,
			Status:  req.Status,
		})
	}

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("pending_requests_list.html", requestsData)
	if err != nil {
		log.Printf("Error rendering pending_requests_list template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load pending requests</p>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// CancelMyJoinRequestHTMLHandler cancels request and returns empty (for HTMX swap removal)
func (h *Handler) CancelMyJoinRequestHTMLHandler(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Error cancelling join request: %v", err)
		http.Error(w, "Failed to cancel request", http.StatusInternalServerError)
		return
	}

	// Return empty response - HTMX will remove the element
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(``))
}
