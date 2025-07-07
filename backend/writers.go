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
