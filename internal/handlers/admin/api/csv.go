package api

import (
	"net/http"

	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// CSVHandler handles CSV import/export
type CSVHandler struct {
	questionService *services.QuestionService
	categoryService *services.CategoryService
}

// NewCSVHandler creates a new CSV handler
func NewCSVHandler(questionService *services.QuestionService, categoryService *services.CategoryService) *CSVHandler {
	return &CSVHandler{
		questionService: questionService,
		categoryService: categoryService,
	}
}

// ExportQuestionsCSV exports questions to CSV
func (h *CSVHandler) ExportQuestionsCSV(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=questions.csv")
	return c.String(http.StatusOK, "ID,Category ID,Category Name,Language,Question Text,Created At\n")
}

// ImportQuestionsCSV imports questions from CSV
func (h *CSVHandler) ImportQuestionsCSV(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "Not yet implemented")
}

// GetImportTemplate downloads CSV template
func (h *CSVHandler) GetImportTemplate(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=questions_template.csv")
	return c.String(http.StatusOK, "Category ID,Language,Question Text\n")
}

// ExportCategoriesCSV exports categories to CSV
func (h *CSVHandler) ExportCategoriesCSV(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=categories.csv")
	return c.String(http.StatusOK, "ID,Key,Icon,Created At\n")
}



