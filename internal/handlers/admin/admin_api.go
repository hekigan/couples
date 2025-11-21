package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/handlers"
	"github.com/yourusername/couple-card-game/internal/models"
	"github.com/yourusername/couple-card-game/internal/services"
)

// AdminAPIHandler handles admin API requests
type AdminAPIHandler struct {
	handler        *handlers.Handler
	adminService   *services.AdminService
	questionService *services.QuestionService
}

// NewAdminAPIHandler creates a new admin API handler
func NewAdminAPIHandler(h *handlers.Handler, adminService *services.AdminService, questionService *services.QuestionService) *AdminAPIHandler {
	return &AdminAPIHandler{
		handler:        h,
		adminService:   adminService,
		questionService: questionService,
	}
}

// ListUsersHandler returns an HTML fragment with the users list
func (ah *AdminAPIHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse pagination parameters
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 50
	offset := (page - 1) * limit

	users, err := ah.adminService.ListAllUsers(ctx, limit, offset)
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		log.Printf("Error listing users: %v", err)
		return
	}

	totalCount, _ := ah.adminService.GetUserCount(ctx)

	// Render HTML fragment
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div id="users-list">
	<form id="bulk-users-form">
		<!-- Bulk Actions Bar (Top) -->
		<div class="bulk-actions-bar">
			<input type="checkbox" id="select-all-users-top" onclick="toggleAllCheckboxes(this, 'user-checkbox')" />
			<label for="select-all-users-top">Select All</label>
			<select name="bulk_action" id="bulk-action-users">
				<option value="">Bulk Actions...</option>
				<option value="delete">Delete Selected</option>
			</select>
			<button type="button" onclick="submitBulkAction('users')" class="btn btn-sm">Apply</button>
		</div>

	<table>
		<thead>
			<tr>
				<th class="checkbox-col"></th>
				<th>Username</th>
				<th>Email</th>
				<th>Type</th>
				<th>Admin</th>
				<th>Created</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>`)

	for _, user := range users {
		email := ""
		if user.Email != nil {
			email = html.EscapeString(*user.Email)
		}

		userType := "Registered"
		if user.IsAnonymous {
			userType = "Anonymous"
		}

		adminBadge := ""
		if user.IsAdmin {
			adminBadge = `<span class="admin-check">✓</span>`
		}

		fmt.Fprintf(w, `
			<tr>
				<td><input type="checkbox" class="user-checkbox" name="user_ids[]" value="%s" /></td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>
					<button hx-post="/admin/api/users/%s/toggle-admin"
					        hx-target="#users-list"
					        hx-swap="outerHTML"
					        class="btn btn-sm">
						Toggle Admin
					</button>
					<button hx-delete="/admin/api/users/%s"
					        hx-target="#users-list"
					        hx-swap="outerHTML"
					        hx-confirm="Are you sure you want to delete this user?"
					        class="btn btn-sm btn-danger">
						Delete
					</button>
				</td>
			</tr>`,
			user.ID.String(),
			html.EscapeString(user.Username),
			email,
			userType,
			adminBadge,
			user.CreatedAt.Format("2006-01-02"),
			user.ID.String(),
			user.ID.String(),
		)
	}

	fmt.Fprintf(w, `
		</tbody>
	</table>

		<!-- Bulk Actions Bar (Bottom) -->
		<div class="bulk-actions-bar bulk-actions-bottom">
			<input type="checkbox" id="select-all-users-bottom" onclick="toggleAllCheckboxes(this, 'user-checkbox')" />
			<label for="select-all-users-bottom">Select All</label>
			<select name="bulk_action_bottom">
				<option value="">Bulk Actions...</option>
				<option value="delete">Delete Selected</option>
			</select>
			<button type="button" onclick="submitBulkAction('users')" class="btn btn-sm">Apply</button>
			<span class="item-count">Total: %d users (Page %d)</span>
		</div>
	</form>
