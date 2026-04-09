package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dshills/lattice/internal/api"
	"github.com/dshills/lattice/internal/graph"
	mysqlstore "github.com/dshills/lattice/internal/store/mysql"
)

func main() {
	dbHost := os.Getenv("LATTICE_DB_HOST")
	dbPort := os.Getenv("LATTICE_DB_PORT")
	dbUser := os.Getenv("LATTICE_DB_USER")
	dbPass := os.Getenv("LATTICE_DB_PASSWORD")
	dbName := os.Getenv("LATTICE_DB_NAME")
	if dbHost == "" || dbUser == "" || dbName == "" {
		log.Fatal("LATTICE_DB_HOST, LATTICE_DB_USER, and LATTICE_DB_NAME environment variables are required")
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		url.QueryEscape(dbUser), url.QueryEscape(dbPass), dbHost, dbPort, dbName)
	addr := os.Getenv("LATTICE_ADDR")
	if addr == "" {
		addr = ":8090"
	}
	migrationsDir := os.Getenv("LATTICE_MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	ctx := context.Background()
	if err := mysqlstore.MigrateUp(ctx, db, migrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	log.Println("migrations applied")

	h := &api.Handler{
		WorkItems:     mysqlstore.NewWorkItemStore(db),
		Relationships: mysqlstore.NewRelationshipStore(db),
		Cycles:        graph.NewCycleDetector(db),
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Apply middleware chain: logging → role extraction → content-type check.
	handler := api.LoggingMiddleware(api.RoleMiddleware(api.JSONContentType(mux)))

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	log.Println("server stopped")
}
