package etl

import (
	"database/sql"
	"fmt"
	"log" // Import log package
	"strings"

	_ "github.com/marcboeker/go-duckdb"
	_ "github.com/mattn/go-sqlite3"
)

// FullRefresh copies all relevant tables from SQLite to DuckDB, replacing OLAP data.
func FullRefresh(sqlitePath, duckdbPath string) error {
	log.Printf("[ETL FullRefresh] Starting: SQLite='%s', DuckDB='%s'", sqlitePath, duckdbPath) // Add logging
	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		log.Printf("[ETL FullRefresh] Error opening SQLite: %v", err) // Add logging
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer sqliteDB.Close()

	duckDB, err := sql.Open("duckdb", duckdbPath)
	if err != nil {
		log.Printf("[ETL FullRefresh] Error opening DuckDB: %v", err) // Add logging
		return fmt.Errorf("open duckdb: %w", err)
	}
	defer duckDB.Close()

	// Ensure all required tables are included, especially inventory_actions
	tables := []string{"eggs", "inventory_actions", "species", "egg_colors", "egg_sizes", "coops"}
	log.Printf("[ETL FullRefresh] Tables to copy: %v", tables) // Add logging
	for _, tbl := range tables {
		log.Printf("[ETL FullRefresh] Copying table: %s", tbl) // Add logging
		if err := copyTable(sqliteDB, duckDB, tbl); err != nil {
			log.Printf("[ETL FullRefresh] Error copying table %s: %v", tbl, err) // Add logging
			return fmt.Errorf("copy table %s: %w", tbl, err)
		}
		log.Printf("[ETL FullRefresh] Successfully copied table: %s", tbl) // Add logging
	}
	log.Println("[ETL FullRefresh] Completed successfully.") // Add logging
	return nil
}

// IncrementalRefresh copies only new or updated records from SQLite to DuckDB based on created_at/updated_at timestamps.
func IncrementalRefresh(sqlitePath, duckdbPath string, since string) error {
	log.Printf("[ETL IncrementalRefresh] Starting: SQLite='%s', DuckDB='%s', Since='%s'", sqlitePath, duckdbPath, since) // Add logging
	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		log.Printf("[ETL IncrementalRefresh] Error opening SQLite: %v", err) // Add logging
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer sqliteDB.Close()

	duckDB, err := sql.Open("duckdb", duckdbPath)
	if err != nil {
		log.Printf("[ETL IncrementalRefresh] Error opening DuckDB: %v", err) // Add logging
		return fmt.Errorf("open duckdb: %w", err)
	}
	defer duckDB.Close()

	// Ensure all required tables are included, especially inventory_actions
	tables := []string{"eggs", "inventory_actions", "species", "egg_colors", "egg_sizes", "coops"}
	log.Printf("[ETL IncrementalRefresh] Tables to upsert: %v", tables) // Add logging
	for _, tbl := range tables {
		// Check if table exists in DuckDB first for incremental, create if not
		var exists int
		err := duckDB.QueryRow("SELECT 1 FROM information_schema.tables WHERE table_name = ?", tbl).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("[ETL IncrementalRefresh] Error checking if table %s exists in DuckDB: %v", tbl, err)
			return fmt.Errorf("check table exists %s: %w", tbl, err)
		}

		if err == sql.ErrNoRows {
			log.Printf("[ETL IncrementalRefresh] Table %s does not exist in DuckDB, performing initial copy.", tbl)
			if err := copyTable(sqliteDB, duckDB, tbl); err != nil {
				log.Printf("[ETL IncrementalRefresh] Error copying table %s during incremental setup: %v", tbl, err)
				return fmt.Errorf("initial copy table %s: %w", tbl, err)
			}
			log.Printf("[ETL IncrementalRefresh] Successfully performed initial copy for table: %s", tbl)
		} else {
			log.Printf("[ETL IncrementalRefresh] Upserting table: %s", tbl) // Add logging
			if err := upsertTable(sqliteDB, duckDB, tbl, since); err != nil {
				log.Printf("[ETL IncrementalRefresh] Error upserting table %s: %v", tbl, err) // Add logging
				return fmt.Errorf("upsert table %s: %w", tbl, err)
			}
			log.Printf("[ETL IncrementalRefresh] Successfully upserted table: %s", tbl) // Add logging
		}
	}
	log.Println("[ETL IncrementalRefresh] Completed successfully.") // Add logging
	return nil
}

