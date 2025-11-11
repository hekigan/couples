package handlers

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/models"
	"github.com/yourusername/couple-card-game/internal/services"
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

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(r.Context(), parsedUserID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

	// Get friends
	friends, err := h.FriendService.GetFriends(r.Context(), parsedUserID)
	if err != nil {
		log.Printf("Error getting friends: %v", err)
		friends = []models.FriendWithUserInfo{} // Empty list
	}

	// Get pending friend requests
	pendingRequests, err := h.FriendService.GetPendingRequests(r.Context(), parsedUserID)
	if err != nil {
		log.Printf("Error getting pending requests: %v", err)
		pendingRequests = []models.FriendWithUserInfo{} // Empty list
	}

	data := &TemplateData{
		Title: "Friends",
		User:  currentUser,
		Data: map[string]interface{}{
			"Friends":            friends,
			"PendingInvitations": pendingRequests,
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

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get current user for template
	currentUser, err := h.UserService.GetUserByID(r.Context(), parsedUserID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch current user: %v", err)
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		data := &TemplateData{
			Title: "Add Friend",
			User:  currentUser,
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
			User:  currentUser,
			Error: "Invalid form data",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
		return
	}

	friendIdentifier := r.FormValue("friend_identifier")
	if friendIdentifier == "" {
		h.RenderTemplate(w, "friends/add.html", &TemplateData{
			Title: "Add Friend",
			User:  currentUser,
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
			User:  currentUser,
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
			User:  currentUser,
			Error: "Failed to send friend request",
			Data:  map[string]interface{}{"CurrentUserID": userID},
		})
		return
	}

	h.RenderTemplate(w, "friends/add.html", &TemplateData{
		Title:   "Add Friend",
		User:    currentUser,
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

	err = h.FriendService.DeclineFriendRequest(r.Context(), requestID)
	if err != nil {
		log.Printf("Error declining friend request: %v", err)
		http.Error(w, "Failed to decline friend request", http.StatusInternalServerError)
		return
	}

	log.Printf("Declined friend request: %s", requestID)
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

	err = h.FriendService.RemoveFriend(r.Context(), friendshipID)
	if err != nil {
		log.Printf("Error removing friend: %v", err)
		http.Error(w, "Failed to remove friend", http.StatusInternalServerError)
		return
	}

	log.Printf("Removed friendship: %s", friendshipID)

	// Return success JSON for AJAX requests
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Friend removed"}`))
}

// GetFriendsAPIHandler returns friends list as JSON (for room invitations)
func (h *Handler) GetFriendsAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get friends
	friends, err := h.FriendService.GetFriends(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting friends: %v", err)
		http.Error(w, "Failed to get friends", http.StatusInternalServerError)
		return
	}

	// Build JSON response manually
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonStr := `[`
	for i, friend := range friends {
		if i > 0 {
			jsonStr += ","
		}
		// Determine which ID is the actual friend (not the current user)
		friendIDStr := friend.FriendID.String()
		if friend.FriendID == userID {
			friendIDStr = friend.UserID.String()
		}
		jsonStr += `{"id":"` + friendIDStr + `","username":"` + friend.Username + `","name":"` + friend.Name + `"}`
	}
	jsonStr += `]`

	w.Write([]byte(jsonStr))
}

// GetFriendsHTMLHandler returns friends list as HTML fragment (for HTMX)
func (h *Handler) GetFriendsHTMLHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get room ID from query parameter
	roomID := r.URL.Query().Get("room_id")

	// Get friends
	friendsList, err := h.FriendService.GetFriends(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting friends: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load friends</p>`))
		return
	}

	// Build FriendsListData
	friends := make([]services.FriendInfo, 0, len(friendsList))
	for _, friend := range friendsList {
		// Determine which ID is the actual friend (not the current user)
		friendIDStr := friend.FriendID.String()
		if friend.FriendID == userID {
			friendIDStr = friend.UserID.String()
		}
		friends = append(friends, services.FriendInfo{
			ID:       friendIDStr,
			Username: friend.Username,
		})
	}

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("friends_list.html", services.FriendsListData{
		Friends: friends,
		RoomID:  roomID,
	})
	if err != nil {
		log.Printf("Error rendering friends list template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load friends</p>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
