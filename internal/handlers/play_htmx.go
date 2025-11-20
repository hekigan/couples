package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/models"
	"github.com/yourusername/couple-card-game/internal/services"
)

// GetTurnIndicatorHandler returns HTML fragment for turn indicator
func (h *Handler) GetTurnIndicatorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Get room to check current turn
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="turn-indicator"><span>Loading...</span></div>`))
		return
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

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("turn_indicator.html", services.TurnIndicatorData{
		IsMyTurn:        isMyTurn,
		OtherPlayerName: otherPlayerName,
	})
	if err != nil {
		log.Printf("Error rendering turn_indicator template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="turn-indicator"><span>Loading...</span></div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// GetQuestionCardHandler returns HTML fragment for current question
func (h *Handler) GetQuestionCardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get room to check current question
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="question-card"><p class="question-text">Waiting to start...</p></div>`))
		return
	}

	questionText := "Waiting for question..."
	if room.CurrentQuestionID != nil {
		// Get the question text
		question, err := h.QuestionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err == nil && question != nil {
			questionText = question.Text
		}
	}

	// Render HTML fragment
	html, err := h.TemplateService.RenderFragment("question_card.html", services.QuestionCardData{
		QuestionText: questionText,
	})
	if err != nil {
		log.Printf("Error rendering question_card template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="question-card"><p class="question-text">Loading...</p></div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// GetGameContentHandler returns HTML fragment for the full game content area
// This includes turn indicator, question card, and game forms
func (h *Handler) GetGameContentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Render the full game content template which includes
	// turn indicator, question card, and game forms
	html, err := h.TemplateService.RenderFragment("game_content.html", services.GameContentData{
		RoomID: roomID.String(),
	})
	if err != nil {
		log.Printf("Error rendering game_content template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="loading">Loading game interface...</div>`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// GetGameFormsHandler returns HTML fragment for answer form, waiting UI, or answer review
func (h *Handler) GetGameFormsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	log.Printf("🎮 GetGameFormsHandler called for room %s by user %s", roomID, userID)

	// Get room to check current turn and state
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="loading">Loading game interface...</div>`))
		return
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

		html, err = h.TemplateService.RenderFragment("answer_review.html", services.AnswerReviewData{
			RoomID:          roomID.String(),
			AnswerText:      lastAnswer.AnswerText,
			ActionType:      lastAnswer.ActionType,
			ShowNextButton:  isMyTurn, // New active player can draw next question
			OtherPlayerName: otherPlayerName,
		})
		if err != nil {
			log.Printf("Error rendering answer_review template: %v", err)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="loading">Loading answer...</div>`))
			return
		}
	} else if isMyTurn {
		// No answer yet and it's my turn - show answer form
		questionID := ""
		if room.CurrentQuestionID != nil {
			questionID = room.CurrentQuestionID.String()
		}

		html, err = h.TemplateService.RenderFragment("answer_form.html", services.AnswerFormData{
			RoomID:     roomID.String(),
			QuestionID: questionID,
		})
		if err != nil {
			log.Printf("Error rendering answer_form template: %v", err)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="loading">Loading form...</div>`))
			return
		}
	} else {
		// No answer yet and it's not my turn - show waiting UI
		html, err = h.TemplateService.RenderFragment("waiting_ui.html", services.WaitingUIData{
			OtherPlayerName: otherPlayerName,
		})
		if err != nil {
			log.Printf("Error rendering waiting_ui template: %v", err)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="loading">Loading...</div>`))
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// NextQuestionHTMLHandler handles drawing the next question and returns game content
func (h *Handler) NextQuestionHTMLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Verify it's the user's turn
	room, err := h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if room.CurrentTurn == nil || *room.CurrentTurn != userID {
		http.Error(w, "Not your turn", http.StatusForbidden)
		return
	}

	// Clear the current question ID before drawing a new one
	// This is necessary because DrawQuestion has an idempotency check
	// that returns the existing question if CurrentQuestionID is set
	room.CurrentQuestionID = nil
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("Error clearing current question: %v", err)
		http.Error(w, "Failed to clear current question", http.StatusInternalServerError)
		return
	}
	log.Printf("✅ Cleared current_question_id for room %s before drawing next question", roomID)

	// Draw next question (this also broadcasts question_drawn via SSE)
	_, err = h.GameService.DrawQuestion(ctx, roomID)
	if err != nil {
		log.Printf("Error drawing question: %v", err)
		http.Error(w, fmt.Sprintf("Failed to draw question: %v", err), http.StatusInternalServerError)
		return
	}

	// Return updated game forms (will show answer form to active player)
	h.GetGameFormsHandler(w, r)
}
