package services

import (
	"github.com/google/uuid"
)

// ToStringSlice converts a slice of UUIDs to a slice of strings
// Eliminates 69 UUID conversion duplications across services
// Example usage:
//   categoryIDStrings := ToStringSlice(categoryIDs)
//   query.In("category_id", categoryIDStrings)
func ToStringSlice(uuids []uuid.UUID) []string {
	if len(uuids) == 0 {
		return []string{}
	}

	result := make([]string, len(uuids))
	for i, id := range uuids {
		result[i] = id.String()
	}
	return result
}

// UUIDToStringOrNil converts a UUID pointer to either a string or nil
// Used for optional UUID fields in database operations
// Example usage:
//   data["guest_id"] = UUIDToStringOrNil(room.GuestID)
func UUIDToStringOrNil(id *uuid.UUID) interface{} {
	if id == nil {
		return nil
	}
	return id.String()
}

// BuildFilter creates a filter map from key-value pairs
// Merges multiple filter maps into one
// Example usage:
//   filters := BuildFilter(
//       WithStatus("active"),
//       WithUserID(userID),
//   )
func BuildFilter(filters ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, f := range filters {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}

// WithStatus creates a status filter
// Example usage:
//   filters := BuildFilter(WithStatus("pending"))
func WithStatus(status string) map[string]interface{} {
	return map[string]interface{}{"status": status}
}

// WithUserID creates a user_id filter
// Example usage:
//   filters := BuildFilter(WithUserID(userID))
func WithUserID(userID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{"user_id": userID.String()}
}

// WithOwnerID creates an owner_id filter
// Example usage:
//   filters := BuildFilter(WithOwnerID(ownerID))
func WithOwnerID(ownerID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{"owner_id": ownerID.String()}
}

// WithRoomID creates a room_id filter
// Example usage:
//   filters := BuildFilter(WithRoomID(roomID))
func WithRoomID(roomID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{"room_id": roomID.String()}
}

// WithFriendID creates a friend_id filter
// Example usage:
//   filters := BuildFilter(WithFriendID(friendID))
func WithFriendID(friendID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{"friend_id": friendID.String()}
}

// WithCategoryID creates a category_id filter
// Example usage:
//   filters := BuildFilter(WithCategoryID(categoryID))
func WithCategoryID(categoryID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{"category_id": categoryID.String()}
}

// WithLanguageCode creates a language_code filter
// Example usage:
//   filters := BuildFilter(WithLanguageCode("en"))
func WithLanguageCode(langCode string) map[string]interface{} {
	return map[string]interface{}{"language_code": langCode}
}
