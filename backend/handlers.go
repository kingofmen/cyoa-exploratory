// Package handlers implements the API internals.
package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

// txnError attempts to roll back the transaction and returns a commented error.
func txnError(comment string, txn *sql.Tx, err error) error {
	if rerr := txn.Rollback(); rerr != nil {
		return fmt.Errorf("%s: %w; rollback failed: %w", comment, err, rerr)
	}
	return fmt.Errorf("%s: %w", comment, err)
}

func createLocationImpl(ctx context.Context, db *sql.DB, loc *storypb.Location) (*spb.CreateLocationResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `INSERT INTO Locations (title, content) VALUES (?, ?)`, loc.GetTitle(), loc.GetContent()); err != nil {
		return nil, txnError("could not insert into Locations", txn, err)
	}

	var lid int64
	row := txn.QueryRowContext(ctx, `SELECT LAST_INSERT_ID()`)
	if err := row.Scan(&lid); err != nil {
		return nil, txnError("could not read back created location ID", txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("could not write to database: %w", err)
	}

	return &spb.CreateLocationResponse{
		Location: &storypb.Location{
			Id:      proto.Int64(lid),
			Title:   proto.String(loc.GetTitle()),
			Content: proto.String(loc.GetContent()),
		},
	}, nil
}

func updateLocationImpl(ctx context.Context, db *sql.DB, id int64, loc *storypb.Location) (*spb.UpdateLocationResponse, error) {
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

	return &spb.UpdateLocationResponse{
		Location: &storypb.Location{
			Id:      proto.Int64(id),
			Title:   proto.String(loc.GetTitle()),
			Content: proto.String(loc.GetContent()),
		},
	}, nil
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

func createStoryImpl(ctx context.Context, db *sql.DB, str *storypb.Story) (*spb.CreateStoryResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	if _, err := txn.ExecContext(ctx, `INSERT INTO Stories (title, description) VALUES (?, ?)`, str.GetTitle(), str.GetDescription()); err != nil {
		return nil, txnError("could not insert into Stories", txn, err)
	}
	var sid int64
	row := txn.QueryRowContext(ctx, `SELECT LAST_INSERT_ID()`)
	if err := row.Scan(&sid); err != nil {
		return nil, txnError("could not read back created ID", txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	str.Id = proto.Int64(sid)
	return &spb.CreateStoryResponse{
		Story: str,
	}, nil
}

func updateStoryImpl(ctx context.Context, db *sql.DB, str *storypb.Story) (*spb.UpdateStoryResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	sid := str.GetId()
	row := txn.QueryRowContext(ctx, `SELECT * FROM Stories AS s WHERE s.id = ?`, sid)
	var slocid sql.NullInt64
	var title string
	var desc sql.NullString
	if err := row.Scan(&sid, &title, &desc, &slocid); err != nil {
		return nil, txnError(fmt.Sprintf("could not read story %d", sid), txn, err)
	}
	if nt := str.GetTitle(); len(nt) > 0 {
		title = nt
	}
	if nd := str.GetDescription(); len(nd) > 0 {
		desc.String = nd
		desc.Valid = true
	}
	if nsl := str.GetStartLocationId(); nsl > 0 {
		slocid.Int64 = nsl
		slocid.Valid = true
	}

	if _, err := txn.ExecContext(ctx, `UPDATE Stories SET title = ?, description = ?, start_location = ? WHERE id = ?`, title, desc, slocid, sid); err != nil {
		return nil, txnError(fmt.Sprintf("could not update story %d", sid), txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	ret := &storypb.Story{
		Id:    proto.Int64(sid),
		Title: proto.String(title),
	}
	if desc.Valid {
		ret.Description = proto.String(desc.String)
	}
	if slocid.Valid {
		ret.StartLocationId = proto.Int64(slocid.Int64)
	}

	return &spb.UpdateStoryResponse{
		Story: ret,
	}, nil
}
