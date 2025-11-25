package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/models"
	"github.com/hekigan/couples/internal/services"
)

// ListCategoriesHandler returns an HTML fragment with the categories list
func (ah *AdminAPIHandler) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper for pagination
	page, perPage, offset := handlers.ParsePaginationParams(r)

	categories, err := ah.questionService.ListCategories(ctx, perPage, offset)
	if err != nil {
		http.Error(w, "Failed to list categories", http.StatusInternalServerError)
		log.Printf("Error listing categories: %v", err)
		return
	}

	totalCount, _ := ah.questionService.GetCategoryCount(ctx)

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

	html, err := ah.handler.TemplateService.RenderFragment("categories_list", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering categories list: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetCategoryCreateFormHandler returns an HTML form for creating a new category
func (ah *AdminAPIHandler) GetCategoryCreateFormHandler(w http.ResponseWriter, r *http.Request) {
	// Create empty form data for create mode
	data := services.CategoryFormData{
		ID:    "", // Empty ID indicates create mode
		Key:   "",
		Label: "",
	}

	html, err := ah.handler.TemplateService.RenderFragment("category_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering category create form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetCategoryEditFormHandler returns an HTML form for editing a category
func (ah *AdminAPIHandler) GetCategoryEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	categoryID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := ah.questionService.GetCategoryByID(ctx, categoryID)
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	data := services.CategoryFormData{
		ID:    category.ID.String(),
		Key:   category.Key,
		Label: category.Label,
	}

	html, err := ah.handler.TemplateService.RenderFragment("category_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering category edit form: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// UpdateCategoryHandler updates a category
func (ah *AdminAPIHandler) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	categoryID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	category := &models.Category{
		ID:    categoryID,
		Key:   r.FormValue("key"),
		Label: r.FormValue("label"),
	}

	if err := ah.questionService.UpdateCategory(ctx, category); err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		log.Printf("Error updating category: %v", err)
		return
	}

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}

// CreateCategoryHandler creates a new category
func (ah *AdminAPIHandler) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	category := &models.Category{
		Key:   r.FormValue("key"),
		Label: r.FormValue("label"),
	}

	if err := ah.questionService.CreateCategory(ctx, category); err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		log.Printf("Error creating category: %v", err)
		return
	}

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}

// DeleteCategoryHandler deletes a category
func (ah *AdminAPIHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Use helper to extract ID
	vars := mux.Vars(r)
	categoryID, err := handlers.ExtractIDFromRoute(vars, "id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ah.questionService.DeleteCategory(ctx, categoryID); err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		log.Printf("Error deleting category: %v", err)
		return
	}

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}
