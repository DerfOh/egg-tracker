package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Migrate(db *sql.DB) error {
	const userTable = `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.Exec(userTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	const eggTable = `
    CREATE TABLE IF NOT EXISTS eggs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date_laid DATE NOT NULL,
        species TEXT,
        deleted BOOLEAN NOT NULL DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err = db.Exec(eggTable)
	if err != nil {
		return fmt.Errorf("failed to create eggs table: %w", err)
	}

	const inventoryActionTable = `
    CREATE TABLE IF NOT EXISTS inventory_actions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        quantity INTEGER NOT NULL,
        species TEXT NOT NULL,
        coop TEXT,
        egg_color TEXT,
        egg_size TEXT,
        action TEXT NOT NULL,
        notes TEXT,
        date DATE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err = db.Exec(inventoryActionTable)
	if err != nil {
		return fmt.Errorf("failed to create inventory_actions table: %w", err)
	}

	const speciesTable = `
    CREATE TABLE IF NOT EXISTS species (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        active BOOLEAN NOT NULL DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err = db.Exec(speciesTable)
	if err != nil {
		return fmt.Errorf("failed to create species table: %w", err)
	}

	const eggColorTable = `
    CREATE TABLE IF NOT EXISTS egg_colors (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        active BOOLEAN NOT NULL DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err = db.Exec(eggColorTable)
	if err != nil {
		return fmt.Errorf("failed to create egg_colors table: %w", err)
	}

	const eggSizeTable = `
    CREATE TABLE IF NOT EXISTS egg_sizes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        active BOOLEAN NOT NULL DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err = db.Exec(eggSizeTable)
	if err != nil {
		return fmt.Errorf("failed to create egg_sizes table: %w", err)
	}

	const coopTable = `
    CREATE TABLE IF NOT EXISTS coops (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        active BOOLEAN NOT NULL DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err = db.Exec(coopTable)
	if err != nil {
		return fmt.Errorf("failed to create coops table: %w", err)
	}

	return nil
}

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}
