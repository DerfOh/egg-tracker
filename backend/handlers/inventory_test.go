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

func setupInventoryTestDB() (*sql.DB, func()) {
	testDBPath := "test_inventory.db"
	dbase, _ := sql.Open("sqlite3", testDBPath)
	db.Migrate(dbase)
	return dbase, func() {
		dbase.Close()
		os.Remove(testDBPath)
	}
}

func TestInventoryCRUD(t *testing.T) {
	dbase, cleanup := setupInventoryTestDB()
	defer cleanup()
	router := gin.Default()
	router.POST("/api/inventory", CreateInventoryHandler(dbase))
	router.GET("/api/inventory", ListInventoryHandler(dbase))
	router.PUT("/api/inventory/:id", UpdateInventoryHandler(dbase))
	router.DELETE("/api/inventory/:id", DeleteInventoryHandler(dbase))

	// Create
	payload := map[string]interface{}{
		"quantity":  5,
		"species":   "Goose",
		"coop":      "Main Coop",
		"egg_color": "White",
		"egg_size":  "Large",
		"action":    "collected",
		"date":      "2024-05-01",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/inventory", bytes.NewBuffer(body))
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

	// Update
	updatePayload := map[string]interface{}{
		"quantity":  7,
		"species":   "Goose",
		"coop":      "Main Coop",
		"egg_color": "White",
		"egg_size":  "Large",
		"action":    "sold",
		"date":      "2024-05-02",
	}
	updateBody, _ := json.Marshal(updatePayload)
	req, _ = http.NewRequest("PUT", "/api/inventory/"+idStr, bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", w.Code)
	}

	// List
	req, _ = http.NewRequest("GET", "/api/inventory", nil)
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
		t.Fatalf("expected 1 inventory action, got %d", len(listResp))
	}
	if int(listResp[0]["quantity"].(float64)) != 7 {
		t.Errorf("expected quantity 7 after update, got %v", listResp[0]["quantity"])
	}
	if listResp[0]["action"] != "sold" {
		t.Errorf("expected action 'sold' after update, got %v", listResp[0]["action"])
	}

	// Delete
	req, _ = http.NewRequest("DELETE", "/api/inventory/"+idStr, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d", w.Code)
	}

	// List again to confirm deletion
	req, _ = http.NewRequest("GET", "/api/inventory", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on list after delete, got %d", w.Code)
	}
	listResp = nil
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to unmarshal list response after delete: %v", err)
	}
	if len(listResp) != 0 {
		t.Fatalf("expected 0 inventory actions after delete, got %d", len(listResp))
	}
}

func TestInventoryInvalidInput(t *testing.T) {
	dbase, cleanup := setupInventoryTestDB()
	defer cleanup()
	router := gin.Default()
	router.POST("/api/inventory", CreateInventoryHandler(dbase))

	// Missing required field
	payload := map[string]interface{}{
		"quantity": 5,
		"species":  "Goose",
		"action":   "collected",
		// "date" is missing
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/inventory", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing date, got %d", w.Code)
	}
}

func TestInventoryInvalidDate(t *testing.T) {
	dbase, cleanup := setupInventoryTestDB()
	defer cleanup()
	router := gin.Default()
	router.POST("/api/inventory", CreateInventoryHandler(dbase))

	// Invalid date format
	payload := map[string]interface{}{
		"quantity": 5,
		"species":  "Goose",
		"action":   "collected",
		"date":     "05-01-2024", // wrong format
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/inventory", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid date, got %d", w.Code)
	}
}

func TestInventoryUpdateNotFound(t *testing.T) {
	dbase, cleanup := setupInventoryTestDB()
	defer cleanup()
	router := gin.Default()
	router.PUT("/api/inventory/:id", UpdateInventoryHandler(dbase))

	updatePayload := map[string]interface{}{
		"quantity":  7,
		"species":   "Goose",
		"coop":      "Main Coop",
		"egg_color": "White",
		"egg_size":  "Large",
		"action":    "sold",
		"date":      "2024-05-02",
	}
	updateBody, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/api/inventory/9999", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Should still return 200, but nothing is updated
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on update non-existent, got %d", w.Code)
	}
}

func TestInventoryDeleteNotFound(t *testing.T) {
	dbase, cleanup := setupInventoryTestDB()
	defer cleanup()
	router := gin.Default()
	router.DELETE("/api/inventory/:id", DeleteInventoryHandler(dbase))

	req, _ := http.NewRequest("DELETE", "/api/inventory/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete non-existent, got %d", w.Code)
	}
}
