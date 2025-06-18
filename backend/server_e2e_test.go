package handlers

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"google.golang.org/protobuf/proto"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
	storypb "github.com/kingofmen/cyoa-exploratory/story/proto"
)

var db *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0"),
		mysql.WithDatabase("test_db"),
		mysql.WithUsername("test_user"),
		mysql.WithPassword("test_password"),
	)
	if err != nil {
		log.Fatalf("could not start mysql container: %v", err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("could not stop mysql container: %v", err)
		}
	}()

	connStr, err := container.ConnectionString(ctx, "multiStatements=true")
	if err != nil {
		log.Fatalf("could not get connection string: %v", err)
	}

	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("could not open database connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}
	if err := goose.Up(db, "../db/migrations"); err != nil {
		log.Fatalf("failed to run goose migrations: %v", err)
	}
	log.Println("Goose migrations applied successfully.")

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestStoryE2E(t *testing.T) {
	ctx := context.Background()
	srv := New(db)
	csresp, err := srv.CreateStory(ctx, &spb.CreateStoryRequest{
		Story: &storypb.Story{
			Title:       proto.String("E2E test story"),
			Description: proto.String("Story for end-to-end testing"),
		},
	})
	if err != nil {
		t.Fatalf("CreateStory() => %v, want nil", err)
	}
	stid := csresp.GetStory().GetId()
	if stid != 1 {
		t.Fatalf("CreateStory() returned story ID %d, want 1", stid)
	}

	if _, err := srv.CreateLocation(ctx, &spb.CreateLocationRequest{
		Location: &storypb.Location{
			Title:   proto.String("Choose Character"),
			Content: proto.String("Choose which character to play as."),
		},
	}); err != nil {
		t.Fatalf("CreateLocation() => %v, want nil", err)
	}

}
