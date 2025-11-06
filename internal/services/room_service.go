package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/yourusername/couple-card-game/internal/models"
)

// RoomService handles room-related operations
type RoomService struct {
	client          *supabase.Client
	realtimeService *RealtimeService
}

// NewRoomService creates a new room service
func NewRoomService(client *supabase.Client, realtimeService *RealtimeService) *RoomService {
	return &RoomService{
		client:          client,
		realtimeService: realtimeService,
	}
}

// GetRoomByID retrieves a room by ID (stub implementation)
func (s *RoomService) GetRoomByID(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	// TODO: Implement proper Supabase query
	return &models.Room{
		ID:       id,
		Name:     "Test Room",
		OwnerID:  uuid.New(),
		Status:   "waiting",
		Language: "en",
	}, nil
}

// CreateRoom creates a new room (stub implementation)
func (s *RoomService) CreateRoom(ctx context.Context, room *models.Room) error {
	// TODO: Implement proper Supabase insert
	return nil
}

// UpdateRoom updates a room (stub implementation)
func (s *RoomService) UpdateRoom(ctx context.Context, room *models.Room) error {
	// TODO: Implement proper Supabase update
	s.realtimeService.BroadcastRoomUpdate(room.ID, room)
	return nil
}

// DeleteRoom deletes a room (stub implementation)
func (s *RoomService) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	s.realtimeService.BroadcastRoomDeleted(id)
	// TODO: Implement proper Supabase delete
	return nil
}

// CountUserRooms counts active rooms for a user (stub implementation)
func (s *RoomService) CountUserRooms(ctx context.Context, userID uuid.UUID) (int, error) {
	// TODO: Implement proper Supabase count
	return 0, nil
}
