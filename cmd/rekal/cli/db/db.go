package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/marcboeker/go-duckdb"
)

// OpenData opens (or creates) the data DB at <gitRoot>/.rekal/data.db.
func OpenData(gitRoot string) (*sql.DB, error) {
	path := filepath.Join(gitRoot, ".rekal", "data.db")
	return open(path)
}

// OpenIndex opens (or creates) the index DB at <gitRoot>/.rekal/index.db.
func OpenIndex(gitRoot string) (*sql.DB, error) {
	path := filepath.Join(gitRoot, ".rekal", "index.db")
	return open(path)
}

func open(path string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database %s: %w", path, err)
	}
	return db, nil
}
