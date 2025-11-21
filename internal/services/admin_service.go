package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
	"github.com/yourusername/couple-card-game/internal/models"
)

// AdminService handles admin-specific operations
type AdminService struct {
	client *supabase.Client
}

// NewAdminService creates a new AdminService
func NewAdminService(client *supabase.Client) *AdminService {
	return &AdminService{
		client: client,
	}
}

// DashboardStats contains statistics for the admin dashboard
type DashboardStats struct {
	TotalUsers      int
	AnonymousUsers  int
	RegisteredUsers int
	AdminUsers      int
	TotalRooms      int
	ActiveRooms     int
	CompletedRooms  int
	TotalQuestions  int
	TotalCategories int
}

// GetDashboardStats retrieves statistics for the admin dashboard
func (s *AdminService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Count total users
	data, _, err := s.client.From("users").Select("id", "", false).Execute()
	if err != nil {
		log.Printf("Error counting total users: %v", err)
		return nil, err
	}
	var totalUsers []map[string]interface{}
	if err := json.Unmarshal(data, &totalUsers); err == nil {
		stats.TotalUsers = len(totalUsers)
	}

	// Count anonymous users
	data, _, err = s.client.From("users").Select("id", "", false).Eq("is_anonymous", "true").Execute()
	if err == nil {
		var anonymousUsers []map[string]interface{}
		if err := json.Unmarshal(data, &anonymousUsers); err == nil {
			stats.AnonymousUsers = len(anonymousUsers)
		}
	}

	// Count registered users
	data, _, err = s.client.From("users").Select("id", "", false).Eq("is_anonymous", "false").Execute()
	if err == nil {
		var registeredUsers []map[string]interface{}
		if err := json.Unmarshal(data, &registeredUsers); err == nil {
			stats.RegisteredUsers = len(registeredUsers)
		}
	}

	// Count admin users
	data, _, err = s.client.From("users").Select("id", "", false).Eq("is_admin", "true").Execute()
	if err == nil {
		var adminUsers []map[string]interface{}
		if err := json.Unmarshal(data, &adminUsers); err == nil {
			stats.AdminUsers = len(adminUsers)
		}
	}

	// Count total rooms
	data, _, err = s.client.From("rooms").Select("id", "", false).Execute()
	if err == nil {
		var totalRooms []map[string]interface{}
		if err := json.Unmarshal(data, &totalRooms); err == nil {
			stats.TotalRooms = len(totalRooms)
		}
	}

	// Count active rooms
	data, _, err = s.client.From("rooms").Select("id", "", false).Eq("status", "active").Execute()
	if err == nil {
		var activeRooms []map[string]interface{}
		if err := json.Unmarshal(data, &activeRooms); err == nil {
			stats.ActiveRooms = len(activeRooms)
		}
	}

	// Count completed rooms
	data, _, err = s.client.From("rooms").Select("id", "", false).Eq("status", "completed").Execute()
	if err == nil {
		var completedRooms []map[string]interface{}
		if err := json.Unmarshal(data, &completedRooms); err == nil {
			stats.CompletedRooms = len(completedRooms)
		}
	}

	// Count total questions
	data, _, err = s.client.From("questions").Select("id", "", false).Execute()
	if err == nil {
		var totalQuestions []map[string]interface{}
		if err := json.Unmarshal(data, &totalQuestions); err == nil {
			stats.TotalQuestions = len(totalQuestions)
		}
	}

	// Count total categories
	data, _, err = s.client.From("categories").Select("id", "", false).Execute()
	if err == nil {
		var totalCategories []map[string]interface{}
		if err := json.Unmarshal(data, &totalCategories); err == nil {
			stats.TotalCategories = len(totalCategories)
		}
	}

	return stats, nil
}

// ListAllUsers retrieves all users with pagination
func (s *AdminService) ListAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := s.client.From("users").
		Select("*", "", false).
		Order("created_at", &postgrest.OrderOpts{Ascending: false})

	if limit > 0 {
		query = query.Range(offset, offset+limit-1, "")
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var users []*models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse users: %w", err)
	}

	return users, nil
}

// ToggleUserAdmin toggles the is_admin flag for a user
func (s *AdminService) ToggleUserAdmin(ctx context.Context, userID uuid.UUID) error {
	// First get the current user
	data, _, err := s.client.From("users").Select("*", "", false).Eq("id", userID.String()).Execute()
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	var users []*models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("failed to parse user: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("user not found")
	}

	user := users[0]

	// Toggle the is_admin flag
	updateData := map[string]interface{}{
		"is_admin": !user.IsAdmin,
	}

	_, _, err = s.client.From("users").Update(updateData, "", "").Eq("id", userID.String()).Execute()
	if err != nil {
		return fmt.Errorf("failed to toggle admin status: %w", err)
	}

	return nil
}

// DeleteUser soft deletes a user by setting deleted_at
func (s *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	_, _, err := s.client.From("users").Delete("", "").Eq("id", userID.String()).Execute()
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListAllRooms retrieves all rooms with player information
func (s *AdminService) ListAllRooms(ctx context.Context, limit, offset int) ([]*models.RoomWithPlayers, error) {
	query := s.client.From("rooms_with_players").
		Select("*", "", false).
		Order("created_at", &postgrest.OrderOpts{Ascending: false})

	if limit > 0 {
		query = query.Range(offset, offset+limit-1, "")
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	var rooms []*models.RoomWithPlayers
	if err := json.Unmarshal(data, &rooms); err != nil {
		return nil, fmt.Errorf("failed to parse rooms: %w", err)
	}

	return rooms, nil
}

// ForceCloseRoom closes a room regardless of its state
func (s *AdminService) ForceCloseRoom(ctx context.Context, roomID uuid.UUID) error {
	updateData := map[string]interface{}{
		"status": "completed",
	}

	_, _, err := s.client.From("rooms").Update(updateData, "", "").Eq("id", roomID.String()).Execute()
	if err != nil {
		return fmt.Errorf("failed to close room: %w", err)
	}

	return nil
}

// GetUserCount returns the total number of users
func (s *AdminService) GetUserCount(ctx context.Context) (int, error) {
	data, _, err := s.client.From("users").Select("id", "", false).Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(data, &users); err != nil {
		return 0, fmt.Errorf("failed to parse users: %w", err)
	}

	return len(users), nil
}
