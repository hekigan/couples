package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/hekigan/couples/internal/services"
	"github.com/labstack/echo/v4"
)

// GetDashboardStatsHandler returns dashboard statistics as HTML for HTMX
func (ah *AdminAPIHandler) GetDashboardStatsHandler(c echo.Context) error {
	ctx := context.Background()

	stats, err := ah.adminService.GetDashboardStats(ctx)
	if err != nil {
		log.Printf("Error getting dashboard stats: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get dashboard stats")
	}

	data := services.DashboardStatsData{
		TotalUsers:      stats.TotalUsers,
		AnonymousUsers:  stats.AnonymousUsers,
		RegisteredUsers: stats.RegisteredUsers,
		AdminUsers:      stats.AdminUsers,
		TotalRooms:      stats.TotalRooms,
		ActiveRooms:     stats.ActiveRooms,
		CompletedRooms:  stats.CompletedRooms,
		TotalQuestions:  stats.TotalQuestions,
		TotalCategories: stats.TotalCategories,
	}

	html, err := ah.handler.TemplateService.RenderFragment("dashboard_stats.html", data)
	if err != nil {
		log.Printf("Error rendering dashboard stats: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, html)
}
