// Program migrate runs a Goose migration of the CloudSQL database.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

func main() {
	// dbUser should be the service account.
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	instanceConnectionName := os.Getenv("INSTANCE_CONNECTION_NAME")

	dsn := fmt.Sprintf("%s@unix(%s)/%s?parseTime=true", dbUser, instanceConnectionName, dbName)
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		log.Fatalf("failed to parse DSN: %v", err)
	}

	// Cloud SQL Go Connector with IAM auth.
	// Note this will only work in the Cloud Run job.
	d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithIAMAuthN())
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

	if err := goose.Up(db, "./"); err != nil {
		log.Fatalf("goose up failed: %v", err)
	}
	log.Println("Successful database migration.")
}
