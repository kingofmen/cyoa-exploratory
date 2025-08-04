// Package handlers implements the CYOA server API.
package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/kingofmen/cyoa-exploratory/narrate"
	"github.com/kingofmen/cyoa-exploratory/story"
	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

const (
	debugTellerKey = "debug"
)

type Server struct {
	db        *sql.DB
	tellers   map[string]*narrateInfo
	tellerKey string
}

type narrateInfo struct {
	narrate.Narrator
}

func New(db *sql.DB) *Server {
	return &Server{
		db: db,
		tellers: map[string]*narrateInfo{
			"noop":         &narrateInfo{Narrator: narrate.NewNoop()},
			debugTellerKey: &narrateInfo{Narrator: narrate.NewDebug()},
		},
		tellerKey: debugTellerKey,
	}
}

func (s *Server) WithNarrator(key string, n narrate.Narrator) *Server {
	if s == nil {
		s = New(nil)
	}
	s.tellers[key] = &narrateInfo{Narrator: n}
	s.tellerKey = key
	return s
}

func (s *Server) CreateLocation(ctx context.Context, req *spb.CreateLocationRequest) (*spb.CreateLocationResponse, error) {
	loc := req.GetLocation()
	if loc == nil {
		return nil, fmt.Errorf("CreateLocation called with nil location")
	}
	if len(loc.GetTitle()) < 1 {
		return nil, fmt.Errorf("cannot create location with empty title")
	}
	resp, err := createLocationImpl(ctx, s.db, loc)
	if err != nil {
		return nil, fmt.Errorf("CreateLocation error: %w", err)
	}
	return resp, nil
}

func (s *Server) UpdateLocation(ctx context.Context, req *spb.UpdateLocationRequest) (*spb.UpdateLocationResponse, error) {
	loc := req.GetLocation()
	if loc == nil {
		return nil, fmt.Errorf("UpdateLocation called with nil location")
	}
	if len(loc.GetTitle()) < 1 {
		return nil, fmt.Errorf("cannot update location to have empty title")
	}
	lid := req.GetLocationId()
	if err := uuid.Validate(lid); err != nil {
		return nil, fmt.Errorf("invalid location ID %q: %w", lid, err)
	}
	resp, err := updateLocationImpl(ctx, s.db, lid, loc)
	if err != nil {
		return nil, fmt.Errorf("UpdateLocation error: %w", err)
	}
	return resp, nil
}

func (s *Server) DeleteLocation(ctx context.Context, req *spb.DeleteLocationRequest) (*spb.DeleteLocationResponse, error) {
	lid := req.GetLocationId()
	if err := uuid.Validate(lid); err != nil {
		return nil, fmt.Errorf("invalid location ID %q: %w", lid, err)
	}
	resp, err := deleteLocationImpl(ctx, s.db, lid)
	if err != nil {
		return nil, fmt.Errorf("DeleteLocation error: %w", err)
	}
	return resp, nil
}

func (s *Server) GetLocation(ctx context.Context, req *spb.GetLocationRequest) (*spb.GetLocationResponse, error) {
	lid := req.GetLocationId()
	if err := uuid.Validate(lid); err != nil {
		return nil, fmt.Errorf("invalid location ID %q: %w", lid, err)
	}
	resp, err := getLocationImpl(ctx, s.db, lid)
	if err != nil {
		return nil, fmt.Errorf("GetLocation error: %w", err)
	}
	return resp, nil
}

func (s *Server) ListLocations(ctx context.Context, req *spb.ListLocationsRequest) (*spb.ListLocationsResponse, error) {
	resp, err := listLocationsImpl(ctx, s.db, req)
	if err != nil {
		return nil, fmt.Errorf("ListLocations error: %w", err)
	}
	return resp, nil
}

func validateContent(content *spb.StoryContent) ([]*storypb.Location, []*storypb.Action, error) {
	locs, acts := content.GetLocations(), content.GetActions()
	lids, aids := make(map[string]bool), make(map[string]bool)
	for _, loc := range locs {
		lid := loc.GetId()
		if err := uuid.Validate(lid); err != nil {
			return nil, nil, fmt.Errorf("invalid location ID %q for %q: %w", lid, loc.GetTitle(), err)
		}
		lids[lid] = true
	}
	for _, act := range acts {
		aid := act.GetId()
		if err := uuid.Validate(aid); err != nil {
			return nil, nil, fmt.Errorf("invalid action ID %q for %q: %w", aid, act.GetTitle(), err)
		}
		aids[aid] = true
		for tidx, trg := range act.GetTriggers() {
			for eidx, eff := range trg.GetEffects() {
				if nlid := eff.GetNewLocationId(); len(nlid) > 0 && !lids[nlid] {
					return nil, nil, fmt.Errorf("action %q trigger %d/%d has bad location ID %q", act.GetTitle(), tidx, eidx, nlid)
				}
			}
		}
	}

	for _, loc := range locs {
		for cidx, cand := range loc.GetPossibleActions() {
			if caid := cand.GetActionId(); !aids[caid] {
				return nil, nil, fmt.Errorf("location %q possible action %d has bad action ID %q", loc.GetTitle(), cidx, caid)
			}
		}
	}
	return locs, acts, nil
}

