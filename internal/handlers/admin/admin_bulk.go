package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// BulkDeleteUsersHandler deletes multiple users at once
func (ah *AdminAPIHandler) BulkDeleteUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Debug: log all form values
	log.Printf("Form values: %+v", r.Form)

	userIDs := r.Form["user_ids[]"]
	log.Printf("Received %d user IDs: %v", len(userIDs), userIDs)

	if len(userIDs) == 0 {
		log.Printf("No users selected for bulk delete")
		http.Error(w, "No users selected", http.StatusBadRequest)
		return
	}

	deletedCount := 0
	for _, idStr := range userIDs {
		userID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid user ID: %s", idStr)
			continue
		}

		if err := ah.adminService.DeleteUser(ctx, userID); err != nil {
			log.Printf("Error deleting user %s: %v", idStr, err)
			continue
		}
		deletedCount++
		log.Printf("Deleted user: %s", idStr)
	}

	log.Printf("✅ Bulk deleted %d users out of %d requested", deletedCount, len(userIDs))

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// BulkDeleteQuestionsHandler deletes multiple questions at once
func (ah *AdminAPIHandler) BulkDeleteQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	questionIDs := r.Form["question_ids[]"]
	if len(questionIDs) == 0 {
		http.Error(w, "No questions selected", http.StatusBadRequest)
		return
	}

	deletedCount := 0
	for _, idStr := range questionIDs {
		questionID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid question ID: %s", idStr)
			continue
		}

		if err := ah.questionService.DeleteQuestion(ctx, questionID); err != nil {
			log.Printf("Error deleting question %s: %v", idStr, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Bulk deleted %d questions", deletedCount)

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}

// BulkDeleteCategoriesHandler deletes multiple categories at once
func (ah *AdminAPIHandler) BulkDeleteCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryIDs := r.Form["category_ids[]"]
	if len(categoryIDs) == 0 {
		http.Error(w, "No categories selected", http.StatusBadRequest)
		return
	}

	deletedCount := 0
	for _, idStr := range categoryIDs {
		categoryID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid category ID: %s", idStr)
			continue
		}

		if err := ah.questionService.DeleteCategory(ctx, categoryID); err != nil {
			log.Printf("Error deleting category %s: %v", idStr, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Bulk deleted %d categories", deletedCount)

	// Return updated categories list
	ah.ListCategoriesHandler(w, r)
}

// BulkCloseRoomsHandler closes multiple rooms at once
func (ah *AdminAPIHandler) BulkCloseRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	roomIDs := r.Form["room_ids[]"]
	if len(roomIDs) == 0 {
		http.Error(w, "No rooms selected", http.StatusBadRequest)
		return
	}

	closedCount := 0
	for _, idStr := range roomIDs {
		roomID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid room ID: %s", idStr)
			continue
		}

		if err := ah.adminService.ForceCloseRoom(ctx, roomID); err != nil {
			log.Printf("Error closing room %s: %v", idStr, err)
			continue
		}
		closedCount++
	}

	log.Printf("Bulk closed %d rooms", closedCount)

	// Return updated rooms list
	ah.ListRoomsHandler(w, r)
}
