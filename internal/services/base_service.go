package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	supabase "github.com/supabase-community/supabase-go"
)

// BaseService provides common database query patterns
// Eliminates 61+ query pattern duplications across all services
// Uses interface{} instead of generics for Go <1.18 compatibility
type BaseService struct {
	client *supabase.Client
	logger *ServiceLogger
}

// NewBaseService creates a new base service instance
func NewBaseService(client *supabase.Client, serviceName string) *BaseService {
	return &BaseService{
		client: client,
		logger: NewServiceLogger(serviceName),
	}
}

// GetSingleRecord fetches a single record by ID
// Caller must pass a pointer to the struct where the result will be unmarshaled
// Example usage:
//   var room models.Room
//   err := base.GetSingleRecord(ctx, "rooms", roomID, &room)
func (b *BaseService) GetSingleRecord(
	ctx context.Context,
	table string,
	id uuid.UUID,
	result interface{},
) error {
	data, _, err := b.client.From(table).
		Select("*", "", false).
		Eq("id", id.String()).
		Single().
		Execute()

	if err != nil {
		return fmt.Errorf("failed to fetch %s by id %s: %w", table, id.String(), err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to parse %s data: %w", table, err)
	}

	return nil
}

// GetRecords fetches multiple records with filters
// Caller must pass a pointer to a slice where results will be unmarshaled
// Example usage:
//   var rooms []models.Room
//   filters := map[string]interface{}{"owner_id": userID.String(), "status": "active"}
//   err := base.GetRecords(ctx, "rooms", filters, &rooms)
func (b *BaseService) GetRecords(
	ctx context.Context,
	table string,
	filters map[string]interface{},
	result interface{},
) error {
	query := b.client.From(table).Select("*", "", false)

	// Apply filters
	for key, value := range filters {
		// Convert value to string for Eq method
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case fmt.Stringer:
			strValue = v.String()
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		query = query.Eq(key, strValue)
	}

	data, _, err := query.Execute()
	if err != nil {
		return fmt.Errorf("failed to fetch %s records: %w", table, err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to parse %s data: %w", table, err)
	}

	return nil
}

// GetRecordsWithLimit fetches multiple records with filters and pagination
// Example usage:
//   var rooms []models.Room
//   filters := map[string]interface{}{"status": "active"}
//   err := base.GetRecordsWithLimit(ctx, "rooms", filters, 25, 0, &rooms)
func (b *BaseService) GetRecordsWithLimit(
	ctx context.Context,
	table string,
	filters map[string]interface{},
	limit int,
	offset int,
	result interface{},
) error {
	query := b.client.From(table).Select("*", "", false)

	// Apply filters
	for key, value := range filters {
		// Convert value to string for Eq method
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case fmt.Stringer:
			strValue = v.String()
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		query = query.Eq(key, strValue)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit, "")
	}
	if offset > 0 {
		query = query.Range(offset, offset+limit-1, "")
	}

	data, _, err := query.Execute()
	if err != nil {
		return fmt.Errorf("failed to fetch %s records with limit: %w", table, err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to parse %s data: %w", table, err)
	}

	return nil
}

// InsertRecord inserts a new record
// Example usage:
//   data := map[string]interface{}{
//       "id": uuid.New().String(),
//       "name": "Test Room",
//       "owner_id": userID.String(),
//   }
//   err := base.InsertRecord(ctx, "rooms", data)
func (b *BaseService) InsertRecord(
	ctx context.Context,
	table string,
	data map[string]interface{},
) error {
	_, _, err := b.client.From(table).
		Insert(data, false, "", "", "").
		Execute()

	if err != nil {
		return fmt.Errorf("failed to insert into %s: %w", table, err)
	}

	return nil
}

// UpdateRecord updates a record by ID
// Example usage:
//   data := map[string]interface{}{
//       "status": "completed",
//       "updated_at": time.Now(),
//   }
//   err := base.UpdateRecord(ctx, "rooms", roomID, data)
func (b *BaseService) UpdateRecord(
	ctx context.Context,
	table string,
	id uuid.UUID,
	data map[string]interface{},
) error {
	_, _, err := b.client.From(table).
		Update(data, "", "").
		Eq("id", id.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to update %s by id %s: %w", table, id.String(), err)
	}

	return nil
}

// UpdateRecordsWithFilter updates multiple records matching filter
// Example usage:
//   data := map[string]interface{}{"status": "cancelled"}
//   filters := map[string]interface{}{"owner_id": userID.String()}
//   err := base.UpdateRecordsWithFilter(ctx, "rooms", filters, data)
func (b *BaseService) UpdateRecordsWithFilter(
	ctx context.Context,
	table string,
	filters map[string]interface{},
	data map[string]interface{},
) error {
	query := b.client.From(table).Update(data, "", "")

	// Apply filters
	for key, value := range filters {
		// Convert value to string for Eq method
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case fmt.Stringer:
			strValue = v.String()
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		query = query.Eq(key, strValue)
	}

	_, _, err := query.Execute()
	if err != nil {
		return fmt.Errorf("failed to update %s records: %w", table, err)
	}

	return nil
}

// DeleteRecord deletes a record by ID
// Example usage:
//   err := base.DeleteRecord(ctx, "rooms", roomID)
func (b *BaseService) DeleteRecord(
	ctx context.Context,
	table string,
	id uuid.UUID,
) error {
	_, _, err := b.client.From(table).
		Delete("", "").
		Eq("id", id.String()).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to delete %s by id %s: %w", table, id.String(), err)
	}

	return nil
}

// DeleteRecordsWithFilter deletes multiple records matching filter
// Example usage:
//   filters := map[string]interface{}{"user_id": userID.String(), "status": "pending"}
//   err := base.DeleteRecordsWithFilter(ctx, "notifications", filters)
func (b *BaseService) DeleteRecordsWithFilter(
	ctx context.Context,
	table string,
	filters map[string]interface{},
) error {
	query := b.client.From(table).Delete("", "")

	// Apply filters
	for key, value := range filters {
		// Convert value to string for Eq method
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case fmt.Stringer:
			strValue = v.String()
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		query = query.Eq(key, strValue)
	}

	_, _, err := query.Execute()
	if err != nil {
		return fmt.Errorf("failed to delete %s records: %w", table, err)
	}

	return nil
}

// CountRecords counts records matching filters
// Example usage:
//   filters := map[string]interface{}{"status": "active"}
//   count, err := base.CountRecords(ctx, "rooms", filters)
func (b *BaseService) CountRecords(
	ctx context.Context,
	table string,
	filters map[string]interface{},
) (int, error) {
	query := b.client.From(table).Select("id", "exact", false)

	// Apply filters
	for key, value := range filters {
		// Convert value to string for Eq method
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case fmt.Stringer:
			strValue = v.String()
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		query = query.Eq(key, strValue)
	}

	_, count, err := query.Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to count %s records: %w", table, err)
	}

	return int(count), nil
}
