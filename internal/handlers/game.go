package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/models"
	"github.com/yourusername/couple-card-game/internal/services"
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

// LeaveRoomHandler handles a guest leaving a room
func (h *Handler) LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room to check if user is the guest
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Left room successfully"}`))
}

// RoomWithUsername is a room enriched with the other player's username
type RoomWithUsername struct {
	*models.Room
	OtherPlayerUsername string
	IsOwner             bool
}

// ListRoomsHandler lists all rooms for the current user
func (h *Handler) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	log.Printf("DEBUG ListRooms: Fetching rooms for user %s", userID)

	// Use database view to get rooms with player info
	// This eliminates N+1 queries (2 room queries + N user queries → 2 room queries with player info)
	roomsWithPlayers, err := h.RoomService.GetRoomsByUserIDWithPlayers(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load rooms", http.StatusInternalServerError)
		return
	}

	log.Printf("DEBUG ListRooms: Found %d rooms (via view)", len(roomsWithPlayers))

	// Convert to RoomWithUsername format for template
	enrichedRooms := make([]RoomWithUsername, 0, len(roomsWithPlayers))
	for _, roomWithPlayers := range roomsWithPlayers {
		log.Printf("DEBUG ListRooms: Room %s - OwnerID: %s, GuestID: %v, Status: %s",
			roomWithPlayers.ID, roomWithPlayers.OwnerID, roomWithPlayers.GuestID, roomWithPlayers.Status)

		isOwner := roomWithPlayers.OwnerID == userID
		enrichedRoom := RoomWithUsername{
			Room:    &roomWithPlayers.Room,
			IsOwner: isOwner,
		}

		// Determine other player's username from the view data (no extra query needed!)
		if isOwner {
			// Current user is owner, so other player is guest
			if roomWithPlayers.GuestUsername != nil {
				enrichedRoom.OtherPlayerUsername = *roomWithPlayers.GuestUsername
			}
		} else {
			// Current user is guest, so other player is owner
			if roomWithPlayers.OwnerUsername != nil {
				enrichedRoom.OtherPlayerUsername = *roomWithPlayers.OwnerUsername
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

	// Get current user ID from context
	currentUserID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use database view to get room with player info in single query
	// This replaces 3 queries (room + owner + guest) with 1 query
	roomWithPlayers, err := h.RoomService.GetRoomWithPlayers(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Extract usernames from the view result
	ownerUsername := ""
	if roomWithPlayers.OwnerUsername != nil {
		ownerUsername = *roomWithPlayers.OwnerUsername
	}

	guestUsername := ""
	if roomWithPlayers.GuestUsername != nil {
		guestUsername = *roomWithPlayers.GuestUsername
	}

	// Check if current user is the owner
	isOwner := currentUserID == roomWithPlayers.OwnerID

	// Debug logging
	log.Printf("🔍 DEBUG RoomHandler: roomID=%s, currentUserID=%s, ownerID=%s, isOwner=%v, status=%s, guestReady=%v (fetched via view)",
		roomWithPlayers.ID, currentUserID, roomWithPlayers.OwnerID, isOwner, roomWithPlayers.Status, roomWithPlayers.GuestReady)

	// Get pending join requests count for the badge
	joinRequestsCount := 0
	if isOwner {
		joinRequests, err := h.RoomService.GetJoinRequestsByRoom(ctx, roomID)
		if err == nil {
			// Count pending requests only
			for _, req := range joinRequests {
				if req.Status == "pending" {
					joinRequestsCount++
				}
			}
		}
	}

	data := &TemplateData{
		Title:             "Room - " + roomWithPlayers.Name,
		Data:              &roomWithPlayers.Room, // Pass the embedded Room struct
		OwnerUsername:     ownerUsername,
		GuestUsername:     guestUsername,
		IsOwner:           isOwner,
		JoinRequestsCount: joinRequestsCount,
	}

	// HTMX refactoring complete - using HTMX version as default
	h.RenderTemplate(w, "game/room.html", data)
}

// PlayHandler handles the game play screen
func (h *Handler) PlayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, _ := uuid.Parse(vars["id"])

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Create a map to pass both room and current user ID
	playData := map[string]interface{}{
		"Room":          room,
		"CurrentUserID": userID.String(),
	}

	data := &TemplateData{
		Title: "Play - " + room.Name,
		Data:  playData,
	}
	h.RenderTemplate(w, "game/play.html", data)
}

// AnswerWithDetails contains an answer with its question and user info
type AnswerWithDetails struct {
	Answer       *models.Answer
	Question     *models.Question
	Username     string
	ActionType   string
}

// GameFinishedHandler shows game results with full history
func (h *Handler) GameFinishedHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, _ := uuid.Parse(vars["id"])

	ctx := context.Background()
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get all answers for this room
	answers, err := h.AnswerService.GetAnswersByRoom(ctx, roomID)
	if err != nil {
		http.Error(w, "Failed to load answers", http.StatusInternalServerError)
		return
	}

	// Get owner and guest info
	owner, _ := h.UserService.GetUserByID(ctx, room.OwnerID)
	var guest *models.User
	if room.GuestID != nil {
		guest, _ = h.UserService.GetUserByID(ctx, *room.GuestID)
	}

	// Enrich answers with question and user details
	answerDetails := make([]AnswerWithDetails, 0, len(answers))
	for _, answer := range answers {
		// Get question
		question, err := h.QuestionService.GetQuestionByID(ctx, answer.QuestionID)
		if err != nil {
			continue // Skip if question not found
		}

		// Get username
		var username string
		if answer.UserID == room.OwnerID && owner != nil {
			username = owner.Username
		} else if guest != nil && answer.UserID == *room.GuestID {
			username = guest.Username
		}

		answerDetails = append(answerDetails, AnswerWithDetails{
			Answer:     &answer,
			Question:   question,
			Username:   username,
			ActionType: answer.ActionType,
		})
	}

	// Calculate statistics
	totalQuestions := len(answerDetails)
	passedCount := 0
	answeredCount := 0
	for _, ad := range answerDetails {
		if ad.ActionType == "passed" {
			passedCount++
		} else {
			answeredCount++
		}
	}

	data := &TemplateData{
		Title: "Game Finished",
		Data: map[string]interface{}{
			"Room":           room,
			"Answers":        answerDetails,
			"TotalQuestions": totalQuestions,
			"PassedCount":    passedCount,
			"AnsweredCount":  answeredCount,
			"OwnerUsername":  owner.Username,
			"GuestUsername":  "",
		},
	}

	if guest != nil {
		data.Data.(map[string]interface{})["GuestUsername"] = guest.Username
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

// StartGameAPIHandler starts the game for a room
func (h *Handler) StartGameAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Only the owner can start the game
	if room.OwnerID != userID {
		http.Error(w, "Only the room owner can start the game", http.StatusForbidden)
		return
	}

	// Verify room is ready (both players present)
	if room.Status != "ready" {
		http.Error(w, "Room is not ready to start. Wait for both players to join.", http.StatusBadRequest)
		return
	}

	// Verify guest is ready
	if !room.GuestReady {
		http.Error(w, "Guest is not ready yet. Wait for guest to click Ready button.", http.StatusBadRequest)
		return
	}

	// Start the game
	if err := h.GameService.StartGame(ctx, roomID); err != nil {
		http.Error(w, "Failed to start game: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return HTMX redirect header to navigate to play page
	w.Header().Set("HX-Redirect", fmt.Sprintf("/game/play/%s", roomID))
	w.WriteHeader(http.StatusOK)
}

// SetGuestReadyAPIHandler marks the guest as ready
func (h *Handler) SetGuestReadyAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify user is the guest
	if room.GuestID == nil || *room.GuestID != userID {
		http.Error(w, "Only the guest can mark themselves as ready", http.StatusForbidden)
		return
	}

	// Mark guest as ready
	room.GuestReady = true
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("❌ Failed to update guest ready status: %v", err)
		http.Error(w, "Failed to update ready status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Guest %s marked as ready in room %s", userID, roomID)

	// Render HTML fragment for updated button state
	html, err := h.TemplateService.RenderFragment("guest_ready_button.html", services.GuestReadyButtonData{
		RoomID:     roomID.String(),
		GuestReady: true,
	})
	if err != nil {
		log.Printf("Error rendering guest ready button template: %v", err)
		http.Error(w, "Failed to render button", http.StatusInternalServerError)
		return
	}

	// Broadcast guest ready event via SSE to update owner's start button
	h.RoomService.BroadcastRoomUpdate(roomID, map[string]interface{}{
		"guest_ready": true,
	})

	log.Printf("📡 Broadcasting guest_ready=true to room %s", roomID)

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// PlayerTypingAPIHandler broadcasts typing status to other players in the room
func (h *Handler) PlayerTypingAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Parse request body
	var req struct {
		IsTyping bool `json:"is_typing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the room to verify user is a participant
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify user is part of this room
	isOwner := room.OwnerID == userID
	isGuest := room.GuestID != nil && *room.GuestID == userID
	if !isOwner && !isGuest {
		http.Error(w, "You are not a member of this room", http.StatusForbidden)
		return
	}

	// Verify it's the user's turn
	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		// Silently ignore typing events if it's not their turn
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ignored"}`))
		return
	}

	// Broadcast typing status to other players in the room
	h.RoomService.BroadcastPlayerTyping(roomID, userID, req.IsTyping)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

// DrawQuestionAPIHandler draws a random question for the room
func (h *Handler) DrawQuestionAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify user is a participant
	if room.OwnerID != userID && (room.GuestID == nil || *room.GuestID != userID) {
		http.Error(w, "You are not a participant in this room", http.StatusForbidden)
		return
	}

	// Verify game is playing
	if room.Status != "playing" {
		http.Error(w, "Game is not in progress", http.StatusBadRequest)
		return
	}

	// Check if a question already exists (prevent race condition)
	var question *models.Question
	if room.CurrentQuestionID != nil {
		// Question already exists, fetch it instead of drawing a new one
		log.Printf("⚠️ Question already exists for room %s, returning existing question %s", roomID, *room.CurrentQuestionID)
		existingQuestion, err := h.QuestionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err != nil {
			log.Printf("❌ Failed to fetch existing question: %v", err)
			http.Error(w, "Failed to fetch existing question: "+err.Error(), http.StatusInternalServerError)
			return
		}
		question = existingQuestion
	} else {
		// No question exists, draw a new one
		log.Printf("🎴 Drawing new question for room %s", roomID)
		newQuestion, err := h.GameService.DrawQuestion(ctx, roomID)
		if err != nil {
			http.Error(w, "Failed to draw question: "+err.Error(), http.StatusInternalServerError)
			return
		}
		question = newQuestion
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"question": question,
	})
}

