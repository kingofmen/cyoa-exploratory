// Package handlers implements the API internals.
package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
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
	lid := loc.GetId()
	if len(lid) < 1 {
		lid = uuid.New().String()
	}
	blob, err := proto.Marshal(loc)
	if err != nil {
		return nil, fmt.Errorf("could not marshal Location: %w", err)
	}
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `INSERT INTO Locations (id, title, proto) VALUES (?, ?, ?)`, lid, loc.GetTitle(), blob); err != nil {
		return nil, txnError("could not insert into Locations", txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	loc.Id = proto.String(lid)
	return &spb.CreateLocationResponse{
		Location: loc,
	}, nil
}

func updateLocationImpl(ctx context.Context, db *sql.DB, lid string, loc *storypb.Location) (*spb.UpdateLocationResponse, error) {
	blob, err := proto.Marshal(loc)
	if err != nil {
		return nil, fmt.Errorf("could not marshal Location: %w", err)
	}
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `UPDATE Locations SET title = ?, proto = ? WHERE id = ?`, loc.GetTitle(), blob, lid); err != nil {
		return nil, txnError(fmt.Sprintf("could not update Location %s", lid), txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	return &spb.UpdateLocationResponse{
		Location: &storypb.Location{
			Id:          proto.String(lid),
			Title:       proto.String(loc.GetTitle()),
			Description: proto.String(loc.GetDescription()),
		},
	}, nil
}

func deleteLocationImpl(ctx context.Context, db *sql.DB, lid string) (*spb.DeleteLocationResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `DELETE FROM Locations WHERE id = ?`, lid); err != nil {
		return nil, txnError(fmt.Sprintf("could not delete Location %s", lid), txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write deletion to database", txn, err)
	}

	return &spb.DeleteLocationResponse{}, nil
}

func getLocationImpl(ctx context.Context, db *sql.DB, lid string) (*spb.GetLocationResponse, error) {
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	loc, err := loadLocation(ctx, txn, lid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find location %s", lid), txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit query", txn, err)
	}
	loc.Id = proto.String(lid)
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
		var lid string
		blob := []byte{}
		if err := rows.Scan(&lid, &blob); err != nil {
			return nil, txnError("error scanning location", txn, err)
		}
		loc := &storypb.Location{}
		if err := proto.Unmarshal(blob, loc); err != nil {
			return nil, txnError(fmt.Sprintf("could not unmarshal location %s", lid), txn, err)
		}
		loc.Id = proto.String(lid)
		resp.Locations = append(resp.Locations, loc)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit query", txn, err)
	}
	return resp, nil
}

func getStoryImpl(ctx context.Context, db *sql.DB, sid int64, view spb.StoryView) (*spb.GetStoryResponse, error) {
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	str, err := loadStory(ctx, txn, sid)
	if err != nil {
		return nil, txnError(fmt.Sprintf("could not find story %d", sid), txn, err)
	}
	resp := &spb.GetStoryResponse{
		Story: str,
	}

	if view == spb.StoryView_VIEW_CONTENT {
		locs, err := loadStoryLocations(ctx, txn, sid)
		if err != nil {
			return nil, txnError(fmt.Sprintf("could not load story %d locations", sid), txn, err)
		}
		acts, err := loadStoryActions(ctx, txn, sid)
		if err != nil {
			return nil, txnError(fmt.Sprintf("could not load story %d actions", sid), txn, err)
		}
		resp.Content = &spb.StoryContent{
			Locations: locs,
			Actions:   acts,
		}
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit query", txn, err)
	}
	str.Id = proto.Int64(sid)
	return resp, nil
}

func deleteStoryImpl(ctx context.Context, db *sql.DB, id int64) (*spb.DeleteStoryResponse, error) {
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	if _, err := txn.ExecContext(ctx, `DELETE FROM Stories WHERE id = ?`, id); err != nil {
		return nil, txnError(fmt.Sprintf("could not delete Story %d", id), txn, err)
	}

	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit deletion", txn, err)
	}
	return &spb.DeleteStoryResponse{}, nil
}

// updateStoryImpl writes the provided story to the transaction, creating it if needed.
func updateStoryImpl(ctx context.Context, txn *sql.Tx, upd *storypb.Story) (*spb.UpdateStoryResponse, error) {
	sid := upd.GetId()
	var wrt *storypb.Story
	var err error
	if sid == 0 {
		if wrt, err = createStory(ctx, txn, upd); err != nil {
			return nil, fmt.Errorf("could not create new story: %w", err)
		}
	} else if wrt, err = loadStory(ctx, txn, sid); err != nil {
		return nil, fmt.Errorf("could not read story %d: %w", sid, err)
	}

	// Merge everything except events, which are overwritten.
	proto.Merge(wrt, upd)
	if upd.GetEvents() != nil {
		wrt.Events = upd.GetEvents()
	}

	blob, err := proto.Marshal(wrt)
	if err != nil {
		return nil, fmt.Errorf("could not marshal updated story %d: %w", sid, err)
	}

	if _, err = txn.ExecContext(ctx, `UPDATE Stories SET title = ?, proto = ? WHERE id = ?`, wrt.GetTitle(), blob, wrt.GetId()); err != nil {
		return nil, fmt.Errorf("could not update stories table: %w", err)
	}

	return &spb.UpdateStoryResponse{
		Story: wrt,
	}, nil
}

