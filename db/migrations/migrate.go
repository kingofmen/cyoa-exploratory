// Program migrate runs a Goose migration of the CloudSQL database.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

func printDebugInfo(migrationFiles string) error {
	// Debug info.
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %v", err)
	}
	log.Printf("Working directory: %q", cwd)

	entries, err := os.ReadDir(filepath.FromSlash(cwd))
	if err != nil {
		return fmt.Errorf("could not read working directory: %v", err)
	}
	for idx, entry := range entries {
		log.Printf("Workdir entry %d: %v", idx, entry)
	}
	entries, err = os.ReadDir(filepath.FromSlash(migrationFiles))
	if err != nil {
		fmt.Errorf("could not read migration directory %q: %v", migrationFiles, err)
	}
	for idx, entry := range entries {
		log.Printf("Migration entry %d: %v", idx, entry)
	}

	return nil
}

func main() {
	// dbUser should be the service account.
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	instanceConnectionName := os.Getenv("INSTANCE_CONNECTION_NAME")
	migrationFiles := os.Getenv("GOOSE_MIGRATION_FILES")
	if len(migrationFiles) == 0 {
		log.Fatalf("migration file location not set")
	}

	dsn := fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true", dbUser, instanceConnectionName, dbName)
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		log.Fatalf("failed to parse DSN: %v", err)
	}

	// Cloud SQL Go Connector with IAM auth.
	// Note this will only work in the Cloud Run job.
	ctx := context.Background()
	d, err := cloudsqlconn.NewDialer(ctx, cloudsqlconn.WithIAMAuthN())
	if err != nil {
		log.Fatalf("failed to initialize dialer: %v", err)
	}
	defer d.Close()

	mysql.RegisterDialContext("cloudsql", func(ctx context.Context, addr string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName)
	})

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	if err := printDebugInfo(migrationFiles); err != nil {
		log.Printf("error getting debug info: %v", err)
	}

	if err := goose.UpContext(ctx, db, filepath.FromSlash(migrationFiles)); err != nil {
		log.Fatalf("goose up (dsn %q, directory %q) failed: %v", dsn, migrationFiles, err)
	}
	log.Println("Successful database migration.")
}
