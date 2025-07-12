package server

import (
	"fmt"
	"net/http"
)

// playData holds data for the playthrough template.
type playData struct {
	Timestamp string
}

// CreatePlaythroughHandler creates a new playthrough for the requested story.
func (h *Handler) CreatePlaythroughHandler(w http.ResponseWriter, req *http.Request) {
	_, err := getStoryId(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot get story ID to play: %v", err), http.StatusBadRequest)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