// SubmitAnswerAPIHandler handles answer submission
func (h *Handler) SubmitAnswerAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify user is a participant
	if room.OwnerID != userID && (room.GuestID == nil || *room.GuestID != userID) {
		http.Error(w, "You are not a participant in this room", http.StatusForbidden)
		return
	}

	// Verify it's the user's turn
	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		http.Error(w, "It's not your turn", http.StatusBadRequest)
		return
	}

	// Parse form data (handles both URL-encoded and multipart forms)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// If multipart parsing fails, try regular form parsing
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
	}

	questionIDStr := r.FormValue("question_id")
	log.Printf("🔍 Received question_id from form: '%s' (length: %d)", questionIDStr, len(questionIDStr))
	answerText := r.FormValue("answer_text")
	passed := r.FormValue("passed") == "true"

	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		log.Printf("❌ Failed to parse question ID '%s': %v", questionIDStr, err)
		http.Error(w, "Invalid question ID format", http.StatusBadRequest)
		return
	}

	log.Printf("🔍 Validating question %s for room %s", questionID, roomID)

	// Verify question exists in database
	question, err := h.QuestionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		log.Printf("❌ Question %s not found in database: %v", questionID, err)
		http.Error(w, fmt.Sprintf("Question not found in database (ID: %s)", questionID), http.StatusBadRequest)
		return
	}
	log.Printf("✅ Question exists: %s", question.Text)

	// Verify question matches room's current question
	room, err = h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("❌ Failed to get room %s: %v", roomID, err)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if room.CurrentQuestionID == nil {
		log.Printf("❌ Room %s has no current question", roomID)
		http.Error(w, "No active question for this room", http.StatusBadRequest)
		return
	}

	if *room.CurrentQuestionID != questionID {
		log.Printf("❌ Question mismatch: room current=%s, submitted=%s", *room.CurrentQuestionID, questionID)
		http.Error(w, fmt.Sprintf("Question mismatch: you're answering question %s but current question is %s", questionID, *room.CurrentQuestionID), http.StatusBadRequest)
		return
	}
	log.Printf("✅ Question matches room's current question")

	// Get action type from form (either "answered" or "passed")
	actionType := r.FormValue("action_type")
	if actionType == "" {
		// Default to "answered" for backwards compatibility
		actionType = "answered"
		if passed {
			actionType = "passed"
		}
	}

	answer := &models.Answer{
		ID:         uuid.New(),
		RoomID:     roomID,
		QuestionID: questionID,
		UserID:     userID,
		AnswerText: answerText,
		ActionType: actionType,
	}

	if err := h.GameService.SubmitAnswer(ctx, answer); err != nil {
		http.Error(w, "Failed to submit answer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Answer submitted by user %s in room %s (action: %s)", userID, roomID, actionType)

	// Clear the current question ID after answer submission
	// This allows the next player to draw a fresh question
	room, err = h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("⚠️ Failed to fetch room for clearing question: %v", err)
	} else {
		room.CurrentQuestionID = nil
		if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
			log.Printf("⚠️ Failed to clear current_question_id: %v", err)
			// Don't fail the request - answer is already saved
		} else {
			log.Printf("✅ Cleared current_question_id for room %s", roomID)
		}
	}

	// Change turn immediately after answer submission
	// The new active player will draw the next question when they click "Next Question"
	if actionType == "answered" {
		log.Printf("🔄 Switching turn after answer in room %s", roomID)
		if err := h.GameService.ChangeTurn(ctx, roomID); err != nil {
			log.Printf("❌ Failed to change turn: %v", err)
			http.Error(w, "Failed to change turn: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Turn switched. New active player will draw next question.")
	}
	// Note: For "passed" action, turn doesn't change but question drawing is deferred to next question click

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Answer submitted"}`))
}

