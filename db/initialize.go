// Package initialize contains utility methods for initializing the database connection.
package initialize

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // For local connection.
)

// Config stores database configuration data to be used for setup.
type Config struct {
	user     string
	password string
	host     string
	port     int
	name     string
	driver   string
}

// validate returns an error if the configuration is obviously bad.
func (c *Config) validate() error {
	if c == nil {
		return fmt.Errorf("uninitialized local configuration")
	}
	unspecified := []string{}
	if len(c.user) < 1 {
		unspecified = append(unspecified, "user")
	}
	if len(c.password) < 1 {
		unspecified = append(unspecified, "password")
	}
	if len(c.host) < 1 {
		unspecified = append(unspecified, "host")
	}
	if len(c.name) < 1 {
		unspecified = append(unspecified, "name")
	}
	if len(c.driver) < 1 {
		unspecified = append(unspecified, "driver")
	}
	if c.driver == "mysql" && c.port < 1 {
		unspecified = append(unspecified, "port")
	}
	if len(unspecified) > 0 {
		return fmt.Errorf("local configuration did not specify %v", unspecified)
	}
	return nil
}

// String returns the connection string for the configuration,
// and an error if the configuration is invalid.
func (c *Config) String() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}
	if c.driver == "mysql" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", c.user, c.password, c.host, c.port, c.name), nil
	}
	if c.driver == "cloudsql-mysql" {
		return fmt.Sprintf("%s:%s@cloudsql-mysql(%s)/%s?parseTime=true", c.user, c.password, c.host, c.name), nil
	}
	return "", fmt.Errorf("unknown DB driver %q", c.driver)
}

// FromEnv returns a Config object initialized to the environment.
func FromEnv(env string) *Config {
	cfg := &Config{}
	if env == "local" {
		cfg.host = "localhost"
		cfg.port = 3306
		cfg.driver = "mysql"
	} else if env == "cloud" {
		cfg.driver = "cloudsql-mysql"
	}

	return cfg
}

func (c *Config) WithUser(user string) *Config {
	if c == nil {
		c = &Config{}
	}
	c.user = user
	return c
}
func (c *Config) WithPassword(password string) *Config {
	if c == nil {
		c = &Config{}
	}
	c.password = password
	return c
}
func (c *Config) WithName(name string) *Config {
	if c == nil {
		c = &Config{}
	}
	c.name = name
	return c
}
func (c *Config) WithHost(host string) *Config {
	if c == nil {
		c = &Config{}
	}
	c.host = host
	return c
}
func (c *Config) WithPort(port int) *Config {
	if c == nil {
		c = &Config{}
	}
	c.port = port
	return c
}

func ConnectionPool(cfg *Config) (*sql.DB, func() error, error) {
	cleanup := func() error { return nil } // Default no-op cleanup.
	connString := ""
	var err error

	log.Println("Initializing database connection.")
	connString, err = cfg.String()
	if err != nil {
		return nil, cleanup, err
	}

	db, err := sql.Open(cfg.driver, connString)
	if err != nil {
		// Ensure cleanup is called if Open fails.
		if cErr := cleanup(); cErr != nil {
			log.Printf("Error during cleanup after sql.Open failure: %v", cErr)
		}
		return nil, cleanup, fmt.Errorf("sql.Open(%s) failed: %w", cfg.driver, err)
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
