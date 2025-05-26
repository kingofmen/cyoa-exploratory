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

func str(s string) *string {
	copy := s
	return &copy
}

func (s *Server) ListLocations(ctx context.Context, req *spb.ListLocationsRequest) (*spb.ListLocationsResponse, error) {
	resp, err := listLocationsImpl(ctx, s.db, req)
	if err != nil {
		return nil, fmt.Errorf("ListLocations error: %w", err)
	}
	return resp, nil
}
