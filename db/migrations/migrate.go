// Program migrate runs a Goose migration of the CloudSQL database.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kingofmen/cyoa-exploratory/db"
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

func migrate() error {
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbConn := os.Getenv("DB_CONN_TYPE")
	dbAddr := os.Getenv("INSTANCE_CONNECTION_NAME")
	migrationFiles := os.Getenv("GOOSE_MIGRATION_FILES")
	if len(migrationFiles) == 0 {
		return fmt.Errorf("migration file location not set")
	}

	// No password or port for Cloud SQL IAM auth.
	config, err := initialize.FromEnv(dbUser, "", dbConn, dbAddr, "", dbName)
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %v", err)
	}

	ctx := context.Background()
	db, cleanup, err := initialize.ConnectionPool(ctx, config)
	if err != nil {
		return fmt.Errorf("Could not initialize DB connection: %v", err)
	}
	defer cleanup()

	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %v", err)
	}

	if err := printDebugInfo(migrationFiles); err != nil {
		log.Printf("error getting debug info: %v", err)
	}

	if err := goose.UpContext(ctx, db, filepath.FromSlash(migrationFiles)); err != nil {
		return fmt.Errorf("goose up (dsn %q, directory %q) failed: %v", config.FormatDSN(), migrationFiles, err)
	}
	return nil
}

func main() {
	if err := migrate(); err != nil {
		log.Fatalf("error migrating database: %v", err)
	}
	log.Println("Successful database migration.")
}
