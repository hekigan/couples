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
		Channel: make(chan RealtimeEvent, 100), // Increased buffer size to prevent dropped events
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

	broadcastCount := 0
	for _, client := range s.clients {
		if client.RoomID == roomID {
			select {
			case client.Channel <- event:
				broadcastCount++
			default:
				// Channel full, skip this client
				fmt.Printf("⚠️ Warning: Skipped broadcasting %s event to client %s (channel full)\n", event.Type, client.ID)
			}
		}
	}
	if broadcastCount > 0 {
		fmt.Printf("📡 Broadcasted %s event to %d clients in room %s\n", event.Type, broadcastCount, roomID)
	}
}

// BroadcastToUser sends an event to all connections for a specific user
func (s *RealtimeService) BroadcastToUser(userID uuid.UUID, event RealtimeEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fmt.Printf("📢 Broadcasting to user %s: event type '%s'\n", userID, event.Type)

	matchedClients := 0
	sentCount := 0
	skippedCount := 0

	for _, client := range s.clients {
		if client.UserID == userID {
			matchedClients++
			select {
			case client.Channel <- event:
				sentCount++
				fmt.Printf("  ✅ Sent to client %s (room: %s)\n", client.ID, client.RoomID)
			default:
				skippedCount++
				fmt.Printf("  ⚠️ Channel full for client %s - event dropped!\n", client.ID)
			}
		}
	}

	if matchedClients == 0 {
		fmt.Printf("  ⚠️ No clients found for user %s\n", userID)
	} else {
		fmt.Printf("  📊 Summary: %d matched, %d sent, %d skipped\n", matchedClients, sentCount, skippedCount)
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
		Data: map[string]interface{}{
			"question": question,
		},
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

// BroadcastJoinRequest broadcasts a join_request event
func (s *RealtimeService) BroadcastJoinRequest(roomID uuid.UUID, userID uuid.UUID) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "join_request",
		Data: map[string]interface{}{
			"user_id": userID.String(),
		},
	})
}

// BroadcastJoinRequestWithDetails broadcasts a join_request event with full request details
func (s *RealtimeService) BroadcastJoinRequestWithDetails(roomID, requestID, userID uuid.UUID, username string, createdAt interface{}) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "join_request",
		Data: map[string]interface{}{
			"id":         requestID.String(),
			"user_id":    userID.String(),
			"username":   username,
			"created_at": createdAt,
			"status":     "pending",
		},
	})
}

// BroadcastRequestAccepted broadcasts a request_accepted event to room
func (s *RealtimeService) BroadcastRequestAccepted(roomID uuid.UUID, guestUsername string) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "request_accepted",
		Data: map[string]interface{}{
			"guest_username": guestUsername,
		},
	})
}

// BroadcastRequestAcceptedToGuest notifies guest that their request was accepted
func (s *RealtimeService) BroadcastRequestAcceptedToGuest(guestUserID, roomID uuid.UUID) {
	s.BroadcastToUser(guestUserID, RealtimeEvent{
		Type: "my_request_accepted",
		Data: map[string]interface{}{
			"room_id": roomID.String(),
			"status":  "accepted",
		},
	})
}

// BroadcastRequestRejectedToGuest notifies guest that their request was rejected
func (s *RealtimeService) BroadcastRequestRejectedToGuest(guestUserID, roomID uuid.UUID) {
	s.BroadcastToUser(guestUserID, RealtimeEvent{
		Type: "my_request_rejected",
		Data: map[string]interface{}{
			"room_id": roomID.String(),
			"status":  "rejected",
		},
	})
}

// BroadcastPlayerTyping broadcasts a player_typing event to other players in the room
func (s *RealtimeService) BroadcastPlayerTyping(roomID, userID uuid.UUID, isTyping bool) {
	s.Broadcast(roomID, RealtimeEvent{
		Type: "player_typing",
		Data: map[string]interface{}{
			"user_id":   userID.String(),
			"is_typing": isTyping,
		},
	})
}

// EventToSSE converts an event to Server-Sent Events format
func EventToSSE(event RealtimeEvent) string {
	data, _ := json.Marshal(event.Data)
	// Include event type in SSE format for proper event handling
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, string(data))
}


