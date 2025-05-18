// Package main contains the request handler.
package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/soheilhy/cmux"
	//"google.golang.org/grpc"
	//pb "github.com/your-username/your-repo-root/helloworld" // Adjust path
	"github.com/kingofmen/cyoa-exploratory/db"
	"github.com/kingofmen/cyoa-exploratory/frontend"
)

var dbPool *sql.DB // Global variable to hold the connection pool

func main() {
	// --- Port Configuration ---
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080" // Default port for Cloud Run
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid PORT environment variable: %v", err)
	}
	addr := ":" + strconv.Itoa(port)

	// TODO: Read from, like, actual config.
	dbCfg := &initialize.Config{
		Direct: &initialize.Local{
			User:     os.Getenv("CYOA_DB_USER"),
			Password: os.Getenv("CYOA_DB_PASSWD"),
			Host:     "localhost",
			Port:     3306,
			Name:     os.Getenv("CYOA_DB_NAME"),
		},
	}
	dbPool, cleanup, err := initialize.ConnectionPool(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if cleanup != nil {
		defer cleanup() // Ensure Cloud SQL connector resources are cleaned up
	}
	defer dbPool.Close() // Close the connection pool on shutdown

	// --- Database Connection (Placeholder) ---
	// Establish DB connection pool early
	/*
		var cleanup func() error
		dbPool, cleanup, err = db.ConnectDB() // Call the function from db/connection.go
		if dbPool!= nil {
			log.Println("Database connection pool established.")
			// Optional: Ping DB to verify connection early
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := dbPool.PingContext(ctx); err!= nil {
				log.Printf("Warning: Failed to ping database: %v", err)
				// Decide if this should be fatal or just a warning
			} else {
				log.Println("Database ping successful.")
			}
		} else {
			log.Println("Database connection pool is nil (running without DB).")
		}
	*/
	// --- Main Listener ---
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", addr, err)
	}
	log.Printf("Listening on %s", addr)

	// --- Multiplexer Setup (cmux) ---
	m := cmux.New(lis)

	// Match gRPC requests (HTTP/2 with specific header)
	//grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	//log.Println("Matcher created for gRPC")

	// Match HTTP/1.1 requests
	httpL := m.Match(cmux.HTTP1Fast())
	log.Println("Matcher created for HTTP/1.1")

	// Match any other HTTP/2 requests (optional, could be for grpc-web without content-type?)
	// http2L := m.Match(cmux.HTTP2())
	// log.Println("Matcher created for other HTTP/2")

	// --- gRPC Server Setup ---
	// Remove for initial hello-world implementation.
	/*
		grpcS := grpc.NewServer()
		pb.RegisterGreeterServer(grpcS, NewGreeterServer()) // Register our greeter implementation
		log.Println("gRPC server configured")
	*/
	// --- HTTP Server Setup ---
	httpMux := http.NewServeMux()
	// Frontend server.
	//fsRoot := http.FileServer(http.Dir("./static"))
	feRoot := server.NewHandler()
	httpMux.Handle("/", feRoot) // Serve static files at the root

	// Add other HTTP handlers here if needed
	// httpMux.HandleFunc("/api/resource", handleAPIResource)

	httpS := &http.Server{
		Handler: httpMux,
	}
	log.Println("HTTP server configured")

	// --- Start Servers in Goroutines ---
	/*
		go func() {
			log.Println("Starting gRPC server...")
			if err := grpcS.Serve(grpcL); err!= nil && err!= cmux.ErrListenerClosed {
				log.Fatalf("gRPC server error: %v", err)
			}
			log.Println("gRPC server stopped.")
		}()
	*/

	go func() {
		log.Println("Starting HTTP server...")
		if err := httpS.Serve(httpL); err != nil && err != http.ErrServerClosed && err != cmux.ErrListenerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("HTTP server stopped.")
	}()

	// go func() { // If matching other HTTP/2
	// 	log.Println("Starting HTTP/2 server...")
	// 	if err := httpS.Serve(http2L); err!= nil && err!= http.ErrServerClosed && err!= cmux.ErrListenerClosed {
	// 		log.Fatalf("HTTP/2 server error: %v", err)
	// 	}
	// 	log.Println("HTTP/2 server stopped.")
	// }()

	// --- Start Multiplexer ---
	log.Println("Starting cmux server...")
	go func() {
		if err := m.Serve(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Fatalf("cmux Serve error: %v", err)
		}
		log.Println("cmux server stopped.")
	}()

	// --- Graceful Shutdown Handling ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Waiting for shutdown signal...")

	<-quit // Block until signal received

	log.Println("Shutdown signal received, initiating graceful shutdown...")

	// Gracefully stop gRPC server
	//grpcS.GracefulStop()
	//log.Println("gRPC server gracefully stopped.")

	// Gracefully stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10-second timeout
	defer cancel()
	if err := httpS.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server gracefully stopped.")
	}

	// Close the main listener (implicitly closes cmux listeners)
	// No need to explicitly close m or sub-listeners if lis is closed.
	// lis.Close() // Closing the listener might happen automatically via cmux shutdown or signal handling

	log.Println("Application shut down gracefully.")
}
