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
	blob, err := proto.Marshal(loc)
	if err != nil {
		return nil, fmt.Errorf("could not marshal Location: %w", err)
	}
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `INSERT INTO Locations (title, proto) VALUES (?, ?)`, loc.GetTitle(), blob); err != nil {
		return nil, txnError("could not insert into Locations", txn, err)
	}

	var lid int64
	row := txn.QueryRowContext(ctx, `SELECT LAST_INSERT_ID()`)
	if err := row.Scan(&lid); err != nil {
		return nil, txnError("could not read back created location ID", txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	loc.Id = proto.Int64(lid)
	return &spb.CreateLocationResponse{
		Location: loc,
	}, nil
}

func updateLocationImpl(ctx context.Context, db *sql.DB, id int64, loc *storypb.Location) (*spb.UpdateLocationResponse, error) {
	blob, err := proto.Marshal(loc)
	if err != nil {
		return nil, fmt.Errorf("could not marshal Location: %w", err)
	}
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `UPDATE Locations SET title = ?, proto = ? WHERE id = ?`, loc.GetTitle(), blob, id); err != nil {
		return nil, txnError(fmt.Sprintf("could not update Location %d", id), txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
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
		return nil, txnError(fmt.Sprintf("could not delete Location %d", id), txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	return &spb.DeleteLocationResponse{}, nil
}

func getLocationImpl(ctx context.Context, db *sql.DB, id int64) (*spb.GetLocationResponse, error) {
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	loc, err := loadLocation(ctx, txn, id)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find location %d", id), txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit query", txn, err)
	}
	loc.Id = proto.Int64(id)
	return &spb.GetLocationResponse{
		Location: loc,
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
	rows, err := txn.QueryContext(ctx, `SELECT l.id, l.proto FROM Locations AS l ORDER BY l.id ASC`)
	if err != nil {
		return nil, txnError("could not list locations", txn, err)
	}
	for rows.Next() {
		var id int64
		blob := []byte{}
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, txnError("error scanning location", txn, err)
		}
		loc := &storypb.Location{}
		if err := proto.Unmarshal(blob, loc); err != nil {
			return nil, txnError(fmt.Sprintf("could not unmarshal location %d", id), txn, err)
		}
		loc.Id = proto.Int64(id)
		resp.Locations = append(resp.Locations, loc)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit query", txn, err)
	}
	return resp, nil
}

func createStoryImpl(ctx context.Context, db *sql.DB, str *storypb.Story) (*spb.CreateStoryResponse, error) {
	blob, err := proto.Marshal(str)
	if err != nil {
		return nil, fmt.Errorf("could not marshal story: %w", err)
	}
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	if _, err := txn.ExecContext(ctx, `INSERT INTO Stories (title, proto) VALUES (?, ?)`, str.GetTitle(), blob); err != nil {
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
	old, err := loadStory(ctx, txn, sid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not read story %d", sid), txn, err)
	}

	if nt := str.GetTitle(); len(nt) == 0 {
		str.Title = proto.String(old.GetTitle())
	}
	if nd := str.GetDescription(); len(nd) == 0 {
		str.Description = proto.String(old.GetDescription())
	}
	if nsl := str.GetStartLocationId(); nsl < 1 {
		str.StartLocationId = proto.Int64(old.GetStartLocationId())
	}
	if evts := str.GetEvents(); len(evts) == 0 {
		str.Events = old.GetEvents()
	}

	blob, err := proto.Marshal(str)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not marshal updated story %d", sid), txn, err)
	}

	if _, err := txn.ExecContext(ctx, `UPDATE Stories SET title = ?, proto = ? WHERE id = ?`, str.GetTitle(), blob, sid); err != nil {
		return nil, txnError(fmt.Sprintf("could not update story %d", sid), txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	return &spb.UpdateStoryResponse{
		Story: str,
	}, nil
}

func createActionImpl(ctx context.Context, db *sql.DB, act *storypb.Action) (*spb.CreateActionResponse, error) {
	blob, err := proto.Marshal(act)
	if err != nil {
		return nil, fmt.Errorf("could not marshal Action: %v", err)
	}
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	if _, err := txn.ExecContext(ctx, `INSERT INTO Actions (proto) VALUES (?)`, blob); err != nil {
		return nil, txnError("could not insert into Actions", txn, err)
	}
	var aid int64
	row := txn.QueryRowContext(ctx, `SELECT LAST_INSERT_ID()`)
	if err := row.Scan(&aid); err != nil {
		return nil, txnError("could not read back created ID", txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	act.Id = proto.Int64(aid)
	return &spb.CreateActionResponse{
		Action: act,
	}, nil
}

func updateActionImpl(ctx context.Context, db *sql.DB, act *storypb.Action) (*spb.UpdateActionResponse, error) {
	// TODO: Implement me.
	return &spb.UpdateActionResponse{
		Action: act,
	}, nil
}

func createGameImpl(ctx context.Context, db *sql.DB, sid int64) (*spb.CreateGameResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	str, err := loadStory(ctx, txn, sid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find story %d", sid), txn, err)
	}

	ngame := &storypb.Playthrough{
		StoryId: proto.Int64(str.GetId()),
	}
	if loc := str.GetStartLocationId(); loc > 0 {
		ngame.LocationId = proto.Int64(loc)
	}
	// TODO: Set starting values from Story object.
	ngame.State = storypb.RunState_RS_ACTIVE.Enum()
	blob, err := proto.Marshal(ngame)
	if err != nil {
		return nil, txnError("could not marshal new game", txn, err)
	}
	if _, err := txn.ExecContext(ctx, `INSERT INTO Playthroughs (proto) VALUES (?)`, blob); err != nil {
		return nil, txnError("could not insert into Playthroughs", txn, err)
	}
	var gid int64
	row := txn.QueryRowContext(ctx, `SELECT LAST_INSERT_ID()`)
	if err := row.Scan(&gid); err != nil {
		return nil, txnError("could not read back created ID", txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	return &spb.CreateGameResponse{
		GameId: proto.Int64(gid),
	}, nil
}

// validateAction loads the action, location, game, and story for a player input.
// It is read-only.
func validateAction(ctx context.Context, db *sql.DB, gid, aid int64) (*storypb.GameEvent, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	game, narration, err := loadGame(ctx, txn, gid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find game %d", gid), txn, err)
	}
	sid := game.GetStoryId()
	str, err := loadStory(ctx, txn, sid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find story %d for playthrough %d", sid, gid), txn, err)
	}
	act, err := loadAction(ctx, txn, aid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find action %d for playthrough %d of story %d", aid, gid, sid), txn, err)
	}
	lid := game.GetLocationId()
	loc, err := loadLocation(ctx, txn, lid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find location %d for playthrough %d of story %d", lid, gid, sid), txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	return &storypb.GameEvent{
		Action:       act,
		Location:     loc,
		GameSnapshot: game,
		Story:        str,
		Narration:    proto.String(narration),
	}, nil
}

func writeAction(ctx context.Context, db *sql.DB, event *storypb.GameEvent) error {
	game := event.GetGameSnapshot()
	aid, gid, sid := event.GetAction().GetId(), game.GetId(), event.GetStory().GetId()
	txn, err := db.BeginTx(ctx, nil)
	blob, err := proto.Marshal(game)
	if err != nil {
		return txnError(fmt.Sprintf("could not marshal updated playthrough %d of story %d after action %d", gid, sid, aid), txn, err)
	}
	if _, err := txn.ExecContext(ctx, `UPDATE Playthroughs SET proto = ?, narration = ? WHERE id = ?`, blob, event.GetNarration(), gid); err != nil {
		return txnError(fmt.Sprintf("could not update playthrough %d of story %d after action %d", gid, sid, aid), txn, err)
	}
	if err := txn.Commit(); err != nil {
		return txnError("could not write to database", txn, err)
	}
	return nil
}
