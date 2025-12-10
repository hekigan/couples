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

// GetTurnIndicatorHandler returns HTML fragment for turn indicator
func (h *Handler) GetTurnIndicatorHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Get room to check current turn
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		return c.HTML(http.StatusOK, `<div class="turn-indicator"><span>Loading...</span></div>`)
	}

	// Determine other player name
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

	isMyTurn := room.CurrentTurn != nil && *room.CurrentTurn == userID

	// Render templ component
	html, err := h.RenderTemplFragment(c, playFragments.TurnIndicator(&services.TurnIndicatorData{
		IsMyTurn:        isMyTurn,
		OtherPlayerName: otherPlayerName,
	}))
	if err != nil {
		log.Printf("Error rendering turn_indicator template: %v", err)
		return c.HTML(http.StatusOK, `<div class="turn-indicator"><span>Loading...</span></div>`)
	}

	return c.HTML(http.StatusOK, html)
}

// GetQuestionCardHandler returns HTML fragment for current question
func (h *Handler) GetQuestionCardHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()

	// Get room to check current question
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		return c.HTML(http.StatusOK, `<div class="question-card"><p class="question-text">Waiting to start...</p></div>`)
	}

	questionText := "Waiting for question..."
	if room.CurrentQuestionID != nil {
		// Get the question text
		question, err := h.QuestionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err == nil && question != nil {
			questionText = question.Text
		}
	}

	// Render templ component
	html, err := h.RenderTemplFragment(c, playFragments.QuestionCard(&services.QuestionCardData{
		QuestionText: questionText,
	}))
	if err != nil {
		log.Printf("Error rendering question_card template: %v", err)
		return c.HTML(http.StatusOK, `<div class="question-card"><p class="question-text">Loading...</p></div>`)
	}

	return c.HTML(http.StatusOK, html)
}

// GetGameContentHandler returns HTML fragment for the full game content area
// This includes turn indicator, question card, and game forms
func (h *Handler) GetGameContentHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	// Render templ component for full game content
	html, err := h.RenderTemplFragment(c, playFragments.GameContent(roomID.String()))
	if err != nil {
		log.Printf("Error rendering game_content template: %v", err)
		return c.HTML(http.StatusOK, `<div class="loading">Loading game interface...</div>`)
	}

	return c.HTML(http.StatusOK, html)
}

// GetGameFormsHandler returns HTML fragment for answer form, waiting UI, or answer review
func (h *Handler) GetGameFormsHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	log.Printf("🎮 GetGameFormsHandler called for room %s by user %s", roomID, userID)

	// Get room to check current turn and state
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		return c.HTML(http.StatusOK, `<div class="loading">Loading game interface...</div>`)
	}

	isMyTurn := room.CurrentTurn != nil && *room.CurrentTurn == userID

	// Determine other player name
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

	var html string

	// Check if there's an answer for the current question
	// This determines if we show answer review vs answer form
	var lastAnswer *models.Answer
	if room.CurrentQuestionID != nil {
		log.Printf("🔍 Checking for answer to question %s in room %s", room.CurrentQuestionID, roomID)
		lastAnswer, _ = h.AnswerService.GetLastAnswerForQuestion(ctx, roomID, *room.CurrentQuestionID)
		if lastAnswer != nil {
			log.Printf("✅ Found answer: %s (action: %s)", lastAnswer.AnswerText, lastAnswer.ActionType)
		} else {
			log.Printf("❌ No answer found for question %s", room.CurrentQuestionID)
		}
	} else {
		log.Printf("⚠️ Room has no current question ID")
	}

	// Determine what UI to show based on turn and answer state
	if lastAnswer != nil {
		// An answer exists for the current question - show answer review to both players
		// The new active player sees "Next Question" button
		// The previous active player sees waiting message
		log.Printf("📝 Answer found for question %s: %s (action: %s)", room.CurrentQuestionID, lastAnswer.AnswerText, lastAnswer.ActionType)

		// Get the username of the player who answered
		answeredPlayerName := "Unknown Player"
		answeredUser, err := h.UserService.GetUserByID(ctx, lastAnswer.UserID)
		if err == nil && answeredUser != nil {
			answeredPlayerName = answeredUser.Username
		}

		html, err = h.RenderTemplFragment(c, playFragments.AnswerReview(&services.AnswerReviewData{
			RoomID:               roomID.String(),
			AnswerText:           lastAnswer.AnswerText,
			ActionType:           lastAnswer.ActionType,
			ShowNextButton:       isMyTurn, // New active player can draw next question
			AnsweredByPlayerName: answeredPlayerName,
			OtherPlayerName:      otherPlayerName,
		}))
		if err != nil {
			log.Printf("Error rendering answer_review template: %v", err)
			return c.HTML(http.StatusOK, `<div class="loading">Loading answer...</div>`)
		}
	} else if isMyTurn {
		// No answer yet and it's my turn - show answer form
		questionID := ""
		if room.CurrentQuestionID != nil {
			questionID = room.CurrentQuestionID.String()
		}

		html, err = h.RenderTemplFragment(c, playFragments.AnswerForm(&services.AnswerFormData{
			RoomID:     roomID.String(),
			QuestionID: questionID,
		}))
		if err != nil {
			log.Printf("Error rendering answer_form template: %v", err)
			return c.HTML(http.StatusOK, `<div class="loading">Loading form...</div>`)
		}
	} else {
		// No answer yet and it's not my turn - show waiting UI
		html, err = h.RenderTemplFragment(c, playFragments.WaitingUI(&services.WaitingUIData{
			OtherPlayerName: otherPlayerName,
		}))
		if err != nil {
			log.Printf("Error rendering waiting_ui template: %v", err)
			return c.HTML(http.StatusOK, `<div class="loading">Loading...</div>`)
		}
	}

	return c.HTML(http.StatusOK, html)
}

// GetProgressCounterHandler returns HTML fragment for progress counter
func (h *Handler) GetProgressCounterHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()

	// Get room to check current question count
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		return c.HTML(http.StatusOK, `Question 0 of 0`)
	}

	// Render templ component
	html, err := h.RenderTemplFragment(c, playFragments.ProgressCounter(&services.ProgressCounterData{
		CurrentQuestion: room.CurrentQuestion,
		MaxQuestions:    room.MaxQuestions,
	}))
	if err != nil {
		log.Printf("Error rendering progress_counter template: %v", err)
		return c.HTML(http.StatusOK, fmt.Sprintf(`Question %d of %d`, room.CurrentQuestion, room.MaxQuestions))
	}

	return c.HTML(http.StatusOK, html)
}

// NextQuestionHTMLHandler handles drawing the next question and returns game content
func (h *Handler) NextQuestionHTMLHandler(c echo.Context) error {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid room ID")
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Verify it's the user's turn
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Room not found")
	}

	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		return echo.NewHTTPError(http.StatusForbidden, "Not your turn")
	}

	// Clear the current question ID before drawing a new one
	// This is necessary because DrawQuestion has an idempotency check
	// that returns the existing question if CurrentQuestionID is set
	room.CurrentQuestionID = nil
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("Error clearing current question: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clear current question")
	}
	log.Printf("✅ Cleared current_question_id for room %s before drawing next question", roomID)

	// Draw next question (this also broadcasts question_drawn via SSE)
	_, err = h.GameService.DrawQuestion(ctx, roomID)
	if err != nil {
		log.Printf("Error drawing question: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to draw question: %v", err))
	}

	// Return updated game forms (will show answer form to active player)
	return h.GetGameFormsHandler(c)
}
