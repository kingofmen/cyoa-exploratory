package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

// playData holds data for the playthrough template.
type playData struct {
	Timestamp   string
	StoryTitle  string
	Narration   string
	Description string
	Actions     []*storypb.Action
}

// CreatePlaythroughHandler creates a new playthrough for the requested story.
func (h *Handler) CreatePlaythroughHandler(w http.ResponseWriter, req *http.Request) {
	sid, err := getStoryId(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot parse story ID to create playthrough: %v", err), http.StatusBadRequest)
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
	gid, err := getGameId(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot parse game ID to play: %v", err), http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	resp, err := h.client.GameState(ctx, &spb.GameStateRequest{
		GameId: proto.Int64(gid),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot load playthrough %d: %v", gid, err), http.StatusInternalServerError)
		return
	}

	state := resp.GetState()
	data := &playData{
		Timestamp:   fmt.Sprintf("%s", time.Now()),
		StoryTitle:  state.GetStory().GetTitle(),
		Narration:   state.GetNarration(),
		Description: state.GetLocation().GetContent(),
		Actions:     make([]*storypb.Action, 0, len(state.GetLocation().GetPossibleActions())),
	}
	for _, act := range state.GetLocation().GetPossibleActions() {
		data.Actions = append(data.Actions, &storypb.Action{Title: proto.String(act.GetActionId())})
	}
	if err := h.playTmpl.Execute(w, data); err != nil {
		log.Printf("Play template execution error: %v", err)
	}
}
