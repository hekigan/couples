package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/yourusername/couple-card-game/internal/models"
)

// GameService handles game logic
type GameService struct {
	client          *supabase.Client
	roomService     *RoomService
	questionService *QuestionService
	answerService   *AnswerService
	realtimeService *RealtimeService
	templateService *TemplateService
}

// NewGameService creates a new game service
func NewGameService(
	client *supabase.Client,
	roomService *RoomService,
	questionService *QuestionService,
	answerService *AnswerService,
	realtimeService *RealtimeService,
	templateService *TemplateService,
) *GameService {
	return &GameService{
		client:          client,
		roomService:     roomService,
		questionService: questionService,
		answerService:   answerService,
		realtimeService: realtimeService,
		templateService: templateService,
	}
}

// StartGame starts a game for a room with random first turn
func (s *GameService) StartGame(ctx context.Context, roomID uuid.UUID) error {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Randomly decide who goes first
	rand.Seed(time.Now().UnixNano())
	if room.GuestID != nil && rand.Intn(2) == 0 {
		room.CurrentTurn = room.GuestID
	} else {
		room.CurrentTurn = &room.OwnerID
	}

	room.Status = "playing"
	room.CurrentQuestion = 0

	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return err
	}

	// Broadcast game started event with HTML fragment for redirect
	// Render the game_started template which contains the redirect script
	html, err := s.templateService.RenderFragment("game_started.html", GameStartedData{
		RoomID: roomID.String(),
	})
	if err != nil {
		// Fallback to JSON event if template fails
		s.realtimeService.BroadcastGameStarted(roomID)
	} else {
		// Broadcast HTML fragment that triggers redirect on all clients
		s.realtimeService.BroadcastHTMLFragment(roomID, HTMLFragmentEvent{
			Type:       "game_started",
			Target:     "#game-start-redirect",
			SwapMethod: "innerHTML",
			HTML:       html,
		})
	}

	// Draw the first question immediately
	_, err = s.DrawQuestion(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to draw first question: %w", err)
	}

	return nil
}

// GetCurrentQuestion retrieves the current active question for a room if it exists
func (s *GameService) GetCurrentQuestion(ctx context.Context, roomID uuid.UUID) (*models.Question, error) {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// If no current question ID is set, return nil (no error)
	if room.CurrentQuestionID == nil {
		return nil, nil
	}

	// Fetch the current question
	question, err := s.questionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
	if err != nil {
		return nil, err
	}

	return question, nil
}

// DrawQuestion draws a new question for the room filtered by selected categories
func (s *GameService) DrawQuestion(ctx context.Context, roomID uuid.UUID) (*models.Question, error) {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// IDEMPOTENCY CHECK: If a question is already drawn, return it
	if room.CurrentQuestionID != nil {
		fmt.Printf("⚠️ Question already exists for room %s, returning existing question %s\n", roomID, *room.CurrentQuestionID)
		fmt.Printf("🔍 DEBUG: CurrentQuestionID type: %T, value: %v\n", *room.CurrentQuestionID, *room.CurrentQuestionID)

		existingQuestion, err := s.questionService.GetQuestionByID(ctx, *room.CurrentQuestionID)
		if err != nil {
			fmt.Printf("❌ Failed to fetch existing question: %v\n", err)
			// Clear the invalid question ID and draw a new one
			fmt.Printf("🔧 Clearing invalid CurrentQuestionID and drawing new question\n")
			room.CurrentQuestionID = nil
			// Fall through to draw a new question
		} else {
			// Return the existing question without broadcasting again
			return existingQuestion, nil
		}
	}

	// Get random question filtered by categories and history
	question, err := s.questionService.GetRandomQuestion(ctx, roomID, room.Language, room.SelectedCategories)
	if err != nil {
		return nil, err
	}

	// Mark question as asked to prevent repetition
	if err := s.questionService.MarkQuestionAsked(ctx, roomID, question.ID); err != nil {
		// Check if this is a duplicate key error (constraint violation)
		if isConstraintViolation(err) {
			fmt.Printf("⚠️ Duplicate question marking detected (race condition), continuing anyway. Room: %s, Question: %s\n", roomID, question.ID)
			// Question is already marked, this is fine - continue
		} else {
			// Real error, return it
			return nil, err
		}
	}

	// Increment current question counter and save current question ID
	room.CurrentQuestion++
	room.CurrentQuestionID = &question.ID
	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return nil, err
	}

	// Get room with players to have usernames for HTML fragment
	roomWithPlayers, err := s.roomService.GetRoomWithPlayers(ctx, roomID)
	if err != nil {
		fmt.Printf("⚠️ Failed to get room with players: %v (falling back to JSON SSE)\n", err)
		s.realtimeService.BroadcastQuestionDrawn(roomID, question)
		return question, nil
	}

	// Get category for the question
	categories, err := s.questionService.GetCategories(ctx)
	var categoryKey string
	var categoryLabel string
	if err == nil {
		for _, cat := range categories {
			if cat.ID == question.CategoryID {
				categoryKey = cat.Key
				categoryLabel = cat.Label
				break
			}
		}
	}
	if categoryKey == "" {
		categoryKey = "unknown"
		categoryLabel = "Unknown"
	}

	// Determine current player username
	var currentPlayerUsername string
	if room.CurrentTurn != nil {
		if *room.CurrentTurn == roomWithPlayers.OwnerID && roomWithPlayers.OwnerUsername != nil {
			currentPlayerUsername = *roomWithPlayers.OwnerUsername
		} else if roomWithPlayers.GuestID != nil && *room.CurrentTurn == *roomWithPlayers.GuestID && roomWithPlayers.GuestUsername != nil {
			currentPlayerUsername = *roomWithPlayers.GuestUsername
		}
	}

	// Render HTML fragment for question drawn
	if s.templateService != nil {
		html, err := s.templateService.RenderFragment("question_drawn.html", QuestionDrawnData{
			RoomID:                roomID.String(),
			QuestionNumber:        room.CurrentQuestion,
			MaxQuestions:          room.MaxQuestions,
			Category:              categoryKey,
			CategoryLabel:         categoryLabel,
			QuestionText:          question.Text,
			IsMyTurn:              false, // Will be determined client-side or per-user
			CurrentPlayerUsername: currentPlayerUsername,
		})
		if err != nil {
			fmt.Printf("⚠️ Failed to render question_drawn template: %v (falling back to JSON SSE)\n", err)
			s.realtimeService.BroadcastQuestionDrawn(roomID, question)
		} else {
			// Broadcast HTML fragment via SSE
			s.realtimeService.BroadcastHTMLFragment(roomID, HTMLFragmentEvent{
				Type:       "question_drawn",
				Target:     "#current-question",
				SwapMethod: "outerHTML", // Replace entire question card
				HTML:       html,
			})
		}
	} else {
		// No template service, fall back to JSON
		s.realtimeService.BroadcastQuestionDrawn(roomID, question)
	}

	return question, nil
}