func (s *Server) UpdateStory(ctx context.Context, req *spb.UpdateStoryRequest) (*spb.UpdateStoryResponse, error) {
	str := req.GetStory()
	if str == nil {
		return nil, fmt.Errorf("UpdateStory called with nil story")
	}
	locs, acts, err := validateContent(req.GetContent())
	if err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}

	txn, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	resp, err := updateStoryImpl(ctx, txn, str)
	if err != nil {
		return nil, txnError("could not update story", txn, err)
	}

	locIds := make(map[string]bool)
	for idx, loc := range locs {
		locs[idx], err = createOrUpdateLocation(ctx, txn, loc.GetId(), loc)
		if err != nil {
			return nil, txnError(fmt.Sprintf("could not update location %q", loc.GetTitle()), txn, err)
		}
		locIds[locs[idx].GetId()] = true
	}

	if err := updateStoryLocationsTable(ctx, txn, resp.GetStory().GetId(), slices.Collect(maps.Keys(locIds))); err != nil {
		return nil, txnError("could not update story-location relationships", txn, err)
	}

	actIds := make(map[string]bool)
	for idx, act := range acts {
		acts[idx], err = createOrUpdateAction(ctx, txn, act.GetId(), act)
		if err != nil {
			return nil, txnError(fmt.Sprintf("could not update action %q", act.GetTitle()), txn, err)
		}
		actIds[acts[idx].GetId()] = true
	}

	if err := updateStoryActionsTable(ctx, txn, resp.GetStory().GetId(), slices.Collect(maps.Keys(actIds))); err != nil {
		return nil, txnError("could not update story-action relationships", txn, err)
	}

	resp.Content = req.GetContent()
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit to database", txn, err)
	}

	return resp, nil
}

func (s *Server) DeleteStory(ctx context.Context, req *spb.DeleteStoryRequest) (*spb.DeleteStoryResponse, error) {
	sid := req.GetId()
	if sid < 1 {
		return nil, fmt.Errorf("DeleteStory called with invalid story ID %d", sid)
	}
	return deleteStoryImpl(ctx, s.db, sid)
}

func (s *Server) GetStory(ctx context.Context, req *spb.GetStoryRequest) (*spb.GetStoryResponse, error) {
	sid := req.GetId()
	if sid < 1 {
		return nil, fmt.Errorf("GetStory called with invalid story ID %d", sid)
	}
	return getStoryImpl(ctx, s.db, sid, req.GetView())
}

func (s *Server) ListStories(ctx context.Context, req *spb.ListStoriesRequest) (*spb.ListStoriesResponse, error) {
	resp, err := listStoriesImpl(ctx, s.db, req)
	if err != nil {
		return nil, fmt.Errorf("ListStories error: %w", err)
	}
	return resp, nil
}

func (s *Server) CreateAction(ctx context.Context, req *spb.CreateActionRequest) (*spb.CreateActionResponse, error) {
	act := req.GetAction()
	if act == nil {
		return nil, fmt.Errorf("CreateAction called with nil action")
	}
	if len(act.GetTitle()) < 1 {
		return nil, fmt.Errorf("cannot create action with empty title")
	}
	resp, err := createActionImpl(ctx, s.db, act)
	if err != nil {
		return nil, fmt.Errorf("CreateAction error: %w", err)
	}
	return resp, nil
}

func (s *Server) UpdateAction(ctx context.Context, req *spb.UpdateActionRequest) (*spb.UpdateActionResponse, error) {
	act := req.GetAction()
	if act == nil {
		return nil, fmt.Errorf("UpdateAction called with nil action")
	}
	aid := act.GetId()
	if err := uuid.Validate(aid); err != nil {
		return nil, fmt.Errorf("invalid action ID %q: %w", aid, err)
	}
	resp, err := updateActionImpl(ctx, s.db, act)
	if err != nil {
		return nil, fmt.Errorf("UpdateAction error: %w", err)
	}
	return resp, nil
}

