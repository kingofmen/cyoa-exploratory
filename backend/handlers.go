// Package handlers implements the API internals.
package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

func createLocationImpl(ctx context.Context, db *sql.DB, loc *storypb.Location) (*spb.CreateLocationResponse, error) {
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

func updateLocationImpl(ctx context.Context, db *sql.DB, id int64, loc *storypb.Location) (*spb.UpdateLocationResponse, error) {
	log.Printf("Updating location %d: %q %q", id, loc.GetTitle(), loc.GetContent())
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `UPDATE Locations SET title = ?, content = ? WHERE id = ?`, loc.GetTitle(), loc.GetContent(), id); err != nil {
		if rerr := txn.Rollback(); rerr != nil {
			return nil, fmt.Errorf("could not write to transaction: %w; rollback failed: %w", err, rerr)
		}
		return nil, fmt.Errorf("could not write to transaction: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("could not write to database: %w", err)
	}

	return &spb.UpdateLocationResponse{}, nil
}

func deleteLocationImpl(ctx context.Context, db *sql.DB, id int64) (*spb.DeleteLocationResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `DELETE FROM Locations WHERE id = ?`, id); err != nil {
		if rerr := txn.Rollback(); rerr != nil {
			return nil, fmt.Errorf("could not write to transaction: %w; rollback failed: %w", err, rerr)
		}
		return nil, fmt.Errorf("could not write to transaction: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("could not write to database: %w", err)
	}

	return &spb.DeleteLocationResponse{}, nil
}

func getLocationImpl(ctx context.Context, db *sql.DB, id int64) (*spb.GetLocationResponse, error) {
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	row := txn.QueryRowContext(ctx, `SELECT l.id, l.title, l.content FROM Locations AS l WHERE l.id = ?`, id)
	var title, content string
	if err := row.Scan(&id, &title, &content); err != nil {
		if rerr := txn.Rollback(); rerr != nil {
			return nil, fmt.Errorf("error scanning location: %w; rollback failed: %w", err, rerr)
		}
		return nil, fmt.Errorf("error scanning location: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("could not write to database: %w", err)
	}
	return &spb.GetLocationResponse{
		Location: &storypb.Location{
			Id:      proto.Int64(id),
			Title:   proto.String(title),
			Content: proto.String(content),
		},
	}, nil
}

func listLocationsImpl(ctx context.Context, db *sql.DB, req *spb.ListLocationsRequest) (*spb.ListLocationsResponse, error) {
	resp := &spb.ListLocationsResponse{
		Locations: make([]*storypb.Location, 0, 10),
	}
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	rows, err := txn.QueryContext(ctx, `SELECT l.id, l.title, l.content FROM Locations AS l ORDER BY l.id ASC`)
	if err != nil {
		if rerr := txn.Rollback(); rerr != nil {
			return nil, fmt.Errorf("database error listing locations: %w; rollback failed: %w", err, rerr)
		}
		return nil, fmt.Errorf("database error listing locations: %w", err)
	}
	for rows.Next() {
		var id int64
		var title, content string
		if err := rows.Scan(&id, &title, &content); err != nil {
			if rerr := txn.Rollback(); rerr != nil {
				return nil, fmt.Errorf("error scanning location: %w; rollback failed: %w", err, rerr)
			}
			return nil, fmt.Errorf("error scanning location: %w", err)
		}
		resp.Locations = append(resp.Locations, &storypb.Location{
			Id:      proto.Int64(id),
			Title:   proto.String(title),
			Content: proto.String(content),
		})
	}
	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("error committing query transaction: %w", err)
	}
	return resp, nil
}
