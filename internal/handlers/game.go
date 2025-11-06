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

	// Start the game
	if err := h.GameService.StartGame(ctx, roomID); err != nil {
		http.Error(w, "Failed to start game: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Game started"}`))
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

	// Draw a question
	question, err := h.GameService.DrawQuestion(ctx, roomID)
	if err != nil {
		http.Error(w, "Failed to draw question: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Return question as JSON (simple encoding for now)
	response := `{"status":"success","question":{"id":"` + question.ID.String() + `","text":"` + question.Text + `"}}`
	w.Write([]byte(response))
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

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	questionIDStr := r.FormValue("question_id")
	answerText := r.FormValue("answer_text")
	passed := r.FormValue("passed") == "true"

	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	// Create answer
	actionType := "answered"
	if passed {
		actionType = "passed"
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

	// Parse category IDs from request
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	categoryIDStrings := r.Form["category_ids"]
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

