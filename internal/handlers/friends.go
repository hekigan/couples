package handlers

import (
	"log"
	"net/http"
	
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/models"
)

// FriendsHandler displays the friends list
func (h *Handler) FriendsHandler(w http.ResponseWriter, r *http.Request) {
	h.FriendListHandler(w, r)
}

// FriendListHandler shows the user's friends list
func (h *Handler) FriendListHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	session, _ := middleware.Store.Get(r, "couple-card-game-session")
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get friends (using stub service for now)
	friends, err := h.FriendService.GetFriends(r.Context(), parsedUserID)
	if err != nil {
		log.Printf("Error getting friends: %v", err)
		friends = []models.Friend{} // Empty list
	}

	data := &TemplateData{
		Title: "Friends",
		Data: map[string]interface{}{
			"Friends":            friends,
			"PendingInvitations": []interface{}{},
			"CurrentUserID":      userID,
		},
	}

	h.RenderTemplate(w, "friends/list.html", data)
}

// AddFriendHandler handles sending friend invitations
func (h *Handler) AddFriendHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	session, _ := middleware.Store.Get(r, "couple-card-game-session")
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		data := &TemplateData{
			Title: "Add Friend",
			Data: map[string]interface{}{
				"CurrentUserID": userID,
			},
		}
		h.RenderTemplate(w, "friends/add.html", data)
		return
	}

	// Handle POST request
	if err := r.ParseForm(); err != nil {
		h.RenderTemplate(w, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			Error: "Invalid form data",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
		return
	}

	friendIdentifier := r.FormValue("friend_identifier")
	if friendIdentifier == "" {
		h.RenderTemplate(w, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			Error: "Friend identifier is required",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
		return
	}

	// Parse user IDs
	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid current user ID", http.StatusBadRequest)
		return
	}

	friendID, err := uuid.Parse(friendIdentifier)
	if err != nil {
		h.RenderTemplate(w, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			Error: "Invalid friend ID format. Please use UUID format.",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
		return
	}

	// Create friend request
	err = h.FriendService.CreateFriendRequest(r.Context(), currentUserID, friendID)
	if err != nil {
		h.RenderTemplate(w, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			Error: "Failed to send friend request",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
		return
	}

	h.RenderTemplate(w, "friends/add.html", &TemplateData{
		Title:   "Add Friend",
		Success: "Friend invitation sent!",
		Data:    map[string]interface{}{"CurrentUserID": userID},
	})
}

// AcceptFriendHandler accepts a friend invitation
func (h *Handler) AcceptFriendHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	err = h.FriendService.AcceptFriendRequest(r.Context(), requestID)
	if err != nil {
		log.Printf("Error accepting friend request: %v", err)
		http.Error(w, "Failed to accept friend request", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/friends", http.StatusSeeOther)
}

// DeclineFriendHandler declines a friend invitation
func (h *Handler) DeclineFriendHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	// For now, just redirect (service method not implemented)
	log.Printf("Would decline friend request: %s", requestID)
	http.Redirect(w, r, "/friends", http.StatusSeeOther)
}

// SendFriendRequestHandler sends a friend request
func (h *Handler) SendFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	h.AddFriendHandler(w, r)
}

// AcceptFriendRequestHandler accepts a friend request
func (h *Handler) AcceptFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	h.AcceptFriendHandler(w, r)
}

// RejectFriendRequestHandler rejects a friend request
func (h *Handler) RejectFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	h.DeclineFriendHandler(w, r)
}

// RemoveFriendHandler removes a friendship
func (h *Handler) RemoveFriendHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	friendshipID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid friendship ID", http.StatusBadRequest)
		return
	}

	// For now, just redirect (service method not implemented)
	log.Printf("Would remove friendship: %s", friendshipID)
	http.Redirect(w, r, "/friends", http.StatusSeeOther)
}
