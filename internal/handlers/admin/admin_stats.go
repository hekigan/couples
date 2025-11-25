package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/hekigan/couples/internal/services"
)

// GetDashboardStatsHandler returns dashboard statistics as HTML for HTMX
func (ah *AdminAPIHandler) GetDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	stats, err := ah.adminService.GetDashboardStats(ctx)
	if err != nil {
		http.Error(w, "Failed to get dashboard stats", http.StatusInternalServerError)
		log.Printf("Error getting dashboard stats: %v", err)
		return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering dashboard stats: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
