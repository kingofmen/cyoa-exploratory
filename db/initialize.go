// Package initialize contains utility methods for initializing the database connection.
package initialize

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // For local connection.
)

// Local stores config data for a locally hosted database.
type Local struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
}

// validate returns an error if the configuration is obviously bad.
func (l *Local) validate() error {
	if l == nil {
		return fmt.Errorf("uninitialized local configuration")
	}
	unspecified := []string{}
	if len(l.User) < 1 {
		unspecified = append(unspecified, "user")
	}
	if len(l.Password) < 1 {
		unspecified = append(unspecified, "password")
	}
	if len(l.Host) < 1 {
		unspecified = append(unspecified, "host")
	}
	if len(l.Name) < 1 {
		unspecified = append(unspecified, "name")
	}
	if l.Port < 1 {
		unspecified = append(unspecified, "port")
	}
	if len(unspecified) > 0 {
		return fmt.Errorf("local configuration did not specify %v", unspecified)
	}
	return nil
}

// String returns the connection string for the configuration,
// and an error if the configuration is invalid.
func (l *Local) String() (string, error) {
	if err := l.validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", l.User, l.Password, l.Host, l.Port, l.Name), nil
}

// Config stores database configuration data to be used for setup.
type Config struct {
	Direct *Local
	// TODO: Add Cloud config.
}

func ConnectionPool(cfg *Config) (*sql.DB, func() error, error) {
	cleanup := func() error { return nil } // Default no-op cleanup.
	driverName := "mysql"
	connString := ""
	var err error

	if cfg.Direct != nil {
		log.Println("Initializing local database connection")
		connString, err = cfg.Direct.String()
		if err != nil {
			return nil, cleanup, err
		}
	} else {
		return nil, cleanup, fmt.Errorf("No database configuration found.")
	}

	db, err := sql.Open(driverName, connString)
	if err != nil {
		// Ensure cleanup is called if Open fails after RegisterDriver succeeded
		if cErr := cleanup(); cErr != nil {
			log.Printf("Error during cleanup after sql.Open failure: %v", cErr)
		}
		return nil, cleanup, fmt.Errorf("sql.Open(%s) failed: %w", driverName, err)
	}

	// Configure connection pool settings (optional but recommended)
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

	return db, cleanup, nil
}
