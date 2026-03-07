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
	ensureDigestColumns(db)
	fmt.Println("executed schema file")
	return db
}

func ensureDigestColumns(db *sql.DB) {
	rows, err := db.Query("PRAGMA table_info(digests)")
	if err != nil {
		log.Fatalf("failed to inspect digests table: %v", err)
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			log.Fatalf("failed scanning digests table info: %v", err)
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("failed reading digests table info: %v", err)
	}

	if !cols["start_id"] {
		if _, err := db.Exec("ALTER TABLE digests ADD COLUMN start_id INTEGER"); err != nil {
			log.Fatalf("failed adding digests.start_id: %v", err)
		}
	}
	if !cols["end_id"] {
		if _, err := db.Exec("ALTER TABLE digests ADD COLUMN end_id INTEGER"); err != nil {
			log.Fatalf("failed adding digests.end_id: %v", err)
		}
	}
}
