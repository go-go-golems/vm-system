package vmstore

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// VMStore manages VM-related data in SQLite
type VMStore struct {
	db *sql.DB
}

// NewVMStore creates a new VMStore
func NewVMStore(dbPath string) (*VMStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &VMStore{db: db}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *VMStore) Close() error {
	return s.db.Close()
}
