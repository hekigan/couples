package api

import (
	"net/http"

	"github.com/yourusername/couple-card-game/internal/services"
)

// CSVHandler handles CSV import/export
type CSVHandler struct {
	questionService *services.QuestionService
}

// NewCSVHandler creates a new CSV handler
func NewCSVHandler(questionService *services.QuestionService) *CSVHandler {
	return &CSVHandler{
		questionService: questionService,
	}
}

// ExportQuestionsCSV exports questions to CSV
func (h *CSVHandler) ExportQuestionsCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=questions.csv")
	w.Write([]byte("ID,Category ID,Category Name,Language,Question Text,Created At\n"))
}

// ImportQuestionsCSV imports questions from CSV
func (h *CSVHandler) ImportQuestionsCSV(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

// GetImportTemplate downloads CSV template
func (h *CSVHandler) GetImportTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=questions_template.csv")
	w.Write([]byte("Category ID,Language,Question Text\n"))
}

// ExportCategoriesCSV exports categories to CSV
func (h *CSVHandler) ExportCategoriesCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=categories.csv")
	w.Write([]byte("ID,Key,Icon,Created At\n"))
}



