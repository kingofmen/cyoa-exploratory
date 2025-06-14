// Package initialize contains utility methods for initializing the database connection.
package initialize

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/go-sql-driver/mysql"
)

// FromEnv returns a MySQL configuration object based on the
// provided environment strings.
func FromEnv(user, pwd, network, addr, port, dbname string) (*mysql.Config, error) {
	if len(pwd) > 0 {
		pwd = fmt.Sprintf(":%s", pwd)
	}
	if len(port) > 0 {
		port = fmt.Sprintf(":%s", port)
	}

	dsn := fmt.Sprintf("%s%s@%s(%s%s)/%s?parseTime=true", user, pwd, network, addr, port, dbname)
	return mysql.ParseDSN(dsn)
}

func ConnectionPool(ctx context.Context, cfg *mysql.Config) (*sql.DB, func() error, error) {
	cleanup := func() error { return nil } // Default no-op cleanup.
	connString := cfg.FormatDSN()
	log.Printf("Initializing database connection.")

	if cfg.Net == "cloudsqlconn" {
		// Cloud SQL Go Connector with IAM auth.
		// Note this will only work in the Cloud Run job, or if
		// the Auth Proxy has been set up locally.
		d, err := cloudsqlconn.NewDialer(ctx, cloudsqlconn.WithIAMAuthN())
		if err != nil {
			log.Fatalf("failed to initialize dialer: %v", err)
		}
		cleanup = d.Close

		mysql.RegisterDialContext(cfg.Net, func(ctx context.Context, addr string) (net.Conn, error) {
			log.Printf("Dialing %q", addr)
			return d.Dial(ctx, addr)
		})
	}

	db, err := sql.Open("mysql", connString)
	if err != nil {
		// Ensure cleanup is called if Open fails.
		if cErr := cleanup(); cErr != nil {
			log.Printf("Error during cleanup after sql.Open failure: %v", cErr)
		}
		return nil, cleanup, fmt.Errorf("sql.Open(%s) failed: %w", connString, err)
	}

	// Configure connection pool settings.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection.
	if err = db.Ping(); err != nil {
		// Ensure cleanup is called if Ping fails
		if cErr := cleanup(); cErr != nil {
			log.Printf("Error during cleanup after db.Ping failure: %v", cErr)
		}
		db.Close() // Close the pool handle as well
		return nil, cleanup, fmt.Errorf("db.Ping failed: %w", err)
	}

	log.Println("Database initialization succeeded.")
	return db, cleanup, nil
}
