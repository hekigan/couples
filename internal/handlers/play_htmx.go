package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
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

	if isMyTurn {
		// Active player - show answer form
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
		// Passive player - show waiting UI
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

	// Draw next question
	question, err := h.GameService.DrawQuestion(ctx, roomID)
	if err != nil {
		log.Printf("Error drawing question: %v", err)
		http.Error(w, fmt.Sprintf("Failed to draw question: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast question_drawn event via SSE
	h.RoomService.GetRealtimeService().BroadcastQuestionDrawn(roomID, question)

	// Return updated game forms (will show answer form to active player)
	h.GetGameFormsHandler(w, r)
}
