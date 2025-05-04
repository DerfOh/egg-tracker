package handlers

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func TestBackupHandler_CreatesBackupFiles(t *testing.T) {
	// Setup: create dummy db files
	sqliteFile := "eggtracker.db"
	duckdbFile := "eggtracker.duckdb"
	backupDir := "backups"
	os.RemoveAll(backupDir)
	defer os.RemoveAll(backupDir)
	ioutil.WriteFile(sqliteFile, []byte("sqlite"), 0644)
	ioutil.WriteFile(duckdbFile, []byte("duckdb"), 0644)
	defer os.Remove(sqliteFile)
	defer os.Remove(duckdbFile)

	r := gin.Default()
	r.POST("/api/backup", BackupHandler())

	req, _ := http.NewRequest("POST", "/api/backup", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("failed to read backup dir: %v", err)
	}
	var foundSqlite, foundDuckdb bool
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "eggtracker.db") {
			foundSqlite = true
		}
		if strings.HasSuffix(f.Name(), "eggtracker.duckdb") {
			foundDuckdb = true
		}
	}
	if !foundSqlite || !foundDuckdb {
		t.Fatalf("expected both backup files, found: %v", files)
	}
}
