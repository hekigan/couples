package handlers

import (
	"net/http"
)

// ListJoinRequestsHandler lists join requests for a room
func (h *Handler) ListJoinRequestsHandler(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Title: "Join Requests",
	}
	h.RenderTemplate(w, "game/join-requests.html", data)
}

// CreateJoinRequestHandler creates a join request
func (h *Handler) CreateJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

// AcceptJoinRequestHandler accepts a join request
func (h *Handler) AcceptJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

// RejectJoinRequestHandler rejects a join request
func (h *Handler) RejectJoinRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

// GetJoinRequestsCountHandler gets count of pending requests
func (h *Handler) GetJoinRequestsCountHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"count":0}`))
}

