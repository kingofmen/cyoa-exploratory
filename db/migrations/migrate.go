// Program migrate runs a Goose migration of the CloudSQL database.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kingofmen/cyoa-exploratory/db"
	"github.com/pressly/goose/v3"
)

var (
	schemaRedo = flag.Int64("schema_redo", -1, "If non-negative, rolls back the database before re-applying migrations.")
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

func migrate(rollback int64) error {
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbConn := os.Getenv("DB_CONN_TYPE")
	dbAddr := os.Getenv("INSTANCE_CONNECTION_NAME")
	migrationFiles := os.Getenv("GOOSE_MIGRATION_FILES")
	if len(migrationFiles) == 0 {
		return fmt.Errorf("migration file location not set")
	}

	fullEnv := os.Environ()
	log.Println("Environment:")
	for _, e := range fullEnv {
		log.Println(e)
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

	if rollback >= 0 {
		log.Printf("Rolling back to database version %d", rollback)
		if err := goose.DownToContext(ctx, db, filepath.FromSlash(migrationFiles), rollback); err != nil {
			return fmt.Errorf("goose down to %d (dsn %q, directory %q) failed: %v", rollback, config.FormatDSN(), migrationFiles, err)
		}
	}

	if err := goose.UpContext(ctx, db, filepath.FromSlash(migrationFiles)); err != nil {
		return fmt.Errorf("goose up (dsn %q, directory %q) failed: %v", config.FormatDSN(), migrationFiles, err)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := migrate(*schemaRedo); err != nil {
		log.Fatalf("error migrating database: %v", err)
	}
	log.Println("Successful database migration.")
}
