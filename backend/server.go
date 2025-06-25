// Package handlers implements the CYOA server API.
package handlers

import (
	"context"
	"database/sql"
	"fmt"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

type Server struct {
	db *sql.DB
}

func New(db *sql.DB) *Server {
	return &Server{
		db: db,
	}
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
	if req.GetLocationId() < 1 {
		return nil, fmt.Errorf("location ID not specified")
	}
	resp, err := updateLocationImpl(ctx, s.db, req.GetLocationId(), loc)
	if err != nil {
		return nil, fmt.Errorf("UpdateLocation error: %w", err)
	}
	return resp, nil
}

func (s *Server) DeleteLocation(ctx context.Context, req *spb.DeleteLocationRequest) (*spb.DeleteLocationResponse, error) {
	if req.GetLocationId() < 1 {
		return nil, fmt.Errorf("location ID not specified")
	}
	resp, err := deleteLocationImpl(ctx, s.db, req.GetLocationId())
	if err != nil {
		return nil, fmt.Errorf("DeleteLocation error: %w", err)
	}
	return resp, nil
}

func (s *Server) GetLocation(ctx context.Context, req *spb.GetLocationRequest) (*spb.GetLocationResponse, error) {
	if req.GetLocationId() < 1 {
		return nil, fmt.Errorf("location ID not specified")
	}
	resp, err := getLocationImpl(ctx, s.db, req.GetLocationId())
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

func (s *Server) CreateStory(ctx context.Context, req *spb.CreateStoryRequest) (*spb.CreateStoryResponse, error) {
	str := req.GetStory()
	if str == nil {
		return nil, fmt.Errorf("CreateStory called with nil story")
	}
	if len(str.GetTitle()) < 1 {
		return nil, fmt.Errorf("cannot create story with empty title")
	}
	resp, err := createStoryImpl(ctx, s.db, str)
	if err != nil {
		return nil, fmt.Errorf("CreateStory error: %w", err)
	}
	return resp, nil
}

func (s *Server) UpdateStory(ctx context.Context, req *spb.UpdateStoryRequest) (*spb.UpdateStoryResponse, error) {
	str := req.GetStory()
	if str == nil {
		return nil, fmt.Errorf("UpdateStory called with nil story")
	}
	if id := str.GetId(); id < 1 {
		return nil, fmt.Errorf("UpdateStory called with invalid story ID %d", id)
	}
	resp, err := updateStoryImpl(ctx, s.db, str)
	if err != nil {
		return nil, fmt.Errorf("UpdateStory error: %w", err)
	}
	return resp, nil
}

func (s *Server) DeleteStory(ctx context.Context, req *spb.DeleteStoryRequest) (*spb.DeleteStoryResponse, error) {
	return &spb.DeleteStoryResponse{}, nil
}

func (s *Server) GetStory(ctx context.Context, req *spb.GetStoryRequest) (*spb.GetStoryResponse, error) {
	return &spb.GetStoryResponse{}, nil
}

func (s *Server) ListStories(ctx context.Context, req *spb.ListStoriesRequest) (*spb.ListStoriesResponse, error) {
	return &spb.ListStoriesResponse{}, nil
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
	if id := act.GetId(); id < 1 {
		return nil, fmt.Errorf("UpdateAction called with invalid action ID %d", id)
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
	return &spb.PlayerActionResponse{}, nil
}
