package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// GetRoomByID retrieves a room by ID from Supabase
func (s *RoomService) GetRoomByID(ctx context.Context, id uuid.UUID) (*models.Room, error) {
	// Query Supabase for the room
	data, _, err := s.client.From("rooms").
		Select("*", "", false).
		Eq("id", id.String()).
		Single().
		Execute()
	
	if err != nil {
		fmt.Printf("ERROR: Failed to fetch room %s: %v\n", id.String(), err)
		return nil, fmt.Errorf("room not found: %w", err)
	}
	
	// Parse the response
	var room models.Room
	if err := json.Unmarshal(data, &room); err != nil {
		fmt.Printf("ERROR: Failed to parse room data: %v\n", err)
		return nil, fmt.Errorf("failed to parse room: %w", err)
	}
	
	fmt.Printf("DEBUG: Room found: %s (owner: %s, status: %s)\n", room.ID, room.OwnerID, room.Status)
	return &room, nil
}

// CreateRoom creates a new room in Supabase
func (s *RoomService) CreateRoom(ctx context.Context, room *models.Room) error {
	// Set timestamps
	now := time.Now()
	room.CreatedAt = now
	room.UpdatedAt = now
	
	// Prepare data for insert
	data := map[string]interface{}{
		"id":         room.ID.String(),
		"owner_id":   room.OwnerID.String(),
		"status":     room.Status,
		"created_at": room.CreatedAt,
		"updated_at": room.UpdatedAt,
	}
	
	// Add optional fields if present
	if room.GuestID != nil {
		data["guest_id"] = room.GuestID.String()
	}
	
	fmt.Printf("DEBUG: Creating room in database: %+v\n", data)
	
	// Insert into Supabase
	responseData, count, err := s.client.From("rooms").Insert(data, false, "", "", "").Execute()
	if err != nil {
		fmt.Printf("ERROR: Failed to create room in Supabase: %v\n", err)
		return fmt.Errorf("failed to create room: %w", err)
	}
	
	fmt.Printf("DEBUG: Room created successfully. Response: %s, Count: %d\n", string(responseData), count)
	return nil
}

// UpdateRoom updates a room in Supabase
func (s *RoomService) UpdateRoom(ctx context.Context, room *models.Room) error {
	room.UpdatedAt = time.Now()

	data := map[string]interface{}{
		"status":     room.Status,
		"updated_at": room.UpdatedAt,
	}

	if room.GuestID != nil {
		data["guest_id"] = room.GuestID.String()
	}

	fmt.Printf("DEBUG: Updating room %s with data: %+v\n", room.ID, data)

	_, _, err := s.client.From("rooms").
		Update(data, "", "").
		Eq("id", room.ID.String()).
		Execute()

	if err != nil {
		fmt.Printf("ERROR: Failed to update room in Supabase: %v\n", err)
		return fmt.Errorf("failed to update room: %w", err)
	}

	fmt.Printf("DEBUG: Room %s updated successfully in database\n", room.ID)
	
	// Broadcast to realtime service
	s.realtimeService.BroadcastRoomUpdate(room.ID, room)
	return nil
}

// DeleteRoom deletes a room from Supabase
func (s *RoomService) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	fmt.Printf("DEBUG: Deleting room %s from database\n", id)
	
	// Delete from Supabase
	_, _, err := s.client.From("rooms").
		Delete("", "").
		Eq("id", id.String()).
		Execute()
	
	if err != nil {
		fmt.Printf("ERROR: Failed to delete room from Supabase: %v\n", err)
		return fmt.Errorf("failed to delete room: %w", err)
	}
	
	fmt.Printf("DEBUG: Room %s deleted successfully from database\n", id)
	
	// Broadcast deletion to realtime service
	s.realtimeService.BroadcastRoomDeleted(id)
	return nil
}

// CountUserRooms counts active rooms for a user (stub implementation)
func (s *RoomService) CountUserRooms(ctx context.Context, userID uuid.UUID) (int, error) {
	// TODO: Implement proper Supabase count
	return 0, nil
}

// GetRoomsByUserID gets all rooms for a user from Supabase (where user is owner OR guest)
func (s *RoomService) GetRoomsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Room, error) {
	userIDStr := userID.String()
	
	// Query for rooms where user is owner
	ownerData, _, err := s.client.From("rooms").
		Select("*", "", false).
		Eq("owner_id", userIDStr).
		Execute()
	
	var ownerRooms []models.Room
	if err == nil {
		json.Unmarshal(ownerData, &ownerRooms)
	}
	
	// Query for rooms where user is guest
	guestData, _, err := s.client.From("rooms").
		Select("*", "", false).
		Eq("guest_id", userIDStr).
		Execute()
	
	var guestRooms []models.Room
	if err == nil {
		json.Unmarshal(guestData, &guestRooms)
	}
	
	// Combine both lists
	allRooms := append(ownerRooms, guestRooms...)
	
	fmt.Printf("DEBUG: Found %d rooms for user %s (owner: %d, guest: %d)\n", 
		len(allRooms), userID, len(ownerRooms), len(guestRooms))
	
	return allRooms, nil
}

