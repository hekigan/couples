package services

import (
	"context"
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
}

// NewGameService creates a new game service
func NewGameService(
	client *supabase.Client,
	roomService *RoomService,
	questionService *QuestionService,
	answerService *AnswerService,
	realtimeService *RealtimeService,
) *GameService {
	return &GameService{
		client:          client,
		roomService:     roomService,
		questionService: questionService,
		answerService:   answerService,
		realtimeService: realtimeService,
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

	s.realtimeService.BroadcastGameStarted(roomID)
	return nil
}

// DrawQuestion draws a new question for the room filtered by selected categories
func (s *GameService) DrawQuestion(ctx context.Context, roomID uuid.UUID) (*models.Question, error) {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Get random question filtered by categories and history
	question, err := s.questionService.GetRandomQuestion(ctx, roomID, room.Language, room.SelectedCategories)
	if err != nil {
		return nil, err
	}

	// Mark question as asked to prevent repetition
	if err := s.questionService.MarkQuestionAsked(ctx, roomID, question.ID); err != nil {
		return nil, err
	}

	// Increment current question counter
	room.CurrentQuestion++
	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return nil, err
	}

	s.realtimeService.BroadcastQuestionDrawn(roomID, question)
	return question, nil
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