func copyTable(src, dst *sql.DB, table string) error {
	// Drop and recreate table in DuckDB
	var schema string
	log.Printf("[ETL copyTable] Getting schema for table: %s", table) // Add logging
	row := src.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table)
	if err := row.Scan(&schema); err != nil {
		// If schema is missing (e.g., eggs table removed), log and skip? Or error out?
		// For now, error out as the spec implies these tables exist.
		log.Printf("[ETL copyTable] Could not get schema for table %s from sqlite_master: %v", table, err) // Use log package
		return fmt.Errorf("get schema for %s: %w", table, err)
	}
	log.Printf("[ETL copyTable] Original schema for %s: %s", table, schema) // Add logging

	// Clean schema for DuckDB compatibility
	schema = cleanSchemaForDuckDB(schema)
	log.Printf("[ETL copyTable] Cleaned schema for %s: %s", table, schema) // Add logging

	log.Printf("[ETL copyTable] Dropping table %s if exists in DuckDB", table) // Add logging
	if _, err := dst.Exec("DROP TABLE IF EXISTS " + table); err != nil {
		log.Printf("[ETL copyTable] Failed to drop table %s: %v", table, err) // Use log package
		return fmt.Errorf("drop table %s: %w", table, err)
	}

	log.Printf("[ETL copyTable] Creating table %s in DuckDB with schema: %s", table, schema) // Use log package and show final schema
	if _, err := dst.Exec(schema); err != nil {
		log.Printf("[ETL copyTable] Failed to create table %s: %v\nSchema used: %s", table, err, schema) // Use log package
		return fmt.Errorf("create table %s: %w", table, err)
	}
	log.Printf("[ETL copyTable] Successfully created table %s in DuckDB", table) // Add logging

	// Copy data
	log.Printf("[ETL copyTable] Selecting data from source table %s", table) // Add logging
	rows, err := src.Query("SELECT * FROM " + table)
	if err != nil {
		log.Printf("[ETL copyTable] Failed to select data from source table %s: %v", table, err) // Use log package
		return fmt.Errorf("select from %s: %w", table, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Printf("[ETL copyTable] Failed to get columns for source table %s: %v", table, err)
		return fmt.Errorf("get columns for %s: %w", table, err)
	}
	if len(cols) == 0 {
		log.Printf("[ETL copyTable] No columns found for source table %s, skipping data copy.", table)
		return nil // Or return error? Table exists but is empty/unstructured?
	}

	// Check if DuckDB table columns match SQLite columns
	duckRows, err := dst.Query("PRAGMA table_info(" + table + ")")
	if err == nil {
		var duckCols []string
		for duckRows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			_ = duckRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
			duckCols = append(duckCols, name)
		}
		duckRows.Close()
		if len(duckCols) != len(cols) {
			log.Printf("[ETL copyTable] WARNING: DuckDB table %s column count (%d) does not match SQLite (%d)", table, len(duckCols), len(cols))
		}
	}

	vals := make([]interface{}, len(cols))
	scanArgs := make([]interface{}, len(cols)) // Use separate slice for Scan arguments
	for i := range vals {
		scanArgs[i] = &vals[i] // Point scanArgs to elements in vals
	}

	insertSQL := "INSERT INTO " + table + " (" + joinCols(cols) + ") VALUES (" + placeholders(len(cols)) + ")"
	log.Printf("[ETL copyTable] Prepared insert statement for %s: %s", table, insertSQL) // Add logging

	rowCount := 0
	tx, err := dst.Begin() // Use transaction for bulk insert
	if err != nil {
		log.Printf("[ETL copyTable] Failed to begin transaction for table %s: %v", table, err)
		return fmt.Errorf("begin transaction for %s: %w", table, err)
	}
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		log.Printf("[ETL copyTable] Failed to prepare insert statement for table %s: %v", table, err)
		tx.Rollback()
		return fmt.Errorf("prepare insert for %s: %w", table, err)
	}
	defer stmt.Close()

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			log.Printf("[ETL copyTable] Failed to scan row for table %s: %v", table, err) // Use log package
			tx.Rollback()
			return fmt.Errorf("scan row %d for %s: %w", rowCount+1, table, err)
		}
		// Convert types if necessary (e.g., time.Time to string for DuckDB TIMESTAMP)
		// For now, assume direct mapping works or DuckDB handles it.
		if _, err := stmt.Exec(vals...); err != nil {
			log.Printf("[ETL copyTable] Failed to execute prepared insert for row %d into table %s: %v", rowCount+1, table, err) // Use log package
			log.Printf("[ETL copyTable] Failed data: %v", vals)                                                                  // Log the data
			tx.Rollback()
			return fmt.Errorf("exec insert row %d for %s: %w", rowCount+1, table, err)
		}
		rowCount++
	}

	if err := rows.Err(); err != nil { // Check for errors during iteration
		log.Printf("[ETL copyTable] Error during row iteration for table %s: %v", table, err)
		tx.Rollback()
		return fmt.Errorf("rows iteration for %s: %w", table, err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[ETL copyTable] Failed to commit transaction for table %s: %v", table, err)
		return fmt.Errorf("commit transaction for %s: %w", table, err)
	}

	log.Printf("[ETL copyTable] Successfully inserted %d rows into table %s", rowCount, table) // Add logging
	return nil
}

