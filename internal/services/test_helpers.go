package services

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
	"github.com/yourusername/couple-card-game/internal/models"
)

// SetupTestDatabase initializes a test database connection
// It looks for TEST_SUPABASE_URL and TEST_SUPABASE_KEY environment variables
// If not found, it attempts to load from .env.test file
func SetupTestDatabase(t *testing.T) *supabase.Client {
	// Try to load test environment file (ignore errors if not present)
	_ = godotenv.Load("../../.env.test")

	testURL := os.Getenv("TEST_SUPABASE_URL")
	testKey := os.Getenv("TEST_SUPABASE_KEY")

	// If not set, provide helpful skip message
	if testURL == "" || testKey == "" {
		t.Skip("Test database not configured. See TEST_DATABASE_SETUP.md for setup instructions.\n" +
			"Quick start: Run 'supabase start' and set TEST_SUPABASE_URL and TEST_SUPABASE_KEY")
	}

	client, err := supabase.NewClient(testURL, testKey, nil)
	if err != nil {
		t.Fatalf("Failed to create test database client: %v\nCheck your TEST_SUPABASE_URL and TEST_SUPABASE_KEY", err)
	}

	return client
}

// CleanupTestData cleans up all test data from the database
// This should be called with defer after SetupTestDatabase
// Example:
//
//	client := SetupTestDatabase(t)
//	defer CleanupTestData(t, client)
func CleanupTestData(t *testing.T, client *supabase.Client) {
	// Delete in reverse dependency order to avoid foreign key constraints
	tables := []string{
		"question_history",   // References: questions, rooms
		"answers",            // References: questions, rooms, users
		"room_join_requests", // References: rooms, users
		"room_invitations",   // References: rooms, users
		"notifications",      // References: users
		"friends",            // References: users
		"rooms",              // References: users
		"users",              // No dependencies
	}

	for _, table := range tables {
		_, _, err := client.From(table).Delete("", "").Execute()
		if err != nil {
			t.Logf("Warning: Failed to clean table %s: %v (this may be expected if table is empty)", table, err)
		}
	}
}

// CreateTestUser creates a test user and returns the user object
func CreateTestUser(t *testing.T, client *supabase.Client, username, name string, isAnonymous bool) *UserTestData {
	userService := NewUserService(client)

	var user *UserTestData
	if isAnonymous {
		createdUser, err := userService.CreateAnonymousUser(context.Background())
		if err != nil {
			t.Fatalf("Failed to create anonymous test user: %v", err)
		}
		user = &UserTestData{
			ID:          createdUser.ID,
			Username:    createdUser.Username,
			Name:        createdUser.Name,
			IsAnonymous: true,
		}
	} else {
		// For non-anonymous users, we'll need to use direct insert
		// since there's no CreateUser method for non-anonymous
		t.Skip("Non-anonymous user creation not yet implemented in test helpers")
	}

	return user
}

// UserTestData holds test user information
type UserTestData struct {
	ID          uuid.UUID
	Username    string
	Name        string
	IsAnonymous bool
}

// CreateTestRoom creates a test room and returns the room ID
func CreateTestRoom(t *testing.T, client *supabase.Client, ownerID uuid.UUID, language string) uuid.UUID {
	roomService := NewRoomService(client, nil) // nil realtimeService for tests

	room := &models.Room{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		Language: language,
		Status:   "waiting",
	}

	err := roomService.CreateRoom(context.Background(), room)
	if err != nil {
		t.Fatalf("Failed to create test room: %v", err)
	}

	return room.ID
}

// CreateTestCategory creates a test category and returns the category ID
func CreateTestCategory(t *testing.T, client *supabase.Client, name string) uuid.UUID {
	categoryData := map[string]interface{}{
		"id":   uuid.New().String(),
		"name": name,
	}

	data, _, err := client.From("categories").Insert(categoryData, false, "", "", "").Execute()
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	var categories []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &categories); err != nil {
		t.Fatalf("Failed to parse category response: %v", err)
	}

	categoryID, _ := uuid.Parse(categories[0].ID)
	return categoryID
}

// CreateTestQuestion creates a test question and returns the question ID
func CreateTestQuestion(t *testing.T, client *supabase.Client, categoryID uuid.UUID, langCode, text string) uuid.UUID {
	questionData := map[string]interface{}{
		"id":            uuid.New().String(),
		"category_id":   categoryID.String(),
		"lang_code":     langCode,
		"question_text": text,
	}

	data, _, err := client.From("questions").Insert(questionData, false, "", "", "").Execute()
	if err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	var questions []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &questions); err != nil {
		t.Fatalf("Failed to parse question response: %v", err)
	}

	questionID, _ := uuid.Parse(questions[0].ID)
	return questionID
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotNil fails the test if value is nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: expected non-nil value", message)
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: expected true, got false", message)
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Fatalf("%s: expected false, got true", message)
	}
}
