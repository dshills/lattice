package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	mysqlstore "github.com/dshills/lattice/internal/store/mysql"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down> [n]")
		os.Exit(1)
	}

	db := openDB()
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	dir := os.Getenv("LATTICE_MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}

	switch os.Args[1] {
	case "up":
		if err := mysqlstore.MigrateUp(ctx, db, dir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrations applied")
	case "down":
		n := 1
		if len(os.Args) >= 3 {
			var err error
			n, err = strconv.Atoi(os.Args[2])
			if err != nil || n < 1 {
				log.Fatalf("invalid count: %s", os.Args[2])
			}
		}
		if err := mysqlstore.MigrateDown(ctx, db, dir, n); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		log.Printf("rolled back %d migration(s)", n)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func openDB() *sql.DB {
	host := os.Getenv("LATTICE_DB_HOST")
	port := os.Getenv("LATTICE_DB_PORT")
	user := os.Getenv("LATTICE_DB_USER")
	pass := os.Getenv("LATTICE_DB_PASSWORD")
	name := os.Getenv("LATTICE_DB_NAME")
	if host == "" || user == "" || name == "" {
		log.Fatal("LATTICE_DB_HOST, LATTICE_DB_USER, and LATTICE_DB_NAME are required")
	}
	if port == "" {
		port = "3306"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		url.QueryEscape(user), url.QueryEscape(pass), host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}
	return db
}