// cleanSchemaForDuckDB attempts to convert SQLite schema syntax to be DuckDB compatible.
func cleanSchemaForDuckDB(schema string) string {
	// Remove AUTOINCREMENT (case-insensitive)
	schema = replaceCaseInsensitive(schema, " AUTOINCREMENT", "")

	// Remove DEFAULT CURRENT_TIMESTAMP (case-insensitive) - DuckDB uses now() or similar, handle defaults separately if needed
	schema = replaceCaseInsensitive(schema, " DEFAULT CURRENT_TIMESTAMP", "")

	// Replace DATETIME with TIMESTAMP (common type difference)
	schema = replaceCaseInsensitive(schema, "DATETIME", "TIMESTAMP")

	// Replace INTEGER PRIMARY KEY with BIGINT PRIMARY KEY (DuckDB often prefers BIGINT for PKs)
	// Only if AUTOINCREMENT was not present, otherwise it was already handled.
	// This is a bit heuristic. A more robust solution involves parsing the schema.
	if !strings.Contains(strings.ToUpper(schema), "AUTOINCREMENT") { // Check original schema maybe? This check is flawed.
		// Let's just replace INTEGER PRIMARY KEY directly for now.
		schema = replaceCaseInsensitive(schema, "INTEGER PRIMARY KEY", "BIGINT PRIMARY KEY")
	}

	// Add more replacements as needed based on observed errors
	// e.g., constraints, specific data types

	return schema
}

// replaceCaseInsensitive performs a case-insensitive replacement.
func replaceCaseInsensitive(s, old, new string) string {
	oldUpper := strings.ToUpper(old)
	var result strings.Builder
	start := 0
	for {
		index := strings.Index(strings.ToUpper(s[start:]), oldUpper)
		if index == -1 {
			result.WriteString(s[start:])
			break
		}
		result.WriteString(s[start : start+index])
		result.WriteString(new)
		start += index + len(old)
	}
	return result.String()
}

