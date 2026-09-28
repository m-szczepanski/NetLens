// Package api implements the REST/WebSocket boundary between the Go backend
// and the SvelteKit frontend. The API depends on domain/application services
// and must never reach into scanner or capture internals. Every endpoint and
// event shape is part of the frontend/backend contract and must stay
// documented.
package api

import (
	"encoding/json"
	"net/http"
)

// Status is the response body of GET /api/status.
type Status struct {
	App      string `json:"app"`
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

// Handler serves the API. Deps (device state, scan manager, ...) are added as
// features require them.
type Handler struct {
	status Status
}

// NewHandler builds the top-level HTTP handler for the public /api routes.
func NewHandler(status Status) http.Handler {
	mux := http.NewServeMux()
	h := &Handler{status: status}
	mux.HandleFunc("GET /api/status", h.getStatus)
	return mux
}

func (h *Handler) getStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.status); err != nil {
		// Headers are already sent; nothing else visible to do here.
		return
	}
}
