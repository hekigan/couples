package services

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// RealtimeEvent represents a real-time event
type RealtimeEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// RealtimeClient represents a connected client
type RealtimeClient struct {
	ID      string
	RoomID  uuid.UUID
	UserID  uuid.UUID
	Channel chan RealtimeEvent
}

// RealtimeService manages real-time connections
type RealtimeService struct {
	clients map[string]*RealtimeClient
	mu      sync.RWMutex
}

// NewRealtimeService creates a new realtime service
func NewRealtimeService() *RealtimeService {
	return &RealtimeService{
		clients: make(map[string]*RealtimeClient),
	}
}

// Subscribe adds a new client
func (s *RealtimeService) Subscribe(roomID, userID uuid.UUID) *RealtimeClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := &RealtimeClient{
		ID:      uuid.New().String(),
		RoomID:  roomID,
		UserID:  userID,
		Channel: make(chan RealtimeEvent, 10),
	}

	s.clients[client.ID] = client
	return client
}

// Unsubscribe removes a client
func (s *RealtimeService) Unsubscribe(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, exists := s.clients[clientID]; exists {
		close(client.Channel)
		delete(s.clients, clientID)
	}
}

// Broadcast sends an event to all clients in a room
func (s *RealtimeService) Broadcast(roomID uuid.UUID, event RealtimeEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		if client.RoomID == roomID {
			select {
			case client.Channel <- event:
			default:
				// Channel full, skip this client
			}
		}
	}
}

// BroadcastRoomUpdate broadcasts a room_update event
func (s *RealtimeService) BroadcastRoomUpdate(roomID uuid.UUID, room interface{}) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "room_update",
		Data: room,
	})
}

// BroadcastPlayerJoined broadcasts a player_joined event
func (s *RealtimeService) BroadcastPlayerJoined(roomID uuid.UUID, player interface{}) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "player_joined",
		Data: player,
	})
}

// BroadcastPlayerLeft broadcasts a player_left event
func (s *RealtimeService) BroadcastPlayerLeft(roomID uuid.UUID, playerID uuid.UUID) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "player_left",
		Data: map[string]interface{}{
			"player_id": playerID,
		},
	})
}

// BroadcastGameStarted broadcasts a game_started event
func (s *RealtimeService) BroadcastGameStarted(roomID uuid.UUID) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "game_started",
		Data: map[string]interface{}{
			"room_id": roomID,
		},
	})
}

// BroadcastQuestionDrawn broadcasts a question_drawn event
func (s *RealtimeService) BroadcastQuestionDrawn(roomID uuid.UUID, question interface{}) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "question_drawn",
		Data: question,
	})
}

// BroadcastAnswerSubmitted broadcasts an answer_submitted event
func (s *RealtimeService) BroadcastAnswerSubmitted(roomID uuid.UUID, answer interface{}) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "answer_submitted",
		Data: answer,
	})
}

// BroadcastTurnChanged broadcasts a turn_changed event
func (s *RealtimeService) BroadcastTurnChanged(roomID uuid.UUID, currentTurn interface{}) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "turn_changed",
		Data: currentTurn,
	})
}

// BroadcastGameFinished broadcasts a game_finished event
func (s *RealtimeService) BroadcastGameFinished(roomID uuid.UUID, results interface{}) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "game_finished",
		Data: results,
	})
}

// BroadcastRoomDeleted broadcasts a room_deleted event
func (s *RealtimeService) BroadcastRoomDeleted(roomID uuid.UUID) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "room_deleted",
		Data: map[string]interface{}{
			"room_id": roomID,
			"message": "This room has been deleted by the owner",
		},
	})
}

// EventToSSE converts an event to Server-Sent Events format
func EventToSSE(event RealtimeEvent) string {
	data, _ := json.Marshal(event)
	return fmt.Sprintf("data: %s\n\n", string(data))
}

