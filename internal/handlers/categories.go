package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
)

// UpdateCategoriesAPIHandler updates selected categories for a room
func (h *Handler) UpdateCategoriesAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room and verify participation
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Use helper to verify participant
	if err := h.VerifyRoomParticipant(room, userID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Parse category IDs from request (multipart form data from FormData)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	categoryIDStrings := r.MultipartForm.Value["category_ids"]
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
		http.Error(w, "Failed to update categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast category update to all room participants via SSE
	h.RoomService.BroadcastCategoriesUpdated(roomID, categoryIDs)

	// Use helper to write JSON response
	err = WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Categories updated",
	})
	if err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

// GetCategoriesAPIHandler returns all available categories
func (h *Handler) GetCategoriesAPIHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	categories, err := h.CategoryService.GetCategories(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	// Manual JSON encoding for simplicity
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonStr := `{"categories":[`
	for i, cat := range categories {
		if i > 0 {
			jsonStr += ","
		}
		jsonStr += `{"id":"` + cat.ID.String() + `","key":"` + cat.Key + `"}`
	}
	jsonStr += `]}`

	w.Write([]byte(jsonStr))
}

// GetRoomCategoriesHTMLHandler returns room categories as HTML fragment (for HTMX)
func (h *Handler) GetRoomCategoriesHTMLHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()

	// Get all categories
	allCategories, err := h.CategoryService.GetCategories(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		if err := h.RenderHTMLFragment(w, "error_message.html", "Failed to load categories"); err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<p style="color: #6b7280;">Failed to load categories</p>`))
		}
		return
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
	if err := h.RenderHTMLFragment(w, "categories_grid.html", services.CategoriesGridData{
		Categories: categoryInfos,
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	}); err != nil {
		log.Printf("Error rendering categories grid template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load categories</p>`))
	}
}

// ToggleCategoryAPIHandler toggles a single category selection (for HTMX)
func (h *Handler) ToggleCategoryAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()

	// Get category ID from form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryIDStr := r.FormValue("category_id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
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
		http.Error(w, "Failed to update categories", http.StatusInternalServerError)
		return
	}

	// Broadcast categories update via SSE
	h.RoomService.BroadcastCategoriesUpdated(roomID, newCategories)

	// Return success (HTMX will handle via hx-swap="none")
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!-- Category toggled successfully -->`))
}
