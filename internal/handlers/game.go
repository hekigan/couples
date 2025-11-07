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

	rooms, err := h.RoomService.GetRoomsByUserID(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load rooms", http.StatusInternalServerError)
		return
	}

	log.Printf("DEBUG ListRooms: Found %d rooms", len(rooms))

	// Enrich rooms with the other player's username and ownership info
	enrichedRooms := make([]RoomWithUsername, 0, len(rooms))
	for _, room := range rooms {
		log.Printf("DEBUG ListRooms: Room %s - OwnerID: %s, GuestID: %v, Status: %s",
			room.ID, room.OwnerID, room.GuestID, room.Status)

		enrichedRoom := RoomWithUsername{
			Room:    &room,
			IsOwner: room.OwnerID == userID,
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

	// Debug logging
	log.Printf("🔍 DEBUG RoomHandler: roomID=%s, currentUserID=%s, ownerID=%s, isOwner=%v, status=%s, guestReady=%v",
		room.ID, currentUserID, room.OwnerID, isOwner, room.Status, room.GuestReady)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Game started"}`))
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

	// Broadcast guest ready event via SSE
	h.RoomService.BroadcastRoomUpdate(roomID, map[string]interface{}{
		"guest_ready": true,
	})

	log.Printf("📡 Broadcasting guest_ready=true to room %s", roomID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Guest marked as ready"}`))
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

	// Handle action based on type
	if actionType == "answered" {
		// Switch turn to the other player
		log.Printf("🔄 Switching turn after answer in room %s", roomID)
		if err := h.GameService.ChangeTurn(ctx, roomID); err != nil {
			log.Printf("❌ Failed to change turn: %v", err)
			http.Error(w, "Failed to change turn: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Draw a new question for the new active player
		log.Printf("🎴 Drawing new question after turn change in room %s", roomID)
		if _, err := h.GameService.DrawQuestion(ctx, roomID); err != nil {
			log.Printf("❌ Failed to draw question: %v", err)
			http.Error(w, "Failed to draw question: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if actionType == "passed" {
		// Same player continues, just draw a new question
		log.Printf("⏭️ Passing question, drawing new card for same player in room %s", roomID)
		if _, err := h.GameService.DrawQuestion(ctx, roomID); err != nil {
			log.Printf("❌ Failed to draw question after pass: %v", err)
			http.Error(w, "Failed to draw question: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Answer submitted"}`))
}

// NextCardAPIHandler changes turn and draws the next question
func (h *Handler) NextCardAPIHandler(w http.ResponseWriter, r *http.Request) {
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

	// Verify it's NOT the user's turn (the other player clicks "Next Card")
	if room.CurrentTurn != nil && *room.CurrentTurn == userID {
		http.Error(w, "You cannot draw the next card on your own turn", http.StatusBadRequest)
		return
	}

	// Change turn to this user
	if err := h.GameService.ChangeTurn(ctx, roomID); err != nil {
		http.Error(w, "Failed to change turn: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Turn changed, ready to draw next card"}`))
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