</div>`, totalCount, page)
}

// ToggleUserAdminHandler toggles a user's admin status
func (ah *AdminAPIHandler) ToggleUserAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := ah.adminService.ToggleUserAdmin(ctx, userID); err != nil {
		http.Error(w, "Failed to toggle admin status", http.StatusInternalServerError)
		log.Printf("Error toggling admin status: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// DeleteUserHandler deletes a user
func (ah *AdminAPIHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := ah.adminService.DeleteUser(ctx, userID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		log.Printf("Error deleting user: %v", err)
		return
	}

	// Return updated users list
	ah.ListUsersHandler(w, r)
}

// ListQuestionsHandler returns an HTML fragment with the questions list
func (ah *AdminAPIHandler) ListQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse filters
	var categoryID *uuid.UUID
	if catID := r.URL.Query().Get("category_id"); catID != "" {
		if parsed, err := uuid.Parse(catID); err == nil {
			categoryID = &parsed
		}
	}

	var langCode *string
	if lang := r.URL.Query().Get("lang_code"); lang != "" {
		langCode = &lang
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 50
	offset := (page - 1) * limit

	questions, err := ah.questionService.ListQuestions(ctx, limit, offset, categoryID, langCode)
	if err != nil {
		http.Error(w, "Failed to list questions", http.StatusInternalServerError)
		log.Printf("Error listing questions: %v", err)
		return
	}

	// Get categories for dropdown
	categories, _ := ah.questionService.GetCategories(ctx)

	// Render HTML fragment
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div id="questions-list">
	<div class="filters">
		<select hx-get="/admin/api/questions/list"
		        hx-target="#questions-list"
		        hx-swap="outerHTML"
		        hx-include="[name='lang_code']"
		        name="category_id">
			<option value="">All Categories</option>`)

	for _, cat := range categories {
		selected := ""
		if categoryID != nil && *categoryID == cat.ID {
			selected = "selected"
		}
		fmt.Fprintf(w, `<option value="%s" %s>%s %s</option>`,
			cat.ID.String(), selected, cat.Icon, html.EscapeString(cat.Label))
	}

	selectedLang := ""
	if langCode != nil {
		selectedLang = *langCode
	}

	fmt.Fprintf(w, `
		</select>
		<select hx-get="/admin/api/questions/list"
		        hx-target="#questions-list"
		        hx-swap="outerHTML"
		        hx-include="[name='category_id']"
		        name="lang_code">
			<option value="">All Languages</option>
			<option value="en" %s>English</option>
			<option value="fr" %s>French</option>
			<option value="ja" %s>Japanese</option>
		</select>
	</div>
	<table>
		<thead>
			<tr>
				<th>Question Text</th>
				<th>Category</th>
				<th>Language</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>`,
		tern(selectedLang == "en", "selected", ""),
		tern(selectedLang == "fr", "selected", ""),
		tern(selectedLang == "ja", "selected", ""),
	)

	// Create category map for lookups
	categoryMap := make(map[uuid.UUID]models.Category)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}

	for _, q := range questions {
		categoryLabel := "Unknown"
		if cat, ok := categoryMap[q.CategoryID]; ok {
			categoryLabel = fmt.Sprintf("%s %s", cat.Icon, cat.Label)
		}

		fmt.Fprintf(w, `
			<tr>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>
					<button hx-get="/admin/api/questions/%s/edit-form"
					        hx-target="#edit-modal-content"
					        hx-swap="innerHTML"
					        onclick="document.getElementById('edit-modal').style.display='block'"
					        class="btn btn-sm">
						Edit
					</button>
					<button hx-delete="/admin/api/questions/%s"
					        hx-target="#questions-list"
					        hx-swap="outerHTML"
					        hx-confirm="Are you sure you want to delete this question?"
					        class="btn btn-sm btn-danger">
						Delete
					</button>
				</td>
			</tr>`,
			html.EscapeString(q.Text),
			html.EscapeString(categoryLabel),
			html.EscapeString(q.LanguageCode),
			q.ID.String(),
			q.ID.String(),
		)
	}

	fmt.Fprintf(w, `
		</tbody>
	</table>
</div>`)
}

