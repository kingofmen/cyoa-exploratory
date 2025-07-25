package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

func loadStory(ctx context.Context, txn *sql.Tx, sid int64) (*storypb.Story, error) {
	row := txn.QueryRowContext(ctx, `SELECT s.id, s.proto FROM Stories AS s WHERE s.id = ?`, sid)
	blob := []byte{}
	if err := row.Scan(&sid, &blob); err != nil {
		return nil, err
	}
	str := &storypb.Story{}
	if err := proto.Unmarshal(blob, str); err != nil {
		return nil, fmt.Errorf("could not unmarshal story %d: %w", sid, err)
	}
	str.Id = proto.Int64(sid)
	return str, nil
}

func loadGame(ctx context.Context, txn *sql.Tx, gid int64) (*storypb.Playthrough, string, error) {
	row := txn.QueryRowContext(ctx, `SELECT * FROM Playthroughs AS p WHERE p.id = ?`, gid)
	blob := []byte{}
	var text sql.NullString
	if err := row.Scan(&gid, &blob, &text); err != nil {
		return nil, "", err
	}
	game := &storypb.Playthrough{}
	if err := proto.Unmarshal(blob, game); err != nil {
		return nil, "", fmt.Errorf("could not unmarshal game %d: %w", gid, err)
	}
	game.Id = proto.Int64(gid)
	return game, text.String, nil
}

func loadActions(ctx context.Context, txn *sql.Tx, aids ...string) ([]*storypb.Action, error) {
	if len(aids) == 0 {
		return []*storypb.Action{}, nil
	}

	placeholders := strings.Repeat("?,", len(aids)-1) + "?"
	query := fmt.Sprintf(`SELECT id, proto FROM Actions WHERE id IN (%s)`, placeholders)

	// Convert to interface{} for variadic argument.
	args := make([]interface{}, len(aids))
	for idx, aid := range aids {
		args[idx] = aid
	}
	rows, err := txn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query actions %v: %w", aids, err)
	}
	defer rows.Close()

	var actions []*storypb.Action
	for rows.Next() {
		var id string
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, fmt.Errorf("could not scan action row: %w", err)
		}

		action := &storypb.Action{}
		if err := proto.Unmarshal(blob, action); err != nil {
			return nil, fmt.Errorf("could not unmarshal action %s: %w", id, err)
		}

		action.Id = proto.String(id)
		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating action rows: %w", err)
	}
	return actions, nil
}

func loadAction(ctx context.Context, txn *sql.Tx, aid string) (*storypb.Action, error) {
	row := txn.QueryRowContext(ctx, `SELECT * FROM Actions AS a WHERE a.id = ?`, aid)
	blob := []byte{}
	if err := row.Scan(&aid, &blob); err != nil {
		return nil, err
	}
	action := &storypb.Action{}
	if err := proto.Unmarshal(blob, action); err != nil {
		return nil, fmt.Errorf("could not unmarshal action %s: %w", aid, err)
	}
	action.Id = proto.String(aid)
	return action, nil
}

func loadLocation(ctx context.Context, txn *sql.Tx, lid string) (*storypb.Location, error) {
	row := txn.QueryRowContext(ctx, `SELECT l.id, l.proto FROM Locations AS l WHERE l.id = ?`, lid)
	blob := []byte{}
	if err := row.Scan(&lid, &blob); err != nil {
		return nil, err
	}
	loc := &storypb.Location{}
	if err := proto.Unmarshal(blob, loc); err != nil {
		return nil, fmt.Errorf("could not unmarshal location %s: %w", lid, err)
	}
	loc.Id = proto.String(lid)
	return loc, nil
}

func loadStoryLocations(ctx context.Context, txn *sql.Tx, sid int64) ([]*storypb.Location, error) {
	rows, err := txn.QueryContext(ctx, `SELECT l.id, l.proto
                                      FROM StoryLocations AS sl
                                      JOIN Locations AS l
                                      ON sl.location_id = l.id
                                      WHERE sl.story_id = ?`,
		sid)
	if err != nil {
		return nil, fmt.Errorf("story %d locations query failed: %w", sid, err)
	}
	ret := make([]*storypb.Location, 0, 10)
	for rows.Next() {
		var lid string
		blob := []byte{}
		if err := rows.Scan(&lid, &blob); err != nil {
			return nil, fmt.Errorf("error scanning location for story %d: %w", sid, err)
		}
		loc := &storypb.Location{}
		if err := proto.Unmarshal(blob, loc); err != nil {
			return nil, fmt.Errorf("could not unmarshal location %s for story %d: %w", lid, sid, err)
		}
		loc.Id = proto.String(lid)
		ret = append(ret, loc)

	}
	return ret, nil
}

func loadStoryActions(ctx context.Context, txn *sql.Tx, sid int64) ([]*storypb.Action, error) {
	rows, err := txn.QueryContext(ctx, `SELECT a.id, a.proto
                                      FROM StoryActions AS sa
                                      JOIN Actions AS a
                                      ON sa.action_id = a.id
                                      WHERE sa.story_id = ?`,
		sid)
	if err != nil {
		return nil, fmt.Errorf("story %d actions query failed: %w", sid, err)
	}
	ret := make([]*storypb.Action, 0, 10)
	for rows.Next() {
		var aid string
		blob := []byte{}
		if err := rows.Scan(&aid, &blob); err != nil {
			return nil, fmt.Errorf("error scanning action for story %d: %w", sid, err)
		}
		act := &storypb.Action{}
		if err := proto.Unmarshal(blob, act); err != nil {
			return nil, fmt.Errorf("could not unmarshal action %s for story %d: %w", aid, sid, err)
		}
		act.Id = proto.String(aid)
		ret = append(ret, act)
	}
	return ret, nil
}
