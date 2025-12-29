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
	*BaseService
	client *supabase.Client
}

// NewAdminService creates a new AdminService
func NewAdminService(client *supabase.Client) *AdminService {
	return &AdminService{
		BaseService: NewBaseService(client, "AdminService"),
		client:      client,
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
		if count, err := s.BaseService.CountRecords(ctx, "users", map[string]interface{}{}); err == nil {
			mu.Lock()
			stats.TotalUsers = count
			mu.Unlock()
		}
	}()

	// Count anonymous users
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "users", map[string]interface{}{"is_anonymous": "true"}); err == nil {
			mu.Lock()
			stats.AnonymousUsers = count
			mu.Unlock()
		}
	}()

	// Count registered users
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "users", map[string]interface{}{"is_anonymous": "false"}); err == nil {
			mu.Lock()
			stats.RegisteredUsers = count
			mu.Unlock()
		}
	}()

	// Count admin users
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "users", map[string]interface{}{"is_admin": "true"}); err == nil {
			mu.Lock()
			stats.AdminUsers = count
			mu.Unlock()
		}
	}()

	// Count total rooms
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "rooms", map[string]interface{}{}); err == nil {
			mu.Lock()
			stats.TotalRooms = count
			mu.Unlock()
		}
	}()

	// Count active rooms
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "rooms", map[string]interface{}{"status": "active"}); err == nil {
			mu.Lock()
			stats.ActiveRooms = count
			mu.Unlock()
		}
	}()

	// Count completed rooms
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "rooms", map[string]interface{}{"status": "completed"}); err == nil {
			mu.Lock()
			stats.CompletedRooms = count
			mu.Unlock()
		}
	}()

	// Count total questions
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "questions", map[string]interface{}{}); err == nil {
			mu.Lock()
			stats.TotalQuestions = count
			mu.Unlock()
		}
	}()

	// Count total categories
	wg.Add(1)
	go func() {
		defer wg.Done()
		if count, err := s.BaseService.CountRecords(ctx, "categories", map[string]interface{}{}); err == nil {
			mu.Lock()
			stats.TotalCategories = count
			mu.Unlock()
		}
	}()

	// Wait for all queries to complete
	wg.Wait()

	return stats, nil
}

// ListAllUsers retrieves all users with pagination
func (s *AdminService) ListAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	// Custom query - uses Order which is not supported by BaseService
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
	var user models.User
	if err := s.BaseService.GetSingleRecord(ctx, "users", userID, &user); err != nil {
		return err
	}

	// Toggle the is_admin flag
	updateData := map[string]interface{}{
		"is_admin": !user.IsAdmin,
	}

	return s.BaseService.UpdateRecord(ctx, "users", userID, updateData)
}

// DeleteUser deletes a user
func (s *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.BaseService.DeleteRecord(ctx, "users", userID)
}

// GetUserByID retrieves a user by ID
func (s *AdminService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.BaseService.GetSingleRecord(ctx, "users", userID, &user); err != nil {
		return nil, err
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

	if err := s.BaseService.InsertRecord(ctx, "users", userMap); err != nil {
		return nil, err
	}

	// Return the created user
	return s.GetUserByID(ctx, userID)
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

	return s.BaseService.UpdateRecord(ctx, "users", userID, updateData)
}

// ListAllRooms retrieves rooms with filtering and sorting
func (s *AdminService) ListAllRooms(ctx context.Context, limit, offset int, search string, statuses []string, sortBy, sortOrder string) ([]*models.RoomWithPlayers, error) {
	query := s.client.From("rooms_with_players").
		Select("*", "", false)

	// Apply search filter (case-insensitive across name, owner, guest)
	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		query = query.Or(fmt.Sprintf("name.ilike.%s,owner_username.ilike.%s,guest_username.ilike.%s",
			searchPattern, searchPattern, searchPattern), "")
	}

	// Apply status filter
	if len(statuses) > 0 {
		query = query.In("status", statuses)
	}

	// Apply sorting (default: created_at DESC)
	ascending := sortOrder == "asc"
	switch sortBy {
	case "owner":
		query = query.Order("owner_username", &postgrest.OrderOpts{Ascending: ascending, NullsFirst: !ascending})
	case "guest":
		query = query.Order("guest_username", &postgrest.OrderOpts{Ascending: ascending, NullsFirst: !ascending})
	case "status":
		query = query.Order("status", &postgrest.OrderOpts{Ascending: ascending})
	case "created_at":
		query = query.Order("created_at", &postgrest.OrderOpts{Ascending: ascending})
	default:
		query = query.Order("created_at", &postgrest.OrderOpts{Ascending: false})
	}

	// Apply pagination
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

	return s.BaseService.UpdateRecord(ctx, "rooms", roomID, updateData)
}

// GetUserCount returns the total number of users
func (s *AdminService) GetUserCount(ctx context.Context) (int, error) {
	return s.BaseService.CountRecords(ctx, "users", map[string]interface{}{})
}

// GetRoomCount returns the total number of rooms
func (s *AdminService) GetRoomCount(ctx context.Context) (int, error) {
	return s.BaseService.CountRecords(ctx, "rooms", map[string]interface{}{})
}

// GetFilteredRoomCount returns count of rooms matching filters
func (s *AdminService) GetFilteredRoomCount(ctx context.Context, search string, statuses []string) (int, error) {
	query := s.client.From("rooms_with_players").
		Select("*", "exact", true) // Count query

	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		query = query.Or(fmt.Sprintf("name.ilike.%s,owner_username.ilike.%s,guest_username.ilike.%s",
			searchPattern, searchPattern, searchPattern), "")
	}

	if len(statuses) > 0 {
		query = query.In("status", statuses)
	}

	_, count, err := query.Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to count rooms: %w", err)
	}

	return int(count), nil
}