// GetQuestionEditFormHandler returns an HTML form for editing a question
func (ah *AdminAPIHandler) GetQuestionEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	questionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	question, err := ah.questionService.GetQuestionByID(ctx, questionID)
	if err != nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	categories, _ := ah.questionService.GetCategories(ctx)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<form hx-put="/admin/api/questions/%s"
      hx-target="#questions-list"
      hx-swap="outerHTML"
      onclick="document.getElementById('edit-modal').style.display='none'">
	<label>Question Text:</label>
	<textarea name="question_text" required rows="4">%s</textarea>

	<label>Category:</label>
	<select name="category_id" required>`,
		question.ID.String(),
		html.EscapeString(question.Text))

	for _, cat := range categories {
		selected := ""
		if cat.ID == question.CategoryID {
			selected = "selected"
		}
		fmt.Fprintf(w, `<option value="%s" %s>%s %s</option>`,
			cat.ID.String(), selected, cat.Icon, html.EscapeString(cat.Label))
	}

	fmt.Fprintf(w, `
	</select>

	<label>Language:</label>
	<select name="lang_code" required>
		<option value="en" %s>English</option>
		<option value="fr" %s>French</option>
		<option value="ja" %s>Japanese</option>
	</select>

	<button type="submit" class="btn">Save Changes</button>
	<button type="button" class="btn" onclick="document.getElementById('edit-modal').style.display='none'">Cancel</button>
</form>`,
		tern(question.LanguageCode == "en", "selected", ""),
		tern(question.LanguageCode == "fr", "selected", ""),
		tern(question.LanguageCode == "ja", "selected", ""),
	)
}

// UpdateQuestionHandler updates a question
func (ah *AdminAPIHandler) UpdateQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	questionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryID, err := uuid.Parse(r.FormValue("category_id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	question := &models.Question{
		ID:           questionID,
		CategoryID:   categoryID,
		LanguageCode: r.FormValue("lang_code"),
		Text:         r.FormValue("question_text"),
	}

	if err := ah.questionService.UpdateQuestion(ctx, question); err != nil {
		http.Error(w, "Failed to update question", http.StatusInternalServerError)
		log.Printf("Error updating question: %v", err)
		return
	}

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}

// CreateQuestionHandler creates a new question
func (ah *AdminAPIHandler) CreateQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	categoryID, err := uuid.Parse(r.FormValue("category_id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	question := &models.Question{
		CategoryID:   categoryID,
		LanguageCode: r.FormValue("lang_code"),
		Text:         r.FormValue("question_text"),
	}

	if err := ah.questionService.CreateQuestion(ctx, question); err != nil {
		http.Error(w, "Failed to create question", http.StatusInternalServerError)
		log.Printf("Error creating question: %v", err)
		return
	}

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}

// DeleteQuestionHandler deletes a question
func (ah *AdminAPIHandler) DeleteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	questionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	if err := ah.questionService.DeleteQuestion(ctx, questionID); err != nil {
		http.Error(w, "Failed to delete question", http.StatusInternalServerError)
		log.Printf("Error deleting question: %v", err)
		return
	}

	// Return updated questions list
	ah.ListQuestionsHandler(w, r)
}

// ListCategoriesHandler returns an HTML fragment with the categories list
func (ah *AdminAPIHandler) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	categories, err := ah.questionService.GetCategories(ctx)
	if err != nil {
		http.Error(w, "Failed to list categories", http.StatusInternalServerError)
		log.Printf("Error listing categories: %v", err)
		return
	}

	// Get question counts per category
	counts, _ := ah.questionService.GetQuestionCountsByCategory(ctx, "en")

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div id="categories-list">
	<table>
		<thead>
			<tr>
				<th>Icon</th>
				<th>Label</th>
				<th>Key</th>
				<th>Questions</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>`)

	for _, cat := range categories {
		count := 0
		if c, ok := counts[cat.ID.String()]; ok {
			count = c
		}

		fmt.Fprintf(w, `
			<tr>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>%d</td>
				<td>
					<button hx-get="/admin/api/categories/%s/edit-form"
					        hx-target="#edit-modal-content"
					        hx-swap="innerHTML"
					        onclick="document.getElementById('edit-modal').style.display='block'"
					        class="btn btn-sm">
						Edit
					</button>
					<button hx-delete="/admin/api/categories/%s"
					        hx-target="#categories-list"
					        hx-swap="outerHTML"
					        hx-confirm="Are you sure you want to delete this category?"
					        class="btn btn-sm btn-danger">
						Delete
					</button>
				</td>
			</tr>`,
			html.EscapeString(cat.Icon),
			html.EscapeString(cat.Label),
			html.EscapeString(cat.Key),
			count,
			cat.ID.String(),
			cat.ID.String(),
		)
	}

	fmt.Fprintf(w, `
		</tbody>
	</table>
</div>`)
}