// upsertTable inserts or updates records in DuckDB from SQLite where created_at or updated_at > since.
// Uses DuckDB's INSERT ... ON CONFLICT DO UPDATE syntax.
func upsertTable(src, dst *sql.DB, table, since string) error {
	log.Printf("[ETL upsertTable] Starting upsert for table %s since %s", table, since) // Add logging

	// Get columns from source table to ensure we handle the correct data
	query := fmt.Sprintf("SELECT * FROM %s WHERE created_at > ? OR updated_at > ?", table)
	log.Printf("[ETL upsertTable] Querying source: %s with since=%s", query, since) // Add logging
	rows, err := src.Query(query, since, since)
	if err != nil {
		log.Printf("[ETL upsertTable] Failed to select data from source table %s: %v", table, err) // Add logging
		return fmt.Errorf("select for upsert %s: %w", table, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Printf("[ETL upsertTable] Failed to get columns for source table %s: %v", table, err) // Add logging
		return fmt.Errorf("get columns for upsert %s: %w", table, err)
	}
	if len(cols) == 0 {
		log.Printf("[ETL upsertTable] No columns found for source table %s, skipping upsert.", table)
		return nil
	}
	log.Printf("[ETL upsertTable] Columns for %s: %v", table, cols) // Add logging

	// Find the 'id' column index
	idColIndex := -1
	idColName := ""
	for i, c := range cols {
		// Assuming 'id' is the primary key for conflict resolution
		if strings.ToLower(c) == "id" {
			idColIndex = i
			idColName = c // Keep original case for SQL statement
			break
		}
	}
	if idColIndex == -1 {
		log.Printf("[ETL upsertTable] No 'id' column found in table %s, cannot perform upsert.", table) // Add logging
		return fmt.Errorf("no id column in %s for upsert", table)
	}

	// Prepare the UPSERT statement for DuckDB
	var setClauses []string
	for _, c := range cols {
		if strings.ToLower(c) == "id" {
			continue // Don't update the ID itself
		}
		// Use excluded.colname to refer to the values from the row proposed for insertion
		setClauses = append(setClauses, fmt.Sprintf("%s = excluded.%s", c, c))
	}

	upsertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		table,
		joinCols(cols),                 // Use original column names
		placeholders(len(cols)),        // Placeholders for values
		idColName,                      // Conflict target column
		strings.Join(setClauses, ", "), // Update clauses
	)
	log.Printf("[ETL upsertTable] Prepared upsert statement for %s: %s", table, upsertSQL) // Add logging

	// Execute upsert within a transaction
	tx, err := dst.Begin()
	if err != nil {
		log.Printf("[ETL upsertTable] Failed to begin transaction for table %s: %v", table, err)
		return fmt.Errorf("begin upsert transaction for %s: %w", table, err)
	}
	stmt, err := tx.Prepare(upsertSQL)
	if err != nil {
		log.Printf("[ETL upsertTable] Failed to prepare upsert statement for table %s: %v", table, err)
		tx.Rollback()
		return fmt.Errorf("prepare upsert for %s: %w", table, err)
	}
	defer stmt.Close()

	vals := make([]interface{}, len(cols))
	scanArgs := make([]interface{}, len(cols))
	for i := range vals {
		scanArgs[i] = &vals[i]
	}

	rowCount := 0
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			log.Printf("[ETL upsertTable] Failed to scan row for table %s: %v", table, err) // Add logging
			tx.Rollback()
			return fmt.Errorf("scan upsert row %d for %s: %w", rowCount+1, table, err)
		}

		// Execute the prepared upsert statement
		if _, err := stmt.Exec(vals...); err != nil {
			log.Printf("[ETL upsertTable] Failed to execute upsert for row %d into table %s: %v", rowCount+1, table, err) // Add logging
			log.Printf("[ETL upsertTable] Failed data: %v", vals)                                                         // Add logging
			tx.Rollback()
			return fmt.Errorf("exec upsert row %d for %s: %w", rowCount+1, table, err)
		}
		rowCount++
	}

	if err := rows.Err(); err != nil { // Check for errors during iteration
		log.Printf("[ETL upsertTable] Error during row iteration for table %s: %v", table, err)
		tx.Rollback()
		return fmt.Errorf("rows iteration for upsert %s: %w", table, err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[ETL upsertTable] Failed to commit upsert transaction for table %s: %v", table, err)
		return fmt.Errorf("commit upsert transaction for %s: %w", table, err)
	}

	log.Printf("[ETL upsertTable] Successfully upserted %d rows into table %s", rowCount, table) // Add logging
	return nil
}

func joinCols(cols []string) string {
	// Quote column names in case they are keywords or contain spaces
	quotedCols := make([]string, len(cols))
	for i, c := range cols {
		quotedCols[i] = `"` + c + `"` // Use standard SQL quotes
	}
	return strings.Join(quotedCols, ", ")
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?, ", n-1) + "?"
}
