// Package handlers implements the API internals.
package handlers

import (
	"context"
	"database/sql"
	"fmt"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

func createLocationImpl(ctx context.Context, db *sql.DB, loc *spb.Location) (*spb.CreateLocationResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO Locations (title, content) VALUES ($1, $2)`, loc.GetTitle(), loc.GetContent()); err != nil {
		return nil, fmt.Errorf("could not write to transaction: %w", err)
	}
	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("could not write to database: %w", err)
	}

	return &spb.CreateLocationResponse{}, nil
}

func listLocationsImpl(ctx context.Context, db *sql.DB, req *spb.ListLocationsRequest) (*spb.ListLocationsResponse, error) {
	resp := &spb.ListLocationsResponse{
		Locations: make([]*spb.Location, 0, 10),
	}
	// TODO: Do an actual database lookup, obviously.
	resp.Locations = append(resp.Locations, &spb.Location{Title: str("Fake title"), Content: str("Fake content")})
	return resp, nil
}
