package db

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

func InitDB() *sql.DB {
	// Open Connection
	db, err := sql.Open("sqlite", "../data/ledger.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
		os.Exit(1)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
		os.Exit(1)
	}
	fmt.Println("connected to database")

	// Load Schema
	path := filepath.Join(".", "models.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read schema file: %v", err)
		os.Exit(1)
	}

	schema := string(content)
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("failed to execute schema file: %v", err)
		os.Exit(1)
	}
	fmt.Println("executed schema file")
	return db
}