// NextQuestionAPIHandler draws the next question (called by new active player after seeing answer)
func (h *Handler) NextQuestionAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify user is a participant
	if room.OwnerID != userID && (room.GuestID == nil || *room.GuestID != userID) {
		http.Error(w, "You are not a participant in this room", http.StatusForbidden)
		return
	}

	// Verify it IS the user's turn (only the active player can draw next question)
	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		http.Error(w, "It's not your turn to draw the next question", http.StatusBadRequest)
		return
	}

	// Draw a new question for the active player
	log.Printf("🎴 Drawing next question for active player %s in room %s", userID, roomID)
	question, err := h.GameService.DrawQuestion(ctx, roomID)
	if err != nil {
		log.Printf("❌ Failed to draw question: %v", err)
		http.Error(w, "Failed to draw question: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Next question drawn: %s", question.Text)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  "Next question drawn",
		"question": question,
	})
}

// FinishGameAPIHandler ends the game
func (h *Handler) FinishGameAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify user is a participant
	if room.OwnerID != userID && (room.GuestID == nil || *room.GuestID != userID) {
		http.Error(w, "You are not a participant in this room", http.StatusForbidden)
		return
	}

	// End the game
	if err := h.GameService.EndGame(ctx, roomID); err != nil {
		http.Error(w, "Failed to end game: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Game ended","redirect":"/game/finished/` + roomID.String() + `"}`))
}

// UpdateCategoriesAPIHandler updates selected categories for a room
func (h *Handler) UpdateCategoriesAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get the room
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Verify user is a participant
	if room.OwnerID != userID && (room.GuestID == nil || *room.GuestID != userID) {
		http.Error(w, "You are not a participant in this room", http.StatusForbidden)
		return
	}

	// Parse category IDs from request (multipart form data from FormData)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	categoryIDStrings := r.MultipartForm.Value["category_ids"]
	categoryIDs := make([]uuid.UUID, 0, len(categoryIDStrings))
	for _, idStr := range categoryIDStrings {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue // Skip invalid UUIDs
		}
		categoryIDs = append(categoryIDs, id)
	}

	// Update room with selected categories
	room.SelectedCategories = categoryIDs
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		http.Error(w, "Failed to update categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast category update to all room participants via SSE
	h.RoomService.BroadcastCategoriesUpdated(roomID, categoryIDs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Categories updated"}`))
}