// CreateJoinRequest creates a new join request in Supabase
func (s *RoomService) CreateJoinRequest(ctx context.Context, request *models.RoomJoinRequest) error {
	// Set timestamps
	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now
	
	// Prepare data for insert (without 'message' field - not in schema)
	data := map[string]interface{}{
		"id":         request.ID.String(),
		"room_id":    request.RoomID.String(),
		"user_id":    request.UserID.String(),
		"status":     request.Status,
		"created_at": request.CreatedAt,
		"updated_at": request.UpdatedAt,
	}
	
	fmt.Printf("DEBUG: Attempting to insert join request: %+v\n", data)
	
	// Insert into Supabase
	responseData, count, err := s.client.From("room_join_requests").Insert(data, false, "", "", "").Execute()
	if err != nil {
		fmt.Printf("ERROR: Supabase insert failed: %v\n", err)
		return fmt.Errorf("failed to create join request: %w", err)
	}
	
	fmt.Printf("DEBUG: Insert successful. Response data: %s, Count: %d\n", string(responseData), count)
	return nil
}

// GetJoinRequestsByRoom gets pending join requests for a room from Supabase
func (s *RoomService) GetJoinRequestsByRoom(ctx context.Context, roomID uuid.UUID) ([]models.RoomJoinRequest, error) {
	// Query Supabase for join requests
	data, _, err := s.client.From("room_join_requests").
		Select("*", "", false).
		Eq("room_id", roomID.String()).
		Eq("status", "pending").
		Execute()
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch join requests: %w", err)
	}
	
	// Parse the response
	var requests []models.RoomJoinRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		return nil, fmt.Errorf("failed to parse join requests: %w", err)
	}
	
	return requests, nil
}

// GetAllJoinRequestsByRoom gets ALL join requests (including rejected) for a room
func (s *RoomService) GetAllJoinRequestsByRoom(ctx context.Context, roomID uuid.UUID) ([]models.RoomJoinRequest, error) {
	// Query Supabase for ALL join requests
	data, _, err := s.client.From("room_join_requests").
		Select("*", "", false).
		Eq("room_id", roomID.String()).
		Execute()
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch join requests: %w", err)
	}
	
	// Parse the response
	var requests []models.RoomJoinRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		return nil, fmt.Errorf("failed to parse join requests: %w", err)
	}
	
	return requests, nil
}

// CancelJoinRequest cancels (deletes) a user's join request
func (s *RoomService) CancelJoinRequest(ctx context.Context, roomID, userID uuid.UUID) error {
	_, _, err := s.client.From("room_join_requests").
		Delete("", "").
		Eq("room_id", roomID.String()).
		Eq("user_id", userID.String()).
		Execute()
	
	if err != nil {
		return fmt.Errorf("failed to cancel join request: %w", err)
	}
	
	return nil
}

// GetAcceptedRequestsByUser gets all accepted join requests for a user
func (s *RoomService) GetAcceptedRequestsByUser(ctx context.Context, userID uuid.UUID) ([]models.RoomJoinRequest, error) {
	data, _, err := s.client.From("room_join_requests").
		Select("*", "", false).
		Eq("user_id", userID.String()).
		Eq("status", "accepted").
		Execute()
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accepted requests: %w", err)
	}
	
	var requests []models.RoomJoinRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		return nil, fmt.Errorf("failed to parse accepted requests: %w", err)
	}
	
	return requests, nil
}

// JoinRequestWithUserInfo represents a join request with user information
type JoinRequestWithUserInfo struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetJoinRequestsWithUserInfo gets join requests with user information
func (s *RoomService) GetJoinRequestsWithUserInfo(ctx context.Context, roomID uuid.UUID) ([]JoinRequestWithUserInfo, error) {
	// Get join requests
	requests, err := s.GetJoinRequestsByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	
	// Fetch user info for each request
	var result []JoinRequestWithUserInfo
	for _, req := range requests {
		// Query user from Supabase
		userData, _, err := s.client.From("users").
			Select("username", "", false).
			Eq("id", req.UserID.String()).
			Single().
			Execute()
		
		if err != nil {
			fmt.Printf("WARNING: Failed to fetch user info for request %s: %v\n", req.ID, err)
			continue
		}
		
		var userInfo struct {
			Username string `json:"username"`
		}
		if err := json.Unmarshal(userData, &userInfo); err != nil {
			fmt.Printf("WARNING: Failed to parse user info: %v\n", err)
			continue
		}
		
		result = append(result, JoinRequestWithUserInfo{
			ID:        req.ID,
			RoomID:    req.RoomID,
			UserID:    req.UserID,
			Username:  userInfo.Username,
			Status:    req.Status,
			CreatedAt: req.CreatedAt,
			UpdatedAt: req.UpdatedAt,
		})
	}
	
	return result, nil
}

// AcceptJoinRequest accepts a join request in Supabase
func (s *RoomService) AcceptJoinRequest(ctx context.Context, requestID uuid.UUID) error {
	now := time.Now()
	
	data := map[string]interface{}{
		"status":     "accepted",
		"updated_at": now,
	}
	
	_, _, err := s.client.From("room_join_requests").
		Update(data, "", "").
		Eq("id", requestID.String()).
		Execute()
	
	if err != nil {
		return fmt.Errorf("failed to accept join request: %w", err)
	}
	
	return nil
}

// RejectJoinRequest rejects a join request in Supabase
func (s *RoomService) RejectJoinRequest(ctx context.Context, requestID uuid.UUID) error {
	now := time.Now()
	
	data := map[string]interface{}{
		"status":     "rejected",
		"updated_at": now,
	}
	
	_, _, err := s.client.From("room_join_requests").
		Update(data, "", "").
		Eq("id", requestID.String()).
		Execute()
	
	if err != nil {
		return fmt.Errorf("failed to reject join request: %w", err)
	}
	
	return nil
}
