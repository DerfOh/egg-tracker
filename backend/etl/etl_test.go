package etl

import (
	"database/sql"
	"os"
	"reflect"
	"testing"

	_ "github.com/marcboeker/go-duckdb"
	_ "github.com/mattn/go-sqlite3"
)

func TestFullRefresh_MatchesOLTP(t *testing.T) {
	sqlitePath := "test_etl_sqlite.db"
	duckdbPath := "test_etl_duckdb.db"
	defer os.Remove(sqlitePath)
	defer os.Remove(duckdbPath)

	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer sqliteDB.Close()

	// Create tables and insert sample data
	tables := []string{
		`CREATE TABLE eggs (id INTEGER PRIMARY KEY, date_laid TEXT, species TEXT, deleted BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE inventory_actions (id INTEGER PRIMARY KEY, quantity INTEGER, species TEXT, action TEXT, notes TEXT, date TEXT, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE species (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE egg_colors (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE egg_sizes (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE coops (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
	}
	for _, stmt := range tables {
		if _, err := sqliteDB.Exec(stmt); err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	// Insert sample data
	if _, err := sqliteDB.Exec(`INSERT INTO eggs (id, date_laid, species, deleted, created_at, updated_at) VALUES (1, '2024-05-01', 'Chicken', 0, '2024-05-01', '2024-05-01')`); err != nil {
		t.Fatalf("insert eggs: %v", err)
	}
	if _, err := sqliteDB.Exec(`INSERT INTO inventory_actions (id, quantity, species, action, notes, date, created_at, updated_at) VALUES (1, 10, 'Goose', 'collected', 'note', '2024-05-01', '2024-05-01', '2024-05-01')`); err != nil {
		t.Fatalf("insert inventory: %v", err)
	}
	if _, err := sqliteDB.Exec(`INSERT INTO species (id, name, active, created_at, updated_at) VALUES (1, 'Chicken', 1, '2024-05-01', '2024-05-01')`); err != nil {
		t.Fatalf("insert species: %v", err)
	}
	if _, err := sqliteDB.Exec(`INSERT INTO egg_colors (id, name, active, created_at, updated_at) VALUES (1, 'White', 1, '2024-05-01', '2024-05-01')`); err != nil {
		t.Fatalf("insert egg_colors: %v", err)
	}
	if _, err := sqliteDB.Exec(`INSERT INTO egg_sizes (id, name, active, created_at, updated_at) VALUES (1, 'Large', 1, '2024-05-01', '2024-05-01')`); err != nil {
		t.Fatalf("insert egg_sizes: %v", err)
	}
	if _, err := sqliteDB.Exec(`INSERT INTO coops (id, name, active, created_at, updated_at) VALUES (1, 'Main Coop', 1, '2024-05-01', '2024-05-01')`); err != nil {
		t.Fatalf("insert coops: %v", err)
	}

	// Run ETL
	if err := FullRefresh(sqlitePath, duckdbPath); err != nil {
		t.Fatalf("etl: %v", err)
	}

	duckDB, err := sql.Open("duckdb", duckdbPath)
	if err != nil {
		t.Fatalf("open duckdb: %v", err)
	}
	defer duckDB.Close()

	tableNames := []string{"eggs", "inventory_actions", "species", "egg_colors", "egg_sizes", "coops"}
	for _, tbl := range tableNames {
		srcRows, err := sqliteDB.Query("SELECT * FROM " + tbl)
		if err != nil {
			t.Fatalf("sqlite select %s: %v", tbl, err)
		}
		dstRows, err := duckDB.Query("SELECT * FROM " + tbl)
		if err != nil {
			t.Fatalf("duckdb select %s: %v", tbl, err)
		}
		srcVals := scanAllRows(srcRows)
		dstVals := scanAllRows(dstRows)
		if !reflect.DeepEqual(srcVals, dstVals) {
			t.Errorf("table %s mismatch:\ngot  %v\nwant %v\ngot types:  %v\nwant types: %v", tbl, dstVals, srcVals, typesOfRows(dstVals), typesOfRows(srcVals))
		}
	}
}

func TestIncrementalRefresh_OnlyNewOrChangedRecords(t *testing.T) {
	sqlitePath := "test_etl_sqlite_inc.db"
	duckdbPath := "test_etl_duckdb_inc.db"
	defer os.Remove(sqlitePath)
	defer os.Remove(duckdbPath)

	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer sqliteDB.Close()

	// Create tables and insert initial data
	tables := []string{
		`CREATE TABLE eggs (id INTEGER PRIMARY KEY, date_laid TEXT, species TEXT, deleted BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE inventory_actions (id INTEGER PRIMARY KEY, quantity INTEGER, species TEXT, action TEXT, notes TEXT, date TEXT, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE species (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE egg_colors (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE egg_sizes (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
		`CREATE TABLE coops (id INTEGER PRIMARY KEY, name TEXT, active BOOLEAN, created_at TEXT, updated_at TEXT);`,
	}
	for _, stmt := range tables {
		if _, err := sqliteDB.Exec(stmt); err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	// Insert initial data
	if _, err := sqliteDB.Exec(`INSERT INTO eggs (id, date_laid, species, deleted, created_at, updated_at) VALUES (1, '2024-05-01', 'Chicken', 0, '2024-05-01T00:00:00', '2024-05-01T00:00:00')`); err != nil {
		t.Fatalf("insert eggs: %v", err)
	}

	// Initial full refresh
	if err := FullRefresh(sqlitePath, duckdbPath); err != nil {
		t.Fatalf("etl: %v", err)
	}

	// Insert new and updated records in SQLite
	if _, err := sqliteDB.Exec(`INSERT INTO eggs (id, date_laid, species, deleted, created_at, updated_at) VALUES (2, '2024-05-02', 'Goose', 0, '2024-05-03T00:00:00', '2024-05-03T00:00:00')`); err != nil {
		t.Fatalf("insert new egg: %v", err)
	}
	if _, err := sqliteDB.Exec(`UPDATE eggs SET species = 'Duck', updated_at = '2024-05-04T00:00:00' WHERE id = 1`); err != nil {
		t.Fatalf("update egg: %v", err)
	}

	// Incremental refresh since '2024-05-02T00:00:00'
	if err := IncrementalRefresh(sqlitePath, duckdbPath, "2024-05-02T00:00:00"); err != nil {
		t.Fatalf("incremental etl: %v", err)
	}

	duckDB, err := sql.Open("duckdb", duckdbPath)
	if err != nil {
		t.Fatalf("open duckdb: %v", err)
	}
	defer duckDB.Close()

	// Check that both records are present and updated
	rows, err := duckDB.Query("SELECT id, species FROM eggs ORDER BY id ASC")
	if err != nil {
		t.Fatalf("duckdb select: %v", err)
	}
	defer rows.Close()
	var ids []int
	var species []string
	for rows.Next() {
		var id int
		var sp string
		if err := rows.Scan(&id, &sp); err != nil {
			t.Fatalf("scan: %v", err)
		}
		ids = append(ids, id)
		species = append(species, sp)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 eggs, got %d", len(ids))
	}
	if species[0] != "Duck" {
		t.Errorf("expected updated species 'Duck' for id 1, got %s", species[0])
	}
	if species[1] != "Goose" {
		t.Errorf("expected species 'Goose' for id 2, got %s", species[1])
	}
}

func scanAllRows(rows *sql.Rows) [][]interface{} {
	defer rows.Close()
	cols, _ := rows.Columns()
	var all [][]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		// Normalize types for comparison
		norm := make([]interface{}, len(vals))
		for i, v := range vals {
			switch b := v.(type) {
			case bool:
				if b {
					norm[i] = 1
				} else {
					norm[i] = 0
				}
			case int64:
				norm[i] = int(b)
			case int32:
				norm[i] = int(b)
			case []uint8:
				norm[i] = string(b)
			default:
				norm[i] = b
			}
		}
		all = append(all, norm)
	}
	return all
}

func typesOfRows(rows [][]interface{}) [][]string {
	types := make([][]string, len(rows))
	for i, row := range rows {
		types[i] = make([]string, len(row))
		for j, v := range row {
			if v == nil {
				types[i][j] = "nil"
			} else {
				types[i][j] = reflect.TypeOf(v).String()
			}
		}
	}
	return types
}
