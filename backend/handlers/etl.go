package handlers

import (
	"database/sql"
	"net/http"
	"os"

	"egg-tracker/backend/etl"

	"github.com/gin-gonic/gin"
)

// FullETLHandler triggers a full ETL refresh from SQLite to DuckDB.
func FullETLHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use absolute paths to ensure correct files are used inside the container
		sqlitePath := "/app/data/eggtracker.db"
		duckdbPath := "/app/data/eggtracker.duckdb"
		if _, err := os.Stat(duckdbPath); err == nil {
			os.Remove(duckdbPath) // Remove old DuckDB file for clean rebuild
		}
		err := etl.FullRefresh(sqlitePath, duckdbPath)
		if err != nil {
			// Log error to backend log for debugging
			logMsg := "[ETL ERROR] " + err.Error()
			println(logMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ETL full refresh complete"})
	}
}

// EtlFullRefreshFromMain runs a full ETL from SQLite to DuckDB using default paths.
func EtlFullRefreshFromMain() error {
	sqlitePath := "/app/data/eggtracker.db"
	duckdbPath := "/app/data/eggtracker.duckdb"
	return etl.FullRefresh(sqlitePath, duckdbPath)
}

// EtlFullRefreshFromMainWithPaths runs a full ETL from SQLite to DuckDB using provided paths.
func EtlFullRefreshFromMainWithPaths(sqlitePath, duckdbPath string) error {
	return etl.FullRefresh(sqlitePath, duckdbPath)
}
