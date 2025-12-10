package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
	adminFragments "github.com/hekigan/couples/internal/views/fragments/admin"
	"github.com/labstack/echo/v4"
)

// ListCategoriesHandler returns an HTML fragment with the categories list
func (ah *AdminAPIHandler) ListCategoriesHandler(c echo.Context) error {
	ctx := context.Background()

	// Use helper for pagination
	page, perPage := handlers.ParsePaginationParams(c)
	offset := (page - 1) * perPage

	categories, err := ah.categoryService.ListCategories(ctx, perPage, offset)
	if err != nil {
		log.Printf("Error listing categories: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list categories")
	}

	totalCount, _ := ah.categoryService.GetCategoryCount(ctx)

	// Calculate pagination
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// Get question counts per category
	counts, _ := ah.questionService.GetQuestionCountsByCategory(ctx, "en")

	// Build data for template
	categoryInfos := make([]services.AdminCategoryInfo, len(categories))
	for i, cat := range categories {
		count := 0
		if c, ok := counts[cat.ID.String()]; ok {
			count = c
		}

		categoryInfos[i] = services.AdminCategoryInfo{
			ID:            cat.ID.String(),
			Label:         cat.Label,
			Key:           cat.Key,
			QuestionCount: count,
		}
	}

	data := services.CategoriesListData{
		Categories: categoryInfos,
		// Pagination fields
		TotalCount:      totalCount,
		CurrentPage:     page,
		TotalPages:      totalPages,
		ItemsPerPage:    perPage,
		BaseURL:         "/admin/api/categories/list",
		PageURL:         "/admin/categories",
		Target:          "#categories-list",
		IncludeSelector: "",
		ExtraParams:     "",
		ItemName:        "categories",
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.CategoriesList(&data))
	if err != nil {
		log.Printf("Error rendering categories list: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// GetCategoryCreateFormHandler returns an HTML form for creating a new category
func (ah *AdminAPIHandler) GetCategoryCreateFormHandler(c echo.Context) error {
	// Create empty form data for create mode
	data := services.CategoryFormData{
		ID:    "", // Empty ID indicates create mode
		Key:   "",
		Label: "",
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.CategoryForm(&data))
	if err != nil {
		log.Printf("Error rendering category create form: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// GetCategoryEditFormHandler returns an HTML form for editing a category
func (ah *AdminAPIHandler) GetCategoryEditFormHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	categoryID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	category, err := ah.categoryService.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}

	data := services.CategoryFormData{
		ID:    category.ID.String(),
		Key:   category.Key,
		Label: category.Label,
	}

	html, err := ah.handler.RenderTemplFragment(c, adminFragments.CategoryForm(&data))
	if err != nil {
		log.Printf("Error rendering category edit form: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}

// UpdateCategoryHandler updates a category
func (ah *AdminAPIHandler) UpdateCategoryHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	categoryID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	category := &models.Category{
		ID:    categoryID,
		Key:   c.FormValue("key"),
		Label: c.FormValue("label"),
	}

	if err := ah.categoryService.UpdateCategory(ctx, category); err != nil {
		log.Printf("Error updating category: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update category")
	}

	// Return updated categories list
	return ah.ListCategoriesHandler(c)
}

// CreateCategoryHandler creates a new category
func (ah *AdminAPIHandler) CreateCategoryHandler(c echo.Context) error {
	ctx := context.Background()

	category := &models.Category{
		Key:   c.FormValue("key"),
		Label: c.FormValue("label"),
	}

	if err := ah.categoryService.CreateCategory(ctx, category); err != nil {
		log.Printf("Error creating category: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create category")
	}

	// Return updated categories list
	return ah.ListCategoriesHandler(c)
}

// DeleteCategoryHandler deletes a category
func (ah *AdminAPIHandler) DeleteCategoryHandler(c echo.Context) error {
	ctx := context.Background()

	// Extract ID from route
	categoryID, err := handlers.ExtractIDFromParam(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := ah.categoryService.DeleteCategory(ctx, categoryID); err != nil {
		log.Printf("Error deleting category: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete category")
	}

	// Return updated categories list
	return ah.ListCategoriesHandler(c)
}