// GetCategoriesAPIHandler returns all available categories
func (h *Handler) GetCategoriesAPIHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	categories, err := h.QuestionService.GetCategories(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	// Manual JSON encoding for simplicity
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonStr := `{"categories":[`
	for i, cat := range categories {
		if i > 0 {
			jsonStr += ","
		}
		jsonStr += `{"id":"` + cat.ID.String() + `","key":"` + cat.Key + `"}`
	}
	jsonStr += `]}`

	w.Write([]byte(jsonStr))
}

// GetRoomCategoriesHTMLHandler returns room categories as HTML fragment (for HTMX)
func (h *Handler) GetRoomCategoriesHTMLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get room to check guest_ready status
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Get all categories
	allCategories, err := h.QuestionService.GetCategories(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load categories</p>`))
		return
	}

	// Build CategoryInfo array with selection state
	categoryInfos := make([]services.CategoryInfo, 0, len(allCategories))
	for _, cat := range allCategories {
		isSelected := false
		// Check if this category is in room's selected categories
		for _, selectedID := range room.SelectedCategories {
			if selectedID == cat.ID {
				isSelected = true
				break
			}
		}

		categoryInfos = append(categoryInfos, services.CategoryInfo{
			ID:         cat.ID.String(),
			Key:        cat.Key,
			Label:      cat.Label,
			IsSelected: isSelected,
		})
	}

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("categories_grid.html", services.CategoriesGridData{
		Categories: categoryInfos,
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	})
	if err != nil {
		log.Printf("Error rendering categories grid template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load categories</p>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// ToggleCategoryAPIHandler toggles a single category selection (for HTMX)
func (h *Handler) ToggleCategoryAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get category ID from form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryIDStr := r.FormValue("category_id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Get room to get current selected categories
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Toggle category - if exists, remove it; if not, add it
	selectedCategories := room.SelectedCategories
	found := false
	newCategories := make([]uuid.UUID, 0, len(selectedCategories))
	for _, id := range selectedCategories {
		if id == categoryID {
			found = true
			// Don't add it (remove it)
		} else {
			newCategories = append(newCategories, id)
		}
	}

	if !found {
		// Add it
		newCategories = append(newCategories, categoryID)
	}

	// Update room
	room.SelectedCategories = newCategories
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("Failed to update categories: %v", err)
		http.Error(w, "Failed to update categories", http.StatusInternalServerError)
		return
	}

	// Broadcast categories update via SSE
	h.RoomService.BroadcastCategoriesUpdated(roomID, newCategories)

	// Return success (HTMX will handle via hx-swap="none")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!-- Category toggled successfully -->`))
}

// GetStartGameButtonHTMLHandler returns start game button as HTML fragment (for HTMX)
func (h *Handler) GetStartGameButtonHTMLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get room to check guest_ready status
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("start_game_button.html", services.StartGameButtonData{
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	})
	if err != nil {
		log.Printf("Error rendering start game button template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load start button</p>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// GetGuestReadyButtonHTMLHandler returns guest ready button as HTML fragment (for HTMX)
func (h *Handler) GetGuestReadyButtonHTMLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get room to check guest_ready status
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("guest_ready_button.html", services.GuestReadyButtonData{
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	})
	if err != nil {
		log.Printf("Error rendering guest ready button template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load ready button</p>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

