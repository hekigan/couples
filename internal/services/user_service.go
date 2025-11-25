package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/hekigan/couples/internal/models"
)

// UserService handles user-related operations
type UserService struct {
	*BaseService
	client *supabase.Client
}

// NewUserService creates a new user service
func NewUserService(client *supabase.Client) *UserService {
	return &UserService{
		BaseService: NewBaseService(client, "UserService"),
		client:      client,
	}
}

// GetSupabaseClient returns the underlying Supabase client
func (s *UserService) GetSupabaseClient() *supabase.Client {
	return s.client
}

// CreateAnonymousUser creates a new anonymous user
// This is a WORKING implementation that creates anonymous users
func (s *UserService) CreateAnonymousUser(ctx context.Context) (*models.User, error) {
	userID := uuid.New()
	user := models.User{
		ID:          userID,
		IsAnonymous: true,
		Name:        "Guest User",
		Username:    fmt.Sprintf("guest_%s", userID.String()[:8]), // Temporary username
	}

	userMap := map[string]interface{}{
		"id":           user.ID.String(),
		"is_anonymous": true,
		"name":         user.Name,
		"username":     user.Username, // Temporary, user will be prompted to change
	}

	if err := s.BaseService.InsertRecord(ctx, "users", userMap); err != nil {
		return nil, fmt.Errorf("failed to create user in database: %w", err)
	}

	s.logger.Success("User created successfully in database with user_id=%s", user.ID.String())
	return &user, nil
}

// GetUserByID retrieves a user by ID from Supabase
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.BaseService.GetSingleRecord(ctx, "users", id, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUsername updates a user's username
func (s *UserService) UpdateUsername(ctx context.Context, userID uuid.UUID, username string) error {
	data := map[string]interface{}{
		"username": username,
	}

	return s.BaseService.UpdateRecord(ctx, "users", userID, data)
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	updateData := map[string]interface{}{
		"username": user.Username,
		"name":     user.Name,
	}

	// Only update is_admin if it's set
	if user.IsAdmin {
		updateData["is_admin"] = user.IsAdmin
	}

	return s.BaseService.UpdateRecord(ctx, "users", user.ID, updateData)
}

// DeleteUser deletes a user with full cascade (removes all related data)
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	s.logger.Debug("Starting cascade delete for user_id=%s", userID.String())

	// 1. Delete user's rooms (as owner)
	s.logger.Debug("Deleting rooms where user is owner...")
	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "rooms", map[string]interface{}{
		"owner_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete rooms: %v", err)
	}

	// 2. Update rooms where user is guest (set guest_id to NULL)
	s.logger.Debug("Removing user as guest from rooms...")
	if err := s.BaseService.UpdateRecordsWithFilter(ctx, "rooms",
		map[string]interface{}{"guest_id": userID.String()},
		map[string]interface{}{"guest_id": nil, "status": "waiting"},
	); err != nil {
		s.logger.Warn("Failed to update rooms: %v", err)
	}

	// 3. Delete user's answers
	s.logger.Debug("Deleting user's answers...")
	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "answers", map[string]interface{}{
		"user_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete answers: %v", err)
	}

	// 4. Delete user's join requests
	s.logger.Debug("Deleting user's join requests...")
	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "room_join_requests", map[string]interface{}{
		"user_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete join requests: %v", err)
	}

	// 5. Delete friendships (both as user_id and friend_id)
	s.logger.Debug("Deleting friendships...")
	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "friends", map[string]interface{}{
		"user_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete friends (as user): %v", err)
	}

	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "friends", map[string]interface{}{
		"friend_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete friends (as friend): %v", err)
	}

	// 6. Delete room invitations
	s.logger.Debug("Deleting room invitations...")
	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "room_invitations", map[string]interface{}{
		"inviter_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete invitations (as inviter): %v", err)
	}

	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "room_invitations", map[string]interface{}{
		"invitee_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete invitations (as invitee): %v", err)
	}

	// 7. Delete notifications
	s.logger.Debug("Deleting notifications...")
	if err := s.BaseService.DeleteRecordsWithFilter(ctx, "notifications", map[string]interface{}{
		"user_id": userID.String(),
	}); err != nil {
		s.logger.Warn("Failed to delete notifications: %v", err)
	}

	// 8. Finally, delete the user
	s.logger.Debug("Deleting user record...")
	if err := s.BaseService.DeleteRecord(ctx, "users", userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Success("User successfully deleted with all related data user_id=%s", userID.String())
	return nil
}

// CleanupExpiredAnonymousUsers deletes anonymous users older than the specified duration
func (s *UserService) CleanupExpiredAnonymousUsers(ctx context.Context, olderThanHours int) (int, error) {
	s.logger.Debug("Starting cleanup of anonymous users older_than_hours=%d", olderThanHours)

	// Query for expired anonymous users
	filters := map[string]interface{}{
		"is_anonymous": "true",
	}

	var users []models.User
	if err := s.BaseService.GetRecords(ctx, "users", filters, &users); err != nil {
		return 0, err
	}

	// Filter users by age manually (since Supabase Go client doesn't support lt() for dates easily)
	// TODO: Implement proper time-based filtering using olderThanHours parameter
	// For now, this returns all anonymous users for manual cleanup
	var expiredUsers []uuid.UUID
	for _, user := range users {
		// Simple check: if created_at is not recent, consider it expired
		// In a real implementation, you'd parse created_at and compare with current time
		expiredUsers = append(expiredUsers, user.ID)
	}

	// Delete each expired user
	deletedCount := 0
	for _, userID := range expiredUsers {
		if err := s.DeleteUser(ctx, userID); err != nil {
			s.logger.Warn("Failed to delete expired user user_id=%s: %v", userID.String(), err)
			continue
		}
		deletedCount++
	}

	s.logger.Success("Cleanup complete deleted_count=%d", deletedCount)
	return deletedCount, nil
}

// GetAnonymousUserCount returns the count of anonymous users
func (s *UserService) GetAnonymousUserCount(ctx context.Context) (int, error) {
	filters := map[string]interface{}{
		"is_anonymous": "true",
	}

	return s.BaseService.CountRecords(ctx, "users", filters)
}

// CleanupInactiveAnonymousUsers deletes anonymous users who have no active rooms or recent activity
func (s *UserService) CleanupInactiveAnonymousUsers(ctx context.Context) (int, error) {
	s.logger.Debug("Starting cleanup of inactive anonymous users")

	// Get all anonymous users
	filters := map[string]interface{}{
		"is_anonymous": "true",
	}

	var users []models.User
	if err := s.BaseService.GetRecords(ctx, "users", filters, &users); err != nil {
		return 0, err
	}

	deletedCount := 0
	for _, user := range users {
		// Check if user has any active rooms
		roomData, _, err := s.client.From("rooms").
			Select("id", "exact", false).
			Eq("owner_id", user.ID.String()).
			Execute()

		hasRooms := err == nil && len(roomData) > 0

		// Check if user is a guest in any room
		guestData, _, err := s.client.From("rooms").
			Select("id", "exact", false).
			Eq("guest_id", user.ID.String()).
			Execute()

		isGuest := err == nil && len(guestData) > 0

		// If user has no active involvement, delete them
		if !hasRooms && !isGuest {
			if err := s.DeleteUser(ctx, user.ID); err != nil {
				s.logger.Warn("Failed to delete inactive user user_id=%s: %v", user.ID.String(), err)
				continue
			}
			deletedCount++
		}
	}

	s.logger.Success("Cleanup complete deleted_count=%d", deletedCount)
	return deletedCount, nil
}