// isConstraintViolation checks if the error is a unique constraint violation
func isConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	// Check for common unique constraint violation error messages
	errStr := err.Error()
	return contains(errStr, "23505") || contains(errStr, "duplicate key") || contains(errStr, "unique constraint")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SubmitAnswer submits an answer to a question
func (s *GameService) SubmitAnswer(ctx context.Context, answer *models.Answer) error {
	if err := s.answerService.CreateAnswer(ctx, answer); err != nil {
		return err
	}

	s.realtimeService.BroadcastAnswerSubmitted(answer.RoomID, answer)
	return nil
}

// EndGame ends the game for a room
func (s *GameService) EndGame(ctx context.Context, roomID uuid.UUID) error {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	room.Status = "finished"
	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return err
	}

	s.realtimeService.BroadcastGameFinished(roomID, map[string]interface{}{
		"room_id": roomID,
		"status":  "finished",
	})

	return nil
}

// ChangeTurn changes the current turn to the other player
func (s *GameService) ChangeTurn(ctx context.Context, roomID uuid.UUID) error {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.CurrentTurn != nil && *room.CurrentTurn == room.OwnerID && room.GuestID != nil {
		room.CurrentTurn = room.GuestID
	} else {
		room.CurrentTurn = &room.OwnerID
	}

	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return err
	}

	s.realtimeService.BroadcastTurnChanged(roomID, map[string]interface{}{
		"current_turn": room.CurrentTurn,
	})

	return nil
}

// PauseGame pauses the game due to disconnection
func (s *GameService) PauseGame(ctx context.Context, roomID uuid.UUID, disconnectedUserID uuid.UUID) error {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Only pause if game is currently playing
	if room.Status != "playing" {
		return nil
	}

	now := time.Now()
	room.Status = "paused"
	room.PausedAt = &now
	room.DisconnectedUser = &disconnectedUserID

	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return err
	}

	s.realtimeService.Broadcast(roomID, RealtimeEvent{
		Type: "game_paused",
		Data: map[string]interface{}{
			"room_id":            roomID.String(),
			"disconnected_user":  disconnectedUserID.String(),
			"paused_at":          now,
		},
	})

	return nil
}

// ResumeGame resumes a paused game after reconnection
func (s *GameService) ResumeGame(ctx context.Context, roomID uuid.UUID) error {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Only resume if game is paused
	if room.Status != "paused" {
		return nil
	}

	room.Status = "playing"
	room.PausedAt = nil
	room.DisconnectedUser = nil

	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return err
	}

	s.realtimeService.Broadcast(roomID, RealtimeEvent{
		Type: "game_resumed",
		Data: map[string]interface{}{
			"room_id": roomID.String(),
			"status":  "playing",
		},
	})

	return nil
}

// CheckReconnectionTimeout checks if a paused game has exceeded the reconnection timeout
func (s *GameService) CheckReconnectionTimeout(ctx context.Context, roomID uuid.UUID, timeoutMinutes int) (bool, error) {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return false, err
	}

	// If not paused or no pause time, no timeout
	if room.Status != "paused" || room.PausedAt == nil {
		return false, nil
	}

	// Check if paused time exceeds timeout
	elapsed := time.Since(*room.PausedAt)
	timeout := time.Duration(timeoutMinutes) * time.Minute

	if elapsed > timeout {
		// Timeout exceeded, end the game
		if err := s.EndGame(ctx, roomID); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}



