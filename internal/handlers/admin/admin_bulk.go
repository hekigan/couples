package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BulkDeleteUsersHandler deletes multiple users at once
func (ah *AdminAPIHandler) BulkDeleteUsersHandler(c echo.Context) error {
	ctx := context.Background()

	// Get form values - Echo automatically parses the request
	formParams, err := c.FormParams()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// Debug: log all form values
	log.Printf("Form values: %+v", formParams)

	userIDs := formParams["user_ids[]"]
	log.Printf("Received %d user IDs: %v", len(userIDs), userIDs)

	if len(userIDs) == 0 {
		log.Printf("No users selected for bulk delete")
		return echo.NewHTTPError(http.StatusBadRequest, "No users selected")
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
	return ah.ListUsersHandler(c)
}

// BulkDeleteQuestionsHandler deletes multiple questions at once
func (ah *AdminAPIHandler) BulkDeleteQuestionsHandler(c echo.Context) error {
	ctx := context.Background()

	formParams, err := c.FormParams()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	questionIDs := formParams["question_ids[]"]
	if len(questionIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No questions selected")
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
	return ah.ListQuestionsHandler(c)
}

// BulkDeleteCategoriesHandler deletes multiple categories at once
func (ah *AdminAPIHandler) BulkDeleteCategoriesHandler(c echo.Context) error {
	ctx := context.Background()

	formParams, err := c.FormParams()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	categoryIDs := formParams["category_ids[]"]
	if len(categoryIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No categories selected")
	}

	deletedCount := 0
	for _, idStr := range categoryIDs {
		categoryID, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid category ID: %s", idStr)
			continue
		}

		if err := ah.categoryService.DeleteCategory(ctx, categoryID); err != nil {
			log.Printf("Error deleting category %s: %v", idStr, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Bulk deleted %d categories", deletedCount)

	// Return updated categories list
	return ah.ListCategoriesHandler(c)
}

// BulkCloseRoomsHandler closes multiple rooms at once
func (ah *AdminAPIHandler) BulkCloseRoomsHandler(c echo.Context) error {
	ctx := context.Background()

	formParams, err := c.FormParams()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	roomIDs := formParams["room_ids[]"]
	if len(roomIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No rooms selected")
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
	return ah.ListRoomsHandler(c)
}