func listStoriesImpl(ctx context.Context, db *sql.DB, req *spb.ListStoriesRequest) (*spb.ListStoriesResponse, error) {
	resp := &spb.ListStoriesResponse{
		Stories: make([]*storypb.Story, 0, 10),
	}
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	rows, err := txn.QueryContext(ctx, `SELECT l.id, l.proto FROM Stories AS l ORDER BY l.id ASC`)
	if err != nil {
		return nil, txnError("could not list stories", txn, err)
	}
	for rows.Next() {
		var id int64
		blob := []byte{}
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, txnError("error scanning story", txn, err)
		}
		str := &storypb.Story{}
		if err := proto.Unmarshal(blob, str); err != nil {
			return nil, txnError(fmt.Sprintf("could not unmarshal story %d", id), txn, err)
		}
		// Clone the limited view.
		resp.Stories = append(resp.Stories, &storypb.Story{
			Id:          proto.Int64(id),
			Title:       proto.String(str.GetTitle()),
			Description: proto.String(str.GetDescription()),
		})
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit query", txn, err)
	}
	return resp, nil
}

func createActionImpl(ctx context.Context, db *sql.DB, act *storypb.Action) (*spb.CreateActionResponse, error) {
	aid := act.GetId()
	if len(aid) < 1 {
		aid = uuid.New().String()
	}
	blob, err := proto.Marshal(act)
	if err != nil {
		return nil, fmt.Errorf("could not marshal Action: %v", err)
	}
	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}

	if _, err := txn.ExecContext(ctx, `INSERT INTO Actions (id, proto) VALUES (?, ?)`, aid, blob); err != nil {
		return nil, txnError("could not insert into Actions", txn, err)
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not write to database", txn, err)
	}

	act.Id = proto.String(aid)
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
	if lid := str.GetStartLocationId(); len(lid) > 0 {
		ngame.LocationId = proto.String(lid)
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

func listGamesImpl(ctx context.Context, db *sql.DB, req *spb.ListGamesRequest) (*spb.ListGamesResponse, error) {
	resp := &spb.ListGamesResponse{
		Games: make([]*storypb.Playthrough, 0, 10),
	}
	txn, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	rows, err := txn.QueryContext(ctx, `SELECT l.id, l.proto FROM Playthroughs AS l ORDER BY l.id ASC`)
	if err != nil {
		return nil, txnError("could not list playthroughs", txn, err)
	}
	for rows.Next() {
		var id int64
		blob := []byte{}
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, txnError("error scanning playthrough", txn, err)
		}
		gam := &storypb.Playthrough{}
		if err := proto.Unmarshal(blob, gam); err != nil {
			return nil, txnError(fmt.Sprintf("could not unmarshal playthrough %d", id), txn, err)
		}
		// Clone a limited view.
		resp.Games = append(resp.Games, &storypb.Playthrough{
			Id:      proto.Int64(id),
			StoryId: proto.Int64(gam.GetStoryId()),
		})
	}
	if err := txn.Commit(); err != nil {
		return nil, txnError("could not commit query", txn, err)
	}
	return resp, nil
}

func loadPossibleActions(ctx context.Context, txn *sql.Tx, loc *storypb.Location) ([]*storypb.Action, error) {
	aids := make([]string, 0, len(loc.GetPossibleActions()))
	for _, pact := range loc.GetPossibleActions() {
		aids = append(aids, pact.GetActionId())
	}
	return loadActions(ctx, txn, aids...)
}

// loadStoryState loads the player action, location, state, and story for a game.
// It is read-only.
func loadStoryState(ctx context.Context, txn *sql.Tx, gid int64, aid string) (*storypb.GameEvent, error) {
	game, narration, err := loadGame(ctx, txn, gid)
	if err != nil {
		return nil, fmt.Errorf("could not find game %d: %w", gid, err)
	}
	sid := game.GetStoryId()
	str, err := loadStory(ctx, txn, sid)
	if err != nil {
		return nil, fmt.Errorf("could not find story %d for playthrough %d: %w", sid, gid, err)
	}
	var act *storypb.Action
	if len(aid) > 0 {
		act, err = loadAction(ctx, txn, aid)
		if err != nil {
			return nil, fmt.Errorf("could not find action %s for playthrough %d of story %d: %w", aid, gid, sid, err)
		}
	}
	lid := game.GetLocationId()
	loc, err := loadLocation(ctx, txn, lid)
	if err != nil {
		return nil, fmt.Errorf("could not find location %s for playthrough %d of story %d: %w", lid, gid, sid, err)
	}

	candActs, err := loadPossibleActions(ctx, txn, loc)
	if err != nil {
		return nil, fmt.Errorf("could not load candidate actions for location %s (%s): %w", loc.GetId(), loc.GetTitle(), err)
	}

	return &storypb.GameEvent{
		PlayerAction:     act,
		Location:         loc,
		Values:           game.GetValues(),
		Story:            str,
		Narration:        proto.String(narration),
		CandidateActions: candActs,
		State:            game.GetState().Enum(),
	}, nil
}

func writeAction(ctx context.Context, txn *sql.Tx, gid int64, gstate *storypb.GameEvent, narration string) error {
	game := &storypb.Playthrough{
		Id:         proto.Int64(gid),
		StoryId:    proto.Int64(gstate.GetStory().GetId()),
		LocationId: proto.String(gstate.GetLocation().GetId()),
		Values:     gstate.GetValues(),
		State:      gstate.GetState().Enum(),
	}
	blob, err := proto.Marshal(game)
	if err != nil {
		return fmt.Errorf("could not marshal updated playthrough %d of story %d: %w", gid, game.GetStoryId(), err)
	}
	if _, err := txn.ExecContext(ctx, `UPDATE Playthroughs SET proto = ?, narration = ? WHERE id = ?`, blob, narration, gid); err != nil {
		return fmt.Errorf("could not update playthrough %d: %w", gid, err)
	}
	return nil
}
