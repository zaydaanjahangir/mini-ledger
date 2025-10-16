package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	_"modernc.org/sqlite"
)

func InitDB() *sql.DB{
	// Open Connection
	db, err := sql.Open("sqlite", "../data/ledger.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
		os.Exit(1)
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