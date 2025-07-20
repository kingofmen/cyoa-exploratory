package server

import (
	"fmt"
	"log"
	"net/http"

	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

// playData holds data for the playthrough template.
type playData struct {
	GameId int64
	State  *storypb.GameDisplay
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
	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("bad form: %v", err), http.StatusBadRequest)
			return
		}
		aid := req.FormValue("action_id")
		if len(aid) < 1 {
			http.Error(w, "choose an action", http.StatusBadRequest)
			return
		}
		_, err := h.client.GameState(ctx, &spb.GameStateRequest{
			GameId:   proto.Int64(gid),
			ActionId: proto.String(aid),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot advance playthrough %d with action %q: %v", gid, aid, err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, req, fmt.Sprintf("%s?game_id=%d", PlayGameURL, gid), http.StatusSeeOther)
		return
	}

	resp, err := h.client.GameState(ctx, &spb.GameStateRequest{
		GameId: proto.Int64(gid),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot load playthrough %d: %v", gid, err), http.StatusInternalServerError)
		return
	}

	data := &playData{
		GameId: gid,
		State:  resp.GetState(),
	}
	if err := h.playTmpl.Execute(w, data); err != nil {
		log.Printf("Play template execution error: %v", err)
	}
}
