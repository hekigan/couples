package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/labstack/echo/v4"
)

// NotificationStreamHandler handles SSE connections for real-time notifications
func (h *Handler) NotificationStreamHandler(c echo.Context) error {
	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	ctx := context.Background()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	// Access underlying writer for SSE
	w := c.Response().Writer

	// Create a channel for this connection
	clientChan := make(chan []byte, 10)

	// Register client with realtime service (you may need to add this method)
	// For now, we'll use a simple ticker to check for new notifications

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Keep track of last notification ID to avoid duplicates
	lastNotificationID := uuid.Nil

	// Create a ticker to check for new notifications every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Send periodic pings to keep connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			// Client disconnected
			fmt.Printf("Client disconnected from notification stream: %s\n", userID)
			return nil

		case <-ticker.C:
			// Check for new notifications
			notifications, err := h.NotificationService.GetUserNotifications(ctx, userID, 1)
			if err == nil && len(notifications) > 0 {
				latestNotification := notifications[0]

				// Only send if it's a new notification
				if latestNotification.ID != lastNotificationID {
					lastNotificationID = latestNotification.ID

					// Send the notification
					data, err := json.Marshal(latestNotification)
					if err == nil {
						fmt.Fprintf(w, "event: notification\ndata: %s\n\n", string(data))
						if f, ok := w.(http.Flusher); ok {
							f.Flush()
						}
					}
				}
			}

		case <-pingTicker.C:
			// Send ping to keep connection alive
			fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

		case data := <-clientChan:
			// Send data from channel (for future use with realtime service)
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", string(data))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}



