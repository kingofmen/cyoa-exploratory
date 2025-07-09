package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"

	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

func createStory(ctx context.Context, txn *sql.Tx, str *storypb.Story) (*storypb.Story, error) {
	if _, err := txn.ExecContext(ctx, `INSERT INTO Stories (title) VALUES (?)`, str.GetTitle()); err != nil {
		return nil, fmt.Errorf("could not insert into Stories: %w", err)
	}
	var sid int64
	row := txn.QueryRowContext(ctx, `SELECT LAST_INSERT_ID()`)
	if err := row.Scan(&sid); err != nil {
		return nil, fmt.Errorf("could not read back created ID: %w", err)
	}

	return &storypb.Story{
		Id: proto.Int64(sid),
	}, nil
}

func createOrUpdateLocation(ctx context.Context, txn *sql.Tx, lid string, loc *storypb.Location) (*storypb.Location, error) {
	blob, err := proto.Marshal(loc)
	if err != nil {
		return nil, fmt.Errorf("could not marshal updated location %s (%s): %w", lid, loc.GetTitle(), err)
	}
	if _, err := txn.ExecContext(ctx, `INSERT INTO Locations (id, title, proto)
                                     VALUES (?, ?, ?)
                                     ON DUPLICATE KEY UPDATE title = VALUES(title), proto = VALUES(proto);
                                    `, lid, loc.GetTitle(), blob); err != nil {
		return nil, fmt.Errorf("could not insert into Locations: %w", err)
	}

	loc.Id = proto.String(lid)
	return loc, nil
}

func updateStoryLocationsTable(ctx context.Context, txn *sql.Tx, sid int64, locIds []string) error {
	_, err := txn.ExecContext(ctx, `DELETE FROM StoryLocations WHERE story_id = ?`, sid)
	if err != nil {
		return fmt.Errorf("failed to delete existing story locations: %w", err)
	}
	insrt, err := txn.PrepareContext(ctx, `INSERT INTO StoryLocations (story_id, location_id) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insrt.Close()
	for _, lid := range locIds {
		_, err := insrt.ExecContext(ctx, sid, lid)
		if err != nil {
			return fmt.Errorf("failed to insert story-location association (%d, %s): %w", sid, lid, err)
		}
	}
	return nil
}
