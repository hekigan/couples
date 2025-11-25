package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
)

// StartGameAPIHandler starts the game for a room
func (h *Handler) StartGameAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify ownership
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

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

	ctx := context.Background()
	if err := h.GameService.StartGame(ctx, roomID); err != nil {
		http.Error(w, "Failed to start game: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return HTMX redirect header to navigate to play page
	w.Header().Set("HX-Redirect", fmt.Sprintf("/game/play/%s", roomID))
	w.WriteHeader(http.StatusOK)
}

// DrawQuestionAPIHandler draws a random question for the room
func (h *Handler) DrawQuestionAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
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

	// Use helper to write JSON response
	err = WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"question": question,
	})
	if err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

// SubmitAnswerAPIHandler handles answer submission
// FIXED: Previously called GetGameFormsHandler (handler-to-handler call), now renders directly
func (h *Handler) SubmitAnswerAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
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

	// NOTE: We do NOT clear current_question_id here anymore.
	// The question remains visible so both players can see the question + answer together.
	// current_question_id will be cleared when the new active player clicks "Next Question"

	// Change turn immediately after answer submission
	// The new active player will see the answer and click "Next Question" to draw
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

	// FIXED: Instead of calling GetGameFormsHandler (handler-to-handler call),
	// render the answer review HTML fragment directly
	// The submitter sees the answer they just submitted, with no "Next Question" button
	// since they're no longer the active player

	// Get the username of the player who answered (current user)
	currentUser, err := h.FetchCurrentUser(r)
	if err != nil {
		log.Printf("❌ Failed to fetch current user: %v", err)
		http.Error(w, "Failed to load user information", http.StatusInternalServerError)
		return
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
	html, err := h.TemplateService.RenderFragment("answer_review.html", services.AnswerReviewData{
		RoomID:               roomID.String(),
		AnswerText:           answerText,
		ActionType:           actionType,
		ShowNextButton:       false, // Submitter is no longer the active player, so no "Next Question" button
		AnsweredByPlayerName: currentUser.Username,
		OtherPlayerName:      otherPlayerName,
	})
	if err != nil {
		log.Printf("❌ Error rendering answer_review template: %v", err)
		http.Error(w, "Failed to render answer review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// NextQuestionAPIHandler draws the next question (called by new active player after seeing answer)
func (h *Handler) NextQuestionAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
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

	// Use helper to write JSON response
	err = WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  "Next question drawn",
		"question": question,
	})
	if err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

// FinishGameAPIHandler ends the game
func (h *Handler) FinishGameAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	ctx := context.Background()
	if err := h.GameService.EndGame(ctx, roomID); err != nil {
		http.Error(w, "Failed to end game: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Game ended","redirect":"/game/finished/` + roomID.String() + `"}`))
}

// PlayerTypingAPIHandler broadcasts typing status to other players in the room
func (h *Handler) PlayerTypingAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Parse request body
	var req struct {
		IsTyping bool `json:"is_typing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
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
