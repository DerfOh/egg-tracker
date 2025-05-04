package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// BackupHandler copies SQLite and DuckDB files to /backups/ with timestamps.
func BackupHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlitePath := "eggtracker.db"
		duckdbPath := "eggtracker.duckdb"
		backupDir := "backups"
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup dir"})
			return
		}
		timestamp := time.Now().Format("20060102_150405")
		files := []string{sqlitePath, duckdbPath}
		var backedUp []string
		for _, f := range files {
			if _, err := os.Stat(f); err == nil {
				backupName := filepath.Join(backupDir, fmt.Sprintf("%s_%s", timestamp, filepath.Base(f)))
				src, err := os.Open(f)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file: " + f})
					return
				}
				defer src.Close()
				dst, err := os.Create(backupName)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup: " + backupName})
					return
				}
				if _, err := io.Copy(dst, src); err != nil {
					dst.Close()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to copy file: " + f})
					return
				}
				dst.Close()
				backedUp = append(backedUp, backupName)
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "backup complete", "files": backedUp})
	}
}