// GetCategoryEditFormHandler returns an HTML form for editing a category
func (ah *AdminAPIHandler) GetCategoryEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	categoryID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := ah.questionService.GetCategoryByID(ctx, categoryID)
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<form hx-put="/admin/api/categories/%s"
      hx-target="#categories-list"
      hx-swap="outerHTML"
      onclick="document.getElementById('edit-modal').style.display='none'">
	<label>Key:</label>
	<input type="text" name="key" value="%s" required />

	<label>Label:</label>
	<input type="text" name="label" value="%s" required />

	<label>Icon:</label>
	<input type="text" name="icon" value="%s" required placeholder="e.g., 💬" />

	<button type="submit" class="btn">Save Changes</button>
	<button type="button" class="btn" onclick="document.getElementById('edit-modal').style.display='none'">Cancel</button>
</form>`,
		category.ID.String(),
		html.EscapeString(category.Key),
		html.EscapeString(category.Label),
		html.EscapeString(category.Icon),
	)
}

// UpdateCategoryHandler updates a category
func (ah *AdminAPIHandler) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	categoryID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
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
		Icon:  r.FormValue("icon"),
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
		Icon:  r.FormValue("icon"),
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
	vars := mux.Vars(r)
	categoryID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
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

// ListRoomsHandler returns an HTML fragment with the rooms list
func (ah *AdminAPIHandler) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 50
	offset := (page - 1) * limit

	rooms, err := ah.adminService.ListAllRooms(ctx, limit, offset)
	if err != nil {
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		log.Printf("Error listing rooms: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div id="rooms-list">
	<table>
		<thead>
			<tr>
				<th>Room ID</th>
				<th>Owner</th>
				<th>Guest</th>
				<th>Status</th>
				<th>Created</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>`)

	for _, room := range rooms {
		owner := "Unknown"
		if room.OwnerUsername != nil {
			owner = html.EscapeString(*room.OwnerUsername)
		}

		guest := "Waiting..."
		if room.GuestUsername != nil {
			guest = html.EscapeString(*room.GuestUsername)
		}

		statusColor := ""
		switch room.Status {
		case "active":
			statusColor = "color: #10b981;"
		case "waiting":
			statusColor = "color: #f59e0b;"
		case "completed":
			statusColor = "color: #6b7280;"
		}

		fmt.Fprintf(w, `
			<tr>
				<td><code>%s</code></td>
				<td>%s</td>
				<td>%s</td>
				<td style="%s">%s</td>
				<td>%s</td>
				<td>
					<button hx-post="/admin/api/rooms/%s/close"
					        hx-target="#rooms-list"
					        hx-swap="outerHTML"
					        hx-confirm="Are you sure you want to force close this room?"
					        class="btn btn-sm btn-danger">
						Force Close
					</button>
				</td>
			</tr>`,
			room.ID.String()[:8],
			owner,
			guest,
			statusColor,
			html.EscapeString(room.Status),
			room.CreatedAt.Format("2006-01-02 15:04"),
			room.ID.String(),
		)
	}

	fmt.Fprintf(w, `
		</tbody>
	</table>
</div>`)
}

// CloseRoomHandler forces a room to close
func (ah *AdminAPIHandler) CloseRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	if err := ah.adminService.ForceCloseRoom(ctx, roomID); err != nil {
		http.Error(w, "Failed to close room", http.StatusInternalServerError)
		log.Printf("Error closing room: %v", err)
		return
	}

	// Return updated rooms list
	ah.ListRoomsHandler(w, r)
}

// Helper function for ternary operator
func tern(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

// GetDashboardStatsHandler returns dashboard statistics as JSON for HTMX
func (ah *AdminAPIHandler) GetDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	stats, err := ah.adminService.GetDashboardStats(ctx)
	if err != nil {
		http.Error(w, "Failed to get dashboard stats", http.StatusInternalServerError)
		log.Printf("Error getting dashboard stats: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

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
