package services

import (
	"testing"

	"github.com/google/uuid"
)

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []uuid.UUID
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []uuid.UUID{},
			expected: []string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []string{},
		},
		{
			name: "single UUID",
			input: []uuid.UUID{
				uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			expected: []string{
				"123e4567-e89b-12d3-a456-426614174000",
			},
		},
		{
			name: "multiple UUIDs",
			input: []uuid.UUID{
				uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
				uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
			},
			expected: []string{
				"123e4567-e89b-12d3-a456-426614174000",
				"123e4567-e89b-12d3-a456-426614174001",
				"123e4567-e89b-12d3-a456-426614174002",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToStringSlice(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("ToStringSlice() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ToStringSlice()[%d] = %s, want %s", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestUUIDToStringOrNil(t *testing.T) {
	tests := []struct {
		name     string
		input    *uuid.UUID
		expected interface{}
	}{
		{
			name:     "nil UUID",
			input:    nil,
			expected: nil,
		},
		{
			name: "valid UUID",
			input: func() *uuid.UUID {
				id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				return &id
			}(),
			expected: "123e4567-e89b-12d3-a456-426614174000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UUIDToStringOrNil(tt.input)

			if result != tt.expected {
				t.Errorf("UUIDToStringOrNil() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildFilter(t *testing.T) {
	tests := []struct {
		name     string
		filters  []map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "no filters",
			filters:  []map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "single filter",
			filters: []map[string]interface{}{
				{"status": "active"},
			},
			expected: map[string]interface{}{
				"status": "active",
			},
		},
		{
			name: "multiple filters",
			filters: []map[string]interface{}{
				{"status": "active"},
				{"user_id": "123"},
			},
			expected: map[string]interface{}{
				"status":  "active",
				"user_id": "123",
			},
		},
		{
			name: "overlapping filters - last wins",
			filters: []map[string]interface{}{
				{"status": "active"},
				{"status": "pending"},
			},
			expected: map[string]interface{}{
				"status": "pending",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildFilter(tt.filters...)

			if len(result) != len(tt.expected) {
				t.Errorf("BuildFilter() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("BuildFilter()[%s] = %v, want %v", key, result[key], expectedValue)
				}
			}
		})
	}
}

func TestWithStatus(t *testing.T) {
	result := WithStatus("active")
	expected := map[string]interface{}{"status": "active"}

	if len(result) != 1 {
		t.Errorf("WithStatus() should return map with 1 element, got %d", len(result))
	}

	if result["status"] != expected["status"] {
		t.Errorf("WithStatus() = %v, want %v", result, expected)
	}
}

func TestWithUserID(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	result := WithUserID(userID)
	expected := map[string]interface{}{"user_id": "123e4567-e89b-12d3-a456-426614174000"}

	if len(result) != 1 {
		t.Errorf("WithUserID() should return map with 1 element, got %d", len(result))
	}

	if result["user_id"] != expected["user_id"] {
		t.Errorf("WithUserID() = %v, want %v", result, expected)
	}
}

func TestWithOwnerID(t *testing.T) {
	ownerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	result := WithOwnerID(ownerID)
	expected := map[string]interface{}{"owner_id": "123e4567-e89b-12d3-a456-426614174000"}

	if len(result) != 1 {
		t.Errorf("WithOwnerID() should return map with 1 element, got %d", len(result))
	}

	if result["owner_id"] != expected["owner_id"] {
		t.Errorf("WithOwnerID() = %v, want %v", result, expected)
	}
}

func TestWithRoomID(t *testing.T) {
	roomID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	result := WithRoomID(roomID)
	expected := map[string]interface{}{"room_id": "123e4567-e89b-12d3-a456-426614174000"}

	if len(result) != 1 {
		t.Errorf("WithRoomID() should return map with 1 element, got %d", len(result))
	}

	if result["room_id"] != expected["room_id"] {
		t.Errorf("WithRoomID() = %v, want %v", result, expected)
	}
}

func TestWithFriendID(t *testing.T) {
	friendID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	result := WithFriendID(friendID)
	expected := map[string]interface{}{"friend_id": "123e4567-e89b-12d3-a456-426614174000"}

	if len(result) != 1 {
		t.Errorf("WithFriendID() should return map with 1 element, got %d", len(result))
	}

	if result["friend_id"] != expected["friend_id"] {
		t.Errorf("WithFriendID() = %v, want %v", result, expected)
	}
}

func TestWithCategoryID(t *testing.T) {
	categoryID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	result := WithCategoryID(categoryID)
	expected := map[string]interface{}{"category_id": "123e4567-e89b-12d3-a456-426614174000"}

	if len(result) != 1 {
		t.Errorf("WithCategoryID() should return map with 1 element, got %d", len(result))
	}

	if result["category_id"] != expected["category_id"] {
		t.Errorf("WithCategoryID() = %v, want %v", result, expected)
	}
}

func TestWithLanguageCode(t *testing.T) {
	result := WithLanguageCode("en")
	expected := map[string]interface{}{"language_code": "en"}

	if len(result) != 1 {
		t.Errorf("WithLanguageCode() should return map with 1 element, got %d", len(result))
	}

	if result["language_code"] != expected["language_code"] {
		t.Errorf("WithLanguageCode() = %v, want %v", result, expected)
	}
}

func TestFilterBuildingCombinations(t *testing.T) {
	// Test combining multiple filter builders
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	roomID := uuid.MustParse("223e4567-e89b-12d3-a456-426614174000")

	result := BuildFilter(
		WithUserID(userID),
		WithRoomID(roomID),
		WithStatus("active"),
	)

	if len(result) != 3 {
		t.Errorf("Combined filter should have 3 elements, got %d", len(result))
	}

	if result["user_id"] != userID.String() {
		t.Errorf("user_id not set correctly")
	}

	if result["room_id"] != roomID.String() {
		t.Errorf("room_id not set correctly")
	}

	if result["status"] != "active" {
		t.Errorf("status not set correctly")
	}
}
