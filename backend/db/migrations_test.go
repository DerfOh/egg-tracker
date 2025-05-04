package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestMigrate_CreatesUsersTable(t *testing.T) {
	testDBPath := "test_migrate.db"
	defer os.Remove(testDBPath)

	db, err := sql.Open("sqlite3", testDBPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users';")
	var tableName string
	if err := row.Scan(&tableName); err != nil {
		t.Fatalf("users table not found after migration: %v", err)
	}
	if tableName != "users" {
		t.Fatalf("expected table 'users', got '%s'", tableName)
	}
}
