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
	loc := req.GetStory()
	if loc == nil {
		return nil, fmt.Errorf("CreateStory called with nil story")
	}
	if len(loc.GetTitle()) < 1 {
		return nil, fmt.Errorf("cannot create story with empty title")
	}
	resp, err := createStoryImpl(ctx, s.db, loc)
	if err != nil {
		return nil, fmt.Errorf("CreateStory error: %w", err)
	}
	return resp, nil
}

func (s *Server) UpdateStory(ctx context.Context, req *spb.UpdateStoryRequest) (*spb.UpdateStoryResponse, error) {
	return &spb.UpdateStoryResponse{}, nil
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
