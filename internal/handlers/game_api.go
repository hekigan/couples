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
	playFragments "github.com/hekigan/couples/internal/views/fragments/play"
	"github.com/labstack/echo/v4"
)

// StartGameAPIHandler starts the game for a room
func (h *Handler) StartGameAPIHandler(c echo.Context) error {
	// Use helper to get room and verify ownership
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Only the owner can start the game
	if room.OwnerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only the room owner can start the game")
	}

	// Verify room is ready (both players present)
	if room.Status != "ready" {
		return echo.NewHTTPError(http.StatusBadRequest, "Room is not ready to start. Wait for both players to join.")
	}

	// Verify guest is ready
	if !room.GuestReady {
		return echo.NewHTTPError(http.StatusBadRequest, "Guest is not ready yet. Wait for guest to click Ready button.")
	}

	ctx := context.Background()
	if err := h.GameService.StartGame(ctx, roomID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start game: "+err.Error())
	}

	// Return HTMX redirect header to navigate to play page
	c.Response().Header().Set("HX-Redirect", fmt.Sprintf("/game/play/%s", roomID))
	return c.NoContent(http.StatusOK)
}

// DrawQuestionAPIHandler draws a random question for the room
func (h *Handler) DrawQuestionAPIHandler(c echo.Context) error {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	// Verify game is playing
	if room.Status != "playing" {
		return echo.NewHTTPError(http.StatusBadRequest, "Game is not in progress")
	}

	// Check if a question already exists (prevent race condition)
	var question *models.Question
	if room.CurrentQuestionID != nil {
		// Question already exists, fetch it instead of drawing a new one
		log.Printf("⚠️ Question already exists for room %s, returning existing question %s", roomID, *room.CurrentQuestionID)
		existingQuestion, err := h.QuestionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err != nil {
			log.Printf("❌ Failed to fetch existing question: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch existing question: "+err.Error())
		}
		question = existingQuestion
	} else {
		// No question exists, draw a new one
		log.Printf("🎴 Drawing new question for room %s", roomID)
		newQuestion, err := h.GameService.DrawQuestion(ctx, roomID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to draw question: "+err.Error())
		}
		question = newQuestion
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "success",
		"question": question,
	})
}

