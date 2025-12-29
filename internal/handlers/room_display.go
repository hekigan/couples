package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	gameFragments "github.com/hekigan/couples/internal/views/fragments/game"
	gamePages "github.com/hekigan/couples/internal/views/pages/game"
	"github.com/labstack/echo/v4"
)

// EmptyRoomsStateHandler serves the empty rooms state partial
func (h *Handler) EmptyRoomsStateHandler(c echo.Context) error {
	return h.RenderTemplComponent(c, gameFragments.EmptyRoomsState())
}

// ListRoomsHandler lists all rooms for the current user
func (h *Handler) ListRoomsHandler(c echo.Context) error {
	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	log.Printf("DEBUG ListRooms: Fetching rooms for user %s", userID)

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	// Use database view to get rooms with player info
	// This eliminates N+1 queries (2 room queries + N user queries → 2 room queries with player info)
	roomsWithPlayers, err := h.RoomService.GetRoomsByUserIDWithPlayers(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load rooms")
	}

	log.Printf("DEBUG ListRooms: Found %d rooms (via view)", len(roomsWithPlayers))

	// Convert to RoomWithUsername format for template
	enrichedRooms := make([]services.RoomWithUsername, 0, len(roomsWithPlayers))
	for _, roomWithPlayers := range roomsWithPlayers {
		log.Printf("DEBUG ListRooms: Room %s -  RoomName %s - OwnerID: %s, GuestID: %v, Status: %s",
			roomWithPlayers.ID, roomWithPlayers.Name, roomWithPlayers.OwnerID, roomWithPlayers.GuestID, roomWithPlayers.Status)

		isOwner := roomWithPlayers.OwnerID == userID
		enrichedRoom := services.RoomWithUsername{
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

	data := NewTemplateData(c)
	data.Title = "My Rooms"
	data.User = currentUser
	data.Data = enrichedRooms
	return h.RenderTemplComponent(c, gamePages.RoomsPage(data))
}

// RoomHandler displays the room lobby
func (h *Handler) RoomHandler(c echo.Context) error {
	// Use helper to extract room ID
	roomID, err := ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := context.Background()
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use database view to get room with player info in single query
	// This replaces 3 queries (room + owner + guest) with 1 query
	roomWithPlayers, err := h.RoomService.GetRoomWithPlayers(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
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
	currentUser, err := h.FetchCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	// Render fragments server-side (eliminates hx-trigger="load")
	categoriesHTML, friendsHTML, actionButtonHTML, err := h.RenderRoomFragments(c, &roomWithPlayers.Room, roomID, currentUserID, isOwner)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to render some fragments: %v", err)
		// Continue anyway - fragments will have fallback error messages
	}

	// Get pending join requests and render them (owner only)
	joinRequestsHTML := ""
	joinRequestsCount := 0
	if isOwner {
		joinRequestsHTML, joinRequestsCount, err = h.renderJoinRequests(c, ctx, roomID)
		if err != nil {
			log.Printf("⚠️ Failed to render join requests: %v", err)
			// Continue with empty state
			joinRequestsHTML = ""
			joinRequestsCount = 0
		}
	}

	data := NewTemplateData(c)
	data.Title = "Room - " + roomWithPlayers.Name
	data.User = currentUser
	// Pass room as map for step-based template
	data.Data = map[string]interface{}{
		"room": &roomWithPlayers.Room,
	}
	data.OwnerUsername = ownerUsername
	data.GuestUsername = guestUsername
	data.IsOwner = isOwner
	data.JoinRequestsCount = joinRequestsCount
	// Add pre-rendered fragments
	data.CategoriesGridHTML = categoriesHTML
	data.FriendsListHTML = friendsHTML
	data.ActionButtonHTML = actionButtonHTML
	data.JoinRequestsHTML = joinRequestsHTML

	// HTMX refactoring complete - using HTMX version as default
	return h.RenderTemplComponent(c, gamePages.RoomPage(data))
}

// PlayHandler handles the game play screen
func (h *Handler) PlayHandler(c echo.Context) error {
	// Use helper to extract room ID
	roomID, err := ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
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

	// Create PlayPageData for type-safe rendering
	playData := &services.PlayPageData{
		Room:                 room,
		CurrentUserID:        userID.String(),
		IsMyTurn:             isMyTurn,
		OtherPlayerName:      otherPlayerName,
		QuestionText:         questionText,
		QuestionID:           questionID,
		HasAnswer:            hasAnswer,
		AnswerText:           "",
		ActionType:           "",
		AnsweredByPlayerName: "",
	}

	// If there's an answer, include its details
	if lastAnswer != nil {
		playData.AnswerText = lastAnswer.Answer.AnswerText
		playData.ActionType = lastAnswer.ActionType
		playData.AnsweredByPlayerName = lastAnswer.Username
	}

	data := NewTemplateData(c)
	data.Title = "Play - " + room.Name
	data.User = currentUser
	data.Data = playData
	return h.RenderTemplComponent(c, gamePages.PlayPage(data))
}

// GameFinishedHandler shows game results with full history
func (h *Handler) GameFinishedHandler(c echo.Context) error {
	// Use helper to extract room ID
	roomID, err := ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := context.Background()

	// Use helper to fetch current user
	currentUser, err := h.FetchCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
	}

	// Get all answers for this room
	answers, err := h.AnswerService.GetAnswersByRoom(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load answers")
	}

	// Get owner and guest info
	owner, _ := h.UserService.GetUserByID(ctx, room.OwnerID)
	var guest *models.User
	if room.GuestID != nil {
		guest, _ = h.UserService.GetUserByID(ctx, *room.GuestID)
	}

	// Enrich answers with question and user details
	answerDetails := make([]services.AnswerWithDetails, 0, len(answers))
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

		answerDetails = append(answerDetails, services.AnswerWithDetails{
			Answer:     &answer,
			Question:   question,
			Username:   username,
			ActionType: answer.ActionType,
		})
	}

	// Calculate statistics
	totalQuestions := len(answerDetails)
	skippedCount := 0
	answeredCount := 0
	for _, ad := range answerDetails {
		if ad.ActionType == "skipped" {
			skippedCount++
		} else {
			answeredCount++
		}
	}

	finishedData := &services.GameFinishedData{
		Room:           room,
		Answers:        answerDetails,
		TotalQuestions: totalQuestions,
		SkippedCount:   skippedCount,
		AnsweredCount:  answeredCount,
	}

	data := NewTemplateData(c)
	data.Title = "Game Finished"
	data.User = currentUser
	data.Data = finishedData
	return h.RenderTemplComponent(c, gamePages.FinishedPage(data))
}
