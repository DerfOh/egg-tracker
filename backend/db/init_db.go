package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// InitializeDatabase ensures the necessary tables exist in the SQLite database
// and seeds them with sample data if they are empty.
func InitializeDatabase(dbPath string) error {
	log.Printf("[DB Init] Opening SQLite database at %s for initialization...", dbPath)
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on") // Enable foreign keys
	if err != nil {
		log.Printf("[DB Init] Error opening SQLite DB: %v", err)
		return err
	}
	defer db.Close()

	log.Println("[DB Init] Creating tables if they don't exist...")
	if err := createTables(db); err != nil {
		return err
	}

	log.Println("[DB Init] Seeding database with initial data if empty...")
	if err := seedData(db); err != nil {
		return err
	}

	log.Println("[DB Init] Database initialization complete.")
	return nil
}

func createTables(db *sql.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS species (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS coops (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS egg_colors (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        color TEXT UNIQUE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS egg_sizes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        size TEXT UNIQUE NOT NULL, -- e.g., Small, Medium, Large, Jumbo
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS eggs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        species_id INTEGER NOT NULL,
        coop_id INTEGER NOT NULL,
        collection_date DATE NOT NULL,
        quantity INTEGER NOT NULL DEFAULT 1,
        notes TEXT,
        color_id INTEGER,
        size_id INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (species_id) REFERENCES species(id),
        FOREIGN KEY (coop_id) REFERENCES coops(id),
        FOREIGN KEY (color_id) REFERENCES egg_colors(id),
        FOREIGN KEY (size_id) REFERENCES egg_sizes(id)
    );

    -- This is the crucial table for inventory tracking
    CREATE TABLE IF NOT EXISTS inventory_actions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        egg_id INTEGER NOT NULL,
        action_type TEXT NOT NULL, -- e.g., 'collected', 'sold', 'used', 'broken', 'gifted'
        quantity INTEGER NOT NULL, -- How many eggs involved in this action
        action_date DATETIME NOT NULL,
        notes TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (egg_id) REFERENCES eggs(id)
    );

    -- Triggers to update updated_at timestamps (Optional but good practice)
    CREATE TRIGGER IF NOT EXISTS update_species_updated_at
    AFTER UPDATE ON species
    FOR EACH ROW
    BEGIN
        UPDATE species SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
    END;

    CREATE TRIGGER IF NOT EXISTS update_coops_updated_at
    AFTER UPDATE ON coops
    FOR EACH ROW
    BEGIN
        UPDATE coops SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
    END;

    CREATE TRIGGER IF NOT EXISTS update_egg_colors_updated_at
    AFTER UPDATE ON egg_colors
    FOR EACH ROW
    BEGIN
        UPDATE egg_colors SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
    END;

    CREATE TRIGGER IF NOT EXISTS update_egg_sizes_updated_at
    AFTER UPDATE ON egg_sizes
    FOR EACH ROW
    BEGIN
        UPDATE egg_sizes SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
    END;

    CREATE TRIGGER IF NOT EXISTS update_eggs_updated_at
    AFTER UPDATE ON eggs
    FOR EACH ROW
    BEGIN
        UPDATE eggs SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
    END;

    CREATE TRIGGER IF NOT EXISTS update_inventory_actions_updated_at
    AFTER UPDATE ON inventory_actions
    FOR EACH ROW
    BEGIN
        UPDATE inventory_actions SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
    END;
    `

	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("[DB Init] Error executing schema creation: %v", err)
		return err
	}
	log.Println("[DB Init] Tables created successfully (if they didn't exist).")
	return nil
}

// seedData adds some initial data if tables are empty
func seedData(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[DB Seed] Error beginning transaction: %v", err)
		return err
	}

	// Seed Species if empty
	if err := seedTableIfEmpty(tx, "species", "INSERT INTO species (name) VALUES (?), (?), (?)", "Chicken", "Duck", "Quail"); err != nil {
		tx.Rollback()
		return err
	}

	// Seed Coops if empty
	if err := seedTableIfEmpty(tx, "coops", "INSERT INTO coops (name) VALUES (?), (?)", "Main Coop", "Duck House"); err != nil {
		tx.Rollback()
		return err
	}

	// Seed Egg Colors if empty
	if err := seedTableIfEmpty(tx, "egg_colors", "INSERT INTO egg_colors (color) VALUES (?), (?), (?)", "Brown", "White", "Blue"); err != nil {
		tx.Rollback()
		return err
	}

	// Seed Egg Sizes if empty
	if err := seedTableIfEmpty(tx, "egg_sizes", "INSERT INTO egg_sizes (size) VALUES (?), (?), (?)", "Medium", "Large", "Small"); err != nil {
		tx.Rollback()
		return err
	}

	// Seed Eggs if empty (Requires other tables to be seeded first)
	var eggCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM eggs").Scan(&eggCount)
	if err != nil {
		log.Printf("[DB Seed] Error checking egg count: %v", err)
		tx.Rollback()
		return err
	}
	if eggCount == 0 {
		log.Println("[DB Seed] Seeding eggs table...")
		_, err = tx.Exec(`INSERT INTO eggs (species_id, coop_id, collection_date, quantity, color_id, size_id) VALUES
            (1, 1, '2025-05-01', 5, 1, 2), -- 5 Large Brown Chicken eggs from Main Coop
            (1, 1, '2025-05-02', 4, 3, 1), -- 4 Medium Blue Chicken eggs from Main Coop
            (2, 2, '2025-05-02', 2, 2, 2)  -- 2 Large White Duck eggs from Duck House
        `)
		if err != nil {
			log.Printf("[DB Seed] Error seeding eggs: %v", err)
			tx.Rollback()
			return err
		}
	} else {
		log.Println("[DB Seed] Eggs table already has data, skipping seeding.")
	}

	// Seed Inventory Actions if empty (Requires eggs table to be seeded first)
	var actionCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM inventory_actions").Scan(&actionCount)
	if err != nil {
		log.Printf("[DB Seed] Error checking inventory_actions count: %v", err)
		tx.Rollback()
		return err
	}
	if actionCount == 0 && eggCount == 0 { // Only seed actions if we also seeded eggs
		log.Println("[DB Seed] Seeding inventory_actions table...")
		// Get the IDs of the eggs we just inserted (assuming they are 1, 2, 3)
		_, err = tx.Exec(`INSERT INTO inventory_actions (egg_id, action_type, quantity, action_date) VALUES
            (1, 'collected', 5, '2025-05-01 08:00:00'),
            (2, 'collected', 4, '2025-05-02 08:30:00'),
            (3, 'collected', 2, '2025-05-02 09:00:00'),
            (1, 'sold', 2, '2025-05-02 10:00:00') -- Sold 2 of the first batch
        `)
		if err != nil {
			log.Printf("[DB Seed] Error seeding inventory_actions: %v", err)
			tx.Rollback()
			return err
		}
	} else {
		log.Println("[DB Seed] Inventory_actions table already has data or eggs table was not seeded, skipping seeding actions.")
	}

	log.Println("[DB Seed] Committing seed data transaction.")
	return tx.Commit()
}

// seedTableIfEmpty checks if a table is empty and executes the seed statement if it is.
func seedTableIfEmpty(tx *sql.Tx, tableName, seedStmt string, args ...interface{}) error {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM " + tableName).Scan(&count)
	if err != nil {
		log.Printf("[DB Seed] Error checking count for table %s: %v", tableName, err)
		return err
	}

	if count == 0 {
		log.Printf("[DB Seed] Seeding %s table...", tableName)
		_, err = tx.Exec(seedStmt, args...)
		if err != nil {
			log.Printf("[DB Seed] Error seeding table %s: %v", tableName, err)
			return err
		}
	} else {
		log.Printf("[DB Seed] Table %s already has data, skipping seeding.", tableName)
	}
	return nil
}
