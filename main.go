// Package main contains the request handler.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kingofmen/cyoa-exploratory/backend"
	"github.com/kingofmen/cyoa-exploratory/db"
	"github.com/kingofmen/cyoa-exploratory/frontend"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	spb "github.com/kingofmen/cyoa-exploratory/backend/proto"
)

var dbPool *sql.DB // Global variable to hold the connection pool

// FakeClient implements CyoaClient by just calling the handlers
// library directly.
type FakeClient struct {
	root *handlers.Server
}

func (fc *FakeClient) CreateLocation(ctx context.Context, in *spb.CreateLocationRequest, opts ...grpc.CallOption) (*spb.CreateLocationResponse, error) {
	if fc == nil {
		return nil, fmt.Errorf("nil client")
	}
	if fc.root == nil {
		return nil, fmt.Errorf("nil server")
	}
	return fc.root.CreateLocation(ctx, in)
}

func (fc *FakeClient) UpdateLocation(ctx context.Context, in *spb.UpdateLocationRequest, opts ...grpc.CallOption) (*spb.UpdateLocationResponse, error) {
	if fc == nil {
		return nil, fmt.Errorf("nil client")
	}
	if fc.root == nil {
		return nil, fmt.Errorf("nil server")
	}
	return fc.root.UpdateLocation(ctx, in)
}

func (fc *FakeClient) DeleteLocation(ctx context.Context, in *spb.DeleteLocationRequest, opts ...grpc.CallOption) (*spb.DeleteLocationResponse, error) {
	if fc == nil {
		return nil, fmt.Errorf("nil client")
	}
	if fc.root == nil {
		return nil, fmt.Errorf("nil server")
	}
	return fc.root.DeleteLocation(ctx, in)
}

func (fc *FakeClient) ListLocations(ctx context.Context, in *spb.ListLocationsRequest, opts ...grpc.CallOption) (*spb.ListLocationsResponse, error) {
	if fc == nil {
		return nil, fmt.Errorf("nil client")
	}
	if fc.root == nil {
		return nil, fmt.Errorf("nil server")
	}
	return fc.root.ListLocations(ctx, in)
}

func main() {
	// Read connection config from environment.
	user := os.Getenv("CYOA_DB_USER")
	network := os.Getenv("CYOA_DB_CONN_TYPE")
	instance := os.Getenv("CYOA_DB_INSTANCE")
	dbname := os.Getenv("CYOA_DB_NAME")

	// For local testing.
	passwd := os.Getenv("CYOA_DB_PASSWD")
	dbport := os.Getenv("CYOA_DB_PORT")

	log.Printf("Debug: %q %q", passwd, dbport)

	dbcfg, err := initialize.FromEnv(user, passwd, network, instance, dbport, dbname)
	if err != nil {
		log.Fatalf("Could not initialize DB configuration: %v", err)
	}

	ctx := context.Background()
	dbPool, cleanup, err := initialize.ConnectionPool(ctx, dbcfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if cleanup != nil {
		defer cleanup() // Ensure Cloud SQL connector resources are cleaned up
	}
	defer dbPool.Close() // Close the connection pool on shutdown

	addr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if len(addr) < 2 {
		addr = ":8080" // Default Cloud Run port.
	}
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

	// --- gRPC Server Setup ---
	beRoot := handlers.New(dbPool)
	// TODO: Set up as actual gRPC server with muxer instead of this fakery.
	fcli := &FakeClient{
		root: beRoot,
	}

	// --- HTTP Server Setup ---
	httpMux := http.NewServeMux()
	// Frontend server.
	feRoot := server.NewHandler(fcli)
	httpMux.HandleFunc(server.CreateLocationURL, feRoot.CreateLocation)
	httpMux.HandleFunc(server.UpdateLocationURL, feRoot.UpdateLocationHandler)
	httpMux.Handle("/", feRoot) // Serve static files at the root

	httpS := &http.Server{
		Handler: httpMux,
	}
	log.Println("HTTP server configured")

	go func() {
		log.Println("Starting HTTP server...")
		if err := httpS.Serve(httpL); err != nil && err != http.ErrServerClosed && err != cmux.ErrListenerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("HTTP server stopped.")
	}()

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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // 10-second timeout
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
