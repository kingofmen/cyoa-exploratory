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
	if _, err := txn.ExecContext(ctx, `INSERT INTO Locations (title, content) VALUES (?, ?)`, loc.GetTitle(), loc.GetContent()); err != nil {
		if rerr := txn.Rollback(); rerr != nil {
			return nil, fmt.Errorf("could not write to transaction: %w; rollback failed: %w", err, rerr)
		}
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
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	rows, err := txn.QueryContext(ctx, `SELECT l.title, l.content FROM Locations AS l ORDER BY l.id ASC`)
	if err != nil {
		if rerr := txn.Rollback(); rerr != nil {
			return nil, fmt.Errorf("database error listing locations: %w; rollback failed: %w", err, rerr)
		}
		return nil, fmt.Errorf("database error listing locations: %w", err)
	}
	for rows.Next() {
		var title, content string
		if err := rows.Scan(&title, &content); err != nil {
			if rerr := txn.Rollback(); rerr != nil {
				return nil, fmt.Errorf("error scanning location: %w; rollback failed: %w", err, rerr)
			}
			return nil, fmt.Errorf("error scanning location: %w", err)
		}
		resp.Locations = append(resp.Locations, &spb.Location{Title: &title, Content: &content})
	}
	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("error committing query transaction: %w", err)
	}
	return resp, nil
}
