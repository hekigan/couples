package services

import (
	"context"
	

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

// StartGame starts a game for a room
func (s *GameService) StartGame(ctx context.Context, roomID uuid.UUID) error {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}

	room.Status = "playing"
	room.CurrentQuestion = 0
	room.CurrentTurn = &room.OwnerID

	if err := s.roomService.UpdateRoom(ctx, room); err != nil {
		return err
	}

	s.realtimeService.BroadcastGameStarted(roomID)
	return nil
}

// DrawQuestion draws a new question for the room
func (s *GameService) DrawQuestion(ctx context.Context, roomID uuid.UUID) (*models.Question, error) {
	room, err := s.roomService.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	question, err := s.questionService.GetRandomQuestion(ctx, roomID, room.Language)
	if err != nil {
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

