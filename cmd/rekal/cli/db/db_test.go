package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenData_CreateAndPing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rekalDir := filepath.Join(dir, ".rekal")
	if err := os.MkdirAll(rekalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	db, err := OpenData(dir)
	if err != nil {
		t.Fatalf("OpenData: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenIndex_CreateAndPing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rekalDir := filepath.Join(dir, ".rekal")
	if err := os.MkdirAll(rekalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	db, err := OpenIndex(dir)
	if err != nil {
		t.Fatalf("OpenIndex: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestInitDataSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rekalDir := filepath.Join(dir, ".rekal")
	if err := os.MkdirAll(rekalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	db, err := OpenData(dir)
	if err != nil {
		t.Fatalf("OpenData: %v", err)
	}
	defer db.Close()

	if err := InitDataSchema(db); err != nil {
		t.Fatalf("InitDataSchema: %v", err)
	}

	// Verify tables exist.
	tables := []string{"sessions", "checkpoints", "files_touched", "checkpoint_sessions"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("table %s should exist: %v", table, err)
		}
	}
}

func TestInitIndexSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rekalDir := filepath.Join(dir, ".rekal")
	if err := os.MkdirAll(rekalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	db, err := OpenIndex(dir)
	if err != nil {
		t.Fatalf("OpenIndex: %v", err)
	}
	defer db.Close()

	if err := InitIndexSchema(db); err != nil {
		t.Fatalf("InitIndexSchema: %v", err)
	}

	// Verify tables exist.
	tables := []string{"turns_ft", "tool_calls_index", "files_index", "session_facets", "file_cooccurrence"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("table %s should exist: %v", table, err)
		}
	}
}
