package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"github.com/hekigan/couples/internal/models"
)

// CategoryService handles category-related operations
type CategoryService struct {
	*BaseService
	client *supabase.Client
}

// NewCategoryService creates a new category service
func NewCategoryService(client *supabase.Client) *CategoryService {
	return &CategoryService{
		BaseService: NewBaseService(client, "CategoryService"),
		client:      client,
	}
}

// GetCategories retrieves all categories
func (s *CategoryService) GetCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	if err := s.BaseService.GetRecords(ctx, "categories", map[string]interface{}{}, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

// ListCategories retrieves paginated categories
func (s *CategoryService) ListCategories(ctx context.Context, limit, offset int) ([]models.Category, error) {
	var categories []models.Category
	if err := s.BaseService.GetRecordsWithLimit(ctx, "categories", map[string]interface{}{}, limit, offset, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

// GetCategoryCount returns the total number of categories
func (s *CategoryService) GetCategoryCount(ctx context.Context) (int, error) {
	return s.BaseService.CountRecords(ctx, "categories", map[string]interface{}{})
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, category *models.Category) error {
	categoryMap := map[string]interface{}{
		"key":   category.Key,
		"label": category.Label,
	}

	return s.BaseService.InsertRecord(ctx, "categories", categoryMap)
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(ctx context.Context, category *models.Category) error {
	categoryMap := map[string]interface{}{
		"key":   category.Key,
		"label": category.Label,
	}

	return s.BaseService.UpdateRecord(ctx, "categories", category.ID, categoryMap)
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.BaseService.DeleteRecord(ctx, "categories", id)
}

// GetCategoryByID retrieves a category by ID
func (s *CategoryService) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	var category models.Category
	if err := s.BaseService.GetSingleRecord(ctx, "categories", id, &category); err != nil {
		return nil, err
	}
	return &category, nil
}
