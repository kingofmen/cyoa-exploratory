package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

// playData holds data for the playthrough template.
type playData struct {
	Timestamp  string
	StoryTitle string
}

// CreatePlaythroughHandler creates a new playthrough for the requested story.
func (h *Handler) CreatePlaythroughHandler(w http.ResponseWriter, req *http.Request) {
	sid, err := getStoryId(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot parse story ID to play: %v", err), http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	resp, err := h.client.CreateGame(ctx, &spb.CreateGameRequest{
		StoryId: proto.Int64(sid),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating playthrough: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, fmt.Sprintf("%s?game_id=%d", PlayGameURL, resp.GetGameId()), http.StatusSeeOther)
}

func (h *Handler) PlayGameHandler(w http.ResponseWriter, req *http.Request) {
	data := &playData{
		Timestamp:  fmt.Sprintf("%s", time.Now()),
		StoryTitle: "Placeholder Title",
	}
	if err := h.playTmpl.Execute(w, data); err != nil {
		log.Printf("Play template execution error: %v", err)
	}
}
