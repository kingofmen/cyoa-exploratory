// Package handlers implements the CYOA server API.
package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/kingofmen/cyoa-exploratory/narrate"
	"github.com/kingofmen/cyoa-exploratory/story"
	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

type Server struct {
	db       *sql.DB
	narrator narrate.Narrator
}

func New(db *sql.DB) *Server {
	return &Server{
		db:       db,
		narrator: narrate.NewNoop(),
	}
}

func (s *Server) WithNarrator(n narrate.Narrator) *Server {
	if s == nil {
		s = New(nil)
	}
	s.narrator = n
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

func (s *Server) UpdateStory(ctx context.Context, req *spb.UpdateStoryRequest) (*spb.UpdateStoryResponse, error) {
	str := req.GetStory()
	if str == nil {
		return nil, fmt.Errorf("UpdateStory called with nil story")
	}

	txn, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	resp, err := updateStoryImpl(ctx, txn, str)
	if err != nil {
		return nil, txnError("could not update story", txn, err)
	}

	content := req.GetContent()
	locs := content.GetLocations()
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

	resp.Content = content
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

func (s *Server) PlayerAction(ctx context.Context, req *spb.PlayerActionRequest) (*spb.PlayerActionResponse, error) {
	gid, aid := req.GetGameId(), req.GetActionId()
	if gid < 1 {
		return nil, fmt.Errorf("PlayerAction called with bad game ID %d", gid)
	}
	if err := uuid.Validate(aid); err != nil {
		return nil, fmt.Errorf("invalid action ID %q: %w", aid, err)
	}

	event, err := validateAction(ctx, s.db, gid, aid)
	if err != nil {
		return nil, fmt.Errorf("could not validate action %s in game %d: %w", aid, gid, err)
	}

	updated, err := story.HandleEvent(event)
	if err != nil {
		return nil, fmt.Errorf("could not apply action %s in game %d: %w", aid, gid, err)
	}

	content, err := s.narrator.Event(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("could not narrate action %s in game %d: %w", aid, gid, err)
	}

	event.GameSnapshot = updated
	if nn := event.GetNarration(); len(nn) > 0 {
		content = strings.Join([]string{nn, content}, "\n")
	}
	event.Narration = proto.String(content)
	if err := writeAction(ctx, s.db, event); err != nil {
		return nil, fmt.Errorf("PlayerAction error: %w", err)
	}

	return &spb.PlayerActionResponse{
		GameState: updated,
		Narrative: proto.String(event.GetNarration()),
	}, nil
}
