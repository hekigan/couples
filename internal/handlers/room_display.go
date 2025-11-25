package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
)

// EmptyRoomsStateHandler serves the empty rooms state partial
func (h *Handler) EmptyRoomsStateHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the partial template
	partialPath := "templates/partials/game/empty-rooms-state.html"
	tmpl, err := template.ParseFiles(partialPath)
	if err != nil {
		log.Printf("Error parsing empty rooms state partial: %v", err)
		http.Error(w, "Failed to load empty state", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Execute the defined template by name
	err = tmpl.ExecuteTemplate(w, "partials/game/empty-rooms-state.html", nil)
	if err != nil {
		log.Printf("Error executing empty rooms state template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// ListRoomsHandler lists all rooms for the current user
func (h *Handler) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	log.Printf("DEBUG ListRooms: Fetching rooms for user %s", userID)

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

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
		User:  currentUser,
		Data:  enrichedRooms,
	}
	h.RenderTemplate(w, "game/rooms.html", data)
}

// RoomHandler displays the room lobby
func (h *Handler) RoomHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to extract room ID
	roomID, err := ExtractIDFromRouteVars(r, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
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

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

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
		User:              currentUser,
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
	// Use helper to extract room ID
	roomID, err := ExtractIDFromRouteVars(r, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Debug logging
	fmt.Printf("🎮 PlayHandler: Room %s, Status: %s, CurrentQuestion: %d, CurrentQuestionID: %v\n",
		roomID, room.Status, room.CurrentQuestion, room.CurrentQuestionID)

	// Determine if it's the current user's turn
	isMyTurn := room.CurrentTurn != nil && *room.CurrentTurn == userID

	// Get other player name
	var otherPlayerID uuid.UUID
	if room.OwnerID == userID && room.GuestID != nil {
		otherPlayerID = *room.GuestID
	} else if room.GuestID != nil && *room.GuestID == userID {
		otherPlayerID = room.OwnerID
	}

	otherPlayerName := "other player"
	if otherPlayerID != uuid.Nil {
		otherUser, err := h.UserService.GetUserByID(ctx, otherPlayerID)
		if err == nil && otherUser != nil {
			otherPlayerName = otherUser.Username
		}
	}

	// Get current question text if exists
	questionText := "Waiting for question..."
	questionID := ""
	if room.CurrentQuestionID != nil {
		questionID = room.CurrentQuestionID.String()
		question, err := h.QuestionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err == nil && question != nil {
			questionText = question.Text
		}
	}

	// Check if there's an answer for the current question (for SSR)
	var lastAnswer *AnswerWithDetails
	hasAnswer := false
	if room.CurrentQuestionID != nil {
		answer, err := h.AnswerService.GetLastAnswerForQuestion(ctx, roomID, *room.CurrentQuestionID)
		if err == nil && answer != nil {
			hasAnswer = true
			// Get the username of the player who answered
			answeredUser, err := h.UserService.GetUserByID(ctx, answer.UserID)
			answeredPlayerName := "Unknown Player"
			if err == nil && answeredUser != nil {
				answeredPlayerName = answeredUser.Username
			}

			// Get the question for the answer
			question, _ := h.QuestionService.GetQuestionByID(ctx, answer.QuestionID)

			lastAnswer = &AnswerWithDetails{
				Answer:     answer,
				Question:   question,
				Username:   answeredPlayerName,
				ActionType: answer.ActionType,
			}
		}
	}

	// Create a map to pass all play data for server-side rendering
	playData := map[string]interface{}{
		"Room":                 room,
		"CurrentUserID":        userID.String(),
		"IsMyTurn":             isMyTurn,
		"OtherPlayerName":      otherPlayerName,
		"QuestionText":         questionText,
		"QuestionID":           questionID,
		"HasAnswer":            hasAnswer,
		"AnswerText":           "",
		"ActionType":           "",
		"AnsweredByPlayerName": "",
	}

	// If there's an answer, include its details
	if lastAnswer != nil {
		playData["AnswerText"] = lastAnswer.Answer.AnswerText
		playData["ActionType"] = lastAnswer.ActionType
		playData["AnsweredByPlayerName"] = lastAnswer.Username
	}

	data := &TemplateData{
		Title: "Play - " + room.Name,
		User:  currentUser,
		Data:  playData,
	}
	h.RenderTemplate(w, "game/play.html", data)
}

// GameFinishedHandler shows game results with full history
func (h *Handler) GameFinishedHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to extract room ID
	roomID, err := ExtractIDFromRouteVars(r, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(r)
	if err != nil {
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
	}

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
		User:  currentUser,
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
