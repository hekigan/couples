package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// UpdateCategoriesAPIHandler updates selected categories for a room
func (h *Handler) UpdateCategoriesAPIHandler(c echo.Context) error {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	// Parse category IDs from request (multipart form data from FormData)
	if err := c.Request().ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse form")
	}

	categoryIDStrings := c.Request().MultipartForm.Value["category_ids"]
	categoryIDs := make([]uuid.UUID, 0, len(categoryIDStrings))
	for _, idStr := range categoryIDStrings {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue // Skip invalid UUIDs
		}
		categoryIDs = append(categoryIDs, id)
	}

	// Update room with selected categories
	room.SelectedCategories = categoryIDs
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update categories: "+err.Error())
	}

	// Broadcast category update to all room participants via SSE
	h.RoomService.BroadcastCategoriesUpdated(roomID, categoryIDs)

	// Return JSON response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Categories updated",
	})
}

// GetCategoriesAPIHandler returns all available categories
func (h *Handler) GetCategoriesAPIHandler(c echo.Context) error {
	ctx := context.Background()

	categories, err := h.CategoryService.GetCategories(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch categories")
	}

	// Build response structure
	type CategoryResponse struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}

	categoryList := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		categoryList[i] = CategoryResponse{
			ID:  cat.ID.String(),
			Key: cat.Key,
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"categories": categoryList,
	})
}

// GetRoomCategoriesHTMLHandler returns room categories as HTML fragment (for HTMX)
func (h *Handler) GetRoomCategoriesHTMLHandler(c echo.Context) error {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	ctx := context.Background()

	// Get all categories
	allCategories, err := h.CategoryService.GetCategories(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		if err := h.RenderHTMLFragment(c, "error_message.html", "Failed to load categories"); err != nil {
			return c.HTML(http.StatusOK, `<p style="color: #6b7280;">Failed to load categories</p>`)
		}
		return nil
	}

	// Get question counts per category for the room's language
	questionCounts, err := h.QuestionService.GetQuestionCountsByCategory(ctx, room.Language)
	if err != nil {
		log.Printf("Error fetching question counts: %v", err)
		// Continue without counts - they'll just be 0
		questionCounts = make(map[string]int)
	}

	// Build CategoryInfo array with selection state and question counts
	categoryInfos := make([]services.CategoryInfo, 0, len(allCategories))
	for _, cat := range allCategories {
		isSelected := false
		// Check if this category is in room's selected categories
		for _, selectedID := range room.SelectedCategories {
			if selectedID == cat.ID {
				isSelected = true
				break
			}
		}

		categoryInfos = append(categoryInfos, services.CategoryInfo{
			ID:            cat.ID.String(),
			Key:           cat.Key,
			Label:         cat.Label,
			IsSelected:    isSelected,
			QuestionCount: questionCounts[cat.ID.String()],
		})
	}

	// Use helper to render HTML fragment
	if err := h.RenderHTMLFragment(c, "categories_grid.html", services.CategoriesGridData{
		Categories: categoryInfos,
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	}); err != nil {
		log.Printf("Error rendering categories grid template: %v", err)
		return c.HTML(http.StatusOK, `<p style="color: #6b7280;">Failed to load categories</p>`)
	}
	return nil
}

// ToggleCategoryAPIHandler toggles a single category selection (for HTMX)
func (h *Handler) ToggleCategoryAPIHandler(c echo.Context) error {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	ctx := context.Background()

	// Get category ID from form
	categoryIDStr := c.FormValue("category_id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
	}

	// Toggle category - if exists, remove it; if not, add it
	selectedCategories := room.SelectedCategories
	found := false
	newCategories := make([]uuid.UUID, 0, len(selectedCategories))
	for _, id := range selectedCategories {
		if id == categoryID {
			found = true
			// Don't add it (remove it)
		} else {
			newCategories = append(newCategories, id)
		}
	}

	if !found {
		// Add it
		newCategories = append(newCategories, categoryID)
	}

	// Update room
	room.SelectedCategories = newCategories
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("Failed to update categories: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update categories")
	}

	// Broadcast categories update via SSE
	h.RoomService.BroadcastCategoriesUpdated(roomID, newCategories)

	// Return success (HTMX will handle via hx-swap="none")
	return c.HTML(http.StatusOK, `<!-- Category toggled successfully -->`)
}