func (s *Server) DeleteAction(ctx context.Context, req *spb.DeleteActionRequest) (*spb.DeleteActionResponse, error) {
	return &spb.DeleteActionResponse{}, nil
}

func (s *Server) GetAction(ctx context.Context, req *spb.GetActionRequest) (*spb.GetActionResponse, error) {
	return &spb.GetActionResponse{}, nil
}

func (s *Server) ListActions(ctx context.Context, req *spb.ListActionsRequest) (*spb.ListActionsResponse, error) {
	return &spb.ListActionsResponse{}, nil
}

func (s *Server) CreateGame(ctx context.Context, req *spb.CreateGameRequest) (*spb.CreateGameResponse, error) {
	sid := req.GetStoryId()
	if sid < 1 {
		return nil, fmt.Errorf("CreateGame called with bad story ID %d", sid)
	}
	resp, err := createGameImpl(ctx, s.db, sid)
	if err != nil {
		return nil, fmt.Errorf("CreateGame error: %w", err)
	}
	return resp, nil
}

func (s *Server) ListGames(ctx context.Context, req *spb.ListGamesRequest) (*spb.ListGamesResponse, error) {
	resp, err := listGamesImpl(ctx, s.db, req)
	if err != nil {
		return nil, fmt.Errorf("ListGames error: %w", err)
	}
	return resp, nil
}

func makeGameDisplay(event *storypb.GameEvent) *storypb.GameDisplay {
	display := &storypb.GameDisplay{
		Story:     summarize(event.GetStory()),
		Location:  summarize(event.GetLocation()),
		Narration: proto.String(event.GetNarration()),
		RunState:  event.GetState().Enum(),
	}

	acts := story.PossibleActions(event)
	for _, act := range acts {
		summary := identify(act)
		display.Actions = append(display.Actions, summary)
	}

	return display
}

func (s *Server) GameState(ctx context.Context, req *spb.GameStateRequest) (*spb.GameStateResponse, error) {
	gid, aid := req.GetGameId(), req.GetActionId()
	if gid < 1 {
		return nil, fmt.Errorf("GameState called with bad game ID %d", gid)
	}
	if len(aid) > 0 {
		if err := uuid.Validate(aid); err != nil {
			return nil, fmt.Errorf("GameState called with invalid action ID %q: %w", aid, err)
		}
	}

	txn, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin read transaction for action %s in playthrough %d: %w", aid, gid, err)
	}
	gstate, err := loadStoryState(ctx, txn, gid, aid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not load story state for action %s in playthrough %d", aid, gid), txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError(fmt.Sprintf("could not commit read for action %s in playthrough %d", aid, gid), txn, err)
	}

	if gstate.GetPlayerAction() == nil {
		return &spb.GameStateResponse{
			State: makeGameDisplay(gstate),
		}, nil
	}

	nstate, err := story.HandleEvent(gstate)
	if err != nil {
		return nil, fmt.Errorf("could not apply action %s in game %d: %w", aid, gid, err)
	}

	txn, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin write transaction for action %s in playthrough %d: %w", aid, gid, err)
	}
	if nlid := nstate.GetLocation().GetId(); nlid != gstate.GetLocation().GetId() {
		// New location, load from DB.
		nloc, err := loadLocation(ctx, txn, nlid)
		if err != nil {
			return nil, txnError(fmt.Sprintf("could not load new location %s after action %s", nlid, aid), txn, err)
		}
		nstate.Location = nloc
		acts, err := loadPossibleActions(ctx, txn, nloc)
		if err != nil {
			return nil, txnError(fmt.Sprintf("could not load new candidate actions for location %s after action %s", nlid, aid), txn, err)
		}
		nstate.CandidateActions = acts
	}

	tell, ok := s.tellers[s.tellerKey]
	if !ok {
		log.Printf("No teller %q found, falling back on default %q", s.tellerKey, debugTellerKey)
		tell = s.tellers[debugTellerKey]
	}
	content, err := tell.Event(ctx, gstate, nstate)
	if err != nil {
		return nil, fmt.Errorf("could not narrate action %s in game %d: %w", aid, gid, err)
	}

	if nn := gstate.GetNarration(); len(nn) > 0 {
		content = strings.Join([]string{nn, content}, "\n")
	}
	nstate.Narration = proto.String(content)

	if err := writeAction(ctx, txn, gid, nstate, content); err != nil {
		return nil, txnError(fmt.Sprintf("error writing action %s to playthrough %d", aid, gid), txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError(fmt.Sprintf("could not commit action %s to playthrough %d", aid, gid), txn, err)
	}

	return &spb.GameStateResponse{
		State: makeGameDisplay(nstate),
	}, nil
}