// SubmitAnswerAPIHandler handles answer submission
// FIXED: Previously called GetGameFormsHandler (handler-to-handler call), now renders directly
func (h *Handler) SubmitAnswerAPIHandler(c echo.Context) error {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	// Verify it's the user's turn
	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		return echo.NewHTTPError(http.StatusBadRequest, "It's not your turn")
	}

	// Parse form data (handles both URL-encoded and multipart forms)
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil {
		// If multipart parsing fails, try regular form parsing
		if err := c.Request().ParseForm(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse form")
		}
	}

	questionIDStr := c.FormValue("question_id")
	log.Printf("🔍 Received question_id from form: '%s' (length: %d)", questionIDStr, len(questionIDStr))
	answerText := c.FormValue("answer_text")
	skipped := c.FormValue("skipped") == "true"

	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		log.Printf("❌ Failed to parse question ID '%s': %v", questionIDStr, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid question ID format")
	}

	log.Printf("🔍 Validating question %s for room %s", questionID, roomID)

	// Verify question exists in database
	question, err := h.QuestionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		log.Printf("❌ Question %s not found in database: %v", questionID, err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Question not found in database (ID: %s)", questionID))
	}
	log.Printf("✅ Question exists: %s", question.Text)

	// Verify question matches room's current question
	room, err = h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("❌ Failed to get room %s: %v", roomID, err)
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
	}

	if room.CurrentQuestionID == nil {
		log.Printf("❌ Room %s has no current question", roomID)
		return echo.NewHTTPError(http.StatusBadRequest, "No active question for this room")
	}

	if *room.CurrentQuestionID != questionID {
		log.Printf("❌ Question mismatch: room current=%s, submitted=%s", *room.CurrentQuestionID, questionID)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Question mismatch: you're answering question %s but current question is %s", questionID, *room.CurrentQuestionID))
	}
	log.Printf("✅ Question matches room's current question")

	// Get action type from form (either "answered" or "skipped")
	actionType := c.FormValue("action_type")
	if actionType == "" {
		// Default to "answered" for backwards compatibility
		actionType = "answered"
		if skipped {
			actionType = "skipped"
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit answer: "+err.Error())
	}

	log.Printf("✅ Answer submitted by user %s in room %s (action: %s)", userID, roomID, actionType)

	// NOTE: We do NOT clear current_question_id here anymore.
	// The question remains visible so both players can see the question + answer together.
	// current_question_id will be cleared when the new active player clicks "Next Question"

	// Change turn immediately after answer submission
	// The new active player will see the answer and click "Next Question" to draw
	if actionType == "answered" {
		log.Printf("🔄 Switching turn after answer in room %s", roomID)
		if err := h.GameService.ChangeTurn(ctx, roomID); err != nil {
			log.Printf("❌ Failed to change turn: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to change turn: "+err.Error())
		}
		log.Printf("✅ Turn switched. New active player will draw next question.")
	}
	// Note: For "skipped" action, turn doesn't change but question drawing is deferred to next question click

	// FIXED: Instead of calling GetGameFormsHandler (handler-to-handler call),
	// render the answer review HTML fragment directly
	// The submitter sees the answer they just submitted, with no "Next Question" button
	// since they're no longer the active player

	// Get the username of the player who answered (current user)
	currentUser, err := h.FetchCurrentUser(c)
	if err != nil {
		log.Printf("❌ Failed to fetch current user: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load user information")
	}

	// Determine other player name for waiting message
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

	// Render answer review fragment
	html, err := h.RenderTemplFragment(c, playFragments.AnswerReview(&services.AnswerReviewData{
		RoomID:               roomID.String(),
		AnswerText:           answerText,
		ActionType:           actionType,
		ShowNextButton:       false, // Submitter is no longer the active player, so no "Next Question" button
		AnsweredByPlayerName: currentUser.Username,
		OtherPlayerName:      otherPlayerName,
	}))
	if err != nil {
		log.Printf("❌ Error rendering answer_review template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render answer review")
	}

	return c.HTML(http.StatusOK, html)
}

// NextQuestionAPIHandler draws the next question (called by new active player after seeing answer)
func (h *Handler) NextQuestionAPIHandler(c echo.Context) error {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	// Verify it IS the user's turn (only the active player can draw next question)
	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		return echo.NewHTTPError(http.StatusBadRequest, "It's not your turn to draw the next question")
	}

	// Draw a new question for the active player
	log.Printf("🎴 Drawing next question for active player %s in room %s", userID, roomID)
	question, err := h.GameService.DrawQuestion(ctx, roomID)
	if err != nil {
		log.Printf("❌ Failed to draw question: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to draw question: "+err.Error())
	}

	log.Printf("✅ Next question drawn: %s", question.Text)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  "Next question drawn",
		"question": question,
	})
}

// FinishGameAPIHandler ends the game
func (h *Handler) FinishGameAPIHandler(c echo.Context) error {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	ctx := context.Background()
	if err := h.GameService.EndGame(ctx, roomID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to end game: "+err.Error())
	}

	// Use HTMX redirect header for client-side redirect
	c.Response().Header().Set("HX-Redirect", "/game/finished/"+roomID.String())
	return c.NoContent(http.StatusOK)
}

// PlayerTypingAPIHandler broadcasts typing status to other players in the room
func (h *Handler) PlayerTypingAPIHandler(c echo.Context) error {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Parse request body
	var req struct {
		IsTyping bool `json:"is_typing"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	// Verify it's the user's turn
	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		// Silently ignore typing events if it's not their turn
		return c.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}

	// Broadcast typing status to other players in the room
	h.RoomService.BroadcastPlayerTyping(roomID, userID, req.IsTyping)

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}
