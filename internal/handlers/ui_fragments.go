package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/hekigan/couples/internal/services"
)

// SetGuestReadyAPIHandler marks the guest as ready
func (h *Handler) SetGuestReadyAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	// Verify user is the guest
	if room.GuestID == nil || *room.GuestID != userID {
		http.Error(w, "Only the guest can mark themselves as ready", http.StatusForbidden)
		return
	}

	// Mark guest as ready
	room.GuestReady = true
	if err := h.RoomService.UpdateRoom(ctx, room); err != nil {
		log.Printf("❌ Failed to update guest ready status: %v", err)
		http.Error(w, "Failed to update ready status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Guest %s marked as ready in room %s", userID, roomID)

	// Refetch room to get updated status from database
	room, err = h.RoomService.GetRoomByID(ctx, roomID)
	if err != nil {
		log.Printf("❌ Failed to refetch room: %v", err)
		http.Error(w, "Failed to refetch room", http.StatusInternalServerError)
		return
	}

	log.Printf("🔍 Room status after refetch: %s (GuestReady: %v)", room.Status, room.GuestReady)

	// Use helper to render HTML fragment for updated button state
	html, err := h.TemplateService.RenderFragment("guest_ready_button.html", services.GuestReadyButtonData{
		RoomID:     roomID.String(),
		GuestReady: true,
	})
	if err != nil {
		log.Printf("Error rendering guest ready button template: %v", err)
		http.Error(w, "Failed to render button", http.StatusInternalServerError)
		return
	}

	// Broadcast room status badge update via SSE (so owner sees status change)
	if statusBadgeHTML, err := h.TemplateService.RenderFragment("status_badge.html", services.RoomStatusBadgeData{
		Status: room.Status,
	}); err == nil {
		h.RoomService.GetRealtimeService().BroadcastHTMLFragment(roomID, services.HTMLFragmentEvent{
			Type:       "room_status_update",
			Target:     "#room-status-badge",
			SwapMethod: "outerHTML",
			HTML:       statusBadgeHTML,
		})
		log.Printf("📡 Broadcasted room status badge update to room %s (status: %s)", roomID, room.Status)
	}

	// Broadcast room_update event to trigger start button refresh
	h.RoomService.BroadcastRoomUpdate(roomID, map[string]interface{}{
		"guest_ready": true,
	})

	log.Printf("📡 Broadcasting guest_ready=true to room %s", roomID)

	// Return HTML fragment for HTMX
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// RoomStatusBadgeAPIHandler returns the room status badge HTML fragment
func (h *Handler) RoomStatusBadgeAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("🔍 RoomStatusBadgeAPIHandler: Room %s status: %s", roomID, room.Status)

	// Use helper to render HTML fragment for status badge
	if err := h.RenderHTMLFragment(w, "status_badge.html", services.RoomStatusBadgeData{
		Status: room.Status,
	}); err != nil {
		log.Printf("Error rendering status badge template: %v", err)
		http.Error(w, "Failed to render status badge", http.StatusInternalServerError)
	}
}

// GetStartGameButtonHTMLHandler returns start game button as HTML fragment (for HTMX)
func (h *Handler) GetStartGameButtonHTMLHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Use helper to render HTML fragment
	if err := h.RenderHTMLFragment(w, "start_game_button.html", services.StartGameButtonData{
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	}); err != nil {
		log.Printf("Error rendering start game button template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load start button</p>`))
	}
}

// GetGuestReadyButtonHTMLHandler returns guest ready button as HTML fragment (for HTMX)
func (h *Handler) GetGuestReadyButtonHTMLHandler(w http.ResponseWriter, r *http.Request) {
	// Use helper to get room
	room, roomID, err := h.GetRoomFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Use helper to render HTML fragment
	if err := h.RenderHTMLFragment(w, "guest_ready_button.html", services.GuestReadyButtonData{
		RoomID:     roomID.String(),
		GuestReady: room.GuestReady,
	}); err != nil {
		log.Printf("Error rendering guest ready button template: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<p style="color: #6b7280;">Failed to load ready button</p>`))
	}
}
