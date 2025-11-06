package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/couple-card-game/internal/middleware"
	"github.com/yourusername/couple-card-game/internal/services"
)

// RealtimeHandler handles real-time connections
type RealtimeHandler struct {
	handler         *Handler
	realtimeService *services.RealtimeService
}

// NewRealtimeHandler creates a new realtime handler
func NewRealtimeHandler(h *Handler, realtimeService *services.RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{
		handler:         h,
		realtimeService: realtimeService,
	}
}

// StreamRoomEvents streams room events via SSE
func (h *RealtimeHandler) StreamRoomEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Subscribe to room events
	client := h.realtimeService.Subscribe(roomID, userID)
	defer h.realtimeService.Unsubscribe(client.ID)

	// Stream events
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	fmt.Fprintf(w, "event: connected\ndata: {\"type\":\"connected\",\"room_id\":\"%s\"}\n\n", roomID)
	flusher.Flush()

	// Keep track of last join request count
	ctx := r.Context()
	lastJoinRequestCount := 0
	joinCheckTicker := time.NewTicker(2 * time.Second)
	defer joinCheckTicker.Stop()

	// Stream events until client disconnects
	for {
		select {
		case event := <-client.Channel:
			// Send event from realtime service
			fmt.Fprint(w, services.EventToSSE(event))
			flusher.Flush()
		case <-joinCheckTicker.C:
			// Check for new join requests every 2 seconds
			requests, err := h.handler.RoomService.GetJoinRequestsByRoom(ctx, roomID)
			if err == nil {
				currentCount := len(requests)
				if currentCount != lastJoinRequestCount {
					fmt.Fprintf(w, "event: join_request\ndata: {\"count\":%d}\n\n", currentCount)
					flusher.Flush()
					lastJoinRequestCount = currentCount
				}
			}
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			// Send keepalive ping
			fmt.Fprintf(w, "event: ping\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
			flusher.Flush()
		}
	}
}

// GetRoomPlayers gets current room players
func (h *RealtimeHandler) GetRoomPlayers(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

// GetRoomState gets current room state
func (h *RealtimeHandler) GetRoomState(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

