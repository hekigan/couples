package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
	"github.com/hekigan/couples/internal/models"
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

	// Use sync.WaitGroup to run all queries in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex // Protect stats writes

	// Count total users
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("users").Select("*", "exact", true).Execute()
		if err == nil {
			mu.Lock()
			stats.TotalUsers = int(count)
			mu.Unlock()
		}
	}()

	// Count anonymous users
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("users").Select("*", "exact", true).Eq("is_anonymous", "true").Execute()
		if err == nil {
			mu.Lock()
			stats.AnonymousUsers = int(count)
			mu.Unlock()
		}
	}()

	// Count registered users
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("users").Select("*", "exact", true).Eq("is_anonymous", "false").Execute()
		if err == nil {
			mu.Lock()
			stats.RegisteredUsers = int(count)
			mu.Unlock()
		}
	}()

	// Count admin users
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("users").Select("*", "exact", true).Eq("is_admin", "true").Execute()
		if err == nil {
			mu.Lock()
			stats.AdminUsers = int(count)
			mu.Unlock()
		}
	}()

	// Count total rooms
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("rooms").Select("*", "exact", true).Execute()
		if err == nil {
			mu.Lock()
			stats.TotalRooms = int(count)
			mu.Unlock()
		}
	}()

	// Count active rooms
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("rooms").Select("*", "exact", true).Eq("status", "active").Execute()
		if err == nil {
			mu.Lock()
			stats.ActiveRooms = int(count)
			mu.Unlock()
		}
	}()

	// Count completed rooms
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("rooms").Select("*", "exact", true).Eq("status", "completed").Execute()
		if err == nil {
			mu.Lock()
			stats.CompletedRooms = int(count)
			mu.Unlock()
		}
	}()

	// Count total questions
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("questions").Select("*", "exact", true).Execute()
		if err == nil {
			mu.Lock()
			stats.TotalQuestions = int(count)
			mu.Unlock()
		}
	}()

	// Count total categories
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, count, err := s.client.From("categories").Select("*", "exact", true).Execute()
		if err == nil {
			mu.Lock()
			stats.TotalCategories = int(count)
			mu.Unlock()
		}
	}()

	// Wait for all queries to complete
	wg.Wait()

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

// GetUserByID retrieves a user by ID
func (s *AdminService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	data, _, err := s.client.From("users").
		Select("*", "", false).
		Eq("id", userID.String()).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var user models.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user (for admin panel)
func (s *AdminService) CreateUser(ctx context.Context, username, email string, isAdmin, isAnonymous bool) (*models.User, error) {
	userID := uuid.New()

	userMap := map[string]interface{}{
		"id":           userID.String(),
		"username":     username,
		"is_admin":     isAdmin,
		"is_anonymous": isAnonymous,
	}

	// Only add email if provided (anonymous users may not have email)
	if email != "" {
		userMap["email"] = email
		userMap["name"] = username // Set name same as username for now
	}

	data, _, err := s.client.From("users").Insert(userMap, false, "", "", "").Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	var users []*models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse created user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user returned after creation")
	}

	return users[0], nil
}

// UpdateUser updates a user's information
func (s *AdminService) UpdateUser(ctx context.Context, userID uuid.UUID, username, email string, isAdmin, isAnonymous bool) error {
	updateData := map[string]interface{}{
		"username":     username,
		"is_admin":     isAdmin,
		"is_anonymous": isAnonymous,
	}

	// Only update email if provided
	if email != "" {
		updateData["email"] = email
		updateData["name"] = username
	}

	_, _, err := s.client.From("users").
		Update(updateData, "", "").
		Eq("id", userID.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
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

// GetRoomCount returns the total number of rooms
func (s *AdminService) GetRoomCount(ctx context.Context) (int, error) {
	data, _, err := s.client.From("rooms").Select("id", "", false).Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to count rooms: %w", err)
	}

	var rooms []map[string]interface{}
	if err := json.Unmarshal(data, &rooms); err != nil {
		return 0, fmt.Errorf("failed to parse rooms: %w", err)
	}

	return len(rooms), nil
}
