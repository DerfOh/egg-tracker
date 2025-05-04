package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"egg-tracker/backend/db"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func setupOptionsTestDB() (*sql.DB, func()) {
	testDBPath := "test_options.db"
	dbase, _ := sql.Open("sqlite3", testDBPath)
	db.Migrate(dbase)
	return dbase, func() {
		dbase.Close()
		os.Remove(testDBPath)
	}
}

func TestOptionsCRUD(t *testing.T) {
	dbase, cleanup := setupOptionsTestDB()
	defer cleanup()
	router := gin.Default()
	router.POST("/api/options/:type", AddOptionHandler(dbase))
	router.GET("/api/options/:type", ListOptionsHandler(dbase))
	router.PUT("/api/options/:type/:id", EditOptionHandler(dbase))
	router.POST("/api/options/:type/:id/deactivate", DeactivateOptionHandler(dbase))
	router.POST("/api/options/:type/:id/reactivate", ReactivateOptionHandler(dbase))

	// Add species
	payload := map[string]interface{}{"name": "Chicken"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/options/species", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	idFloat, ok := resp["id"].(float64)
	if !ok {
		t.Fatalf("expected id in response, got: %v", resp)
	}
	id := int(idFloat)
	idStr := strconv.Itoa(id)

	// Edit species
	editPayload := map[string]interface{}{"name": "Goose"}
	editBody, _ := json.Marshal(editPayload)
	req, _ = http.NewRequest("PUT", "/api/options/species/"+idStr, bytes.NewBuffer(editBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on edit, got %d", w.Code)
	}

	// Deactivate
	req, _ = http.NewRequest("POST", "/api/options/species/"+idStr+"/deactivate", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on deactivate, got %d", w.Code)
	}

	// Reactivate
	req, _ = http.NewRequest("POST", "/api/options/species/"+idStr+"/reactivate", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on reactivate, got %d", w.Code)
	}

	// List
	req, _ = http.NewRequest("GET", "/api/options/species", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d", w.Code)
	}
	var listResp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to unmarshal list response: %v", err)
	}
	if len(listResp) != 1 {
		t.Fatalf("expected 1 option, got %d", len(listResp))
	}
	if listResp[0]["name"] != "Goose" {
		t.Errorf("expected name 'Goose' after edit, got %v", listResp[0]["name"])
	}
	if !listResp[0]["active"].(bool) {
		t.Errorf("expected active true after reactivate, got %v", listResp[0]["active"])
	}
}
