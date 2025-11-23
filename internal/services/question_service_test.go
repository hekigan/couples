package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/models"
)

// TestGetRandomQuestion_CategoryFiltering tests that questions are filtered by category
func TestGetRandomQuestion_CategoryFiltering(t *testing.T) {
	// Note: This is an integration test that requires a test Supabase instance
	// Skip if running unit tests only
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		roomID      uuid.UUID
		language    string
		categoryIDs []uuid.UUID
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid request with categories",
			roomID:      uuid.New(),
			language:    "en",
			categoryIDs: []uuid.UUID{uuid.New(), uuid.New()},
			wantErr:     false,
		},
		{
			name:        "valid request without categories",
			roomID:      uuid.New(),
			language:    "en",
			categoryIDs: []uuid.UUID{},
			wantErr:     false,
		},
		{
			name:        "empty language should still work",
			roomID:      uuid.New(),
			language:    "",
			categoryIDs: []uuid.UUID{},
			wantErr:     true,
			errMsg:      "failed to fetch question",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test requires a real Supabase client
			// In production, you would inject a mock client here
			t.Skip("Requires test database setup")
		})
	}
}

// TestGetRandomQuestion_HistoryExclusion tests that already asked questions are excluded
func TestGetRandomQuestion_HistoryExclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("excludes previously asked questions", func(t *testing.T) {
		t.Skip("Requires test database with question history")
	})
}

// TestMarkQuestionAsked tests marking a question as asked
func TestMarkQuestionAsked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		roomID     uuid.UUID
		questionID uuid.UUID
		wantErr    bool
	}{
		{
			name:       "valid mark question as asked",
			roomID:     uuid.New(),
			questionID: uuid.New(),
			wantErr:    false,
		},
		{
			name:       "duplicate mark should not error",
			roomID:     uuid.New(),
			questionID: uuid.New(),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Requires test database setup")
		})
	}
}

// TestGetQuestionByID tests retrieving a question by ID
func TestGetQuestionByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		questionID uuid.UUID
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid question ID",
			questionID: uuid.New(),
			wantErr:    false,
		},
		{
			name:       "non-existent question ID",
			questionID: uuid.New(),
			wantErr:    true,
			errMsg:     "question not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Requires test database setup")
		})
	}
}

// TestCreateQuestion tests creating a new question
func TestCreateQuestion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name     string
		question *models.Question
		wantErr  bool
	}{
		{
			name: "valid question",
			question: &models.Question{
				CategoryID:   uuid.New(),
				LanguageCode: "en",
				Text:         "What is your favorite color?",
			},
			wantErr: false,
		},
		{
			name: "question with empty text",
			question: &models.Question{
				CategoryID:   uuid.New(),
				LanguageCode: "en",
				Text:         "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Requires test database setup")
		})
	}
}

// TestUpdateQuestion tests updating a question
func TestUpdateQuestion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("update existing question", func(t *testing.T) {
		t.Skip("Requires test database setup")
	})

	t.Run("update non-existent question", func(t *testing.T) {
		t.Skip("Requires test database setup")
	})
}

// TestDeleteQuestion tests deleting a question
func TestDeleteQuestion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("delete existing question", func(t *testing.T) {
		t.Skip("Requires test database setup")
	})

	t.Run("delete non-existent question should not error", func(t *testing.T) {
		t.Skip("Requires test database setup")
	})
}

// TestGetCategories tests retrieving all categories
func TestGetCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("get all categories", func(t *testing.T) {
		t.Skip("Requires test database setup")
	})
}

// TestJoinStrings tests the helper function for joining strings
func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "join with comma",
			strs:     []string{"a", "b", "c"},
			sep:      ",",
			expected: "'a','b','c'",
		},
		{
			name:     "empty slice",
			strs:     []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "single element",
			strs:     []string{"test"},
			sep:      ",",
			expected: "'test'",
		},
		{
			name:     "join with space",
			strs:     []string{"foo", "bar"},
			sep:      " ",
			expected: "'foo' 'bar'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("joinStrings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance-critical operations
func BenchmarkGetRandomQuestion(b *testing.B) {
	// Skip benchmark if no test database available
	b.Skip("Requires test database setup")
}

func BenchmarkMarkQuestionAsked(b *testing.B) {
	b.Skip("Requires test database setup")
}
